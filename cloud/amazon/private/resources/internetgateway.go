// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &InternetGateway{}

type InternetGateway struct {
	Shared
}

func (r *InternetGateway) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("internetgateway.Actual")

	newResource := &InternetGateway{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if immutable.ProviderConfig().Network.InternetGW.Identifier != "" {
		output, err := Sdk.Ec2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&immutable.ProviderConfig().Network.InternetGW.Identifier},
		})
		if err != nil {
			return nil, nil, err
		}
		ig := output.InternetGateways[0]

		newResource.Identifier = *ig.InternetGatewayId
		for _, tag := range ig.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InternetGateway) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("internetgateway.Expected")

	newResource := &InternetGateway{
		Shared: Shared{
			Identifier: immutable.ProviderConfig().Network.InternetGW.Identifier,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
			},
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InternetGateway) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("internetgateway.Apply")

	applyResource := expected.(*InternetGateway)
	isEqual, err := compare.IsEqual(actual.(*InternetGateway), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	output, err := Sdk.Ec2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Internet Gateway: %v", err)
	}
	internetGatewayID := output.InternetGateway.InternetGatewayId

	logger.Success("Created Internet Gateway [%s]", *internetGatewayID)

	_, err = Sdk.Ec2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: internetGatewayID,
		VpcId:             &immutable.ProviderConfig().Network.Identifier,
	})
	if err != nil {
		return nil, nil, err
	}

	logger.Info("Waiting for Internet Gateway [%s] to be available", *internetGatewayID)
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 5 * time.Minute
	dng := func() error {
		igErr := fmt.Errorf("Internet Gateway [%s] not available", *internetGatewayID)

		output, err := Sdk.Ec2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{internetGatewayID},
		})
		if err != nil {
			return err
		}

		if len(output.InternetGateways[0].Attachments) < 1 {
			return igErr
		}
		if *output.InternetGateways[0].Attachments[0].State == "available" {
			return nil
		}
		return igErr
	}
	err = backoff.Retry(dng, b)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Attached Internet Gateway [%s] to VPC [%s]", *internetGatewayID, immutable.ProviderConfig().Network.Identifier)

	newResource := &InternetGateway{
		Shared: Shared{
			Identifier: *internetGatewayID,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Internet Gateway: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InternetGateway) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("internetgateway.Delete")

	deleteResource := actual.(*InternetGateway)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Internet Gateway resource without ID [%s]", deleteResource.Name)
	}

	_, err := Sdk.Ec2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
		InternetGatewayId: &deleteResource.Identifier,
		VpcId:             &immutable.ProviderConfig().Network.Identifier,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = Sdk.Ec2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
		InternetGatewayId: &deleteResource.Identifier,
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Internet Gateway [%s]", deleteResource.Identifier)

	newResource := &InternetGateway{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InternetGateway) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("internetgateway.Render")

	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Network.InternetGW.Identifier = newResource.(*InternetGateway).Identifier

	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func (r *InternetGateway) tag(tags map[string]string) error {
	logger.Debug("internetgateway.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Internet Gateway tag [%s] %s", key, val)
		tagInput.Tags = append(tagInput.Tags, &ec2.Tag{
			Key:   S("%s", key),
			Value: S("%s", val),
		})
	}
	_, err := Sdk.Ec2.CreateTags(tagInput)
	if err != nil {
		return err
	}
	return nil
}
