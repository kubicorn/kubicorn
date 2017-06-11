package clustermap

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleGoogleCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name: name,
		ServerPools: []*cluster.ServerPool{
			{
				PoolType: cluster.ServerPoolType_Master,
				Cloud:    cluster.ServerPoolCloud_Google,
				Name:     "google-master",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Public,
					},
				},
			},
			{
				PoolType: cluster.ServerPoolType_Node,
				Cloud:    cluster.ServerPoolCloud_Google,
				Name:     "google-node",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Public,
					},
				},
			},
		},
	}
}
