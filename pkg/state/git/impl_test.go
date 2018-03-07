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

package git

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/state"
	"github.com/kubicorn/kubicorn/profiles/amazon"
)

func TestJsonGit(t *testing.T) {
	testFilePath := ".test/"
	clusterName := "git-test"
	c := amazon.NewUbuntuCluster(clusterName)
	o := &JSONGitStoreOptions{
		BasePath:    testFilePath,
		ClusterName: c.Name,
		CommitConfig: &JSONGitCommitConfig{
			Name:   "Dummy Cluster",
			Email:  "dummy@clustermail.co",
			Remote: "https://github.com/kubicorn/kubicorn",
		},
	}
	git := NewJSONGitStore(o)
	if err := git.Destroy(); err != nil {
		t.Fatalf("Error destroying any existing state: %v", err)
	}
	if git.Exists() {
		t.Fatalf("State shouldn't exist because we just destroyed it, but Exists() returned true")
	}
	if err := git.Commit(c); err != nil {
		t.Fatalf("Error committing cluster: %v", err)
	}
	files, err := git.List()
	if err != nil {
		t.Fatalf("Error listing files: %v", err)
	}
	if len(files) < 1 {
		t.Fatalf("Expected at least one cluster, got: %v", len(files))
	}
	if files[0] != state.ClusterJSONFile {
		t.Fatalf("Expected file name to be %v, got %v", state.ClusterJSONFile, files[0])
	}
	read, err := git.GetCluster()
	if err != nil {
		t.Fatalf("Error getting cluster: %v", err)
	}
	if !reflect.DeepEqual(read, c) {
		t.Fatalf("Cluster in doesn't equal cluster out")
	}
	unmarshalled := &cluster.Cluster{}
	bytes, err := ioutil.ReadFile(filepath.Join(testFilePath, clusterName, state.ClusterJSONFile))
	if err != nil {
		t.Fatalf("Error reading json file: %v", err)
	}
	if err := json.Unmarshal(bytes, unmarshalled); err != nil {
		t.Fatalf("Error unmarshalling json: %v", err)
	}
	if !reflect.DeepEqual(unmarshalled, c) {
		t.Fatalf("Cluster read directly from json file doesn't equal cluster inputted: %v", unmarshalled)
	}
	if err = git.Destroy(); err != nil {
		t.Fatalf("Error cleaning up state: %v", err)
	}
}
