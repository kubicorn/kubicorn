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
	CIDR  string
	VpcID string
	Zone  string
}

func (r *PublicSubnet) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Actual")
	newResource := &PublicSubnet{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.Identifier != "" {
		input := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{S(r.Identifier)},
		}
		output, err := Sdk.Ec2.DescribeSubnets(input)
		if err != nil {
			return nil, nil, err
		}
		lsn := len(output.Subnets)
		if lsn != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Public Subnets for ID [%s]", lsn, r.Identifier)
		}
		subnet := output.Subnets[0]
		newResource.CIDR = *subnet.CidrBlock
		newResource.Identifier = *subnet.SubnetId
		newResource.VpcID = *subnet.VpcId
		newResource.Zone = *subnet.AvailabilityZone
		newResource.Tags = map[string]string{
			"Name":              r.Name,
			"KubernetesCluster": immutable.Name,
		}
		for _, tag := range subnet.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Expected")
	newResource := &PublicSubnet{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
			Identifier: r.Identifier,
			Name:       r.Name,
		},
		CIDR:  r.CIDR,
		VpcID: immutable.ProviderConfig().Network.Identifier,
		Zone:  r.Zone,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicsubnet.Apply")
	applyResource := expected.(*PublicSubnet)
	isEqual, err := compare.IsEqual(actual.(*PublicSubnet), expected.(*PublicSubnet))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	input := &ec2.CreateSubnetInput{
		CidrBlock:        &expected.(*PublicSubnet).CIDR,
		VpcId:            &immutable.ProviderConfig().Network.Identifier,
		AvailabilityZone: &expected.(*PublicSubnet).Zone,
	}
	output, err := Sdk.Ec2.CreateSubnet(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created Public Subnet [%s]", *output.Subnet.SubnetId)
	newResource := &PublicSubnet{}
	newResource.CIDR = *output.Subnet.CidrBlock
	newResource.VpcID = *output.Subnet.VpcId
	newResource.Zone = *output.Subnet.AvailabilityZone
	newResource.Name = applyResource.Name
	newResource.Identifier = *output.Subnet.SubnetId

	// Tag newly created PublicSubnet
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
		return nil, nil, fmt.Errorf("Unable to delete publicsubnet resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSubnetInput{
		SubnetId: &actual.(*PublicSubnet).Identifier,
	}
	_, err := Sdk.Ec2.DeleteSubnet(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted publicsubnet [%s]", actual.(*PublicSubnet).Identifier)

	newResource := &PublicSubnet{}
	newResource.Name = actual.(*PublicSubnet).Name
	newResource.Tags = actual.(*PublicSubnet).Tags
	newResource.CIDR = actual.(*PublicSubnet).CIDR
	newResource.Zone = actual.(*PublicSubnet).Zone

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicSubnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("publicsubnet.Render")

	newCluster := inaccurateCluster
	subnet := &cluster.PublicSubnet{}
	subnet.CIDR = newResource.(*PublicSubnet).CIDR
	subnet.Zone = newResource.(*PublicSubnet).Zone
	subnet.Name = newResource.(*PublicSubnet).Name
	subnet.Identifier = newResource.(*PublicSubnet).Identifier

	providerConfig := newCluster.ProviderConfig()
	for i := 0; i < len(providerConfig.Network.PublicSubnets); i++ {
		subnet := providerConfig.Network.PublicSubnets[i]
		if subnet.Name == newResource.(*PublicSubnet).Name {
			subnet.CIDR = newResource.(*PublicSubnet).CIDR
			subnet.Zone = newResource.(*PublicSubnet).Zone
			subnet.Identifier = newResource.(*PublicSubnet).Identifier
			providerConfig.Network.PublicSubnets[i] = subnet
			newCluster.SetProviderConfig(providerConfig)
		}
	}
	return newCluster
}

func (r *PublicSubnet) tag(tags map[string]string) error {
	logger.Debug("publicsubnet.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering PublicSubnet tag [%s] %s", key, val)
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
