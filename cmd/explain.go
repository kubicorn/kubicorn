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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg"
	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/initapi"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type OutputData struct {
	Actual   *cluster.Cluster
	Expected *cluster.Cluster
}

// ExplainCmd represents the explain command
func ExplainCmd() *cobra.Command {
	var exo = &cli.ExplainOptions{}

	var cmd = &cobra.Command{
		Use:   "explain",
		Short: "Explain cluster",
		Long:  `Output expected and actual state of the given cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				exo.Name = viper.GetString(keyKubicornName)
			case 1:
				exo.Name = args[0]
			default:
				logger.Critical("Too many arguments.")
				os.Exit(1)
			}

			if err := runExplain(exo); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()

	fs.StringVarP(&exo.StateStore, keyStateStore, "s", viper.GetString(keyStateStore), descStateStore)
	fs.StringVarP(&exo.StateStorePath, keyStateStorePath, "S", viper.GetString(keyStateStorePath), descStateStorePath)
	fs.StringVarP(&exo.Output, keyOutput, "o", viper.GetString(keyOutput), descOutput)

	fs.StringVar(&exo.GitRemote, keyGitConfig, viper.GetString(keyGitConfig), descGitConfig)
	fs.StringVar(&exo.S3AccessKey, keyS3Access, viper.GetString(keyS3Access), descS3AccessKey)
	fs.StringVar(&exo.S3SecretKey, keyS3Secret, viper.GetString(keyS3Secret), descS3SecretKey)
	fs.StringVar(&exo.BucketEndpointURL, keyS3Endpoint, viper.GetString(keyS3Endpoint), descS3Endpoints)
	fs.StringVar(&exo.BucketName, keyS3Bucket, viper.GetString(keyS3Bucket), descS3Bucket)

	fs.BoolVar(&exo.BucketSSL, keyS3SSL, viper.GetBool(keyS3SSL), descS3SSL)

	return cmd
}

func runExplain(options *cli.ExplainOptions) error {

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
	} else if !stateStore.Exists() {
		return fmt.Errorf("State store [%s] does not exists, can't edit", name)
	}

	cluster, err := stateStore.GetCluster()
	if err != nil {
		return fmt.Errorf("Unable to get cluster [%s]: %v", name, err)
	}

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

	var d OutputData
	d.Actual, err = reconciler.Actual(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get actual cluster: %v", err)
	}
	d.Expected, err = reconciler.Expected(cluster)
	if err != nil {
		return fmt.Errorf("Unable to get expected cluster: %v", err)
	}

	if options.Output == "json" {
		o, err := json.MarshalIndent(d, "", "\t")
		if err != nil {
			return fmt.Errorf("Unable to parse cluster: %v", err)
		}
		fmt.Printf("%s\n", o)
	} else {
		return fmt.Errorf("Unsupported output format")
	}

	return nil
}
