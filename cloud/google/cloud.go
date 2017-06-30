package google

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
)

type Reconciler struct {
	Expected *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Expected: expected,
	}
}

func (r *Reconciler) GetActual(known *cluster.Cluster) (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}
func (r *Reconciler) GetExpected(known *cluster.Cluster) (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}
func (r *Reconciler) Reconcile(actual, expected *cluster.Cluster) error {
	return nil
}
