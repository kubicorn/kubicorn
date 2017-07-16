package digitalocean

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/digitalocean/resources"
)

func ClusterModel(known *cluster.Cluster) map[int]cloud.Resource {
	r := make(map[int]cloud.Resource)
	i := 0

	for _, serverPool := range known.ServerPools {
		// ---- [Autoscale Group] ----
		r[i] = &resources.Droplet{
			Shared: resources.Shared{
				Name: serverPool.Name,
			},
			ServerPool: serverPool,
		}
		i++
	}
	return r
}
