package amazon

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
	"github.com/kris-nova/kubicorn/logger"
	"strings"
	"time"
	"github.com/kris-nova/kubicorn/cutil/hang"
)

type Reconciler struct {
	Known *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Known: expected,
	}
}

var model map[int]cloud.Resource

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
	for i := 0; i < len(model); i++ {
		resource := model[i]
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
	for i := 0; i < len(model); i++ {
		resource := model[i]
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

func cleanUp(cluster *cluster.Cluster, i int) error {
	for j := i - 1; i >= 0; i-- {
		resource := model[j]
		actualResource, err := resource.Actual(cluster)
		if err != nil {
			err, i = destroyI(err, j)
			if err != nil {
				return err
			}
			continue
		}
		err = resource.Delete(actualResource)
		if err != nil {
			err, i = destroyI(err, j)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (r *Reconciler) Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error) {
	newCluster := newClusterDefaults(r.Known)
	for i := 0; i < len(model); i++ {
		resource := model[i]
		expectedResource, err := resource.Expected(expectedCluster)
		if err != nil {
			return nil, err
		}
		actualResource, err := resource.Actual(actualCluster)
		if err != nil {
			return nil, err
		}
		appliedResource, err := resource.Apply(actualResource, expectedResource, newCluster)
		if err != nil {
			logger.Critical("Error during apply! Attempting cleaning: %v", err)
			err = cleanUp(actualCluster, i)
			if err != nil {
				logger.Critical("Failure during cleanup! Abandoned resources!")
				return nil, err
			}
			return nil,  nil
		}
		newCluster, err = resource.Render(appliedResource, newCluster)
		if err != nil {
			return nil, err
		}
	}
	return newCluster, nil
}

var destroyRetryStrings = []string{
	"DependencyViolation:",
}

var hg = &hang.Hanger{
	Ratio: 1,
}

func destroyI(err error, i int) (error, int) {
	hg.Hang()
	for _, retryString := range destroyRetryStrings {
		if strings.Contains(err.Error(), retryString) {
			logger.Debug("Retry failed delete: %v", err)
			time.Sleep(1 * time.Second)
			return nil, i + 1
		}
	}
	return err, 0
}



func (r *Reconciler) Destroy() error {
	for i := len(model) -1; i >= 0; i-- {
		resource := model[i]
		actualResource, err := resource.Actual(r.Known)
		if err != nil {
			err, i = destroyI(err, i)
			if err != nil {
				return err
			}
			continue
		}
		err = resource.Delete(actualResource)
		if err != nil {
			err, i = destroyI(err, i)
			if err != nil {
				return err
			}
			continue
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
