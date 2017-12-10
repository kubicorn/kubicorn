package crd

import (
	"time"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func (m *CRDManager) CreateClusters() error {

	// ----- Create CRD for Machines -----
	success := false
	for i := 0; i <= RetryAttempts; i++ {
		 _, err := clusterv1.CreateClustersCRD(m.clientSet)
		 if err != nil && !strings.Contains(err.Error(),"already exists"){
			logger.Info("Failure creating clusters CRD: %v", err)
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		success = true
		logger.Info("Clusters CRD created successfully!")
		//logger.Always("You can now `kubectl get clusters`")
		break
	}
	if !success {
		return fmt.Errorf("error creating Clusters CRD")
	}

	// ----- Populate Clusters -----

	cluster := &clusterv1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: m.cluster.Name,
		},
	}
	outputcluster, err := m.client.Clusters().Create(cluster)
	if err != nil {
		logger.Warning("unable to create clusters CRD: %v", err)
	}
	logger.Debug("Ensured cluster: %v", outputcluster.Name)


	return nil
}
