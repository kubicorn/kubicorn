package google

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider"
)

type GoogleCloudProvider struct {
	Expected *cluster.Cluster
}

func NewGoogleCloudProvider(expected *cluster.Cluster) cloudprovider.CloudProvider {
	return &GoogleCloudProvider{
		Expected: expected,
	}
}

func (g *GoogleCloudProvider) GetExpectedServerPoolResources(expectedServerPool *cluster.ServerPool) []cloudprovider.CloudResource {
	var resources []cloudprovider.CloudResource
	return resources
}

func (g *GoogleCloudProvider) ConcurrentApplyResource(resource cloudprovider.CloudResource, errchan chan error) {
}
