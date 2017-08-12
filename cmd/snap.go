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
	//"github.com/spf13/pflag"
	"fmt"
	"os"

	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/snap"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/spf13/cobra"
)

var snapCommand = &cobra.Command{
	Use:   "snap [-n|--name NAME] [-f|--file LABELNAME] [-N|--namespaces NAMESPACES]",
	Short: "Take a snapshot of a Kubernetes cluster",
	Long: `Use this command to take a snapshot of a Kubernetes cluster and save the snapshot to your state store.

This command will create a snapshot of a Kubernetes cluster and save the snapshot in your state store.
Once the snapshots have been saved, they can be optionally modified and used to replicate the original cluster.
One of the powerful features of snapshots is their ability to work cross cloud. A user can snapshot a cluster
in one cloud, and then replicate the cluster in another cloud.

In the case of moving clouds it is important to understand that every public cloud provider is different.
Kubicorn does the best it can to mirror the environment, but in some cases the configurations won't match perfectly.

Snapshots are not backups. They do not backup any persistent data. They only backup Kubernetes resources and definitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunSnap(so)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}

	},
}

type SnapOptions struct {
	Options
	File           string
	KubeconfigPath string
	Namespaces     []string
}

var so = &SnapOptions{}

func init() {

	snapCommand.Flags().StringVarP(&so.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	snapCommand.Flags().StringVarP(&so.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	snapCommand.Flags().StringVarP(&so.Name, "name", "n", strEnvDef("KUBICORN_NAME", ""), "An optional name to use. If empty, will generate a random name.")
	snapCommand.Flags().StringVarP(&so.File, "file", "f", strEnvDef("KUBICORN_FILE", ""), "An optional filename to save the snapshot to")
	snapCommand.Flags().StringVarP(&so.KubeconfigPath, "kube-config-path", "p", strEnvDef("KUBICORN_KUBE_CONFIG_PATH", "~/.kube/config"), "The optional kube config path to use to query Kubernetes")
	snapCommand.Flags().StringSliceVarP(&so.Namespaces, "namespaces", "N", strSliceEnvDef("KUBICORN_NAMESPACES", []string{"*"}), "List of namespaces to use in the query. * is acceptable for query all namespaces.")

	flagApplyAnnotations(snapCommand, "name", "__kubicorn_parse_list")
	flagApplyAnnotations(snapCommand, "profile", "__kubicorn_parse_profiles")

	RootCmd.AddCommand(snapCommand)
}

func RunSnap(options *SnapOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return fmt.Errorf("Empty name")
	}

	// Expand paths
	options.KubeconfigPath = expandPath(options.KubeconfigPath)
	options.File = expandPath(options.File)
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
	if !stateStore.Exists() {
		return fmt.Errorf("State store [%s] doesn't exist", name)
	}

	declaredCluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to read cluster from state store: %v", err)
	}

	// Ensure we are always pulling an accurate representation of the cluster
	reconciler, err := cutil.GetReconciler(declaredCluster)
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

	util := snap.NewSnapShotUtility(actual, stateStore, options.KubeconfigPath)
	snap, err := util.Capture(options.Namespaces, options.File)
	if err != nil {
		return fmt.Errorf("Unable to snapshot cluster: %v", err)
	}

	// Always audit and commit
	err = stateStore.Commit(actual)
	if err != nil {
		return fmt.Errorf("Unable to init state store: %v", err)
	}

	err = snap.WriteCompressedFile()
	if err != nil {
		return fmt.Errorf("Unable to compress and write file: %v", err)
	}
	logger.Always("The snapshot has been successfully saved!")
	logger.Always("%s", snap.AbsolutePath())
	return nil
}
