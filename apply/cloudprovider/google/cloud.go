package google

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/apply/cloudprovider"
)

type GoogleCloudProvider struct {
}

func NewGoogleCloudProvider() cloudprovider.CloudProvider {
	return &GoogleCloudProvider{}
}

func (g *GoogleCloudProvider) GetServerPool() (*cluster.ServerPool, error) {
	return &cluster.ServerPool{}, nil
}

func (g *GoogleCloudProvider) ApplyServerPool(*cluster.ServerPool) error {
	return nil
}
