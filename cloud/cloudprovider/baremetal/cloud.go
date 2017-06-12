package baremetal

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider"
)

type BaremetalCloudProvider struct {
	Expected *cluster.Cluster
}

func NewBaremetalCloudProvider(expected *cluster.Cluster) cloudprovider.CloudProvider {
	return &BaremetalCloudProvider{
		Expected: expected,
	}
}

func (b *BaremetalCloudProvider) GetExpectedServerPoolResources(expectedServerPool *cluster.ServerPool) []cloudprovider.CloudResource {
	var resources []cloudprovider.CloudResource
	return resources
}

func (b *BaremetalCloudProvider) ConcurrentApplyResource(resource cloudprovider.CloudResource, errchan chan error) {
}
