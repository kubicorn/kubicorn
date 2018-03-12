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

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/rackspace/gophercloud/pagination"
)

var _ cloud.Resource = &Subnet{}

type Subnet struct {
	Shared
	ClusterSubnet *cluster.Subnet
	CIDR          string
	NetworkID     string
}

func (r *Subnet) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("subnet.Actual")
	newResource := new(Subnet)

	// Find the subnet by name
	res := subnets.List(Sdk.Network, subnets.ListOpts{
		Name: r.Name,
	})
	if res.Err != nil {
		return nil, nil, err
	}
	err = res.EachPage(func(page pagination.Page) (bool, error) {
		list, err := subnets.ExtractSubnets(page)
		if err != nil {
			return false, err
		}
		if len(list) > 1 {
			return false, fmt.Errorf("Found more than one subnet with name [%s]", newResource.Name)
		}
		if len(list) == 1 {
			newResource.Identifier = list[0].ID
			newResource.CIDR = list[0].CIDR
			newResource.NetworkID = list[0].NetworkID
			newResource.Name = list[0].Name
		}
		return false, nil
	})

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Expected")
	newResource := &Subnet{
		Shared: Shared{
			Name:       r.Name,
			Identifier: r.ClusterSubnet.Identifier,
		},
		CIDR:      r.ClusterSubnet.CIDR,
		NetworkID: immutable.ProviderConfig().Network.Identifier,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Apply")
	subnet := expected.(*Subnet)
	isEqual, err := compare.IsEqual(actual.(*Subnet), expected.(*Subnet))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, subnet, nil
	}
	// Create the subnet
	res := subnets.Create(Sdk.Network, subnets.CreateOpts{
		CIDR:      subnet.CIDR,
		IPVersion: subnets.IPv4,
		Name:      subnet.Name,
		NetworkID: subnet.NetworkID,
	})
	output, err := res.Extract()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Subnet: %v", err)
	}

	logger.Success("Created Subnet [%s]", output.ID)

	newResource := &Subnet{
		Shared: Shared{
			Name:       output.Name,
			Identifier: output.ID,
		},
		NetworkID: output.NetworkID,
		CIDR:      output.CIDR,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("subnet.Delete")
	subnet := actual.(*Subnet)

	// Delete the subnet
	if res := subnets.Delete(Sdk.Network, subnet.Identifier); res.Err != nil {
		return nil, nil, res.Err
	}
	logger.Success("Deleted Subnet [%s]", actual.(*Subnet).Identifier)

	newResource := &Subnet{
		Shared: Shared{
			Name: subnet.Name,
		},
		CIDR: subnet.CIDR,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Subnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("subnet.Render")
	newCluster := inaccurateCluster
	newSubnet := newResource.(*Subnet)
	subnet := new(cluster.Subnet)
	subnet.CIDR = newSubnet.CIDR
	subnet.Name = newSubnet.Name
	subnet.Identifier = newSubnet.Identifier
	found := false

	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Subnets); j++ {
			if machineProviderConfig.ServerPool.Subnets[j].Name == newSubnet.Name {
				machineProviderConfig.ServerPool.Subnets[j].CIDR = newSubnet.CIDR
				machineProviderConfig.ServerPool.Subnets[j].Identifier = newSubnet.Identifier
				found = true
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}
	if !found {
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == newResource.(*Subnet).Name {
				machineProviderConfig.ServerPool.Subnets = append(newCluster.ServerPools()[i].Subnets, subnet)
				found = true
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}
	if !found {

		providerConfig := []*cluster.MachineProviderConfig{
			{
				ServerPool: &cluster.ServerPool{
					Name: newSubnet.Name,
					Subnets: []*cluster.Subnet{
						subnet,
					},
				},
			},
		}
		newCluster.NewMachineSetsFromProviderConfigs(providerConfig)

	}

	return newCluster
}
