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
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeadm"
	"github.com/kubicorn/kubicorn/pkg/ptrconvenient"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewControllerUbuntuCluster creates a simple Ubuntu Amazon cluster
func NewControllerUbuntuCluster(name string) *cluster.Cluster {

	controlPlaneProviderConfig := &cluster.ControlPlaneProviderConfig{
		Cloud:    cluster.CloudDigitalOcean,
		Location: "nyc3",
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
		Components: &cluster.Components{
			ComponentVPN: false,
		},
	}

	machineSetsProviderConfigs := []*cluster.MachineProviderConfig{
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     fmt.Sprintf("%s-master", name),
				MaxCount: 1,
				MinCount: 1,
				Image:    "ubuntu-16-04-x64",
				Size:     "s-2vcpu-2gb",
				BootstrapScripts: []string{
					"bootstrap/digitalocean_k8s_ubuntu_16.04_master.sh",
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s-master", name),
						IngressRules: []*cluster.IngressRule{
							{
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "443",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "1194",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "udp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-node", name),
								IngressProtocol: "tcp",
							},
						},
						EgressRules: []*cluster.EgressRule{
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "tcp",
							},
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "udp",
							},
						},
					},
				},
			},
		},
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s-node", name),
				MaxCount: 1,
				MinCount: 1,
				Image:    "ubuntu-16-04-x64",
				Size:     "s-1vcpu-2gb",
				BootstrapScripts: []string{
					"bootstrap/digitalocean_k8s_ubuntu_16.04_node.sh",
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s-node", name),
						IngressRules: []*cluster.IngressRule{
							{
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "1194",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "udp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-master", name),
								IngressProtocol: "tcp",
							},
						},
						EgressRules: []*cluster.EgressRule{
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "tcp",
							},
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "udp",
							},
						},
					},
				},
			},
		},
	}

	deployment := &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubicorn-controller",
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: ptrconvenient.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kubicorn-controller",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kubicorn-controller",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "kubicorn-controller",
							Image: "kubicorn/controller:latest",
							//Ports: []apiv1.ContainerPort{
							//	{
							//		Name:          "http",
							//		Protocol:      apiv1.ProtocolTCP,
							//		ContainerPort: 80,
							//	},
							//},
						},
					},
				},
			},
		},
	}

	c := cluster.NewCluster(name)
	c.SetProviderConfig(controlPlaneProviderConfig)
	c.NewMachineSetsFromProviderConfigs(machineSetsProviderConfigs)

	//
	//
	// Here we define the replicas for the controller
	//
	//
	cpms := c.MachineSets[1]
	cpms.Spec.Replicas = ptrconvenient.Int32Ptr(3)
	c.MachineSets[1] = cpms
	//
	//
	//
	//
	//

	c.ControllerDeployment = deployment
	return c
}
