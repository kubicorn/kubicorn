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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
)

type JSONFileSystemStoreOptions struct {
	AbsolutePath string
	ClusterName  string
}

func NewJSONFileSystemStoreOptions(clusterName string) *JSONFileSystemStoreOptions {
	return &JSONFileSystemStoreOptions{
		AbsolutePath: state.ClusterJsonPath,
		ClusterName:  clusterName,
	}
}

// JSONFileSystemStore exists to save the cluster at runtime to the file defined
// in the state.ClusterJsonFile constant. We perform this operation so that
// various bash scripts can get the cluster state at runtime without having to
// inject key/value pairs into the script or anything like that.
type JSONFileSystemStore struct {
	options      *JSONFileSystemStoreOptions
	ClusterName  string
	AbsolutePath string
}

func NewJSONFileSystemStore(o *JSONFileSystemStoreOptions) *JSONFileSystemStore {
	return &JSONFileSystemStore{
		options:      o,
		ClusterName:  o.ClusterName,
		AbsolutePath: o.AbsolutePath,
	}
}

func (fs *JSONFileSystemStore) Exists() bool {
	if _, err := os.Stat(fs.AbsolutePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fs *JSONFileSystemStore) write(relativePath string, data []byte) error {
	fqn := fmt.Sprintf("%s/%s", fs.AbsolutePath, relativePath)
	os.MkdirAll(path.Dir(fqn), 0700)
	fo, err := os.Create(fqn)
	if err != nil {
		return err
	}
	defer fo.Close()
	_, err = io.Copy(fo, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	return nil
}

func (fs *JSONFileSystemStore) read(relativePath string) ([]byte, error) {
	fqn := fmt.Sprintf("%s/%s", fs.AbsolutePath, relativePath)
	bytes, err := ioutil.ReadFile(fqn)
	if err != nil {
		return []byte(""), err
	}
	return bytes, nil
}

func (fs *JSONFileSystemStore) Commit(c *cluster.Cluster) error {
	if c == nil {
		return fmt.Errorf("Nil cluster spec")
	}
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	fs.write(state.ClusterJsonFile, bytes)
	return nil
}

func (fs *JSONFileSystemStore) Rename(existingRelativePath, newRelativePath string) error {
	return os.Rename(existingRelativePath, newRelativePath)
}

func (fs *JSONFileSystemStore) Destroy() error {
	logger.Warning("Removing path [%s]", fs.AbsolutePath)
	return os.RemoveAll(fs.AbsolutePath)
}

func (fs *JSONFileSystemStore) GetCluster() (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	configBytes, err := fs.read(state.ClusterJsonFile)
	if err != nil {
		return cluster, err
	}
	err = yaml.Unmarshal(configBytes, cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (fs *JSONFileSystemStore) List() ([]string, error) {

	var stateList []string

	files, err := ioutil.ReadDir(fs.AbsolutePath)
	if err != nil {
		return stateList, err
	}

	for _, file := range files {
		stateList = append(stateList, file.Name())
	}

	return stateList, nil
}
