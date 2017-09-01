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

package cmd

import (
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/kris-nova/kubicorn/state/jsonfs"
	"testing"
)

func TestRunCreate(t *testing.T) {
	clusterName := "test-create"
	statePath := "./_state"
	state := fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
		BasePath:    statePath,
		ClusterName: clusterName,
	})
	jsonState := jsonfs.NewJSONFileSystemStore(&jsonfs.JSONFileSystemStoreOptions{
		BasePath:    statePath,
		ClusterName: clusterName,
	})
	if err := state.Destroy(); err != nil {
		t.Fatalf("Error cleaning up any existing state: %v", err)
	}
	if err := jsonState.Destroy(); err != nil {
		t.Fatalf("Error cleaning up any existing json state: %v", err)
	}

	options := &CreateOptions{
		Profile: "aws",
		Options: Options{
			JSONStateStorePath: statePath,
			Name:               clusterName,
			StateStore:         "fs",
			StateStorePath:     statePath,
		},
	}
	err := RunCreate(options)
	if err != nil {
		t.Fatalf("Error running create cmd: %v", err)
	}

	storedCluster, err := jsonState.GetCluster()
	if err != nil {
		t.Fatalf("Error reading cluster from json state store: %v", err)
	}
	if storedCluster.Name != clusterName {
		t.Fatalf("Expected stored cluster to have name %v, got: %v", clusterName, storedCluster.Name)
	}

	if err := state.Destroy(); err != nil {
		t.Fatalf("Error cleaning up any state: %v", err)
	}
	if err := jsonState.Destroy(); err != nil {
		t.Fatalf("Error cleaning up json state: %v", err)
	}
}
