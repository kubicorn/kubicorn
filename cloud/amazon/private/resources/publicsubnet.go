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

var _ cloud.Resource = &PublicSubnet{}

type PublicSubnet struct {
	Shared
	ClusterPublicSubnet *cluster.PublicSubnet
	CIDR                string
	VpcID               string
	Zone                string
}

func (r *PublicSubnet) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Actual")
	newResource := &PublicSubnet{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterPublicSubnet.Identifier != "" {
		input := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{&r.ClusterPublicSubnet.Identifier},
		}
		output, err := Sdk.Ec2.DescribeSubnets(input)
		if err != nil {
			return nil, nil, err
		}
		lsn := len(output.Subnets)
		if lsn != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Subnets for ID [%s]", lsn, r.ClusterPublicSubnet.Identifier)
		}
		subnet := output.Subnets[0]

		newResource.CIDR = *subnet.CidrBlock
		newResource.Identifier = *subnet.SubnetId
		newResource.VpcID = *subnet.VpcId
		newResource.Zone = *subnet.AvailabilityZone
		for _, tag := range subnet.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	} else {
		newResource.CIDR = r.ClusterPublicSubnet.CIDR
		newResource.Zone = r.ClusterPublicSubnet.Zone
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Expected")
	newResource := &PublicSubnet{
		Shared: Shared{
			Identifier: r.ClusterPublicSubnet.Identifier,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
				"kubernetes.io/role/elb":                  "true",
			},
		},
		CIDR:  r.ClusterPublicSubnet.CIDR,
		VpcID: immutable.ProviderConfig().Network.Identifier,
		Zone:  r.ClusterPublicSubnet.Zone,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Apply")
	applyResource := expected.(*PublicSubnet)
	isEqual, err := compare.IsEqual(actual.(*PublicSubnet), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	input := &ec2.CreateSubnetInput{
		CidrBlock:        &applyResource.CIDR,
		VpcId:            &immutable.ProviderConfig().Network.Identifier,
		AvailabilityZone: &applyResource.Zone,
	}
	output, err := Sdk.Ec2.CreateSubnet(input)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Public Subnet: %v", err)
	}
	waitInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []*string{output.Subnet.SubnetId},
	}
	logger.Info("Waiting for Public Subnet [%s] to be available", *output.Subnet.SubnetId)
	err = Sdk.Ec2.WaitUntilSubnetAvailable(waitInput)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created Public Subnet [%s]", *output.Subnet.SubnetId)

	newResource := &PublicSubnet{
		Shared: Shared{
			Identifier: *output.Subnet.SubnetId,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
		CIDR:  *output.Subnet.CidrBlock,
		VpcID: *output.Subnet.VpcId,
		Zone:  *output.Subnet.AvailabilityZone,
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Public Subnet: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Delete")
	deleteResource := actual.(*PublicSubnet)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Public Subnet resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSubnetInput{
		SubnetId: &deleteResource.Identifier,
	}
	_, err := Sdk.Ec2.DeleteSubnet(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Public Subnet [%s]", deleteResource.Identifier)

	newResource := &PublicSubnet{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
		CIDR: deleteResource.CIDR,
		Zone: deleteResource.Zone,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("publicsubnet.Render")

	newCluster := inaccurateCluster
	providerConfig := inaccurateCluster.ProviderConfig()
	for i := 0; i < len(providerConfig.Network.PublicSubnets); i++ {
		if providerConfig.Network.PublicSubnets[i].Name == newResource.(*PublicSubnet).Name {
			providerConfig.Network.PublicSubnets[i].CIDR = newResource.(*PublicSubnet).CIDR
			providerConfig.Network.PublicSubnets[i].Zone = newResource.(*PublicSubnet).Zone
			providerConfig.Network.PublicSubnets[i].Identifier = newResource.(*PublicSubnet).Identifier
		}
	}

	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func (r *PublicSubnet) tag(tags map[string]string) error {
	logger.Debug("publicsubnet.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Public Subnet tag [%s] %s", key, val)
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
