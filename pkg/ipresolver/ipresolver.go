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

package ipresolver

import (
	"github.com/kubicorn/kubicorn/pkg/local"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetNodeIPAddress obtains node IP address using Kubernetes API.
func GetNodeIPAddress(nodename string) (string, error) {
	// This is a temporary (and probably bad) solution for obtaining IP address of the node.
	// TODO(xmudrii): Improve this function.
	config, err := clientcmd.BuildConfigFromFlags("", local.Expand("~/.kube/config"))
	if err != nil {
		return "", err
	}

	client, err := k8s.NewForConfig(config)
	if err != nil {
		return "", err
	}

	node, err := client.CoreV1().Nodes().Get(nodename, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return node.Status.Addresses[0].Address, nil
}
