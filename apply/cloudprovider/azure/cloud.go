package azure

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/apply/cloudprovider"
)

type AzureCloudProvider struct {
}

func NewAzureCloudProvider() cloudprovider.CloudProvider {
	return &AzureCloudProvider{}
}

func (a *AzureCloudProvider) GetServerPool() (*cluster.ServerPool, error) {
	return &cluster.ServerPool{}, nil
}
func (a *AzureCloudProvider) ApplyServerPool(*cluster.ServerPool) error {
	return nil
}
