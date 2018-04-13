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

var _ cloud.Resource = &Subnet{}

type Subnet struct {
	Shared
	ClusterSubnet *cluster.Subnet
	ServerPool    *cluster.ServerPool
	CIDR          string
	VpcID         string
	Zone          string
}

func (r *Subnet) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Actual")
	newResource := &Subnet{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		input := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{&r.ClusterSubnet.Identifier},
		}
		output, err := Sdk.Ec2.DescribeSubnets(input)
		if err != nil {
			return nil, nil, err
		}
		lsn := len(output.Subnets)
		if lsn != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Subnets for ID [%s]", lsn, r.ClusterSubnet.Identifier)
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
		newResource.CIDR = r.ClusterSubnet.CIDR
		newResource.Zone = r.ClusterSubnet.Zone
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Expected")
	newResource := &Subnet{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Identifier,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
			},
		},
		CIDR:  r.ClusterSubnet.CIDR,
		VpcID: immutable.ProviderConfig().Network.Identifier,
		Zone:  r.ClusterSubnet.Zone,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Apply")
	applyResource := expected.(*Subnet)
	isEqual, err := compare.IsEqual(actual.(*Subnet), applyResource)
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
		return nil, nil, fmt.Errorf("Unable to create new Subnet: %v", err)
	}
	waitInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []*string{output.Subnet.SubnetId},
	}
	logger.Info("Waiting for Subnet [%s] to be available", *output.Subnet.SubnetId)
	err = Sdk.Ec2.WaitUntilSubnetAvailable(waitInput)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created Subnet [%s]", *output.Subnet.SubnetId)

	newResource := &Subnet{
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
		return nil, nil, fmt.Errorf("Unable to tag new Subnet: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Delete")
	deleteResource := actual.(*Subnet)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Subnet resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSubnetInput{
		SubnetId: &deleteResource.Identifier,
	}
	_, err := Sdk.Ec2.DeleteSubnet(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Subnet [%s]", deleteResource.Identifier)

	newResource := &Subnet{
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

func (r *Subnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("subnet.Render")

	subnet := &cluster.Subnet{
		CIDR:       newResource.(*Subnet).CIDR,
		Identifier: newResource.(*Subnet).Identifier,
		Name:       newResource.(*Subnet).Name,
		Zone:       newResource.(*Subnet).Zone,
	}
	found := false

	newCluster := inaccurateCluster
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Subnets); j++ {
			subnet := machineProviderConfig.ServerPool.Subnets[j]
			if subnet.Name == newResource.(*Subnet).Name {
				subnet.CIDR = newResource.(*Subnet).CIDR
				subnet.Zone = newResource.(*Subnet).Zone
				subnet.Identifier = newResource.(*Subnet).Identifier
				machineProviderConfig.ServerPool.Subnets[j] = subnet
				machineProviderConfigs[i] = machineProviderConfig
				found = true
			}
		}
	}

	if !found {
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == newResource.(*Subnet).Name {
				newCluster.ServerPools()[i].Subnets = append(newCluster.ServerPools()[i].Subnets, subnet)
				machineProviderConfig.ServerPool.Subnets = []*cluster.Subnet{subnet}
				machineProviderConfigs[i] = machineProviderConfig
				found = true
			}
		}
	}

	if !found {
		machineProviderConfigs = []*cluster.MachineProviderConfig{
			{
				ServerPool: &cluster.ServerPool{
					Name: newResource.(*Subnet).Name,
					Subnets: []*cluster.Subnet{
						subnet,
					},
				},
			},
		}
	}

	newCluster.SetMachineProviderConfigs(machineProviderConfigs)
	return newCluster
}

func (r *Subnet) tag(tags map[string]string) error {
	logger.Debug("subnet.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Subnet tag [%s] %s", key, val)
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
