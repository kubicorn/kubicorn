package crd

import (
	"github.com/kris-nova/kubicorn/cutil/kubeconfig"
	"fmt"
	"k8s.io/kube-deploy/cluster-api/util"
	client2 "k8s.io/kube-deploy/cluster-api/client"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"github.com/kris-nova/kubicorn/apis/cluster"
)

const (
	//MasterIPAttempts       = 40
	SleepSecondsPerAttempt = 5
	RetryAttempts          = 30
	//DeleteAttempts         = 150
	//DeleteSleepSeconds     = 5
)

type CRDManager struct {
	client *client2.ClusterAPIV1Alpha1Client
	kubeConfigPath string
	clientSet *apiextensionsclient.Clientset
	cluster *cluster.Cluster
}

func NewCRDManager(cluster *cluster.Cluster) (*CRDManager, error) {
	kubeConfigPath, err := kubeconfig.KubeConfigPath()
	if err != nil {
		return nil, fmt.Errorf("unable to find kube config path: %v", err)
	}
	cs, err := util.NewClientSet(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize new client set: %v", err)
	}
	client, err := util.NewApiClient(kubeConfigPath)
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
