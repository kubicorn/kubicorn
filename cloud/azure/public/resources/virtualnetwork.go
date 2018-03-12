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
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &Vnet{}

type Vnet struct {
	Shared
}

func (r *Vnet) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vnet.Actual")

	newResource := &Vnet{
		Shared: Shared{},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vnet) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vnet.Expected")
	newResource := &Vnet{
		Shared: Shared{},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vnet) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vnet.Apply")
	applyResource := expected.(*Vnet)
	isEqual, err := compare.IsEqual(actual.(*Vnet), expected.(*Vnet))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	newResource := &Vnet{
		Shared: Shared{},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *Vnet) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vnet.Delete")
	deleteResource := actual.(*Vnet)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete VPC resource without ID [%s]", deleteResource.Name)
	}

	newResource := &Vnet{
		Shared: Shared{},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vnet) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("vnet.Render")
	newCluster := inaccurateCluster
	return newCluster
}
