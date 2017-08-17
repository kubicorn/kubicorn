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

// Model will model resources to an API representation of infrastructure
type Model interface {

	// Get will return the map of ordered resources against a cloud
	Get() map[int]Resource
}

// Reconciler will create and destroy infrastructure based on an intended state. A Reconciler will
// also audit the expected and actual state.
type Reconciler interface {

	//SetModel will set the model to use for the cloud reconciler
	SetModel(model Model)

	// GetActual will audit a cloud and return the API representation of the current resources in the cloud
	GetActual(known *cluster.Cluster) (*cluster.Cluster, error)

	// GetExpected will audit a state store and return the API representation of the intended resources in the cloud
	GetExpected(known *cluster.Cluster) (*cluster.Cluster, error)

	// Reconcile will take an actual and expected API representation and attempt to ensure the intended state
	Reconcile(actual, expected *cluster.Cluster) (*cluster.Cluster, error)

	// Destroy will take an actual API representation and destroy the resources in the cloud
	Destroy(actual *cluster.Cluster) (*cluster.Cluster, error)
}

// Resource represents a single cloud level resource that can be mutated. Resources are mapped via a model.
type Resource interface {

	// Actual will return the current existing resource in the cloud if it exists.
	Actual(immutable *cluster.Cluster) (Resource, error)

	// Expected will return the anticipated cloud resource.
	Expected(immutable *cluster.Cluster) (Resource, error)

	// Apply will create a cloud resource if needed.
	Apply(actual, expected Resource, immutable *cluster.Cluster) (*cluster.Cluster, Resource, error)

	// Delete will delete a cloud resource if needed.
	Delete(actual Resource, immutable *cluster.Cluster) (*cluster.Cluster, Resource, error)
}
