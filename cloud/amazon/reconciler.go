package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
)

type Reconciler struct {
	Expected *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Expected: expected,
	}
}

var actual = &cluster.Cluster{}
var expected = &cluster.Cluster{}
var vpc = &resources.Vpc{}

func (r *Reconciler) GetActual(known *cluster.Cluster) (*cluster.Cluster, error) {
	sdk, err := awsSdkGo.NewSdk(known.Location)
	if err != nil {
		return nil, err
	}
	vpc.Init(known, actual, expected, sdk)
	vpc.Parse()
	return actual, nil
}

func (r *Reconciler) GetExpected(known *cluster.Cluster) (*cluster.Cluster, error) {
	return expected, nil
}

func (r *Reconciler) Reconcile(actual, expected *cluster.Cluster) error {
	return vpc.Apply()
}
