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

package crd

import (
	"fmt"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kube-deploy/ext-apiserver/pkg/client/clientset_generated/clientset"
	"k8s.io/kube-deploy/ext-apiserver/util"
)

const (
	//MasterIPAttempts       = 40
	SleepSecondsPerAttempt = 5
	RetryAttempts          = 30
	//DeleteAttempts         = 150
	//DeleteSleepSeconds     = 5
)

type CRDManager struct {
	client         *kubernetes.Clientset
	kubeConfigPath string
	clientSet      *clientset.Clientset
	cluster        *cluster.Cluster
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
		clientSet:      cs,
		client:         client,
		cluster:        cluster,
	}, nil
}
