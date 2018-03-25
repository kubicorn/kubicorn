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
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/resourcedeploy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DeployControllerCmd represents the apply command
func DeployControllerCmd() *cobra.Command {
	var dco = &cli.DeployControllerOptions{}
	var deployControllerCmd = &cobra.Command{
		Use:   "deploycontroller <NAME>",
		Short: "Deploy a controller for a given cluster",
		Long: `Use this command to deploy a controller for a given cluster.

As long as a controller is defined, this will create the deployment and the namespace.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				dco.Name = viper.GetString(keyKubicornName)
			case 1:
				dco.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runDeployController(dco); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := deployControllerCmd.Flags()

	fs.StringVarP(&dco.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&dco.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)

	fs.StringVar(&dco.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&dco.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&dco.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&dco.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&dco.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&dco.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return deployControllerCmd
}

func runDeployController(options *cli.DeployControllerOptions) error {

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

	err = resourcedeploy.EnsureNamespace(cluster)
	if err != nil {
		return fmt.Errorf("Unable to ensure namespace: %v", err)
	}

	err = resourcedeploy.DeployClusterControllerDeployment(cluster)
	if err != nil {
		return fmt.Errorf("Unable to deploy controller: %v", err)
	}

	logger.Always("Deployed")
	return nil
}
