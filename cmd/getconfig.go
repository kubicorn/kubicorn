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

	"github.com/kris-nova/kubicorn/cutil/initapi"
	"github.com/kris-nova/kubicorn/cutil/kubeconfig"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/spf13/cobra"
)

// getConfigCmd represents the apply command
var getConfigCmd = &cobra.Command{
	Use:   "getconfig [-n|--name NAME]",
	Short: "Manage Kubernetes configuration",
	Long: `Use this command to pull a kubeconfig file from a cluster so you can use kubectl.

This command will attempt to find a cluster, and append a local kubeconfig file with a kubeconfig `,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunGetConfig(cro)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type GetConfigOptions struct {
	Options
}

var cro = &GetConfigOptions{}

func init() {
	getConfigCmd.Flags().StringVarP(&cro.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	getConfigCmd.Flags().StringVarP(&cro.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	getConfigCmd.Flags().StringVarP(&cro.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "An optional name to use. If empty, will generate a random name.")

	flagApplyAnnotations(getConfigCmd, "name", "__kubicorn_parse_list")

	RootCmd.AddCommand(getConfigCmd)
}

func RunGetConfig(options *GetConfigOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to get config")
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

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}
	logger.Info("Loaded cluster: %s", cluster.Name)

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
