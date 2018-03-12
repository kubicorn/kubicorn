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

package ovh

import (
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeadm"
)

// NewUbuntuCluster creates a simple Ubuntu Openstack cluster.
func NewUbuntuCluster(name string) *cluster.Cluster {
	var (
		masterName = fmt.Sprintf("%s-master", name)
		nodeName   = fmt.Sprintf("%s-node", name)
	)
	controlPlaneProviderConfig := &cluster.ControlPlaneProviderConfig{
		Cloud:    cluster.CloudOVH,
		Location: "BHS1",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "ubuntu",
		},
		Values: &cluster.Values{
			ItemMap: map[string]string{
				"INJECTEDTOKEN": kubeadm.GetRandomToken(),
			},
		},
		KubernetesAPI: &cluster.KubernetesAPI{
			Port: "443",
		},
		Network: &cluster.Network{
			Type: cluster.NetworkTypePrivate,
		},
	}
	machineSetsProviderConfigs := []*cluster.MachineProviderConfig{
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     masterName,
				MaxCount: 1,
				Image:    "Ubuntu 16.04",
				Size:     "b2-15",
				BootstrapScripts: []string{
					"bootstrap/ovh_k8s_ubuntu_16.04_master.sh",
				},
				Subnets: []*cluster.Subnet{
					{
						Name: "internal",
						CIDR: "192.168.0.0/16",
					},
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: masterName,
						IngressRules: []*cluster.IngressRule{
							{
								IngressFromPort: "22",
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressFromPort: "443",
								IngressToPort:   "443",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressSource: "192.168.0.0/16",
							},
						},
					},
				},
			},
		},
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeNode,
				Name:     nodeName,
				MaxCount: 2,
				Image:    "Ubuntu 16.04",
				Size:     "b2-7",
				BootstrapScripts: []string{
					"bootstrap/ovh_k8s_ubuntu_16.04_node.sh",
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: nodeName,
						IngressRules: []*cluster.IngressRule{
							{
								IngressFromPort: "22",
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressSource: "192.168.0.0/16",
							},
						},
					},
				},
			},
		},
	}
	c := cluster.NewCluster(name)
	c.SetProviderConfig(controlPlaneProviderConfig)
	c.NewMachineSetsFromProviderConfigs(machineSetsProviderConfigs)
	return c
}
