package crd

import (
	"fmt"
	"time"
	"github.com/kris-nova/kubicorn/cutil/logger"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"

	"strings"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *CRDManager) CreateMachines() error {

	// ----- Create CRD for Machines -----
	success := false
	for i := 0; i <= RetryAttempts; i++ {
		_, err := clusterv1.CreateMachinesCRD(m.clientSet)
		if err != nil && !strings.Contains(err.Error(),"already exists"){
			logger.Info("Failure creating machines CRD: %v", err)
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		success = true
		logger.Info("Machines CRD created successfully!")
		//logger.Always("You can now `kubectl get machines`")
		break
	}
	if !success {
		return fmt.Errorf("error creating Machines CRD")
	}

	// ----- Populate Machines -----
	for _, serverPool := range m.cluster.ServerPools {
		if strings.Contains(serverPool.Name, "node") {
			for i := 0; i < serverPool.MaxCount; i++ {
				name := fmt.Sprintf("%s-%d", serverPool.Name, i)
				machine := &clusterv1.Machine{
					ObjectMeta: v1.ObjectMeta{
						Name: name,
					},
				}
				outputmachine, err := m.client.Machines().Create(machine)
				if err != nil {
					logger.Warning("unable to create new machine: %v", err)
				}
				logger.Debug("Created machine: %s", outputmachine.Name)
			}
		}
	}


	return nil
}

