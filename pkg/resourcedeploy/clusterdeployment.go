package resourcedeploy

import (
	"github.com/kubicorn/kubicorn/apis/cluster"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"k8s.io/client-go/kubernetes"
	"fmt"
)

const (
	KubicornDefaultNamespace = "kubicorn"
)


func DeployClusterControllerDeployment(cluster *cluster.Cluster) (error) {
	kubeConfigPath := kubeconfig.GetKubeConfigPath(cluster)
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return fmt.Errorf("Unable to load kube config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Unable to load client set: %v", err)
	}
	deploymentsClient := clientset.AppsV1beta2().Deployments(KubicornDefaultNamespace)
	_, err = deploymentsClient.Create(cluster.ControllerDeployment)
	if err != nil {
		return err
	}
	return nil
}
