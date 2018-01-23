/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

// Controls machines on a specific cloud.
type MachineActuator interface {
	Create(*clusterv1.Cluster, *clusterv1.Machine) error
	Delete(*clusterv1.Machine) error
	Update(c *clusterv1.Cluster, oldMachine *clusterv1.Machine, newMachine *clusterv1.Machine) error
	Get(string) (*clusterv1.Machine, error)
	GetIP(machine *clusterv1.Machine) (string, error)
	GetKubeConfig(master *clusterv1.Machine) (string, error)

	// Create and start the machine controller. The list of initial
	// machines don't have to be reconciled as part of this function, but
	// are provided in case the function wants to refer to them (and their
	// ProviderConfigs) to know how to configure the machine controller.
	CreateMachineController(cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error
	PostDelete(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error
}
