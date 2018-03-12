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

package amazon

import (
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeadm"
	"github.com/kubicorn/kubicorn/pkg/uuid"
)

// NewCentosCluster creates a simple CentOS Amazon cluster
func NewCentosCluster(name string) *cluster.Cluster {

	controlPlaneProviderConfig := &cluster.ControlPlaneProviderConfig{
		Cloud:    cluster.CloudAmazon,
		Location: "us-west-2",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "centos",
		},
		KubernetesAPI: &cluster.KubernetesAPI{
			Port: "443",
		},
		Network: &cluster.Network{
			Type:       cluster.NetworkTypePublic,
			CIDR:       "10.0.0.0/16",
			InternetGW: &cluster.InternetGW{},
		},
		Values: &cluster.Values{
			ItemMap: map[string]string{
				"INJECTEDTOKEN": kubeadm.GetRandomToken(),
			},
		},
	}
	machineSetsProviderConfigs := []*cluster.MachineProviderConfig{
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     fmt.Sprintf("%s.master", name),
				MaxCount: 1,
				MinCount: 1,
				Image:    "ami-0c2aba6c",
				Size:     "t2.xlarge",
				BootstrapScripts: []string{
					"bootstrap/amazon_k8s_centos_7_master.sh",
				},
				InstanceProfile: &cluster.IAMInstanceProfile{
					Name: fmt.Sprintf("%s-KubicornMasterInstanceProfile", name),
					Role: &cluster.IAMRole{
						Name: fmt.Sprintf("%s-KubicornMasterRole", name),
						Policies: []*cluster.IAMPolicy{
							{
								Name: "MasterPolicy",
								Document: `{
								  "Version": "2012-10-17",
								  "Statement": [
									 {
										"Effect": "Allow",
										"Action": [
										   "ec2:*",
										   "elasticloadbalancing:*",
										   "ecr:GetAuthorizationToken",
										   "ecr:BatchCheckLayerAvailability",
										   "ecr:GetDownloadUrlForLayer",
										   "ecr:GetRepositoryPolicy",
										   "ecr:DescribeRepositories",
										   "ecr:ListImages",
										   "ecr:BatchGetImage",
										   "autoscaling:DescribeAutoScalingGroups",
										   "autoscaling:UpdateAutoScalingGroup"
										],
										"Resource": "*"
									 }
								  ]
								}`,
							},
						},
					},
				},
				Subnets: []*cluster.Subnet{
					{
						Name: fmt.Sprintf("%s.master", name),
						CIDR: "10.0.0.0/24",
						Zone: "us-west-2a",
					},
				},

				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s.master-external-%s", name, uuid.TimeOrderedUUID()),
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
								IngressFromPort: "0",
								IngressToPort:   "65535",
								IngressSource:   "10.0.100.0/24",
								IngressProtocol: "-1",
							},
						},
					},
				},
			},
		},
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s.node", name),
				MaxCount: 1,
				MinCount: 1,
				Image:    "ami-0c2aba6c",
				Size:     "t2.medium",
				BootstrapScripts: []string{
					"bootstrap/amazon_k8s_centos_7_node.sh",
				},
				InstanceProfile: &cluster.IAMInstanceProfile{
					Name: fmt.Sprintf("%s-KubicornNodeInstanceProfile", name),
					Role: &cluster.IAMRole{
						Name: fmt.Sprintf("%s-KubicornNodeRole", name),
						Policies: []*cluster.IAMPolicy{
							{
								Name: "NodePolicy",
								Document: `{
								  "Version": "2012-10-17",
								  "Statement": [
									 {
										"Effect": "Allow",
										"Action": [
										   "ec2:Describe*",
										   "ec2:AttachVolume",
										   "ec2:DetachVolume",
										   "ecr:GetAuthorizationToken",
										   "ecr:BatchCheckLayerAvailability",
										   "ecr:GetDownloadUrlForLayer",
										   "ecr:GetRepositoryPolicy",
										   "ecr:DescribeRepositories",
										   "ecr:ListImages",
										   "ecr:BatchGetImage",
										   "autoscaling:DescribeAutoScalingGroups",
										   "autoscaling:UpdateAutoScalingGroup"
										],
										"Resource": "*"
									 }
								  ]
								}`,
							},
						},
					},
				},
				Subnets: []*cluster.Subnet{
					{
						Name: fmt.Sprintf("%s.node", name),
						CIDR: "10.0.100.0/24",
						Zone: "us-west-2b",
					},
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s.node-external-%s", name, uuid.TimeOrderedUUID()),
						IngressRules: []*cluster.IngressRule{
							{
								IngressFromPort: "22",
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressFromPort: "0",
								IngressToPort:   "65535",
								IngressSource:   "10.0.0.0/24",
								IngressProtocol: "-1",
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
