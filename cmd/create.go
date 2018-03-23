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
	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/namer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuroyoro/swalker"
)

// CreateCmd represents create command
func CreateCmd() *cobra.Command {
	var co = &cli.CreateOptions{}
	var createCmd = &cobra.Command{
		Use:   "create [NAME] [-p|--profile PROFILENAME] [-c|--cloudid CLOUDID]",
		Short: "Create a Kubicorn API model from a profile",
		Long: `Use this command to create a Kubicorn API model in a defined state store.

	This command will create a cluster API model as a YAML manifest in a state store.
	Once the API model has been created, a user can optionally change the model to their liking.
	After a model is defined and configured properly, the user can then apply the model.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				co.Name = viper.GetString(keyKubicornName)
				if co.Name == "" {
					co.Name = namer.RandomName()
				}
			case 1:
				co.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := RunCreate(co); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := createCmd.Flags()

	fs.StringVarP(&co.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&co.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)
	fs.StringVarP(&co.Profile, keyProfile, "p", viper.GetString(keyProfile), descProfile)
	fs.StringVarP(&co.CloudID, keyCloudID, "c", viper.GetString(keyCloudID), descCloudID)
	fs.StringVarP(&co.Set, keySet, "e", viper.GetString(keySet), descSet)
	fs.StringVarP(&co.GitRemote, keyGitConfig, "g", viper.GetString(keyGitConfig), descGitConfig)

	fs.StringVar(&co.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&co.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&co.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&co.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&co.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	flagApplyAnnotations(createCmd, "profile", "__kubicorn_parse_profiles")
	flagApplyAnnotations(createCmd, "cloudid", "__kubicorn_parse_cloudid")

	createCmd.SetUsageTemplate(cli.UsageTemplate)

	return createCmd
}

// RunCreate is the starting point when a user runs the create command.
func RunCreate(options *cli.CreateOptions) error {
	// Create our cluster resource
	name := options.Name
	var newCluster *cluster.Cluster
	if _, ok := cli.ProfileMapIndexed[options.Profile]; ok {
		newCluster = cli.ProfileMapIndexed[options.Profile].ProfileFunc(name)
	} else {
		return fmt.Errorf("Invalid profile [%s]", options.Profile)
	}

	if options.Set != "" {
		sets := strings.Split(options.Set, ",")
		for _, set := range sets {
			parts := strings.SplitN(set, "=", 2)
			if len(parts) == 1 {
				continue
			}
			providerConfig := newCluster.ProviderConfig()
			err := swalker.Write(strings.Title(parts[0]), providerConfig, parts[1])
			if err != nil {
				return fmt.Errorf("Invalid --set: %v", err)
			}
			newCluster.SetProviderConfig(providerConfig)
		}
	}

	if newCluster.ProviderConfig().Cloud == cluster.CloudGoogle && options.CloudID == "" {
		return fmt.Errorf("CloudID is required for google cloud. Please set it to your project ID")
	}

	providerConfig := newCluster.ProviderConfig()
	providerConfig.CloudId = options.CloudID
	newCluster.SetProviderConfig(providerConfig)

	// Expand state store path
	// Todo (@kris-nova) please pull this into a filepath package or something
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store and check if it exists
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if stateStore.Exists() {
		return fmt.Errorf("State store [%s] exists, will not overwrite. Delete existing profile [%s] and retry", name, options.StateStorePath+"/"+name)
	}

	// Init new state store with the cluster resource
	err = stateStore.Commit(newCluster)
	if err != nil {
		return fmt.Errorf("Unable to init state store: %v", err)
	}

	logger.Always("The state [%s/%s/cluster.yaml] has been created. You can edit the file, then run `kubicorn apply %s`", options.StateStorePath, name, name)
	return nil
}
