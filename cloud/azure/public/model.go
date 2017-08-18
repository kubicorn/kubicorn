package public

import "github.com/kris-nova/kubicorn/cloud"

type Model struct {
	known           *cluster.Cluster
	cachedResources map[int]cloud.Resource
}

func NewAmazonPublicModel(known *cluster.Cluster) cloud.Model {
	return &Model{
		known: known,
	}
}

func (m *Model) Resources() map[int]cloud.Resource {
	cloud.Resource()
}
