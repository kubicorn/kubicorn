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

package cloud

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type Reconciler interface {
	Init() error
	GetActual() (*cluster.Cluster, error)
	GetExpected() (*cluster.Cluster, error)
	Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error)
	Destroy() (*cluster.Cluster, error)
}

type Resource interface {
	Actual(known *cluster.Cluster) (Resource, error)
	Expected(known *cluster.Cluster) (Resource, error)
	Apply(actual, expected Resource, expectedCluster *cluster.Cluster) (Resource, error)
	Delete(actual Resource, known *cluster.Cluster) (Resource, error)
	Render(renderResource Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error)
	Tag(tags map[string]string) error
}
