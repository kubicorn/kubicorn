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
	"github.com/spf13/viper"
)

// GetConfigCmd represents the apply command
func GetConfigCmd() *cobra.Command {
	var cro = &cli.GetConfigOptions{}
	var getConfigCmd = &cobra.Command{
		Use:   "getconfig <NAME>",
		Short: "Manage Kubernetes configuration",
		Long: `Use this command to pull a kubeconfig file from a cluster so you can use kubectl.
	
	This command will attempt to find a cluster, and append a local kubeconfig file with a kubeconfig `,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				cro.Name = viper.GetString(keyKubicornName)
			case 1:
				cro.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runGetConfig(cro); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := getConfigCmd.Flags()

	fs.StringVarP(&cro.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&cro.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)

	fs.StringVar(&cro.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&cro.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&cro.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&cro.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&cro.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&cro.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return getConfigCmd
}

func runGetConfig(options *cli.GetConfigOptions) error {

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

	if err = kubeconfig.GetConfig(cluster); err != nil {
		return err
	}
	logger.Always("Applied kubeconfig")

	return nil
}
