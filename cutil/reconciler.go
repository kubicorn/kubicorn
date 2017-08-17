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

package cutil

import (
	"fmt"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/public"
	"github.com/kris-nova/kubicorn/cloud/digitalocean/droplet"
	"github.com/kris-nova/kubicorn/cloud/google/compute"
)

// GetReconciler gets the correct Reconciler for the cloud provider currenty used.
func GetReconciler(known *cluster.Cluster) (reconciler cloud.Reconciler, err error) {

	switch known.Cloud {
	case cluster.CloudAmazon:
		return cloud.NewAtomicReconciler(known, compute.NewGoogleComputeModel(known)), nil
	case cluster.CloudDigitalOcean:
		return cloud.NewAtomicReconciler(known, droplet.NewDigitalOceanDropletModel(known)), nil
	case cluster.CloudGoogle:
		return cloud.NewAtomicReconciler(known, public.NewAmazonPublicModel(known)), nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", known.Cloud)
	}

}
