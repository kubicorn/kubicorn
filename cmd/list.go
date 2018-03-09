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
	"fmt"
	"os"

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/state"
	"github.com/kubicorn/kubicorn/pkg/state/fs"
	"github.com/kubicorn/kubicorn/pkg/state/git"
	"github.com/kubicorn/kubicorn/pkg/state/jsonfs"
	"github.com/kubicorn/kubicorn/pkg/state/s3"
	"github.com/minio/minio-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var noHeaders bool

// ListCmd represents the list command
func ListCmd() *cobra.Command {
	var lo = &cli.ListOptions{}
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List available states",
		Long:  `List the states available in the _state directory`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runList(lo); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&lo.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&lo.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)

	fs.StringVar(&lo.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&lo.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&lo.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&lo.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVarP(&noHeaders, keyNoHeaders, "n", viper.GetBool(keyNoHeaders), desNoHeaders)

	fs.BoolVar(&lo.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return cmd
}

func runList(options *cli.ListOptions) error {
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	var stateStore state.ClusterStorer
	switch options.StateStore {
	case "fs":
		if !noHeaders {
			logger.Info("Selected [fs] state store")
		}
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			BasePath: options.StateStorePath,
		})

	case "git":
		if !noHeaders {
			logger.Info("Selected [git] state store")
		}
		stateStore = git.NewJSONGitStore(&git.JSONGitStoreOptions{
			BasePath: options.StateStorePath,
		})
	case "jsonfs":
		if !noHeaders {
			logger.Info("Selected [jsonfs] state store")
		}
		stateStore = jsonfs.NewJSONFileSystemStore(&jsonfs.JSONFileSystemStoreOptions{
			BasePath: options.StateStorePath,
		})
	case "s3":
		client, err := minio.New(options.BucketEndpointURL, options.S3AccessKey, options.S3SecretKey, options.BucketSSL)
		if err != nil {
			return err
		}

		logger.Info("Selected [s3] state store")
		stateStore = s3.NewJSONFS3Store(&s3.JSONS3StoreOptions{
			Client:   client,
			BasePath: options.StateStorePath,
			BucketOptions: &s3.S3BucketOptions{
				EndpointURL: options.BucketEndpointURL,
				BucketName:  options.BucketName,
			},
		})
	}

	clusters, err := stateStore.List()
	if err != nil {
		return fmt.Errorf("Unable to list clusters: %v", err)
	}
	for _, cluster := range clusters {
		if !noHeaders {
			logger.Always(cluster)
		} else {
			fmt.Println(cluster)
		}
	}

	return nil
}
