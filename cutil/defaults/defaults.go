package defaults

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewClusterDefaults(base *cluster.Cluster) *cluster.Cluster {
	new := &cluster.Cluster{
		Name:          base.Name,
		Cloud:         base.Cloud,
		Location:      base.Location,
		Network:       &cluster.Network{},
		SSH:           &cluster.SSH{},
		Values:        base.Values,
		KubernetesAPI: base.KubernetesAPI,
		ServerPools:   base.ServerPools,
	}
	return new
}
