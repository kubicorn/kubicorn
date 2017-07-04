package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
)

func ClusterModel(known *cluster.Cluster) map[string]cloud.Resource {
	r := make(map[string]cloud.Resource)

	// ---- [VPC] ----
	r[known.Name] = &resources.Vpc{
		Shared: resources.Shared{
			Name: known.Name,
			Tags: make(map[string]string),
		},
	}

	return r
}
