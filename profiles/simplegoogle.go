package profiles

import "github.com/kris-nova/kubicorn/apis/cluster"

func NewSimpleGoogleCluster(name string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:  name,
		Cloud: cluster.Cloud_Google,
		ServerPools: []*cluster.ServerPool{
			{
				Type: cluster.ServerPoolType_Master,
				Name: "google-master",
			},
			{
				Type: cluster.ServerPoolType_Node,
				Name: "google-node",
			},
		},
	}
}
