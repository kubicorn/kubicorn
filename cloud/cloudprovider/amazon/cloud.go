package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider"
)

type AmazonCloudProvider struct {
	Expected *cluster.Cluster
}

func NewAmazonCloudProvider(expected *cluster.Cluster) cloudprovider.CloudProvider {
	return &AmazonCloudProvider{
		Expected: expected,
	}
}

func (a *AmazonCloudProvider) GetExpectedServerPoolResources(expectedServerPool *cluster.ServerPool) []cloudprovider.CloudResource {

	var resources []cloudprovider.CloudResource
	return resources
}

func (a *AmazonCloudProvider) ConcurrentApplyResource(resource cloudprovider.CloudResource, errchan chan error) {
}
