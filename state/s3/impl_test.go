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

package s3

import (
	"testing"

	"reflect"

	"github.com/kris-nova/kubicorn/profiles/amazon"
	"github.com/minio/minio-go"
)

func TestJsonFileSystem(t *testing.T) {
	client, err := minio.New("localhost:9000", "KubicornAccess", "KubicornSecret", false)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	err = client.MakeBucket("test", "")
	if err != nil {
		exists, err := client.BucketExists("test")
		if err != nil || !exists {
			t.Fatalf("Cannot create bucket: %#v", err)
		}
	}

	testFilePath := ".test/"
	clusterName := "s3-test"
	c := amazon.NewUbuntuCluster(clusterName)

	s3store := NewJSONFS3Store(&JSONS3StoreOptions{
		Client:      client,
		BasePath:    testFilePath,
		ClusterName: c.Name,
		BucketOptions: &S3BucketOptions{
			EndpointURL: "localhost:9000",
			BucketName:  "test",
		},
	})

	if err := s3store.Destroy(); err != nil {
		t.Fatalf("Error destroying any existing state: %v", err)
	}
	if s3store.Exists() {
		t.Fatalf("State shouldn't exist because we just destroyed it, but Exists() returned true")
	}
	if err := s3store.Commit(c); err != nil {
		t.Fatalf("Error committing cluster: %v", err)
	}
	dirs, err := s3store.List()
	if err != nil {
		t.Fatalf("Error listing files: %v", err)
	}
	if len(dirs) < 1 {
		t.Fatalf("Expected at least one cluster, got: %v", len(dirs))
	}
	if dirs[0] != c.Name {
		t.Fatalf("Expected file name to be %v, got %v", c.Name, dirs[0])
	}
	read, err := s3store.GetCluster()
	if err != nil {
		t.Fatalf("Error getting cluster: %v", err)
	}
	if !reflect.DeepEqual(read, c) {
		t.Fatalf("Cluster in doesn't equal cluster out")
	}
	if err = s3store.Destroy(); err != nil {
		t.Fatalf("Error cleaning up state: %v", err)
	}
}
