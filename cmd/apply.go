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

	"github.com/kubicorn/kubicorn/pkg"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"
	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/resourcedeploy"
	"github.com/kubicorn/kubicorn/pkg/state/crd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuroyoro/swalker"
)

// ApplyCmd represents the apply command
func ApplyCmd() *cobra.Command {
	var ao = &cli.ApplyOptions{}
	var applyCmd = &cobra.Command{
		Use:   "apply <NAME>",
		Short: "Apply a cluster resource to a cloud",
		Long: `Use this command to apply an API model in a cloud.
	
	This command will attempt to find an API model in a defined state store, and then apply any changes needed directly to a cloud.
	The apply will run once, and ultimately time out if something goes wrong.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				ao.Name = viper.GetString(keyKubicornName)
			case 1:
				ao.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runApply(ao); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}

	fs := applyCmd.Flags()

	fs.StringVarP(&ao.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&ao.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)
	fs.StringVarP(&ao.Set, keyKubicornSet, "e", viper.GetString(keyKubicornSet), descSet)

	fs.StringVar(&ao.AwsProfile, keyAwsProfile, viper.GetString(keyAwsProfile), descAwsProfile)
	fs.StringVar(&ao.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&ao.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&ao.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&ao.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&ao.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&ao.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return applyCmd
}

func runApply(options *cli.ApplyOptions) error {

	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to apply")
	}

	// Expand state store path
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	// Register state store
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
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

	logger.Info("Init Cluster")
	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		return err
	}

	runtimeParams := &pkg.RuntimeParameters{}

	if len(options.AwsProfile) > 0 {
		runtimeParams.AwsProfile = options.AwsProfile
	}

	reconciler, err := pkg.GetReconciler(cluster, runtimeParams)
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

	if err = stateStore.Commit(newCluster); err != nil {
		return fmt.Errorf("Unable to commit state store: %v", err)
	}

	logger.Info("Updating state store for cluster [%s]", options.Name)
	logger.Info("Hanging while fetching kube config...")
	if err = kubeconfig.RetryGetConfig(newCluster); err != nil {
		return fmt.Errorf("Unable to write kubeconfig: %v", err)
	}

	if newCluster.ControllerDeployment != nil {
		// -------------------------------------------------------------------------------------------------------------
		//
		// Here is where we hook in for the new controller logic
		// This is exclusive to profiles that have a controller defined
		//
		logger.Info("Deploying cluster controller: %s", newCluster.ControllerDeployment.Spec.Template.Spec.Containers[0].Image)
		err := resourcedeploy.DeployClusterControllerDeployment(newCluster)
		if err != nil {
			return fmt.Errorf("Unable to deploy cluster controller: %v", err)
		}
		crdStateStore := crd.NewCRDStore(&crd.CRDStoreOptions{
			BasePath:    options.StateStorePath,
			ClusterName: options.Name,
		})
		crdStateStore.Commit(newCluster)
	}

	logger.Always("The [%s] cluster has applied successfully!", newCluster.Name)
	if path, ok := newCluster.Annotations[kubeconfig.ClusterAnnotationKubeconfigLocalFile]; ok {
		path = local.Expand(path)
		logger.Always("To start using your cluster, you need to run")
		logger.Always("  export KUBECONFIG=\"${KUBECONFIG}:%s\"", path)
	}
	logger.Always("You can now `kubectl get nodes`")
	privKeyPath := strings.Replace(cluster.ProviderConfig().SSH.PublicKeyPath, ".pub", "", 1)
	logger.Always("You can SSH into your cluster ssh -i %s %s@%s", privKeyPath, newCluster.ProviderConfig().SSH.User, newCluster.ProviderConfig().KubernetesAPI.Endpoint)

	return nil
}
