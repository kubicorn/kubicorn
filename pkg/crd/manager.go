package crd

import (
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"k8s.io/kube-deploy/ext-apiserver/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kube-deploy/ext-apiserver/pkg/client/clientset_generated/clientset"
)

const (
	//MasterIPAttempts       = 40
	SleepSecondsPerAttempt = 5
	RetryAttempts          = 30
	//DeleteAttempts         = 150
	//DeleteSleepSeconds     = 5
)

type CRDManager struct {
	client *kubernetes.Clientset
	kubeConfigPath string
	clientSet *clientset.Clientset
	cluster *cluster.Cluster
}

func NewCRDManager(cluster *cluster.Cluster) (*CRDManager, error) {
	kubeConfigPath := kubeconfig.GetKubeConfigPath(cluster)
	cs, err := util.NewClientSet(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize new client set: %v", err)
	}
	client, err := util.NewKubernetesClient(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize API client for machines")
	}
	return &CRDManager{
		kubeConfigPath: kubeConfigPath,
		clientSet: cs,
		client: client,
		cluster: cluster,
	}, nil
}