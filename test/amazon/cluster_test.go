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

package amazon

import (
	"fmt"
	"testing"
	"time"

	"github.com/kris-nova/charlie/network"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/resourcedeploy"
	profile "github.com/kubicorn/kubicorn/profiles/amazon"
	"github.com/kubicorn/kubicorn/test"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAmazon(t *testing.T) {
	logger.TestMode = true
	logger.Level = 4

	distros := []struct {
		cluster *cluster.Cluster
		name    string
	}{
		{
			cluster: profile.NewUbuntuCluster("ubuntu-test"),
			name:    "ubuntu",
		},
		{
			cluster: profile.NewCentosCluster("centos-test"),
			name:    "centos",
		},
	}

	cases := []struct {
		testCase func(*cluster.Cluster) error
		name     string
		retries  int
		sleep    time.Duration
	}{
		{
			testCase: testAPIListen,
			name:     "TestApiListen",
			retries:  40,
			sleep:    5 * time.Second,
		},
		{
			testCase: testGetNodes,
			name:     "TestGetNodes",
			retries:  40,
			sleep:    5 * time.Second,
		},
	}

	test.InitRsaTravis()
	for _, distro := range distros {
		t.Run(fmt.Sprintf("%s amazon cluter test", distro.name), func(t *testing.T) {
			testCluster, err := test.Create(distro.cluster)
			if err != nil {
				t.Fatalf("Unable to create Amazon test cluster: %v\n", err)
			}

			if testCluster.ProviderConfig() == nil {
				t.Fatal("Unable to create Amazon test cluster")
			}

			defer func() {
				_, err := test.Delete(testCluster)
				if err != nil {
					fmt.Println("Failure cleaning up cluster! Abandoned resources!")
				}
			}()

			succeeded := true
			for _, tcase := range cases {
				// Don't short-circuit
				result := t.Run(tcase.name, func(t *testing.T) {
					var err error
					for i := 0; i < tcase.retries; i++ {
						err = tcase.testCase(testCluster)
						if err == nil {
							return
						}
						time.Sleep(tcase.sleep)
					}
					t.Error(err)
				})
				succeeded = succeeded && result
			}

			if !succeeded {
				fmt.Printf("-----------------------------------------------------------------------\n")
				fmt.Printf("[FAILURE]\n")
				fmt.Printf("-----------------------------------------------------------------------\n")
			}

		})
	}
}

func testAPIListen(testCluster *cluster.Cluster) error {
	api := testCluster.ProviderConfig().KubernetesAPI
	_, err := network.AssertTcpSocketAcceptsConnection(fmt.Sprintf("%s:%s", api.Endpoint, api.Port), "opening a new socket connection against the Kubernetes API")
	logger.Info("Attempting to open a socket to the Kubernetes API: %v...\n", err)
	return err
}

func testGetNodes(testCluster *cluster.Cluster) error {
	if err := kubeconfig.GetConfig(testCluster); err != nil {
		return errors.Wrap(err, "failed to retrieve kubeconfig")
	}
	client, err := resourcedeploy.ClientSet(testCluster)
	if err != nil {
		return errors.Wrap(err, "couldn't get kubeconfig: %v")
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to retrieve node list")
	}

	if len(nodes.Items) != 2 {
		return fmt.Errorf("Expected 2 nodes, got %d", len(nodes.Items))
	}
	return nil
}
