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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
)

type JSONGitStoreOptions struct {
	BasePath    string
	ClusterName string
}

// JSONGitStore exists to save the cluster at runtime to the file defined
// in the state.ClusterJSONFile constant. We perform this operation so that
// various bash scripts can get the cluster state at runtime without having to
// inject key/value pairs into the script or anything like that.
type JSONGitStore struct {
	options      *JSONGitStoreOptions
	ClusterName  string
	BasePath     string
	AbsolutePath string
}

func NewJSONGitStore(o *JSONGitStoreOptions) *JSONGitStore {
	return &JSONGitStore{
		options:      o,
		ClusterName:  o.ClusterName,
		BasePath:     o.BasePath,
		AbsolutePath: filepath.Join(o.BasePath, o.ClusterName),
	}
}

func (git *JSONGitStore) Exists() bool {
	if _, err := os.Stat(git.AbsolutePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (git *JSONGitStore) write(relativePath string, data []byte) error {
	fqn := filepath.Join(git.AbsolutePath, relativePath)
	err := os.MkdirAll(path.Dir(fqn), 0700)
	if err != nil {
		return err
	}
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

func (git *JSONGitStore) Read(relativePath string) ([]byte, error) {
	fqn := filepath.Join(git.AbsolutePath, relativePath)
	bytes, err := ioutil.ReadFile(fqn)
	if err != nil {
		return []byte(""), err
	}
	return bytes, nil
}

func (git *JSONGitStore) ReadStore() ([]byte, error) {
	return git.Read(state.ClusterJSONFile)
}

func (git *JSONGitStore) Commit(c *cluster.Cluster) error {
	if c == nil {
		return fmt.Errorf("Nil cluster spec")
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return git.write(state.ClusterJSONFile, bytes)
}

func (git *JSONGitStore) Rename(existingRelativePath, newRelativePath string) error {
	return os.Rename(existingRelativePath, newRelativePath)
}

func (git *JSONGitStore) Destroy() error {
	logger.Warning("Removing path [%s]", git.AbsolutePath)
	return os.RemoveAll(git.AbsolutePath)
}

func (git *JSONGitStore) GetCluster() (*cluster.Cluster, error) {
	configBytes, err := git.Read(state.ClusterJSONFile)
	if err != nil {
		return nil, err
	}

	return git.BytesToCluster(configBytes)
}

func (git *JSONGitStore) BytesToCluster(bytes []byte) (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	err := json.Unmarshal(bytes, cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (git *JSONGitStore) List() ([]string, error) {

	var stateList []string

	files, err := ioutil.ReadDir(git.AbsolutePath)
	if err != nil {
		return stateList, err
	}

	for _, file := range files {
		stateList = append(stateList, file.Name())
	}

	return stateList, nil
}
