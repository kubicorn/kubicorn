package cloudprovider

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type CloudProvider interface {
	GetExpectedServerPoolResources(expectedServerPool *cluster.ServerPool) []CloudResource
	ConcurrentApplyResource(resource CloudResource, errchan chan error)
}

type CloudResource interface {
	Apply() error
}
