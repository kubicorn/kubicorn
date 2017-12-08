package amazon

import (
	"github.com/kris-nova/kubicorn/apis"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"fmt"
	"github.com/kris-nova/kubicorn/cutil/kubeadm"
	"github.com/kris-nova/kubicorn/cutil/uuid"
	//"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewUbuntuCluster creates a basic Azure cluster profile, to bootstrap Kubernetes.
func NewUbuntuCluster(name string) apis.KubicornCluster {

	providerConfig, _ := profiles.SerializeProviderConfig(NewUbuntuControlPlane(name))

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
			ProviderConfig: providerConfig,
			ClusterNetwork: v1alpha1.ClusterNetworkingConfig{
				// --------------------------------------------------------------
			},
		},
	}


	return &cluster

}

func NewUbuntuControlPlane(name string) apis.KubicornCluster {
	return &cluster.Cluster{
		Name:     name,
		Cloud:    cluster.CloudAmazon,
		Location: "us-west-2",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "ubuntu",
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
		ServerPools: []*cluster.ServerPool{
			{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     fmt.Sprintf("%s.master", name),
				MaxCount: 1,
				MinCount: 1,
				Image:    "ami-835b4efa",
				Size:     "t2.xlarge",
				BootstrapScripts: []string{
					"bootstrap/amazon_k8s_ubuntu_16.04_master.sh",
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
				AwsConfiguration: &cluster.AwsConfiguration{},
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
			{

				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s.node", name),
				MaxCount: 0,
				MinCount: 0,
				Image:    "ami-835b4efa",
				Size:     "t2.medium",
				BootstrapScripts: []string{
					"bootstrap/amazon_k8s_ubuntu_16.04_node.sh",
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
				AwsConfiguration: &cluster.AwsConfiguration{},
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
}