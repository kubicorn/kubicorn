package cloud

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type Reconciler interface {
	GetActual(known *cluster.Cluster) (*cluster.Cluster, error)
	GetExpected(known *cluster.Cluster) (*cluster.Cluster, error)
	Reconcile(actual, expected *cluster.Cluster) error
}
