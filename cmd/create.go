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
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/namer"
	"github.com/kris-nova/kubicorn/profiles"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/spf13/cobra"
	"os"
	"os/user"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Kubicorn API model from a profile",
	Long: `Use this command to create a Kubicorn API model in a defined state store.

This command will create a cluster API model as a YAML manifest in a state store.
Once the API model has been created, a user can optionally change the model to their liking.
After a model is defined and configured properly, the user can then apply the model.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunCreate(co)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type CreateOptions struct {
	Options
	Profile string
}

var co = &CreateOptions{}

func init() {
	createCmd.Flags().StringVarP(&co.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	createCmd.Flags().StringVarP(&co.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	createCmd.Flags().StringVarP(&co.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "An optional name to use. If empty, will generate a random name.")
	createCmd.Flags().StringVarP(&co.Profile, "profile", "p", strEnvDef("KUBICORN_PROFILE", "azure"), "The cluster profile to use")
	RootCmd.AddCommand(createCmd)
}

type profileFunc func(name string) *cluster.Cluster

var alias = map[string]profileFunc{
	"baremetal": profiles.NewSimpleBareMetal,
	"metal":     profiles.NewSimpleBareMetal,
	"amazon":    profiles.NewSimpleAmazonCluster,
	"aws":       profiles.NewSimpleAmazonCluster,
	"azure":     profiles.NewSimpleAzureCluster,
	"az":        profiles.NewSimpleAzureCluster,
	"google":    profiles.NewSimpleGoogleCluster,
	"googs":     profiles.NewSimpleGoogleCluster,
	"gce":       profiles.NewSimpleGoogleCluster,
	"gke":       profiles.NewSimpleGoogleCluster,
}

func RunCreate(options *CreateOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		name = namer.RandomName()
	}

	// Create our cluster resource
	var cluster *cluster.Cluster
	if _, ok := alias[options.Profile]; ok {
		cluster = alias[options.Profile](name)
	} else {
		return fmt.Errorf("Invalid profile [%s]", options.Profile)
	}

	// Expand state store path
	// Todo (@kris-nova) please pull this into a filepath package or something
	options.StateStorePath = expandPath(options.StateStorePath)

	// Register state store
	var stateStore state.ClusterStorer
	switch options.StateStore {
	case "fs":
		logger.Info("Selected [fs] state store")
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: name,
		})
	}

	// Check if state store exists
	if stateStore.Exists() {
		return fmt.Errorf("State store [%s] exists, will not overwrite", name)
	}

	// Init new state store with the cluster resource
	err := stateStore.Commit(cluster)
	if err != nil {
		return fmt.Errorf("Unable to init state store: %v", err)
	}

	fmt.Printf(options.StateStorePath + "/" + name + "/" + "cluster.yaml has been created.\nNow run `kubicorn apply`.\n")
	return nil
}

func expandPath(path string) string {
	if path == "." {
		wd, err := os.Getwd()
		if err != nil {
			logger.Critical("Unable to get current working directory: %v", err)
			return ""
		}
		path = wd
	}
	if path == "~" {
		homeVar := os.Getenv("HOME")
		if homeVar == "" {
			homeUser, err := user.Current()
			if err != nil {
				logger.Critical("Unable to use user.Current() for user. Maybe a cross compile issue: %v", err)
				return ""
			}
			path = homeUser.HomeDir
		}
	}
	return path
}
