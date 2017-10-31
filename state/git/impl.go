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
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

type JSONGitStoreOptions struct {
	BasePath    string
	ClusterName string
}

// JSONGitStore exists to save the cluster state as a git change.
type JSONGitStore struct {
	commit       *git.CommitOptions
	push         *git.PushOptions
	options      *JSONGitStoreOptions
	ClusterName  string
	BasePath     string
	AbsolutePath string
}

func NewJSONGitStore(o *JSONGitStoreOptions) *JSONGitStore {
	return &JSONGitStore{
		commit: &git.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			},
		},
		push: &git.PushOptions{
			RemoteName: "",
			Auth: &transport.AuthMethod{
				User: "",
				Pass: "",
			},
		},
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

	//git init here
	log.Printf("\nCreating new git repo into $GOPATH [%s]", fqn)
	r, err := git.PlainOpen(fqn)
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

//Performs a git 'commit' and 'push' of the current cluster changes.
func (git *JSONGitStore) Commit(c *cluster.Cluster) error {
	if c == nil {
		return fmt.Errorf("Nil cluster spec")
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	//writes latest changes to git repo.
	git.write(state.ClusterJSONFile, bytes)

	//commits the changes
	r, err := git.PlainOpen(state.ClusterJSONFile)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(" . ")
	if err != nil {
		return err
	}

	// Commits the current staging are to the repository, with the new files
	// just created. We should provide the object.Signature of Author of the
	// commit.
	commit, err := w.Commit("Adding new cluster changes", JSONGitStore.commit)
	if err != nil {
		return err
	}

	//pushes the changes to remote repo.
	err = r.Push(JSONGitStore.push)
	if err != nil {
		return err
	}
	return nil
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
