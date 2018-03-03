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
	"errors"
	"fmt"
	"os"

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

var cro = &cli.GetConfigOptions{}

// GetConfigCmd represents the apply command
func GetConfigCmd() *cobra.Command {
	var getConfigCmd = &cobra.Command{
		Use:   "getconfig <NAME>",
		Short: "Manage Kubernetes configuration",
		Long: `Use this command to pull a kubeconfig file from a cluster so you can use kubectl.
	
	This command will attempt to find a cluster, and append a local kubeconfig file with a kubeconfig `,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cro.Name = cli.StrEnvDef("KUBICORN_NAME", "")
			} else if len(args) > 1 {
				logger.Critical("Too many arguments.")
				os.Exit(1)
			} else {
				cro.Name = args[0]
			}

			err := RunGetConfig(cro)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	getConfigCmd.Flags().StringVarP(&cro.StateStore, "state-store", "s", cli.StrEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	getConfigCmd.Flags().StringVarP(&cro.StateStorePath, "state-store-path", "S", cli.StrEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")

	// git flags
	getConfigCmd.Flags().StringVar(&cro.GitRemote, "git-config", cli.StrEnvDef("KUBICORN_GIT_CONFIG", "git"), "The git remote url to use")

	// s3 flags
	getConfigCmd.Flags().StringVar(&cro.S3AccessKey, "s3-access", cli.StrEnvDef("KUBICORN_S3_ACCESS_KEY", ""), "The s3 access key.")
	getConfigCmd.Flags().StringVar(&cro.S3SecretKey, "s3-secret", cli.StrEnvDef("KUBICORN_S3_SECRET_KEY", ""), "The s3 secret key.")
	getConfigCmd.Flags().StringVar(&cro.BucketEndpointURL, "s3-endpoint", cli.StrEnvDef("KUBICORN_S3_ENDPOINT", ""), "The s3 endpoint url.")
	getConfigCmd.Flags().BoolVar(&cro.BucketSSL, "s3-ssl", cli.BoolEnvDef("KUBICORN_S3_SSL", true), "The s3 bucket name to be used for saving the git state for the cluster.")
	getConfigCmd.Flags().StringVar(&cro.BucketName, "s3-bucket", cli.StrEnvDef("KUBICORN_S3_BUCKET", ""), "The s3 bucket name to be used for saving the s3 state for the cluster.")

	return getConfigCmd
}

func RunGetConfig(options *cli.GetConfigOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to get config")
	}

	// Expand state store path
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}
	logger.Info("Loaded cluster: %s", cluster.Name)

	logger.Info("Init Cluster")
	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		return err
	}

	err = kubeconfig.GetConfig(cluster)
	if err != nil {
		return err
	}
	logger.Always("Applied kubeconfig")

	return nil
}
