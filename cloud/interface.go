package cloud

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
)

type Reconciler interface {
	Init() error
	GetActual() (*cluster.Cluster, error)
	GetExpected() (*cluster.Cluster, error)
	Reconcile() error
	Destroy() error
}
