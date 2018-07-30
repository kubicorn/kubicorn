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

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &Router{}

type Router struct {
	Shared
}

func (r *Router) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("router.Actual")
	newResource := new(Router)

	// TODO @xmudrii: This is bad. Like, VERY bad. We should we fix this, but let's first get OVH working.
	if immutable.ProviderConfig() != nil && immutable.ProviderConfig().Network != nil &&
		immutable.ProviderConfig().Network.InternetGW != nil && immutable.ProviderConfig().Network.InternetGW.Identifier != "" {
		// Find the router by name
		res := routers.List(Sdk.Network, routers.ListOpts{
			Name: r.Name,
		})
		if res.Err != nil {
			return nil, nil, err
		}
		err = res.EachPage(func(page pagination.Page) (bool, error) {
			list, err := routers.ExtractRouters(page)
			if err != nil {
				return false, err
			}
			if len(list) > 1 {
				return false, fmt.Errorf("Found more than one Router with name [%s]", newResource.Name)
			}
			if len(list) == 1 {
				newResource.Identifier = list[0].ID
				newResource.Name = r.Name
			}
			return false, nil
		})
		if err != nil {
			return nil, nil, err
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Router) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("router.Expected")
	newResource := &Router{
		Shared: Shared{
			Name:       r.Name,
			Identifier: r.Identifier,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Router) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("router.Apply")
	router := expected.(*Router)
	isEqual, err := compare.IsEqual(actual.(*Router), expected.(*Router))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, router, nil
	}
	// Get the public network ID
	// TODO: public network name must be parameterized
	allPages, err := networks.List(Sdk.Network, networks.ListOpts{
		Name: "PublicNetwork",
	}).AllPages()
	if err != nil {
		return nil, nil, err
	}

	list, err := networks.ExtractNetworks(allPages)
	if err != nil {
		return nil, nil, err
	}
	if len(list) == 0 {
		return nil, nil, fmt.Errorf("Public network not found")
	}
	if len(list) > 1 {
		return nil, nil, fmt.Errorf("Found more than one network")
	}

	// Create the router
	gwi := routers.GatewayInfo{
		NetworkID: list[0].ID,
	}
	res := routers.Create(Sdk.Network, routers.CreateOpts{
		Name:        router.Name,
		GatewayInfo: &gwi,
	})
	output, err := res.Extract()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Network: %v", err)
	}

	logger.Success("Created Router [%s]", output.ID)

	newResource := &Router{
		Shared: Shared{
			Name:       router.Name,
			Identifier: output.ID,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Router) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("router.Delete")
	router := actual.(*Router)

	if router.Identifier != "" {
		// Clear all the router interfaces
		allPages, err := ports.List(Sdk.Network, &ports.ListOpts{
			DeviceID: router.Identifier,
		}).AllPages()
		if err != nil {
			return nil, nil, err
		}
		allPorts, err := ports.ExtractPorts(allPages)
		if err != nil {
			return nil, nil, err
		}
		for _, port := range allPorts {
			if res := routers.RemoveInterface(Sdk.Network, router.Identifier, &routers.RemoveInterfaceOpts{
				PortID: port.ID,
			}); res.Err != nil {
				return nil, nil, res.Err
			}
		}

		// Delete the router
		if res := routers.Delete(Sdk.Network, router.Identifier); res.Err != nil {
			return nil, nil, res.Err
		}
		logger.Success("Deleted Router [%s]", actual.(*Router).Identifier)
	}

	newResource := &Router{
		Shared: Shared{
			Name: router.Name,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Router) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("network.Render")
	router := newResource.(*Router)
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Network.InternetGW.Identifier = router.Identifier
	providerConfig.Network.InternetGW.Name = router.Name
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
