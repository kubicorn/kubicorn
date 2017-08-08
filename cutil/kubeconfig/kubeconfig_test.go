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
	"testing"
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
	localPath, err := getKubeConfigPath(localDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		t.Fatal("error creating kubectl directory")
	}

	expectedLocalPath := fmt.Sprintf("%s/.kube/config", tmpdir)
	if localPath != expectedLocalPath {
		t.Fatalf("kubectl config path incorrect, got: %s, expected: %s", localPath, expectedLocalPath)
	}
}
