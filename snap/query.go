package snap

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/cutil/logger"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesQuery struct {
	kubeconfigPath string
	namespaces     []string
	config         *restclient.Config
	clientset      *kubernetes.Clientset
	result         []byte
}

func NewKubernetesQuery(kubeConfigPath string, namespaces []string) *KubernetesQuery {
	return &KubernetesQuery{
		kubeconfigPath: kubeConfigPath,
		namespaces:     namespaces,
	}
}

func (q *KubernetesQuery) Authenticate() error {
	config, err := clientcmd.BuildConfigFromFlags("", q.kubeconfigPath)
	if err != nil {
		return err
	}
	q.config = config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	q.clientset = clientset
	return nil
}

type header struct {
	apiVersion string
	kind       string
}

func (q *KubernetesQuery) Execute() error {

	// Support for wildcard * namespaces
	if len(q.namespaces) == 1 && q.namespaces[0] == "*" {
		nss, err := q.GetAllNamespaces()
		if err != nil {
			return fmt.Errorf("Unable to look up namespaces: %v", err)
		}
		q.namespaces = nss
	}

	for _, namespace := range q.namespaces {
		logger.Info("Querying: %s", namespace)

		// ----- Pods -----
		pods, err := q.clientset.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})
		if err != nil {
			return fmt.Errorf("Invalid namespace [%s]: %v", namespace, err)
		}
		for _, pod := range pods.Items {
			headerData, err := yaml.Marshal(header{
				apiVersion: pods.APIVersion,
				kind:       pods.Kind,
			})
			if err != nil {
				return err
			}
			fmt.Println(pod.Kind)
			fmt.Println(pod.APIVersion)
			bb, _ := pod.Marshal()
			fmt.Println(string(bb))

			data, err := yaml.Marshal(pod)
			if err != nil {
				return fmt.Errorf("Unable to marshal pod: %v", err)
			}

			data = append(headerData, data...)
			logger.Debug("Pod: %s", pod.Name)
			q.result = appendData(q.result, data)
		}
	}
	return nil
}

func appendData(a, b []byte) []byte {
	sepStr := `
---

`
	sep := []byte(sepStr)
	a = append(a, sep...)
	a = append(a, b...)
	return a
}

func (q *KubernetesQuery) GetAllNamespaces() ([]string, error) {
	nsl, err := q.clientset.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var namespaces []string
	for _, namespace := range nsl.Items {
		namespaces = append(namespaces, namespace.Namespace)
	}
	return namespaces, nil
}

func (q *KubernetesQuery) Bytes() []byte {
	return q.result
}
