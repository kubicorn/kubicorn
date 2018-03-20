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
	clusterv1 "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Controls machines on a specific cloud. All methods should be idempotent unless otherwise specified.
type MachineActuator interface {
	// Create the machine.
	Create(*clusterv1.Cluster, *clusterv1.Machine) error
	// Delete the machine.
	Delete(*clusterv1.Machine) error
	// Update the machine to the provided definition.
	Update(c *clusterv1.Cluster, machine *clusterv1.Machine) error
	// Checks if the machine currently exists.
	Exists(*clusterv1.Machine) (bool, error)
}
