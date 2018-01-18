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

package healthcheck

import (
	"fmt"
	"time"

	"github.com/kris-nova/kubicorn/cutil/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

// RetryVerifyNodeCount waits for expected number of nodes to come up.
func RetryVerifyNodeCount(client *k8s.Clientset, expectedNodes int) error {
	for i := 0; i <= retryAttempts; i++ {
		cnt, err := verifyNodeCount(client, expectedNodes)
		if err != nil || cnt != expectedNodes {
			logger.Debug("Waiting for Nodes to be created..")
			time.Sleep(time.Duration(retrySleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timedout waiting nodes to be created.")
}

// verifyNodeCount returns number of nodes in the cluster.
func verifyNodeCount(client *k8s.Clientset, expectedNodes int) (int, error) {
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return -1, err
	}
	return len(nodes.Items), nil
}

// RetryVerifyNodeReadiness returns error in case any node is not ready, after several trys.
func RetryVerifyNodeReadiness(client *k8s.Clientset) error {
	for i := 0; i <= retryAttempts; i++ {
		err := verifyNodeReadiness(client)
		if err != nil {
			logger.Debug("Waiting for Nodes to become ready..")
			time.Sleep(time.Duration(retrySleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timedout waiting nodes to become ready")
}

// verifyNodeReadiness returns error in case any node is not ready.
func verifyNodeReadiness(client *k8s.Clientset) error {
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		status := false
		for _, c := range node.Status.Conditions {
			if c.Type == v1.NodeReady {
				status = true
			}
		}
		if !status {
			return fmt.Errorf("node %s not ready", node.Name)
		}
	}
	return nil
}
