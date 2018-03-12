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

package initapi

import (
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
)

func validateAtLeastOneMachineSet(initCluster *cluster.Cluster) error {
	if len(initCluster.MachineSets) < 1 {
		return fmt.Errorf("cluster %v must have at least one machine set", initCluster.Name)
	}
	return nil
}

func validateMachineSetMaxCountGreaterThan1(initCluster *cluster.Cluster) error {
	providerConfigs := initCluster.MachineProviderConfigs()
	for _, providerConfig := range providerConfigs {
		p := providerConfig.ServerPool
		if p.MaxCount < 1 {
			return fmt.Errorf("server pool %v in cluster %v must have a maximum count greater than 0", p.Name, initCluster.Name)
		}
	}
	return nil
}

func validateSpotPriceOnlyForAwsCluster(initCluster *cluster.Cluster) error {
	providerConfigs := initCluster.MachineProviderConfigs()
	for _, providerConfig := range providerConfigs {
		p := providerConfig.ServerPool
		if p.AwsConfiguration != nil && p.AwsConfiguration.SpotPrice != "" && initCluster.ProviderConfig().Cloud != cluster.CloudAmazon {
			return fmt.Errorf("Spot price provided for server pool %v can only be used with AWS", p.Name)
		}
	}
	return nil
}
