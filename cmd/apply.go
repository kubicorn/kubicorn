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
	clusterInit "github.com/kris-nova/kubicorn/cluster"
	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/kubeconfig"
	"github.com/kris-nova/kubicorn/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a cluster resource to a cloud",
	Long: `Use this command to apply an API model in a cloud.

This command will attempt to find an API model in a defined state store, and then apply any changes needed directly to a cloud.
The apply will run once, and ultimately time out if something goes wrong.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunApply(ao)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type ApplyOptions struct {
	Options
}

var ao = &ApplyOptions{}

func init() {
	RootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&ao.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	applyCmd.Flags().StringVarP(&ao.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	applyCmd.Flags().StringVarP(&ao.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "An optional name to use. If empty, will generate a random name.")
}

func RunApply(options *ApplyOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to apply.")
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

	cluster, err = clusterInit.InitCluster(cluster)
	if err != nil {
		return err
	}

	reconciler, err := cutil.GetReconciler(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get reconciler: %v", err)
	}

	if err := reconciler.Init(); err != nil {
		return fmt.Errorf("Unable to init reconciler: %v", err)
	}
	logger.Info("Query existing resources")
	actual, err := reconciler.GetActual()
	if err != nil {
		return fmt.Errorf("Unable to get actual cluster: %v", err)
	}
	logger.Info("Resolving expected resources")
	expected, err := reconciler.GetExpected()
	if err != nil {
		return fmt.Errorf("Unable to get expected cluster: %v", err)
	}

	logger.Info("Reconciling")
	newCluster, err := reconciler.Reconcile(actual, expected)
	if err != nil {
		return fmt.Errorf("Unable to reconcile cluster: %v", err)
	}

	err = stateStore.Commit(newCluster)
	if err != nil {
		return fmt.Errorf("Unable to commit state store: %v", err)
	}

	logger.Info("Updating state store for cluster [%s]", options.Name)

	err = kubeconfig.RetryGetConfig(newCluster)
	if err != nil {
		return fmt.Errorf("Unable to write kubeconfig: %v", err)
	}

	logger.Always("The [%s] cluster has applied successfully!", cluster.Name)
	logger.Always("You can now `kubectl get nodes`")
	privKeyPath := strings.Replace(cluster.Ssh.PublicKeyPath, ".pub", "", 1)
	logger.Always("You can SSH into your cluster ssh -i %s %s@%s", privKeyPath, cluster.Ssh.User, cluster.KubernetesApi.Endpoint)

	return nil
}
