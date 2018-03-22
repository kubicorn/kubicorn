package resourcedeploy

import (
	"github.com/kubicorn/kubicorn/apis/cluster"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/api/core/v1"
)

const (
	KubicornDefaultNamespace = "kubicorn"
)


func DeployClusterControllerDeployment(cluster *cluster.Cluster) (error) {
	var kubeConfigPath *string
	kstr := kubeconfig.GetConfig(cluster)

	config, err := clientcmd.BuildConfigFromFlags("", kstr)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	deploymentsClient := clientset.AppsV1beta2().Deployments(KubicornDefaultNamespace)
	_, err = deploymentsClient.Create(cluster.ControllerDeployment)
	if err != nil {
		return err
	}
	return nil
}
