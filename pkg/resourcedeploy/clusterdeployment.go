// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resourcedeploy

import (
	"fmt"
	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func EnsureNamespace(cluster *cluster.Cluster) error {
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

func DeployClusterControllerDeployment(cluster *cluster.Cluster) error {
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
