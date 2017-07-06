package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
)

func ClusterModel(known *cluster.Cluster) map[int]cloud.Resource {
	r := make(map[int]cloud.Resource)
	i := 0

	// ---- [VPC] ----
	r[i] = &resources.Vpc{
		Shared: resources.Shared{
			Name: known.Name,
			Tags: make(map[string]string),
		},
	}
	vpcIndex := i
	i++

	for _, serverPool := range known.ServerPools {
		name := serverPool.Name
		for _, subnet := range serverPool.Subnets {
			r[i] = &resources.Subnet{
				Shared: resources.Shared{
					Name:        name,
					Tags:        make(map[string]string),
					TagResource: r[vpcIndex],
				},
				ServerPool: serverPool,
				ClusterSubnet: subnet,
			}
			i++
		}

		// ---- [Launch Configuration] ----
		r[i] = &resources.Lc{
			Shared: resources.Shared{
				Name:        name,
				Tags:        make(map[string]string),
				TagResource: r[vpcIndex],
			},
			ServerPool: serverPool,
		}
		i++

		// ---- [Autoscale Group] ----
		r[i] = &resources.Asg{
			Shared: resources.Shared{
				Name:        name,
				Tags:        make(map[string]string),
				TagResource: r[vpcIndex],
			},
			ServerPool: serverPool,
		}
		i++
	}

	return r
}
