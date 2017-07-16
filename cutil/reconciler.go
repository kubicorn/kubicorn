package cutil

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon"
	"github.com/kris-nova/kubicorn/cloud/digitalocean"
)

func GetReconciler(c *cluster.Cluster) (cloud.Reconciler, error) {
	switch c.Cloud {
	case cluster.Cloud_Amazon:
		return amazon.NewReconciler(c), nil
	case cluster.Cloud_DigitalOcean:
		return digitalocean.NewReconciler(c), nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", c.Cloud)
	}
}
