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

package cli

import (
	"errors"
	"fmt"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/state"
	"github.com/kubicorn/kubicorn/pkg/state/crd"
	"github.com/kubicorn/kubicorn/pkg/state/fs"
	"github.com/kubicorn/kubicorn/pkg/state/git"
	"github.com/kubicorn/kubicorn/pkg/state/jsonfs"
	"github.com/kubicorn/kubicorn/pkg/state/s3"
	minio "github.com/minio/minio-go"
	gg "github.com/tcnksm/go-gitconfig"
)

// NewStateStore returns clusterStorer object based on type.
func (options Options) NewStateStore() (state.ClusterStorer, error) {
	var stateStore state.ClusterStorer

	switch options.StateStore {
	case "fs":
		logger.Info("Selected [fs] state store")
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
		})
	case "crd":
		logger.Info("Selected [crd] state store")
		stateStore = crd.NewCRDStore(&crd.CRDStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
		})
	case "git":
		logger.Info("Selected [git] state store")
		if options.GitRemote == "" {
			return nil, errors.New("empty GitRemote url. Must specify the link to the remote git repo")
		}
		user, _ := gg.Global("user.name")
		email, _ := gg.Email()

		stateStore = git.NewJSONGitStore(&git.JSONGitStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
			CommitConfig: &git.JSONGitCommitConfig{
				Name:   user,
				Email:  email,
				Remote: options.GitRemote,
			},
		})
	case "jsonfs":
		logger.Info("Selected [jsonfs] state store")
		stateStore = jsonfs.NewJSONFileSystemStore(&jsonfs.JSONFileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
		})
	case "s3":
		logger.Info("Selected [s3] state store")
		client, err := minio.New(options.BucketEndpointURL, options.S3AccessKey, options.S3SecretKey, options.BucketSSL)
		if err != nil {
			return nil, err
		}
		stateStore = s3.NewJSONFS3Store(&s3.JSONS3StoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
			Client:      client,
			BucketOptions: &s3.S3BucketOptions{
				EndpointURL: options.BucketEndpointURL,
				BucketName:  options.BucketName,
			},
		})
	default:
		return nil, fmt.Errorf("state store [%s] has an invalid type [%s]", options.Name, options.StateStore)
	}

	return stateStore, nil
}
