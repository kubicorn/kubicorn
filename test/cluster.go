package test

import (
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/initapi"
)

// Create will create a new test cluster
func Create(testCluster *cluster.Cluster) (*cluster.Cluster, error) {
	testCluster, err := initapi.InitCluster(testCluster)
	if err != nil {
		return nil, err
	}
	reconciler, err := cutil.GetReconciler(testCluster)
	if err != nil {
		return nil, err
	}

	err = reconciler.Init()
	if err != nil {
		return nil, err
	}
	expected, err := reconciler.GetExpected()
	if err != nil {
		return nil, err
	}
	actual, err := reconciler.GetActual()
	if err != nil {
		return nil, err
	}
	created, err := reconciler.Reconcile(actual, expected)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Read will read a test cluster
func Read(testCluster *cluster.Cluster) (*cluster.Cluster, error) {
	reconciler, err := cutil.GetReconciler(testCluster)
	if err != nil {
		return nil, err
	}
	err = reconciler.Init()
	if err != nil {
		return nil, err
	}
	actual, err := reconciler.GetActual()
	if err != nil {
		return nil, err
	}
	return actual, nil
}

// Update will update a test cluster
func Update(testCluster *cluster.Cluster) (*cluster.Cluster, error) {
	testCluster, err := initapi.InitCluster(testCluster)
	if err != nil {
		return nil, err
	}
	reconciler, err := cutil.GetReconciler(testCluster)
	if err != nil {
		return nil, err
	}

	err = reconciler.Init()
	if err != nil {
		return nil, err
	}
	expected, err := reconciler.GetExpected()
	if err != nil {
		return nil, err
	}
	actual, err := reconciler.GetActual()
	if err != nil {
		return nil, err
	}
	updated, err := reconciler.Reconcile(actual, expected)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete will delete a test cluster
func Delete(testCluster *cluster.Cluster) (*cluster.Cluster, error) {
	reconciler, err := cutil.GetReconciler(testCluster)
	if err != nil {
		return nil, err
	}
	err = reconciler.Init()
	if err != nil {
		return nil, err
	}
	err = reconciler.Destroy()
	if err != nil {
		return nil, err
	}
	return nil, nil
}
