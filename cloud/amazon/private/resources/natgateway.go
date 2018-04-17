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

var _ cloud.Resource = &NATGateway{}

type NATGateway struct {
	Shared
	ClusterSubnet *cluster.Subnet
}

func (r *NATGateway) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("natgateway.Actual")

	newResource := &NATGateway{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		output, err := Sdk.Ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
			Filter: []*ec2.Filter{
				{
					Name:   S("tag:kubicorn-nat-gateway"),
					Values: []*string{&r.ClusterSubnet.Name},
				},
				{
					Name:   S("vpc-id"),
					Values: []*string{&immutable.ProviderConfig().Network.Identifier},
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
		lng := len(output.NatGateways)
		if lng != 1 {
			return nil, nil, fmt.Errorf("Found [%d] NAT Gateways for ID [%s]", lng, r.ClusterSubnet.Name)
		}
		ng := output.NatGateways[0]

		newResource.Identifier = r.ClusterSubnet.Name
		for _, tag := range ng.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *NATGateway) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("natgateway.Expected")

	newResource := &NATGateway{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Name,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
				"kubicorn-nat-gateway":                    r.ClusterSubnet.Name,
			},
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *NATGateway) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("natgateway.Apply")

	applyResource := expected.(*NATGateway)
	isEqual, err := compare.IsEqual(actual.(*NATGateway), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	aaOutput, err := Sdk.Ec2.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: S("vpc"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to allocate new Elastic IP: %v", err)
	}
	logger.Success("Allocated Elastic IP [%s]", *aaOutput.AllocationId)

	subnetID := ""
	for _, psn := range immutable.ProviderConfig().Network.PublicSubnets {
		if psn.Name == r.Name {
			subnetID = psn.Identifier
		}
	}
	if subnetID == "" {
		return nil, nil, fmt.Errorf("Unable to find Public Subnet ID")
	}

	ngOutput, err := Sdk.Ec2.CreateNatGateway(&ec2.CreateNatGatewayInput{
		AllocationId: aaOutput.AllocationId,
		SubnetId:     &subnetID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new NAT Gateway: %v", err)
	}
	natGatewayID := ngOutput.NatGateway.NatGatewayId

	logger.Info("Waiting for NAT Gateway [%s] to be available", *natGatewayID)
	err = Sdk.Ec2.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{natGatewayID},
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created NAT Gateway [%s] in Public Subnet [%s] with Elastic IP [%s]", *natGatewayID, subnetID, *aaOutput.AllocationId)

	newResource := &NATGateway{
		Shared: Shared{
			Identifier: *natGatewayID,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new NAT Gateway: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *NATGateway) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("natgateway.Delete")

	deleteResource := actual.(*NATGateway)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete NAT Gateway resource without ID [%s]", deleteResource.Name)
	}

	output, err := Sdk.Ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name:   S("tag:kubicorn-nat-gateway"),
				Values: []*string{&r.ClusterSubnet.Name},
			},
			{
				Name:   S("vpc-id"),
				Values: []*string{&immutable.ProviderConfig().Network.Identifier},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	lng := len(output.NatGateways)
	if lng != 1 {
		return nil, nil, fmt.Errorf("Found [%d] NAT Gateways for ID [%s]", lng, deleteResource.Identifier)
	}
	natGatewayID := output.NatGateways[0].NatGatewayId
	allocationID := output.NatGateways[0].NatGatewayAddresses[0].AllocationId

	_, err = Sdk.Ec2.DeleteNatGateway(&ec2.DeleteNatGatewayInput{
		NatGatewayId: natGatewayID,
	})
	if err != nil {
		return nil, nil, err
	}

	// Deletion can take up to five minutes, and there is no WaitUntilNatGatewayDeleted
	logger.Info("Waiting for NAT Gateway [%s] to be deleted", *natGatewayID)
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 5 * time.Minute
	dng := func() error {
		output, err := Sdk.Ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{natGatewayID},
		})
		if err != nil {
			return err
		}

		if *output.NatGateways[0].State == "deleted" {
			return nil
		}

		return fmt.Errorf("NAT Gateway [%s] not deleted", *natGatewayID)
	}
	err = backoff.Retry(dng, b)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted NAT Gateway [%s]", *natGatewayID)

	_, err = Sdk.Ec2.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: allocationID,
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Released Elastic IP [%s]", *allocationID)

	newResource := &NATGateway{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *NATGateway) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("natgateway.Render")

	return inaccurateCluster
}

func (r *NATGateway) tag(tags map[string]string) error {
	logger.Debug("natgateway.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering NAT Gateway tag [%s] %s", key, val)
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
