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

// Package s3 implements S3-compatible state store.
package s3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/state"
	"github.com/minio/minio-go"
)

type JSONS3StoreOptions struct {
	BasePath      string
	ClusterName   string
	Client        *minio.Client
	BucketOptions *S3BucketOptions
}

type S3BucketOptions struct {
	EndpointURL string
	BucketName  string
}

// JSONFS3 exists to save the cluster at runtime to the file defined
// in the state.ClusterJSONFile constant. We perform this operation so that
// various bash scripts can get the cluster state at runtime without having to
// inject key/value pairs into the script or anything like that.
type JSONFS3Store struct {
	options       *JSONS3StoreOptions
	BucketOptions *S3BucketOptions
	Client        *minio.Client
	ClusterName   string
	BasePath      string
	AbsolutePath  string
}

func NewJSONFS3Store(o *JSONS3StoreOptions) *JSONFS3Store {
	return &JSONFS3Store{
		options:       o,
		BucketOptions: o.BucketOptions,
		Client:        o.Client,
		ClusterName:   o.ClusterName,
		BasePath:      o.BasePath,
		AbsolutePath:  filepath.Join(o.BasePath, o.ClusterName),
	}
}

func (fs *JSONFS3Store) Exists() bool {
	_, err := fs.Client.StatObject(fs.BucketOptions.BucketName,
		filepath.Join(fs.AbsolutePath, state.ClusterJSONFile), minio.StatObjectOptions{})
	return err == nil
}

func (fs *JSONFS3Store) write(relativePath string, data []byte) error {
	fqn := filepath.Join(fs.AbsolutePath, relativePath)
	r := bytes.NewReader(data)

	_, err := fs.Client.PutObject(fs.BucketOptions.BucketName, fqn, r, -1, minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		return err
	}

	return nil
}

func (fs *JSONFS3Store) Read(relativePath string) ([]byte, error) {
	fqn := filepath.Join(fs.AbsolutePath, relativePath)

	o, err := fs.Client.GetObject(fs.BucketOptions.BucketName, fqn, minio.GetObjectOptions{})
	if err != nil {
		return []byte(""), err
	}

	b := new(bytes.Buffer)
	b.ReadFrom(o)
	return b.Bytes(), nil
}

func (fs *JSONFS3Store) ReadStore() ([]byte, error) {
	return fs.Read(state.ClusterJSONFile)
}

func (fs *JSONFS3Store) Commit(c *cluster.Cluster) error {
	if c == nil {
		return fmt.Errorf("Nil cluster spec")
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return fs.write(state.ClusterJSONFile, bytes)
}

func (fs *JSONFS3Store) Rename(existingRelativePath, newRelativePath string) error {
	src := minio.NewSourceInfo(fs.BucketOptions.BucketName, existingRelativePath, nil)
	dst, err := minio.NewDestinationInfo(fs.BucketOptions.BucketName, newRelativePath, nil, nil)
	if err != nil {
		return err
	}

	err = fs.Client.CopyObject(dst, src)
	if err != nil {
		return err
	}

	err = fs.Client.RemoveObject(fs.BucketOptions.BucketName, existingRelativePath)
	if err != nil {
		return err
	}

	return nil
}

func (fs *JSONFS3Store) Destroy() error {
	var err error
	logger.Warning("Removing path [%s]", fs.AbsolutePath)

	objectsCh := make(chan string)
	go func() {
		defer close(objectsCh)
		for object := range fs.Client.ListObjects(fs.BucketOptions.BucketName, fs.AbsolutePath, true, nil) {
			if object.Err != nil {
				err = fmt.Errorf("Error encountered while listing objects: %#v", object.Err)
			}
			objectsCh <- object.Key
		}
	}()

	for rErr := range fs.Client.RemoveObjects(fs.BucketOptions.BucketName, objectsCh) {
		err = fmt.Errorf("Error detected during deletion: %#v", rErr)
	}

	return err
}

func (fs *JSONFS3Store) GetCluster() (*cluster.Cluster, error) {
	configBytes, err := fs.Read(state.ClusterJSONFile)
	if err != nil {
		return nil, err
	}

	return fs.BytesToCluster(configBytes)
}

func (fs *JSONFS3Store) BytesToCluster(bytes []byte) (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	err := json.Unmarshal(bytes, cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (fs *JSONFS3Store) List() ([]string, error) {
	var stateList []string

	doneCh := make(chan struct{})
	defer close(doneCh)
	for object := range fs.Client.ListObjects(fs.BucketOptions.BucketName, filepath.Base(fs.BasePath)+"/", false, doneCh) {
		if object.Err != nil {
			return stateList, object.Err
		}
		stateList = append(stateList, filepath.Base(object.Key))
	}

	return stateList, nil
}
