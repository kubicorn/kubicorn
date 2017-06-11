package state

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/state/stores"
)

func DisableStateStore(s stores.ClusterStorer, cluster *cluster.Cluster) error {
	// Todo figure out how to disable
	return nil
}

func DestroyStateStore(s stores.ClusterStorer, cluster *cluster.Cluster) error {
	return s.Destroy()
}
