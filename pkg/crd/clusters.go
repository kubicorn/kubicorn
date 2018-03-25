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

func (m *CRDManager) CreateClusters() error {
	//
	//// ----- Create CRD for Machines -----
	//success := false
	//for i := 0; i <= RetryAttempts; i++ {
	//	_, err := clusterv1.CreateClustersCRD(m.clientSet)
	//	if err != nil && !strings.Contains(err.Error(),"already exists"){
	//		logger.Info("Failure creating clusters CRD: %v", err)
	//		time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
	//		continue
	//	}
	//	success = true
	//	logger.Info("Clusters CRD created successfully!")
	//	//logger.Always("You can now `kubectl get clusters`")
	//	break
	//}
	//if !success {
	//	return fmt.Errorf("error creating Clusters CRD")
	//}
	//
	//// ----- Populate Clusters -----
	//
	//cluster := &clusterv1.Cluster{
	//	ObjectMeta: v1.ObjectMeta{
	//		Name: m.cluster.Name,
	//	},
	//}
	//outputcluster, err := m.client.Clusters().Create(cluster)
	//if err != nil {
	//	logger.Warning("unable to create clusters CRD: %v", err)
	//}
	//logger.Debug("Ensured cluster: %v", outputcluster.Name)

	return nil
}
