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

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
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
	NetworkID     string
}

func (r *Subnet) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("subnet.Actual")
	newResource := new(Subnet)

	if r.ClusterSubnet.Identifier != "" {
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
	} else {
		newResource.CIDR = r.ClusterSubnet.CIDR
		newResource.NetworkID = immutable.ProviderConfig().Network.Identifier
	}

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
		IPVersion: gophercloud.IPv4,
		Name:      subnet.Name,
		NetworkID: subnet.NetworkID,
	})
	output, err := res.Extract()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Subnet: %v", err)
	}

	logger.Success("Created Subnet [%s]", output.ID)

	// Attach the subnet to the router
	router := immutable.ProviderConfig().Network.InternetGW
	if router.Identifier != "" {
		_, err = routers.AddInterface(Sdk.Network, router.Identifier, routers.AddInterfaceOpts{
			SubnetID: output.ID,
		}).Extract()

		if err != nil {
			return nil, nil, fmt.Errorf("Unable to attach the subnet to the router: %v", err)
		}
	}

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
	if subnet.Identifier != "" {
		// Dettach the subnet from the router
		router := immutable.ProviderConfig().Network.InternetGW
		if router.Identifier != "" {
			_, err := routers.RemoveInterface(Sdk.Network, router.Identifier, routers.RemoveInterfaceOpts{
				SubnetID: subnet.Identifier,
			}).Extract()
			if err != nil {
				return nil, nil, fmt.Errorf("Unable to dettach the subnet to the default router: %v", err)
			}
		}

		// Delete the subnet
		if res := subnets.Delete(Sdk.Network, subnet.Identifier); res.Err != nil {
			return nil, nil, res.Err
		}
		logger.Success("Deleted Subnet [%s]", actual.(*Subnet).Identifier)
	}

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
	subnet := &cluster.Subnet{}
	subnet.CIDR = newResource.(*Subnet).CIDR
	subnet.Name = newResource.(*Subnet).Name
	subnet.Identifier = newResource.(*Subnet).Identifier
	newCluster := inaccurateCluster
	found := false

	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Subnets); j++ {
			if machineProviderConfig.ServerPool.Subnets[j].Name == subnet.Name {
				found = true
				machineProviderConfig.ServerPool.Subnets[j].Identifier = subnet.Identifier
				machineProviderConfig.ServerPool.Subnets[j].CIDR = subnet.CIDR
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	if !found {
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == newResource.(*Subnet).Name {
				found = true
				machineProviderConfig.ServerPool.Subnets = append(newCluster.ServerPools()[i].Subnets, subnet)
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	if !found {
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == subnet.Name {
				newCluster.ServerPools()[i].Subnets = append(newCluster.ServerPools()[i].Subnets, subnet)
				machineProviderConfig.ServerPool.Subnets = []*cluster.Subnet{subnet}
				machineProviderConfigs[i] = machineProviderConfig
				found = true
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}
	return newCluster
}
