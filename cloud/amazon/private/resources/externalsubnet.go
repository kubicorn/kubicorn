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

var _ cloud.Resource = &ExternalSubnet{}

type ExternalSubnet struct {
	Shared
	CIDR          string
	ClusterSubnet *cluster.Subnet
	Public        bool
	Zone          string
}

func (r *ExternalSubnet) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("externalsubnet.Actual")

	newResource := &ExternalSubnet{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
		CIDR:   r.ClusterSubnet.CIDR,
		Public: r.Public,
		Zone:   r.ClusterSubnet.Zone,
	}

	if r.ClusterSubnet.Identifier != "" {
		output, err := Sdk.Ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{&r.ClusterSubnet.Identifier},
		})
		if err != nil {
			return nil, nil, err
		}
		subnet := output.Subnets[0]

		newResource.CIDR = *subnet.CidrBlock
		newResource.Identifier = *subnet.SubnetId
		newResource.Zone = *subnet.AvailabilityZone
		for _, tag := range subnet.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *ExternalSubnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("externalsubnet.Expected")

	newResource := &ExternalSubnet{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Identifier,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
			},
		},
		CIDR:   r.ClusterSubnet.CIDR,
		Public: r.Public,
		Zone:   r.ClusterSubnet.Zone,
	}
	if r.Public {
		newResource.Tags["kubernetes.io/role/elb"] = "true"
	} else {
		newResource.Tags["kubernetes.io/role/internal-elb"] = "true"
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *ExternalSubnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("externalsubnet.Apply")

	applyResource := expected.(*ExternalSubnet)
	isEqual, err := compare.IsEqual(actual.(*ExternalSubnet), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	output, err := Sdk.Ec2.CreateSubnet(&ec2.CreateSubnetInput{
		CidrBlock:        &applyResource.CIDR,
		VpcId:            &immutable.ProviderConfig().Network.Identifier,
		AvailabilityZone: &applyResource.Zone,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new External Subnet: %v", err)
	}
	subnetID := output.Subnet.SubnetId

	logger.Info("Waiting for External Subnet [%s] to be available", *subnetID)
	err = Sdk.Ec2.WaitUntilSubnetAvailable(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{subnetID},
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created External Subnet [%s]", *subnetID)

	newResource := &ExternalSubnet{
		Shared: Shared{
			Identifier: *subnetID,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
		CIDR:   *output.Subnet.CidrBlock,
		Public: applyResource.Public,
		Zone:   *output.Subnet.AvailabilityZone,
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new External Subnet: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *ExternalSubnet) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("externalsubnet.Delete")

	deleteResource := actual.(*ExternalSubnet)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete External Subnet resource without ID [%s]", deleteResource.Name)
	}

	_, err := Sdk.Ec2.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: &deleteResource.Identifier,
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted External Subnet [%s]", deleteResource.Identifier)

	newResource := &ExternalSubnet{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
		CIDR:   deleteResource.CIDR,
		Public: deleteResource.Public,
		Zone:   deleteResource.Zone,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *ExternalSubnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("externalsubnet.Render")

	newCluster := inaccurateCluster
	providerConfig := inaccurateCluster.ProviderConfig()

	subnets := providerConfig.Network.PrivateSubnets
	if r.Public {
		subnets = providerConfig.Network.PublicSubnets
	}
	for i := 0; i < len(subnets); i++ {
		if subnets[i].Name == newResource.(*ExternalSubnet).Name {
			subnets[i].CIDR = newResource.(*ExternalSubnet).CIDR
			subnets[i].Identifier = newResource.(*ExternalSubnet).Identifier
			subnets[i].Zone = newResource.(*ExternalSubnet).Zone
		}
	}

	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func (r *ExternalSubnet) tag(tags map[string]string) error {
	logger.Debug("externalsubnet.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering External Subnet tag [%s] %s", key, val)
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
