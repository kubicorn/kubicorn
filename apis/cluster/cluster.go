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

package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Cloud_Amazon       = "amazon"
	Cloud_Azure        = "azure"
	Cloud_Google       = "google"
	Cloud_Baremetal    = "baremetal"
	Cloud_DigitalOcean = "digitalocean"
)

type Cluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Name              string
	ServerPools       []*ServerPool
	Cloud             string
	Location          string
	Ssh               *Ssh
	Network           *Network
	Values            *Values
	KubernetesApi     *KubernetesApi
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Name: name,
	}
}
