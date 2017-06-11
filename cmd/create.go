// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"github.com/kris-nova/kubicorn/api/cluster"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/namer"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/stores"
	"github.com/kris-nova/kubicorn/state/stores/fs"
	"github.com/spf13/cobra"
	"os"
	"os/user"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Kubernetes cluster",
	Long:  `Create a Kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunCreate(o)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type CreateOptions struct {
	StateStore       string
	StateStorePath   string
	StateStorePrefix string
	Name             string
}

var o = &CreateOptions{}

func init() {
	createCmd.Flags().StringVarP(&o.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	createCmd.Flags().StringVarP(&o.StateStorePath, "state-store-path", "p", strEnvDef("KUBICORN_STATE_STORE_PATH", "."), "The state store path to use")
	createCmd.Flags().StringVarP(&o.StateStorePrefix, "state-store-prefix", "x", strEnvDef("KUBICORN_STATE_STORE_PREFIX", "_state"), "Directory name prefix to use for state store")
	createCmd.Flags().StringVarP(&o.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "An optional name to use. If empty, will generate a random name.")
	RootCmd.AddCommand(createCmd)
}

func RunCreate(options *CreateOptions) error {

	// Expand state store path
	options.StateStorePath = expandStateStorePath(options)

	// Register state store
	var stateStore stores.Storer
	switch options.StateStore {
	case "fs":
		logger.Info("Selected [fs] state store")
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			Prefix: options.StateStorePrefix,
			Path:   options.StateStorePath,
		})
	}

	// Check if state store exists
	if stateStore.Exists() {
		return fmt.Errorf("State store [%s/%s] exists, will not overwrite", options.StateStorePath, options.StateStorePrefix)
	}

	// Ensure we have a name
	name := options.Name
	if name == "" {
		name = namer.RandomName()
	}

	// Create our cluster resource
	cluster := cluster.NewCluster(name)

	// Init new state store with the cluster resource
	err := state.InitStateStore(stateStore, cluster)
	if err != nil {
		return fmt.Errorf("Unable to init state store: %v", err)
	}

	return nil
}

func expandStateStorePath(options *CreateOptions) string {
	path := options.StateStorePath
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
