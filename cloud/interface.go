package cloud

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type Reconciler interface {
	GetActual(*cluster.Cluster) (*cluster.Cluster, error)
	GetExpected(cluster *cluster.Cluster) (*cluster.Cluster, error)
	Reconcile(actual, expected *cluster.Cluster) error
}
