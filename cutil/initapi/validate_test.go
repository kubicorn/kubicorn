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
	"github.com/kris-nova/kubicorn/apis/cluster"
	"testing"
)

func TestServerPoolCountsHappy(t *testing.T) {
	c := &cluster.Cluster{
		Name: "c",
		ServerPools: []*cluster.ServerPool{
			{
				Name:     "p",
				MaxCount: 1,
				MinCount: 1,
			},
		},
	}
	err := serverPoolCounts(c)
	if err != nil {
		t.Fatalf("error message incorrect for valid server pool counts"+
			"should be: nil\n"+
			"got:       %v\n", err)
	}
}

func TestServerPoolCountsSad(t *testing.T) {
	tt := []struct {
		name     string
		cluster  *cluster.Cluster
		expected string
	}{
		{"no server pools", emptyCluster(), "cluster c must have at least one server pool"},
		{"min count of 0", badMinCount(), "server pool p in cluster c must have a minimum count greater than 0"},
		{"min count of 0", badMaxCount(), "server pool p in cluster c must have a maximum count greater than 0"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := serverPoolCounts(tc.cluster)
			if actual.Error() != tc.expected {
				t.Fatalf("error message incorrect for %v\n"+
					"should be: %v\n"+
					"got:       %v\n", tc.name, tc.expected, actual)
			}
		})
	}
}

func emptyCluster() *cluster.Cluster {
	return cluster.NewCluster("c")
}

func badMinCount() *cluster.Cluster {
	return &cluster.Cluster{
		Name: "c",
		ServerPools: []*cluster.ServerPool{
			{
				Name:     "p",
				MaxCount: 1,
				MinCount: 0,
			},
		},
	}
}

func badMaxCount() *cluster.Cluster {
	return &cluster.Cluster{
		Name: "c",
		ServerPools: []*cluster.ServerPool{
			{
				Name:     "p",
				MaxCount: 0,
				MinCount: 1,
			},
		},
	}
}
