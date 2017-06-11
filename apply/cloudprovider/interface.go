package cloudprovider

import "github.com/kris-nova/kubicorn/apis/cluster"

type CloudProvider interface {
	GetServerPool() (*cluster.ServerPool, error)
	ApplyServerPool(pool *cluster.ServerPool) error
}
