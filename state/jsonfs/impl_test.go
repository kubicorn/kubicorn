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

package jsonfs

import (
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/state"
	"reflect"
	"testing"
)

func TestJsonFileSystem(t *testing.T) {
	c := profiles.NewSimpleAmazonCluster("jsonfs-test")
	o := &JSONFileSystemStoreOptions{
		AbsolutePath: ".test/",
		ClusterName:  c.Name,
	}
	fs := NewJSONFileSystemStore(o)
	if err := fs.Destroy(); err != nil {
		t.Fatalf("Error destroying any existing state: %v", err)
	}
	if fs.Exists() {
		t.Fatalf("State shouldn't exist because we just destroyed it, but Exists() returned true")
	}
	if err := fs.Commit(c); err != nil {
		t.Fatalf("Error committing cluster: %v", err)
	}
	files, err := fs.List()
	if err != nil {
		t.Fatalf("Error listing files: %v", err)
	}
	if len(files) < 1 {
		t.Fatalf("Expected at least one cluster, got: %v", len(files))
	}
	if files[0] != state.ClusterJsonFile {
		t.Fatalf("Expected file name to be %v, got %v", state.ClusterJsonFile, files[0])
	}
	read, err := fs.GetCluster()
	if err != nil {
		t.Fatalf("Error getting cluster: %v", err)
	}
	if !reflect.DeepEqual(read, c) {
		t.Fatalf("Cluster in doesn't equal cluster out")
	}
	if err = fs.Destroy(); err != nil {
		t.Fatalf("Error cleaning up state: %v", err)
	}
}
