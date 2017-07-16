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
	"github.com/kris-nova/kubicorn/logger"
	"github.com/spf13/cobra"
	"os"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubicorn",
	Short: "Kubernetes cluster management, without any magic",
	Long: fmt.Sprintf(`
%s
`, Unicorn),
	//Run: func(cmd *cobra.Command, args []string) {
	//	cmd.Help()
	//},
}

type Options struct {
	StateStore     string
	StateStorePath string
	Name           string
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	//flags here
	RootCmd.PersistentFlags().IntVarP(&logger.Level, "verbose", "v", 3, "Log level")
	RootCmd.PersistentFlags().BoolVarP(&logger.Color, "color", "C", true, "Toggle colorized logs")

	// register env vars
	registerEnvironmentalVariables()
}

func registerEnvironmentalVariables() {

}
