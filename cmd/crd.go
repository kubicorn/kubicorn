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
	"github.com/kubicorn/kubicorn/pkg/state/crd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func CRDCmd() *cobra.Command {
	var crdo = &cli.CRDOptions{}
	var crdCommand = &cobra.Command{
		Use:   "crd <NAME>",
		Short: "Ensure a CRD for a given cluster",
		Long: `Use this command to upsert a CRD based on a given cluster definition.

All normal state store configuration is applicable for READING a state store, and the state store will ALWAYS be saved as a CRD.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				crdo.Name = viper.GetString(keyKubicornName)
			case 1:
				crdo.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runCRD(crdo); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := crdCommand.Flags()

	fs.StringVarP(&crdo.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&crdo.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)

	fs.StringVar(&crdo.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&crdo.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&crdo.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&crdo.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&crdo.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&crdo.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return crdCommand
}

func runCRD(options *cli.CRDOptions) error {

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

	crdStateStore := crd.NewCRDStore(&crd.CRDStoreOptions{
		BasePath:    options.StateStorePath,
		ClusterName: options.Name,
	})
	err = crdStateStore.Commit(cluster)
	if err != nil {
		return err
	}
	logger.Always("CRDs created")
	return nil
}
