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
	"github.com/kris-nova/kubicorn/cloud/amazon"
	"github.com/kris-nova/kubicorn/cloud/digitalocean"
)

func GetReconciler(c *cluster.Cluster) (cloud.Reconciler, error) {
	switch c.Cloud {
	case cluster.CloudAmazon:
		return amazon.NewReconciler(c), nil
	case cluster.CloudDigitalOcean:
		return digitalocean.NewReconciler(c), nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", c.Cloud)
	}

}
