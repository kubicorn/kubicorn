package resourcedeploy

import (
	"github.com/kubicorn/kubicorn/apis/cluster"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/api/core/v1"
	"fmt"
	"strings"
)

const (
	KubicornDefaultNamespace = "kubicorn"
)

func clientSet(cluster *cluster.Cluster) (*kubernetes.Clientset, error) {
	kubeConfigPath := kubeconfig.GetKubeConfigPath(cluster)
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to load kube config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Unable to load client set: %v", err)
	}
	return clientset, nil
}

func EnsureNamespace(cluster *cluster.Cluster) (error) {
	clientset, err := clientSet(cluster)
	if err != nil {
		return err
	}
	namespaceClient := clientset.CoreV1().Namespaces()
	namespace := &v1.Namespace{}
	namespace.Name = KubicornDefaultNamespace
	_, err = namespaceClient.Create(namespace)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Unable to ensure namespace: %v", err)
	}
	return nil
}

func DeployClusterControllerDeployment(cluster *cluster.Cluster) (error) {
	err := EnsureNamespace(cluster)
	if err != nil {
		return err
	}
	clientset, err := clientSet(cluster)
	if err != nil {
		return err
	}
	deploymentsClient := clientset.AppsV1beta2().Deployments(KubicornDefaultNamespace)
	_, err = deploymentsClient.Create(cluster.ControllerDeployment)
	if err != nil {
		return err
	}
	return nil
}
