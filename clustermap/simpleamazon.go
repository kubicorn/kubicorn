package clustermap

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleAmazonCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name: name,
		ServerPools: []*cluster.ServerPool{
			{
				PoolType: cluster.ServerPoolType_Master,
				Cloud:    cluster.ServerPoolCloud_Amazon,
				Name:     "amazon-master",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Public,
					},
				},
			},
			{
				PoolType: cluster.ServerPoolType_Node,
				Cloud:    cluster.ServerPoolCloud_Amazon,
				Name:     "amazon-node",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Public,
					},
				},
			},
		},
	}
}
