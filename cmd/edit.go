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
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

var eo = &cli.EditOptions{}

// EditCmd represents edit command
func EditCmd() *cobra.Command {
	var editCmd = &cobra.Command{
		Use:   "edit <NAME>",
		Short: "Edit a cluster state",
		Long:  `Use this command to edit a state.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				ao.Name = cli.StrEnvDef("KUBICORN_NAME", "")
			} else if len(args) > 1 {
				logger.Critical("Too many arguments.")
				os.Exit(1)
			} else {
				eo.Name = args[0]
			}

			err := RunEdit(eo)
			if err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	editCmd.Flags().StringVarP(&eo.StateStore, "state-store", "s", cli.StrEnvDef("KUBICORN_STATE_STORE", "fs"), "The state store type to use for the cluster")
	editCmd.Flags().StringVarP(&eo.StateStorePath, "state-store-path", "S", cli.StrEnvDef("KUBICORN_STATE_STORE_PATH", "./_state"), "The state store path to use")
	editCmd.Flags().StringVarP(&eo.Editor, "editor", "e", cli.StrEnvDef("EDITOR", "vi"), "The editor used to edit the state store")

	// git flags
	editCmd.Flags().StringVar(&eo.GitRemote, "git-config", cli.StrEnvDef("KUBICORN_GIT_CONFIG", "git"), "The git remote url to use")

	// s3 flags
	editCmd.Flags().StringVar(&eo.S3AccessKey, "s3-access", cli.StrEnvDef("KUBICORN_S3_ACCESS_KEY", ""), "The s3 access key.")
	editCmd.Flags().StringVar(&eo.S3SecretKey, "s3-secret", cli.StrEnvDef("KUBICORN_S3_SECRET_KEY", ""), "The s3 secret key.")
	editCmd.Flags().StringVar(&eo.BucketEndpointURL, "s3-endpoint", cli.StrEnvDef("KUBICORN_S3_ENDPOINT", ""), "The s3 endpoint url.")
	editCmd.Flags().BoolVar(&eo.BucketSSL, "s3-ssl", cli.BoolEnvDef("KUBICORN_S3_SSL", true), "The s3 bucket name to be used for saving the git state for the cluster.")
	editCmd.Flags().StringVar(&eo.BucketName, "s3-bucket", cli.StrEnvDef("KUBICORN_S3_BUCKET", ""), "The s3 bucket name to be used for saving the s3 state for the cluster.")

	return editCmd
}

func RunEdit(options *cli.EditOptions) error {
	options.StateStorePath = cli.ExpandPath(options.StateStorePath)

	name := options.Name

	// Register state store and check if it exists
	stateStore, err := options.NewStateStore()
	if err != nil {
		return err
	} else if !stateStore.Exists() {
		return fmt.Errorf("State store [%s] does not exists, can't edit", name)
	}
	stateContent, err := stateStore.ReadStore()
	if err != nil {
		return err
	}

	fpath := os.TempDir() + "/kubicorn_cluster.tmp"
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fpath, stateContent, 0664)
	f.Close()

	path, err := exec.LookPath(options.Editor)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	cmd := exec.Command(path, fpath)
	err = cmd.Start()
	if err != nil {
		os.Remove(fpath)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logger.Debug("Error while editing. Error: %v", err)
		os.Remove(fpath)
		return err
	} else {
		logger.Info("Cluster edited")
	}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	cluster, err := stateStore.BytesToCluster(data)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	logger.Info("Init Cluster")
	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		os.Remove(fpath)
		return err
	}

	// Init new state store with the cluster resource
	err = stateStore.Commit(cluster)
	if err != nil {
		os.Remove(fpath)
		return fmt.Errorf("Unable to init state store: %v", err)
	}
	os.Remove(fpath)

	logger.Always("The state [%s/%s/cluster.yaml] has been updated.", options.StateStorePath, name)
	return nil
}
