package cutil

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon"
	"github.com/kris-nova/kubicorn/cloud/azure"
	"github.com/kris-nova/kubicorn/cloud/baremetal"
	"github.com/kris-nova/kubicorn/cloud/google"
)

func GetReconciler(c *cluster.Cluster) (cloud.Reconciler, error) {
	switch c.Cloud {
	case cluster.Cloud_Amazon:
		return amazon.NewReconciler(c), nil
	case cluster.Cloud_Azure:
		return azure.NewReconciler(c), nil
	case cluster.Cloud_Baremetal:
		return baremetal.NewReconciler(c), nil
	case cluster.Cloud_Google:
		return google.NewReconciler(c), nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", c.Cloud)
	}
}
