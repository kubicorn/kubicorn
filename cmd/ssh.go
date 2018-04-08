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

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SSHCmd is used to SSH to the cluster.
// Only master SSH is currently supported.
func SSHCmd() *cobra.Command {
	var ssho = &cli.SSHOptions{}
	var sshCommand = &cobra.Command{
		Use:   "ssh <CLUSTER-NAME>",
		Short: "Run SSH session for a node",
		Long:  `Use this command to connect to the node.`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				ssho.Name = viper.GetString(keyKubicornName)
			case 1:
				ssho.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runSSH(ssho); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}

		},
	}

	fs := sshCommand.Flags()

	bindCommonStateStoreFlags(&ssho.StateStoreOptions, fs)
	bindCommonAwsFlags(&ssho.AwsOptions, fs)

	fs.StringVar(&ssho.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)

	return sshCommand
}

func runSSH(options *cli.SSHOptions) error {
	// Ensure we have a name
	name := options.Name
	if name == "" {
		return errors.New("Empty name. Must specify the name of the cluster to ssh")
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

	logger.Info("Init Cluster")
	cluster, err = initapi.InitCluster(cluster)
	if err != nil {
		return err
	}

	client := ssh.NewSSHClient(cluster.ProviderConfig().KubernetesAPI.Endpoint, "22", "root")
	err = client.Connect()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh for cluster [%s]: %v", name, err)
	}
	err = client.StartInteractiveSession()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh for cluster [%s]: %v", name, err)
	}

	return nil
}
