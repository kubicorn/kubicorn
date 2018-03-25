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
