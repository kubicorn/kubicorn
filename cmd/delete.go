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

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/task"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DeleteCmd represents the delete command
func DeleteCmd() *cobra.Command {
	var do = &cli.DeleteOptions{}
	var deleteCmd = &cobra.Command{
		Use:   "delete <NAME>",
		Short: "Delete a Kubernetes cluster",
		Long: `Use this command to delete cloud resources.
	
	This command will attempt to build the resource graph based on an API model.
	Once the graph is built, the delete will attempt to delete the resources from the cloud.
	After the delete is complete, the state store will be left in tact and could potentially be applied later.
	
	To delete the resource AND the API model in the state store, use --purge.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				do.Name = viper.GetString(keyKubicornName)
			case 1:
				do.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runDelete(do); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := deleteCmd.Flags()

	fs.StringVarP(&do.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&do.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)

	fs.StringVar(&do.AwsProfile, keyAwsProfile, viper.GetString(keyAwsProfile), descAwsProfile)
	fs.StringVar(&do.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&do.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&do.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&do.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&do.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&do.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	fs.BoolVarP(&do.Purge, keyPurge, "p", viper.GetBool(keyPurge), descPurge)

	return deleteCmd
}

func runDelete(options *cli.DeleteOptions) error {
	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to delete")
	}
	// Expand state store path
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store and check if it exists
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if !stateStore.Exists() {
		logger.Info("Cluster [%s] does not exist", name)
		return nil
	}

	expectedCluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}

	runtimeParams := &pkg.RuntimeParameters{}

	if len(options.AwsProfile) > 0 {
		runtimeParams.AwsProfile = options.AwsProfile
	}

	reconciler, err := pkg.GetReconciler(expectedCluster, runtimeParams)
	if err != nil {
		return fmt.Errorf("Unable to get cluster reconciler: %v", err)
	}
	var deleteCluster *cluster.Cluster
	var deleteClusterTask = func() error {
		deleteCluster, err = reconciler.Destroy()
		return err
	}

	err = task.RunAnnotated(deleteClusterTask, fmt.Sprintf("\nDestroying resources for cluster [%s]:\n", options.Name), "")
	if err != nil {
		return fmt.Errorf("Unable to destroy resources for cluster [%s]: %v", options.Name, err)
	}

	if err = stateStore.Commit(deleteCluster); err != nil {
		return fmt.Errorf("Unable to save state store: %v", err)
	}

	if options.Purge {
		err := stateStore.Destroy()
		if err != nil {
			return fmt.Errorf("Unable to remove state store for cluster [%s]: %v", options.Name, err)
		}
	}
	return nil
}
