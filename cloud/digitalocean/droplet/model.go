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

package droplet

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/digitalocean/droplet/resources"
)

type Model struct {
	known           *cluster.Cluster
	cachedResources map[int]cloud.Resource
}

func NewDigitalOceanDropletModel(known *cluster.Cluster) cloud.Model {
	return &Model{
		known: known,
	}
}

// ClusterModel maps cluster info to DigitalOcean Resources.
func (m *Model) Resources() map[int]cloud.Resource {
	if len(m.cachedResources) > 0 {
		return m.cachedResources
	}
	r := make(map[int]cloud.Resource)
	i := 0

	// ---- [SSH Key] ----
	r[i] = &resources.SSH{
		Shared: resources.Shared{
			Name: m.known.Name,
		},
	}
	i++

	for _, serverPool := range m.known.ServerPools {
		// ---- [Droplet] ----
		r[i] = &resources.Droplet{
			Shared: resources.Shared{
				Name: serverPool.Name,
			},
			ServerPool: serverPool,
		}
		i++
	}

	for _, serverPool := range m.known.ServerPools {
		for _, firewall := range serverPool.Firewalls {
			// ---- [Firewall] ----
			f := &resources.Firewall{
				Shared: resources.Shared{
					Name:    serverPool.Name,
					CloudID: firewall.Identifier,
				},
				Tags:       []string{serverPool.Name},
				ServerPool: serverPool,
			}

			for _, rule := range firewall.IngressRules {
				InboundRule := resources.InboundRule{
					Protocol:  rule.IngressProtocol,
					PortRange: rule.IngressToPort,
					Source: &resources.Sources{
						Addresses: []string{rule.IngressSource},
					},
				}
				f.InboundRules = append(f.InboundRules, InboundRule)
			}
			for _, rule := range firewall.EgressRules {
				OutboundRule := resources.OutboundRule{
					Protocol:  rule.EgressProtocol,
					PortRange: rule.EgressToPort,
					Destinations: &resources.Destinations{
						Addresses: []string{rule.EgressDestination},
					},
				}
				f.OutboundRules = append(f.OutboundRules, OutboundRule)
			}
			r[i] = f
			i++
		}
	}

	m.cachedResources = r
	return m.cachedResources
}
