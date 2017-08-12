package snap

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"fmt"
)

type KubernetesQuery struct {
	kubeconfigPath string
	namespaces []string
}

func NewKubernetesQuery(kubeConfigPath string, namespaces []string) *KubernetesQuery {
	return &KubernetesQuery{
		kubeconfigPath: kubeConfigPath,
		namespaces: namespaces,
	}
}

func (q *KubernetesQuery) Execute() error {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", q.kubeconfigPath)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	for _, namespace := range q.namespaces {
		ns, err := clientset.CoreV1().Namespaces().Get(namespace, meta_v1.GetOptions{})
		if err != nil {
			return err
		}
		fmt.Println(ns)
	}
	return nil
}

func (q *KubernetesQuery) Bytes() ([]byte) {
	return []byte("")
}
