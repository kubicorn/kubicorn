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

package ovh

import (
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/cloud/openstack/operator/generic/resources"
	ovhr "github.com/kubicorn/kubicorn/cloud/openstack/operator/ovh/resources"
)

type Model struct {
	known           *cluster.Cluster
	cachedResources map[int]cloud.Resource
}

func NewOvhPublicModel(known *cluster.Cluster) cloud.Model {
	return &Model{
		known: known,
	}
}

func (m *Model) Resources() map[int]cloud.Resource {
	if len(m.cachedResources) > 0 {
		return m.cachedResources
	}

	known := m.known

	r := make(map[int]cloud.Resource)
	i := 0

	// ---- [Key Pair] ----
	r[i] = &resources.KeyPair{
		Shared: resources.Shared{
			Name: known.Name,
		},
	}
	i++

	// ---- [Network] ----
	r[i] = &resources.Network{
		Shared: resources.Shared{
			Name: known.Name,
		},
	}
	i++

	for _, pool := range known.ServerPools() {
		// ---- [Subnet] ----
		for _, subnet := range pool.Subnets {
			r[i] = &resources.Subnet{
				Shared: resources.Shared{
					Name: subnet.Name,
				},
				ClusterSubnet: subnet,
			}
			i++
		}

		// ---- [Security group] ----
		for _, firewall := range pool.Firewalls {
			r[i] = &resources.SecurityGroup{
				Shared: resources.Shared{
					Name: firewall.Name,
				},
				Firewall:   firewall,
				ServerPool: pool,
			}
			i++
		}

		// ---- [Instance] ----
		r[i] = &ovhr.InstanceGroup{
			Shared: resources.Shared{
				Name: pool.Name,
			},
			ServerPool: pool,
		}
		i++
	}

	m.cachedResources = r
	return m.cachedResources
}
