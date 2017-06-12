package azure

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/cloudprovider"
)

type AzureCloudProvider struct {
	Expected *cluster.Cluster
}

func NewAzureCloudProvider(expected *cluster.Cluster) cloudprovider.CloudProvider {
	return &AzureCloudProvider{
		Expected: expected,
	}
}

func (a *AzureCloudProvider) GetExpectedServerPoolResources(expectedServerPool *cluster.ServerPool) []cloudprovider.CloudResource {
	var resources []cloudprovider.CloudResource
	return resources
}

func (a *AzureCloudProvider) ConcurrentApplyResource(resource cloudprovider.CloudResource, errchan chan error) {
}
