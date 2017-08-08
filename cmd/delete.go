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
	"time"

	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [-n|--name NAME]",
	Short: "Delete a Kubernetes cluster",
	Long: `Use this command to delete cloud resources.

This command will attempt to build the resource graph based on an API model.
Once the graph is built, the delete will attempt to delete the resources from the cloud.
After the delete is complete, the state store will be left in tact and could potentially be applied later.

To delete the resource AND the API model in the state store, use --purge.`,
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
	Purge bool
}

var do = &DeleteOptions{}

func init() {
	deleteCmd.Flags().StringVarP(&do.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	deleteCmd.Flags().StringVarP(&do.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	deleteCmd.Flags().StringVarP(&do.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "Cluster name to delete")
	deleteCmd.Flags().BoolVarP(&do.Purge, "purge", "p", false, "Remove the API model from the state store after the resources are deleted.")

	flagApplyAnnotations(deleteCmd, "name", "__kubicorn_parse_list")

	RootCmd.AddCommand(deleteCmd)
}

func RunDelete(options *DeleteOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to delete")
	}
	// Expand state store path
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

	if !stateStore.Exists() {
		logger.Info("Cluster [%s] does not exist", name)
		return nil
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}

	reconciler, err := cutil.GetReconciler(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get cluster reconciler: %v", err)
	}

	if err = reconciler.Init(); err != nil {
		return fmt.Errorf("Unable to init reconciler: %v", err)
	}

	donechan := make(chan bool)
	errchan := make(chan error)

	go func() {
		errchan <- reconciler.Destroy()
	}()

	go func(description string, symbol string, c chan bool) {
		if description != "" {
			logger.Log(description)
		}

		for {
			select {
			case quit := <-c:
				if quit {
					return
				}
			default:
				time.Sleep(200 * time.Millisecond)
				logger.Log(symbol)
			}
		}
	}(fmt.Sprintf("Destroying resources for cluster [%s]:\n", options.Name), ".", donechan)

	err = <-errchan
	donechan <- true
	if err != nil {
		return errors.Errorf("Unable to destroy resources for cluster [%s]: %v", options.Name, err)
	}

	if options.Purge {
		err := stateStore.Destroy()
		if err != nil {
			return fmt.Errorf("Unable to remove state store for cluster [%s]: %v", options.Name, err)
		}
	}
	return nil
}
