package baremetal

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/apply/cloudprovider"
)

type BaremetalCloudProvider struct {
}

func NewBaremetalCloudProvider() cloudprovider.CloudProvider {
	return &BaremetalCloudProvider{}
}

func (b *BaremetalCloudProvider) GetServerPool() (*cluster.ServerPool, error) {
	return &cluster.ServerPool{}, nil
}

func (b *BaremetalCloudProvider) ApplyServerPool(*cluster.ServerPool) error {
	return nil
}
