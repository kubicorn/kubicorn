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

package crd

import (
	"fmt"
	"time"

	"github.com/kubicorn/kubicorn/pkg/logger"

	"strings"

	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

func (m *CRDManager) CreateMachines() error {

	// ----- Create CRD for Machines -----
	success := false
	for i := 0; i <= RetryAttempts; i++ {
		_, err := v1alpha1.CreateMachinesCRD(*m.clientSet)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			logger.Info("Failure creating machines CRD: %v", err)
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		success = true
		logger.Info("Machines CRD created successfully!")
		//logger.Always("You can now `kubectl get machines`")
		break
	}

	if !success {
		return fmt.Errorf("error creating Machines CRD")
	}
	//

	//// ----- Populate Machines -----
	//for _, serverPool := range m.cluster.ServerPools {
	//	if strings.Contains(serverPool.Name, "node") {
	//		for i := 0; i < serverPool.MaxCount; i++ {
	//			name := fmt.Sprintf("%s-%d", serverPool.Name, i)
	//			machine := &clusterv1.Machine{
	//				ObjectMeta: v1.ObjectMeta{
	//					Name: name,
	//				},
	//			}
	//			outputmachine, err := m.client.Machines().Create(machine)
	//			if err != nil {
	//				logger.Warning("unable to create new machine: %v", err)
	//			}
	//			logger.Debug("Created machine: %s", outputmachine.Name)
	//		}
	//	}
	//}

	return nil
}
