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
	"github.com/kris-nova/kubicorn/crd"


	"errors"
	"fmt"
	"os"

	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/state"
	"github.com/kris-nova/kubicorn/state/fs"
	"github.com/kris-nova/kubicorn/state/jsonfs"
	"github.com/spf13/cobra"
	"github.com/kris-nova/kubicorn/profiles"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

const (
	//MasterIPAttempts       = 40
	SleepSecondsPerAttempt = 5
	RetryAttempts          = 30
	//DeleteAttempts         = 150
	//DeleteSleepSeconds     = 5
)


type CRDOptions struct {
	KubeConfigPath string
	Options
}

var crdo = &CRDOptions{}

// GetConfigCmd represents the apply command
func CRDCommand() *cobra.Command {
	var crdCmd = &cobra.Command{
		Use:   "crd <TYPE>",
		Short: "Used to create a clusters and machines CRD in Kubernetes based on a state store",
		Long: `This command will create a machines CRD and clusters CRD based on a written state store.`,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				crdo.Name = strEnvDef("KUBICORN_NAME", "")
			} else if len(args) > 1 {
				logger.Critical("Too many arguments.")
				os.Exit(1)
			} else {
				crdo.Name = args[0]
			}
			err := RunCRDCreate(crdo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	crdCmd.Flags().StringVarP(&crdo.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	crdCmd.Flags().StringVarP(&crdo.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	crdCmd.Flags().StringVarP(&crdo.KubeConfigPath, "kube-config-path", "k", "/Users/knova/.kube/config", "The path to use for the kube config")

	return crdCmd
}

func RunCRDCreate(options *CRDOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to apply")
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
	case "jsonfs":
		logger.Info("Selected [jsonfs] state store")
		stateStore = jsonfs.NewJSONFileSystemStore(&jsonfs.JSONFileSystemStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: name,
		})
	}

	kubicornCluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("unable to get cluster [%s]: %v", name, err)
	}


	// Translate into an API cluster
	apiCluster, ok := kubicornCluster.(*v1alpha1.Cluster)
	if !ok {
		return fmt.Errorf("unable to unmarshal cluster, major error")
	}

	cluster, err := profiles.DeserializeProviderConfig(apiCluster.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("unable to deserialize provider config: %v", err)
	}


	logger.Info("Loaded cluster: %s", cluster.Name)


	manager, err := crd.NewCRDManager(cluster)

	err = manager.CreateMachines()
	if err != nil {
		logger.Critical("Unable to create machines CRD: %v", err)
	}
	err = manager.CreateClusters()
	if err != nil {
		logger.Critical("Unable to create clusters CRD: %v", err)
	}
	return nil
}
