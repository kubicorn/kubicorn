package state

import "github.com/kris-nova/kubicorn/apis/cluster"

type ClusterStorer interface {
	Exists() bool
	Commit(cluster *cluster.Cluster) error
	Destroy() error
	GetCluster() (*cluster.Cluster, error)
	List() ([]string, error)
}
