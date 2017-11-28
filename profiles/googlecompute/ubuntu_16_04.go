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

package googlecompute

import (
	"fmt"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/kubeadm"
	"github.com/kris-nova/kubicorn/apis"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// NewUbuntuCluster creates a basic Azure cluster profile, to bootstrap Kubernetes.
func NewUbuntuClusterA(name string) apis.KubicornCluster {


	cluster := v1alpha1.Cluster{

		ObjectMeta: metav1.ObjectMeta{
			// ------------------------------------------------------------------
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			// ------------------------------------------------------------------
		},
		Spec: v1alpha1.ClusterSpec{
			// ------------------------------------------------------------------
		},
	}


	return &cluster

}

// NewUbuntuCluster creates a basic Ubuntu Google Compute cluster.
func NewUbuntuCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:     name,
		CloudId:  "example-id",
		Cloud:    cluster.CloudGoogle,
		Location: "us-central1-a",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "ubuntu",
		},
		KubernetesAPI: &cluster.KubernetesAPI{
			Port: "443",
		},
		Values: &cluster.Values{
			ItemMap: map[string]string{
				"INJECTEDTOKEN": kubeadm.GetRandomToken(),
			},
		},
		ServerPools: []*cluster.ServerPool{
			{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     fmt.Sprintf("%s-master", name),
				MaxCount: 1,
				Image:    "ubuntu-1604-xenial-v20170811",
				Size:     "n1-standard-1",
				BootstrapScripts: []string{
					"bootstrap/google_compute_k8s_ubuntu_16.04_master.sh",
				},
			},
			{
				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s-node", name),
				MaxCount: 2,
				Image:    "ubuntu-1604-xenial-v20170811",
				Size:     "n1-standard-1",
				BootstrapScripts: []string{
					"bootstrap/google_compute_k8s_ubuntu_16.04_node.sh",
				},
			},
		},
	}
}
