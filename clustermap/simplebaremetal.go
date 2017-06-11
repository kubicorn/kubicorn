package clustermap

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleBareMetal(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name: name,
		ServerPools: []*cluster.ServerPool{
			{
				PoolType: cluster.ServerPoolType_Hybrid,
				Cloud:    cluster.ServerPoolCloud_Baremetal,
				Name:     "baremetal-hybrid",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Local,
					},
				},
			},
		},
	}
}
