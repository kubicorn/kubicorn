package profiles

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleAzureCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:  name,
		Cloud: cluster.Cloud_Azure,
		ServerPools: []*cluster.ServerPool{
			{
				Type:     cluster.ServerPoolType_Master,
				Name:     "azure-master",
				MaxCount: 2,
				MinCount: 2,
			},
			{
				Type:     cluster.ServerPoolType_Node,
				Name:     "azure-node",
				MaxCount: 2,
				MinCount: 2,
			},
		},
	}
}
