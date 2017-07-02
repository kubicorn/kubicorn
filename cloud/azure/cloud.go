package azure

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
)

type Reconciler struct {
	Known *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Known: expected,
	}
}

func (r *Reconciler) Init() error {
	return nil
}
func (r *Reconciler) GetActual() (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}
func (r *Reconciler) GetExpected() (*cluster.Cluster, error) {
	return &cluster.Cluster{}, nil
}
func (r *Reconciler) Reconcile() error {
	return nil
}
func (r *Reconciler) Destroy() error {
	return nil
}
