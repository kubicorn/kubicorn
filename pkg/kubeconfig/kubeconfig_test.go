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

package kubeconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/local"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestSdkHappy(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmpdir)
	}()

	localDir := fmt.Sprintf("%s/.kube", tmpdir)
	localPath, err := getPath(localDir)
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		t.Fatal("error creating kubectl directory")
	}

	expectedLocalPath := fmt.Sprintf("%s/.kube/config", tmpdir)
	if localPath != expectedLocalPath {
		t.Fatalf("kubectl config path incorrect, got: %s, expected: %s", localPath, expectedLocalPath)
	}
}

func TestGetConfigHappy(t *testing.T) {
	dir, err := os.Getwd()
	dir, err = filepath.Abs(dir + "/../../test")
	port, ok := os.LookupEnv(local.TestPort)
	if !ok {
		port = "6666"
	}
	testCluster := cluster.NewCluster("test_cluster")
	providerConfig := &cluster.ControlPlaneProviderConfig{
		SSH: &cluster.SSH{
			User:          "root",
			PublicKeyPath: dir + "/credentials/id_rsa.pub",
			Port:          port,
		},
		KubernetesAPI: &cluster.KubernetesAPI{
			Endpoint: "localhost",
		},
	}
	testCluster.SetProviderConfig(providerConfig)
	os.Setenv(local.TestHome, dir+"/tmp")

	// ignore this error, its expected to do so, however, we should have a config file.
	err = GetConfig(testCluster)
	if err != nil {
		t.Skipf("WARNING, skipping test since its possible local machine may not allow ssh tunnels: \n%+v", err)
	}

	result, err := ioutil.ReadFile(dir + "/tmp/.kube/config")
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(result)) != "kubicorn test data" {
		t.Fatalf("File content is incorrect \"%v\"", strings.TrimSpace(string(result)))
	}

	defer os.RemoveAll(dir + "/tmp/.kube")
}

func TestMergeKubeConfigs(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed; err=%q", err)
	}
	dir, err = filepath.Abs(dir + "/../../test/kubeconfig")
	if err != nil {
		t.Fatalf("filepath.Abs(%q) failed; err=%q", dir+"/../../test/kubeconfig", err)
	}

	testConfigFiles := []string{fmt.Sprintf("%s/%s", dir, "config-1.yaml"), fmt.Sprintf("%s/%s", dir, "config-2.yaml"), fmt.Sprintf("%s/%s", dir, "config-3.yaml")}
	var testConfigs []*clientcmdapi.Config

	for _, tcf := range testConfigFiles {
		fileBytes, err := ioutil.ReadFile(tcf)
		if err != nil {
			t.Fatalf("Failed to read test data from file [%s]", tcf)
		}
		config, err := clientcmd.Load(fileBytes)
		if err != nil {
			t.Fatalf("getKubeConfigFromBytes(%q) failed; error=%q", fileBytes, err)
		}
		testConfigs = append(testConfigs, config)
		testConfigs = append(testConfigs, nil)
	}

	testcases := []struct {
		name                    string
		inputKubeConfigs        []*clientcmdapi.Config
		expectedClustersCount   int
		expectedAuthInfosCount  int
		expectedContextsCount   int
		expectedExtensionsCount int
	}{
		{
			name:                   "success_merge_3_kubeconfigs_without_nils",
			inputKubeConfigs:       testConfigs,
			expectedClustersCount:  3,
			expectedContextsCount:  3,
			expectedAuthInfosCount: 6,
		},
		{
			name:                   "success_merge_3_kubeconfigs_with_nil",
			inputKubeConfigs:       append(testConfigs, nil),
			expectedClustersCount:  3,
			expectedContextsCount:  3,
			expectedAuthInfosCount: 6,
		},
		{
			name:                   "success_merge_1_kubeconfigs_and_nil",
			inputKubeConfigs:       append(testConfigs[0:1], nil...),
			expectedClustersCount:  1,
			expectedContextsCount:  1,
			expectedAuthInfosCount: 2,
		},
	}

	for _, tc := range testcases {
		merged := mergeKubeconfigs(tc.inputKubeConfigs)

		if len(merged.Clusters) != tc.expectedClustersCount {
			t.Fatalf("TestCase:%q; mergeKubeconfigs failed, gotMergedClustersCount=[%d]; wantMergedClustersCount=[%d]", tc.name, len(merged.Clusters), tc.expectedClustersCount)
		}

		if len(merged.Contexts) != tc.expectedContextsCount {
			t.Fatalf("TestCase:%q; mergeKubeconfigs failed, gotMergedContextsCount=[%d]; wantMergedContextsCount=[%d]", tc.name, len(merged.Contexts), tc.expectedContextsCount)
		}

		if len(merged.AuthInfos) != tc.expectedAuthInfosCount {
			t.Fatalf("TestCase:%q; mergeKubeconfigs failed, gotMergedAuthInfosCount=[%d]; wantMergedAuthInfosCount=[%d]", tc.name, len(merged.AuthInfos), tc.expectedAuthInfosCount)
		}
	}
}

func TestGetRemoteKubeconfigPath(t *testing.T) {
	testCases := []struct {
		inputUserName    string
		expectedUserHome string
	}{
		{
			inputUserName:    "root",
			expectedUserHome: "/root/.kube/config",
		},
		{
			inputUserName:    "fabuser",
			expectedUserHome: "/home/fabuser/.kube/config",
		},
	}

	for _, tc := range testCases {
		actualUserKubeConfig := getRemoteKubeconfigPath(tc.inputUserName)
		if actualUserKubeConfig != tc.expectedUserHome {
			t.Fatalf("getUserKubeConfig(%q) failed, got %q; want %q", tc.inputUserName, actualUserKubeConfig, tc.expectedUserHome)
		}
	}
}
