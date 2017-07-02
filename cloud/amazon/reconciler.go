package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
)

type Reconciler struct {
	Known *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Known: expected,
	}
}

var actual = &cluster.Cluster{}
var expected = &cluster.Cluster{}
var vpc = &resources.Vpc{}

func (r *Reconciler) Init() error {
	sdk, err := awsSdkGo.NewSdk(r.Known.Location)
	if err != nil {
		return err
	}
	vpc.Init(r.Known, actual, expected, sdk)
	err = vpc.Parse()
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) GetActual() (*cluster.Cluster, error) {
	vpc.Render()
	return vpc.ActualCluster, nil
}

func (r *Reconciler) GetExpected() (*cluster.Cluster, error) {
	vpc.Render()
	return vpc.ExpectedCluster, nil
}

func (r *Reconciler) Reconcile() error {
	return vpc.Apply()
}

func (r *Reconciler) Destroy() error {
	vpc.Render()
	return vpc.Delete()
}
