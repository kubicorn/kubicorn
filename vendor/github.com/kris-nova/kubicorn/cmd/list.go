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
	"github.com/kubicorn/kubicorn/state"
	"github.com/kubicorn/kubicorn/state/fs"
	"github.com/kubicorn/kubicorn/state/git"
	"github.com/kubicorn/kubicorn/state/jsonfs"
	"github.com/kubicorn/kubicorn/state/s3"
	"github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

var (
	lo        = &cli.ListOptions{}
	noHeaders bool
)

// ListCmd represents the list command
func ListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List available states",
		Long:  `List the states available in the _state directory`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunList(lo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&lo.StateStore, "state-store", "s", cli.StrEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	cmd.Flags().StringVarP(&lo.StateStorePath, "state-store-path", "S", cli.StrEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	cmd.Flags().BoolVarP(&noHeaders, "no-headers", "n", false, "Show the list containing names only")

	// s3 flags
	cmd.Flags().StringVar(&lo.S3AccessKey, "s3-access", cli.StrEnvDef("KUBICORN_S3_ACCESS_KEY", ""), "The s3 access key.")
	cmd.Flags().StringVar(&lo.S3SecretKey, "s3-secret", cli.StrEnvDef("KUBICORN_S3_SECRET_KEY", ""), "The s3 secret key.")
	cmd.Flags().StringVar(&lo.BucketEndpointURL, "s3-endpoint", cli.StrEnvDef("KUBICORN_S3_ENDPOINT", ""), "The s3 endpoint url.")
	cmd.Flags().BoolVar(&lo.BucketSSL, "s3-ssl", cli.BoolEnvDef("KUBICORN_S3_SSL", true), "The s3 bucket name to be used for saving the git state for the cluster.")
	cmd.Flags().StringVar(&lo.BucketName, "s3-bucket", cli.StrEnvDef("KUBICORN_S3_BUCKET", ""), "The s3 bucket name to be used for saving the s3 state for the cluster.")

	return cmd
}

func RunList(options *cli.ListOptions) error {
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
		client, err := minio.New(lo.BucketEndpointURL, lo.S3AccessKey, lo.S3SecretKey, lo.BucketSSL)
		if err != nil {
			return err
		}

		logger.Info("Selected [s3] state store")
		stateStore = s3.NewJSONFS3Store(&s3.JSONS3StoreOptions{
			Client:   client,
			BasePath: options.StateStorePath,
			BucketOptions: &s3.S3BucketOptions{
				EndpointURL: lo.BucketEndpointURL,
				BucketName:  lo.BucketName,
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
