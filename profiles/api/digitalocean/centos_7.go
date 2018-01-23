package digitalocean

import (
	"github.com/kris-nova/kubicorn/apis"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/profiles/legacy/digitalocean"
)

// NewCentosCluster creates a basic Azure cluster profile, to bootstrap Kubernetes.
func NewCentosCluster(name string) apis.KubicornCluster {

	providerConfig, _ := profiles.SerializeProviderConfig(digitalocean.NewCentosCluster(name))

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

