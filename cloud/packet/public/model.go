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

package public

import (
	"encoding/json"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/cloud/packet/public/resources"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

type Model struct {
	known           *cluster.Cluster
	cachedResources map[int]cloud.Resource
}

func NewPacketPublicModel(known *cluster.Cluster) cloud.Model {
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

	// ---- [Project] ----
	projectName := ""
	if known.ProviderConfig().Project != nil {
		projectName = known.ProviderConfig().Project.Name
	}
	r[i] = &resources.Project{
		Shared: resources.Shared{
			Name: projectName,
			Tags: []string{},
		},
	}
	i++

	// ---- [Key Pair] ----
	r[i] = &resources.SSH{
		Shared: resources.Shared{
			Name: m.known.Name,
		},
	}
	i++

	//
	machineConfigs := known.MachineProviderConfigs()
	for _, machineConfig := range machineConfigs {
		serverPool := machineConfig.ServerPool
		name := serverPool.Name

		// ---- [Device] ----
		r[i] = &resources.Device{
			Shared: resources.Shared{
				Name: name,
			},
			ServerPool: serverPool,
		}
		i++
	}

	m.cachedResources = r
	j, _ := json.Marshal(r)
	logger.Debug("PacketModel: %v", string(j))
	return m.cachedResources
}
