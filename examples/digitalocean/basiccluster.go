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

package main

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/initapi"
	"github.com/kris-nova/kubicorn/cutil/kubeadm"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

func main() {
	logger.Level = 4
	cluster := getCluster("mycluster")
	cluster, err := initapi.InitCluster(cluster)
	if err != nil {
		panic(err.Error())
	}
	reconciler, err := cutil.GetReconciler(cluster)
	if err != nil {
		panic(err.Error())
	}

	err = reconciler.Init()
	if err != nil {
		panic(err.Error())
	}
	expected, err := reconciler.GetExpected()
	if err != nil {
		panic(err.Error())
	}
	actual, err := reconciler.GetActual()
	if err != nil {
		panic(err.Error())
	}
	created, err := reconciler.Reconcile(actual, expected)
	logger.Info("Created cluster [%s]", created.Name)
	if err != nil {
		panic(err.Error())
	}
	_, err = reconciler.Destroy()
	if err != nil {
		panic(err.Error())
	}
}

func getCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:     name,
		Cloud:    cluster.CloudDigitalOcean,
		Location: "sfo2",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "root",
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
				Image:    "ubuntu-16-04-x64",
				Size:     "1gb",
				BootstrapScripts: []string{
					"vpn/meshbirdMaster.sh",
					"digitalocean_k8s_ubuntu_16.04_master.sh",
				},
			},
			{
				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s-node", name),
				MaxCount: 3,
				Image:    "ubuntu-16-04-x64",
				Size:     "1gb",
				BootstrapScripts: []string{
					"vpn/meshbirdNode.sh",
					"digitalocean_k8s_ubuntu_16.04_node.sh",
				},
			},
		},
	}
}
