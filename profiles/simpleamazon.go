package profiles

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleAmazonCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:  name,
		Cloud: cluster.Cloud_Amazon,
		ServerPools: []*cluster.ServerPool{
			{
				PoolType: cluster.ServerPoolType_Master,
				Name:     "amazon-master",
				Networks: []*cluster.Network{
					{
						NetworkType: cluster.NetworkType_Public,
					},
				},
			},
			{
				PoolType: cluster.ServerPoolType_Node,
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
