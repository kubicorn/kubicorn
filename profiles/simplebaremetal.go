package profiles

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleBareMetal(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:  name,
		Cloud: cluster.Cloud_Baremetal,
		ServerPools: []*cluster.ServerPool{
			{
				Type:     cluster.ServerPoolType_Hybrid,
				Name:     "baremetal-hybrid",
				MaxCount: 2,
				MinCount: 2,
			},
		},
	}
}
