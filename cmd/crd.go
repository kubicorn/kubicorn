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
	"github.com/golang/glog"
	"time"
	"fmt"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/util"
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
		Short: "Used to create a CRD in Kubernetes",
		Long: `Used to create a CRD in Kubernetes`,
		Run: func(cmd *cobra.Command, args []string) {

			err := RunCRDCreate(crdo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	//getConfigCmd.Flags().StringVarP(&cro.StateStore, "state-store", "s", strEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	//getConfigCmd.Flags().StringVarP(&cro.StateStorePath, "state-store-path", "S", strEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")

	crdCmd.Flags().StringVarP(&crdo.KubeConfigPath, "kube-config-path", "k", "/Users/knova/.kube/config", "The path to use for the kube config")

	return crdCmd
}

func RunCRDCreate(options *CRDOptions) error {


	cs, err := util.NewClientSet(options.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("unable to initialize new client set: %v", err)
	}

	// Create CRD for Machines
	success := false
	for i := 0; i <= RetryAttempts; i++ {
		if _, err = clusterv1.CreateMachinesCRD(cs); err != nil {
			glog.Info("Failure creating Machines CRD (will retry).")
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		success = true
		logger.Info("Machines CRD created successfully!")
		break
	}

	if !success {
		return fmt.Errorf("error creating Machines CRD: %v", err)
	}
	return nil

	return nil
}
