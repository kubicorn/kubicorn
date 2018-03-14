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

package compute

import (
	"testing"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud/google/compute/resources"
)

func TestClusterModelHappy(t *testing.T) {
	machineProviderConfig := []*cluster.MachineProviderConfig{
		{
			ServerPool: &cluster.ServerPool{
				Name: "ServerPool1",
			},
		},
		{
			ServerPool: &cluster.ServerPool{
				Name: "ServerPool2",
			},
		},
	}
	c := cluster.NewCluster("test_cluster")
	c.NewMachineSetsFromProviderConfigs(machineProviderConfig)
	result := NewGoogleComputeModel(c)

	if len(result.Resources()) != 2 {
		t.Fatalf("Amount of serverpools is incorrect")
	}

	if result.Resources()[0].(*resources.InstanceGroup).Name != "ServerPool1" {
		t.Fatalf("Serverpool first name is incorrect")
	}

	if result.Resources()[1].(*resources.InstanceGroup).Name != "ServerPool2" {
		t.Fatalf("Serverpool first name is incorrect")
	}
}
