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
	"github.com/kris-nova/kubicorn/apis/cluster"
	"testing"
)

func TestExpectedHappy(t *testing.T) {
	instance := Instance{
		Shared: Shared{
			Name: "SharedName",
		},
		ServerPool: &cluster.ServerPool{
			Identifier: "ClusterPool1",
			Size:       "5",
			Image:      "server-os-image",
			MaxCount:   5,
			BootstrapScripts: []string{
				"script1.sh",
			},
		},
	}

	knownCluster := &cluster.Cluster{
		Name: "ClusterName",
		SSH: &cluster.SSH{
			PublicKeyFingerprint: "fingerprint",
		},
		Location: "Location-us",
	}

	resource, err := instance.Expected(knownCluster)
	if err != nil {
		t.Fatalf("Error while creating resource %v", err)
	}

	tt := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Shared.cloudId", resource.(*Instance).Shared.CloudID, "ClusterPool1"},
		{"Size", resource.(*Instance).Size, "5"},
		{"Label Amount", len(resource.(*Instance).Labels), 1},
		{"Label group", resource.(*Instance).Labels["group"], "SharedName"},
		{"Location", resource.(*Instance).Location, "Location-us"},
		{"Image", resource.(*Instance).Image, "server-os-image"},
		{"Count", resource.(*Instance).Count, 5},
		{"SSHFingerprint", resource.(*Instance).SSHFingerprint, "fingerprint"},
		{"Bootstrapscript", resource.(*Instance).BootstrapScripts[0], "script1.sh"},
		{"Cache", resource, instance.CachedExpected},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.actual != tc.expected {
				t.Fatalf("%v should be %v got %v\n", tc.name, tc.expected, tc.actual)
			}
		})
	}
}
