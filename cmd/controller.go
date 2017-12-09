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
	"github.com/spf13/cobra"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"os"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-deploy/cluster-api/client"
	//clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type ControllerOptions struct {
	KubeConfigPath string
	Options
}

var ctlo = &ControllerOptions{}

// GetConfigCmd represents the apply command
func ControllerCmd() *cobra.Command {
	var ctlCmd = &cobra.Command{
		Use:   "controller <cloud>",
		Short: "Run kubicorn as a node controller",
		Long: `Run kubicorn as a node controller`,
		Run: func(cmd *cobra.Command, args []string) {

			err := RunController(ctlo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	//getConfigCmd.Flags().StringVarP(&cro.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	//getConfigCmd.Flags().StringVarP(&cro.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")

	//crdCmd.Flags().StringVarP(&crdo.KubeConfigPath, "kube-config-path", "k", "/Users/knova/.kube/config", "The path to use for the kube config")
	return ctlCmd
}

func RunController(options *ControllerOptions) error {


	// Config
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfigPath)

	// Client
	cs, err := client.NewForConfig(config)



	machines, err := cs.Machines().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	cachedCount := -1


	// Loop indefinitely
	for {

		// Hard code AWS for demo
		sdk, err := awsSdkGo.NewSdk("us-west-2", "default")
		if err != nil {
			logger.Critical(err.Error())
			continue
		}
		machineCount := len(machines.Items)
		name := ""
		m := int64(machineCount)
		if machineCount != cachedCount {
			input := &autoscaling.UpdateAutoScalingGroupInput{
				MaxSize: &m,
				MinSize: &m,
				LaunchConfigurationName: &name,
			}
			_, err := sdk.ASG.UpdateAutoScalingGroup(input)
			if err != nil {
				logger.Critical("unable to update ASG: %v", err)

			}else {
				cachedCount = machineCount
				logger.Info("updated machine count [%d]", machineCount)
			}
		}
	}



}
