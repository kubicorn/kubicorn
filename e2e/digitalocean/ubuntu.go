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

package digitalocean

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/initapi"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/e2e/tutil/clustername"
	"github.com/kris-nova/kubicorn/profiles/digitalocean"
)

// CreateDOUbuntuCluster creates new Ubuntu cluster with 3 nodes
// on the DigitalOcean platform and returns cluster object and reconciler.
func CreateDOUbuntuCluster() (*cluster.Cluster, cloud.Reconciler, error) {

	// Logger level
	logger.Level = 4

	// Create new cluster named e2e-cluster-do and initialize reconciler.
	cluster := digitalocean.NewUbuntuCluster(clustername.GetClusterName("do"))
	cluster, err := initapi.InitCluster(cluster)
	if err != nil {
		return nil, nil, err
	}
	reconciler, err := cutil.GetReconciler(cluster, nil)
	if err != nil {
		return nil, nil, err
	}

	// Get expected and actual states.
	expected, err := reconciler.Expected(cluster)
	if err != nil {
		return nil, nil, err
	}
	actual, err := reconciler.Actual(cluster)
	if err != nil {
		return nil, nil, err
	}

	// Reconcile the cluster.
	c, err := reconciler.Reconcile(actual, expected)
	return c, reconciler, err
}

// DestroyDOUbuntuCluster destroys provided Ubuntu cluster on the
// DigitalOcean platform. In case of failure, error is returned.
func DestroyDOUbuntuCluster(reconciler cloud.Reconciler) error {
	_, err := reconciler.Destroy()
	if err != nil {
		return err
	}
	return nil
}
