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

	"github.com/kubicorn/kubicorn/profiles/amazon"
	"github.com/minio/minio-go"
)

const (
	endpoint   = "localhost:9000"
	bucketName = "test"
)

func TestJsonFileSystem(t *testing.T) {
	client, err := minio.New(endpoint, "KubicornAccess", "KubicornSecret", false)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// If tests fails, you may need to run locally:
	// docker run -p 9000:9000 --name minio1   -e "MINIO_ACCESS_KEY=KubicornAccess"   -e "MINIO_SECRET_KEY=KubicornSecret" minio/minio server /data
	// to get this test to pass. For that reason will be skipping this step if it fails.
	err = client.MakeBucket(bucketName, "")
	if err != nil {
		t.Skipf("could not create bucket, skipping test, please check if minio server is running: %#v", err)

		// Commenting below out for error message above and skipping tests if locally fails for a dev.
		//exists, err := client.BucketExists(bucketName)
		//if err != nil || !exists {
		//	t.Fatalf("Cannot create bucket: %#v", err)
		//}
	}

	testFilePath := ".test/"
	clusterName := "s3-test"
	c := amazon.NewUbuntuCluster(clusterName)

	s3store := NewJSONFS3Store(&JSONS3StoreOptions{
		Client:      client,
		BasePath:    testFilePath,
		ClusterName: c.Name,
		BucketOptions: &S3BucketOptions{
			EndpointURL: endpoint,
			BucketName:  bucketName,
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
