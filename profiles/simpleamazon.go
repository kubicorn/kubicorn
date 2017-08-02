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

package profiles

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/kubeadm"
)

func NewSimpleAmazonCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:     name,
		Cloud:    cluster.Cloud_Amazon,
		Location: "us-west-2",
		Ssh: &cluster.Ssh{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "ubuntu",
		},
		KubernetesApi: &cluster.KubernetesApi{
			Port: "443",
		},
		Network: &cluster.Network{
			Type: cluster.NetworkType_Public,
			CIDR: "10.0.0.0/16",
		},
		Values: &cluster.Values{
			ItemMap: map[string]string{
				"INJECTEDTOKEN": kubeadm.GetRandomToken(),
			},
		},
		ServerPools: []*cluster.ServerPool{
			{
				Type:            cluster.ServerPoolType_Master,
				Name:            fmt.Sprintf("%s.master", name),
				MaxCount:        1,
				MinCount:        1,
				Image:           "ami-835b4efa",
				Size:            "t2.medium",
				BootstrapScript: "amazon_k8s_1.7.0_ubuntu_16.04_master.sh",
				Subnets: []*cluster.Subnet{
					{
						Name:     fmt.Sprintf("%s.master", name),
						CIDR:     "10.0.0.0/24",
						Location: "us-west-2a",
					},
				},

				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s.master-external", name),
						Rules: []*cluster.Rule{
							{
								IngressFromPort: 22,
								IngressToPort:   22,
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressFromPort: 443,
								IngressToPort:   443,
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
						},
					},
				},
			},
			{
				Type:            cluster.ServerPoolType_Node,
				Name:            fmt.Sprintf("%s.node", name),
				MaxCount:        1,
				MinCount:        1,
				Image:           "ami-835b4efa",
				Size:            "t2.medium",
				BootstrapScript: "amazon_k8s_1.7.0_ubuntu_16.04_node.sh",
				Subnets: []*cluster.Subnet{
					{
						Name:     fmt.Sprintf("%s.node", name),
						CIDR:     "10.0.100.0/24",
						Location: "us-west-2b",
					},
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s.node-external", name),
						Rules: []*cluster.Rule{
							{
								IngressFromPort: 22,
								IngressToPort:   22,
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
						},
					},
				},
			},
		},
	}
}
