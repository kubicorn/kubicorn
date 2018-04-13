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

	"github.com/aws/aws-sdk-go/service/ec2"
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
		input := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&immutable.ProviderConfig().Network.InternetGW.Identifier},
		}
		output, err := Sdk.Ec2.DescribeInternetGateways(input)
		if err != nil {
			return nil, nil, err
		}
		lig := len(output.InternetGateways)
		if lig != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Internet Gateways for ID [%s]", lig, immutable.ProviderConfig().Network.InternetGW.Identifier)
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

	input := &ec2.CreateInternetGatewayInput{}
	output, err := Sdk.Ec2.CreateInternetGateway(input)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Internet Gateway: %v", err)
	}
	logger.Success("Created Internet Gateway [%s]", *output.InternetGateway.InternetGatewayId)

	attachInput := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: output.InternetGateway.InternetGatewayId,
		VpcId:             &immutable.ProviderConfig().Network.Identifier,
	}
	_, err = Sdk.Ec2.AttachInternetGateway(attachInput)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Attached Internet Gateway [%s] to VPC [%s]", *output.InternetGateway.InternetGatewayId, immutable.ProviderConfig().Network.Identifier)

	newResource := &InternetGateway{
		Shared: Shared{
			Identifier: *output.InternetGateway.InternetGatewayId,
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

	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []*string{&deleteResource.Identifier},
	}
	output, err := Sdk.Ec2.DescribeInternetGateways(input)
	if err != nil {
		return nil, nil, err
	}
	lig := len(output.InternetGateways)
	if lig == 0 {
		return nil, nil, fmt.Errorf("Found [%d] Internet Gateways for ID [%s]", lig, deleteResource.Identifier)
	}
	ig := output.InternetGateways[0]

	detInput := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: ig.InternetGatewayId,
		VpcId:             &immutable.ProviderConfig().Network.Identifier,
	}
	_, err = Sdk.Ec2.DetachInternetGateway(detInput)
	if err != nil {
		return nil, nil, err
	}

	dInput := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: ig.InternetGatewayId,
	}
	_, err = Sdk.Ec2.DeleteInternetGateway(dInput)
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
