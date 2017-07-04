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

var model map[string]cloud.Resource

func (r *Reconciler) Init() error {
	sdk, err := awsSdkGo.NewSdk(r.Known.Location)
	if err != nil {
		return err
	}
	resources.Sdk = sdk
	model = ClusterModel(r.Known)
	return nil
}

func (r *Reconciler) GetActual() (*cluster.Cluster, error) {
	actualCluster := newClusterDefaults(r.Known)
	for _, resource := range model {
		actualResource, err := resource.Actual(r.Known)
		if err != nil {
			return nil, err
		}
		actualCluster, err = resource.Render(actualResource, actualCluster)
		if err != nil {
			return nil, err
		}
	}
	return actualCluster, nil
}

func (r *Reconciler) GetExpected() (*cluster.Cluster, error) {
	expectedCluster := newClusterDefaults(r.Known)
	for _, resource := range model {
		expectedResource, err := resource.Expected(r.Known)
		if err != nil {
			return nil, err
		}
		expectedCluster, err = resource.Render(expectedResource, expectedCluster)
		if err != nil {
			return nil, err
		}
	}
	return expectedCluster, nil
}

func (r *Reconciler) Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error) {
	newCluster := newClusterDefaults(r.Known)
	for _, resource := range model {
		expectedResource, err := resource.Expected(expectedCluster)
		if err != nil {
			return nil, err
		}
		actualResource, err := resource.Actual(actualCluster)
		if err != nil {
			return nil, err
		}
		appliedResource, err := resource.Apply(actualResource, expectedResource)
		if err != nil {
			return nil, err
		}
		newCluster, err = resource.Render(appliedResource, newCluster)
		if err != nil {
			return nil, err
		}
	}
	return newCluster, nil
}

func (r *Reconciler) Destroy() error {
	for _, resource := range model {
		actualResource, err := resource.Actual(r.Known)
		if err != nil {
			return err
		}
		err = resource.Delete(actualResource)
		if err != nil {
			return err
		}
	}
	return nil
}

func newClusterDefaults(base *cluster.Cluster) *cluster.Cluster {
	new := &cluster.Cluster{
		Name:     base.Name,
		Cloud:    base.Cloud,
		Location: base.Location,
		Network:  &cluster.Network{},
	}
	return new
}
