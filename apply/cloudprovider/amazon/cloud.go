package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/apply/cloudprovider"
)

type AmazonCloudProvider struct {
}

func NewAmazonCloudProvider() cloudprovider.CloudProvider {
	return &AmazonCloudProvider{}
}

func (a *AmazonCloudProvider) GetServerPool() (*cluster.ServerPool, error) {
	return &cluster.ServerPool{}, nil
}

func (a *AmazonCloudProvider) ApplyServerPool(*cluster.ServerPool) error {
	return nil
}
