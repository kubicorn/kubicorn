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
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/stores"
	"github.com/kris-nova/kubicorn/state/stores/fs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"fmt"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Kubernetes cluster",
	Long:  `Delete a Kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunDelete(do)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type DeleteOptions struct {
	Options
	Disable bool
}

var do = &DeleteOptions{}

func init() {
	deleteCmd.Flags().StringVarP(&do.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	deleteCmd.Flags().StringVarP(&do.StateStorePath, "state-store-path", "p", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	deleteCmd.Flags().StringVarP(&do.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "Cluster name to delete")
	deleteCmd.Flags().BoolVarP(&do.Disable, "disable", "d", false, "Disable the state instead of destroying it. Will retain state store data.")
	RootCmd.AddCommand(deleteCmd)
}

func RunDelete(options *DeleteOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to delete.")
	}
	// Expand state store path
	options.StateStorePath = expandPath(options.StateStorePath)

	// Register state store
	var stateStore stores.ClusterStorer
	switch options.StateStore {
	case "fs":
		logger.Info("Selected [fs] state store")
		stateStore = fs.NewFileSystemStore(&fs.FileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: name,
		})
	}

	if !stateStore.Exists() {
		logger.Info("Cluster [%s] does not exist", name)
		return nil
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}

	if options.Disable {
		err := state.DisableStateStore(stateStore, cluster)
		if err != nil {
			return err
		}
	} else {
		err := state.DestroyStateStore(stateStore, cluster)
		if err != nil {
			return err
		}
	}
	return nil
}
