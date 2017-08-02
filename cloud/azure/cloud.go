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

package azure

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
)

type Reconciler struct {
	Known *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Known: expected,
	}
}

func (r *Reconciler) Init() error {
	return nil
}
func (r *Reconciler) GetActual() (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}
func (r *Reconciler) GetExpected() (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}

func (r *Reconciler) Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error) {
	return nil, nil
}
func (r *Reconciler) Destroy() error {
	return nil
}
