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
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/pagination"
)

var _ cloud.Resource = &Network{}

type Network struct {
	Shared
}

func (r *Network) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("network.Actual")
	newResource := new(Network)

	if immutable.ProviderConfig().Network.Identifier != "" {
		// Find the network by name
		res := networks.List(Sdk.Network, networks.ListOpts{
			Name: r.Name,
		})
		if res.Err != nil {
			return nil, nil, err
		}
		err = res.EachPage(func(page pagination.Page) (bool, error) {
			list, err := networks.ExtractNetworks(page)
			if err != nil {
				return false, err
			}
			if len(list) > 1 {
				return false, fmt.Errorf("Found more than one network with name [%s]", newResource.Name)
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

func (r *Network) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("network.Expected")
	newResource := &Network{
		Shared: Shared{
			Name:       r.Name,
			Identifier: immutable.ProviderConfig().Network.Identifier,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Network) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("network.Apply")
	network := expected.(*Network)
	isEqual, err := compare.IsEqual(actual.(*Network), expected.(*Network))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, network, nil
	}
	// Create the network
	res := networks.Create(Sdk.Network, networks.CreateOpts{
		Name:         network.Name,
		AdminStateUp: networks.Up,
	})
	output, err := res.Extract()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Network: %v", err)
	}

	logger.Success("Created Network [%s]", output.ID)

	newResource := &Network{
		Shared: Shared{
			Name:       network.Name,
			Identifier: output.ID,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Network) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("network.Delete")
	network := actual.(*Network)

	// Delete the network
	if res := networks.Delete(Sdk.Network, network.Identifier); res.Err != nil {
		return nil, nil, res.Err
	}
	logger.Success("Deleted Network [%s]", actual.(*Network).Identifier)

	newResource := &Network{
		Shared: Shared{
			Name: network.Name,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Network) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("network.Render")
	network := newResource.(*Network)
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Network.Identifier = network.Identifier
	providerConfig.Network.Name = network.Name
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
