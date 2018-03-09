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
	"github.com/spf13/viper"
)

// EditCmd represents edit command
func EditCmd() *cobra.Command {
	var eo = &cli.EditOptions{}
	var editCmd = &cobra.Command{
		Use:   "edit <NAME>",
		Short: "Edit a cluster state",
		Long:  `Use this command to edit a state.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				eo.Name = viper.GetString(keyKubicornName)
			case 1:
				eo.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runEdit(eo); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := editCmd.Flags()

	fs.StringVarP(&eo.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&eo.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)
	fs.StringVarP(&eo.Editor, keyEditor, "e", viper.GetString(keyEditor), descEditor)

	fs.StringVar(&eo.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&eo.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&eo.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&eo.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&eo.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&eo.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return editCmd
}

func runEdit(options *cli.EditOptions) error {
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
	}

	logger.Info("Cluster edited")

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
