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
	"strings"

	"github.com/kris-nova/kubicorn/cutil"
	"github.com/kris-nova/kubicorn/cutil/agent"
	"github.com/kris-nova/kubicorn/cutil/initapi"
	"github.com/kris-nova/kubicorn/cutil/kubeconfig"
	"github.com/kris-nova/kubicorn/cutil/local"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/spf13/cobra"
	"github.com/yuroyoro/swalker"
)

type ApplyOptions struct {
	Options
}

var ao = &ApplyOptions{}

// ApplyCmd represents the apply command
func ApplyCmd() *cobra.Command {
	var applyCmd = &cobra.Command{
		Use:   "apply <NAME>",
		Short: "Apply a cluster resource to a cloud",
		Long: `Use this command to apply an API model in a cloud.
	
	This command will attempt to find an API model in a defined state store, and then apply any changes needed directly to a cloud.
	The apply will run once, and ultimately time out if something goes wrong.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				ao.Name = strEnvDef("KUBICORN_NAME", "")
			} else if len(args) > 1 {
				logger.Critical("Too many arguments.")
				os.Exit(1)
			} else {
				ao.Name = args[0]
			}

			err := RunApply(ao)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	applyCmd.Flags().StringVarP(&ao.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	applyCmd.Flags().StringVarP(&ao.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	applyCmd.Flags().StringVarP(&ao.Set, "set", "e", strEnvDef("KUBICORN_SET", ""), "set cluster setting")
	applyCmd.Flags().StringVar(&ao.AwsProfile, "aws-profile", strEnvDef("AWS_PROFILE", ""), "The profile to be used as defined in $HOME/.aws/credentials")
	applyCmd.Flags().StringVar(&ao.GitRemote, "git-config", strEnvDef("KUBICORN_GIT_CONFIG", "git"), "The git remote url to be used for saving the git state for the cluster.")

	return applyCmd
}

func RunApply(options *ApplyOptions) error {

	// Ensure we have SSH agent
	agent := agent.NewAgent()

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to apply")
	}

	// Expand state store path
	options.StateStorePath = expandPath(options.StateStorePath)

	// Register state store and check if it exists
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if stateStore.Exists() {
		return fmt.Errorf("State store [%s] exists, will not overwrite. Delete existing profile [%s] and retry", name, options.StateStorePath+"/"+name)
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}
	logger.Info("Loaded cluster: %s", cluster.Name)

	if options.Set != "" {
		sets := strings.Split(options.Set, ",")
		for _, set := range sets {
			parts := strings.SplitN(set, "=", 2)
			if len(parts) == 1 {
				continue
			}
			err := swalker.Write(strings.Title(parts[0]), cluster, parts[1])
			if err != nil {
				logger.Critical("Error expanding set flag: %#v", err)
			}
		}
	}

	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		return err
	}

	runtimeParams := &cutil.RuntimeParameters{}

	if len(ao.AwsProfile) > 0 {
		runtimeParams.AwsProfile = ao.AwsProfile
	}

	reconciler, err := cutil.GetReconciler(cluster, runtimeParams)
	if err != nil {
		return fmt.Errorf("Unable to get reconciler: %v", err)
	}

	logger.Info("Query existing resources")
	actual, err := reconciler.Actual(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get actual cluster: %v", err)
	}
	logger.Info("Resolving expected resources")
	expected, err := reconciler.Expected(cluster)
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

	err = kubeconfig.RetryGetConfig(newCluster, agent)
	if err != nil {
		return fmt.Errorf("Unable to write kubeconfig: %v", err)
	}

	logger.Always("The [%s] cluster has applied successfully!", newCluster.Name)
	if path, ok := newCluster.Annotations[kubeconfig.ClusterAnnotationKubeconfigLocalFile]; ok {
		path = local.Expand(path)
		logger.Always("To start using your cluster, you need to run")
		logger.Always("  export KUBECONFIG=\"${KUBECONFIG}:%s\"", path)
	}
	logger.Always("You can now `kubectl get nodes`")
	privKeyPath := strings.Replace(cluster.SSH.PublicKeyPath, ".pub", "", 1)
	logger.Always("You can SSH into your cluster ssh -i %s %s@%s", privKeyPath, newCluster.SSH.User, newCluster.KubernetesAPI.Endpoint)

	return nil
}
