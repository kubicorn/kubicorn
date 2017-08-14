// Copyright © 2017 The Kamp Authors
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
	"github.com/Nivenly/kamp/local"
	"github.com/Nivenly/kamp/runner"
	"github.com/Nivenly/kamp/runner/kubernetes"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a container in a Kubernetes cluster",
	Long:  KampBannerMessage("Run and attach to an arbitrary container in a Kubernetes cluster."),
	Run: func(cmd *cobra.Command, args []string) {
		local.LogLevel = O.Verbosity
		if len(os.Args) < 3 {
			cmd.Help()
			os.Exit(0)
		}
		image := os.Args[2]
		if strings.Contains("--", image) {
			color.Red("Invalid image [%s]", image)
			cmd.Help()
		}
		//name := os.Args[3]
		//if strings.Contains("--", name) {
		//	color.Red("Invalid image [%s]", image)
		//	cmd.Help()
		//}
		runOpt.ImageQuery = image
		err := RunRun(runOpt)
		Check(err)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringSliceVarP(&runOpt.Command, "cmd", "c", []string{"/bin/bash"}, "The command to execute in the container.")
	runCmd.Flags().StringVarP(&runOpt.KubernetesNamespace, "namespace", "n", "default", "The Kubernetes namespace to run the container in.")
	runCmd.Flags().StringVarP(&runOpt.Volume, "volume", "V", "", "The volume string to mount CIFS volumes with. <local>:<remote>")
	runCmd.Flags().StringVarP(&runOpt.Name, "name", "N", "kamper", "The name of your kamp pod")
	runCmd.SetUsageTemplate(UsageTemplate)
}

type RunOptions struct {
	Options
	ImageQuery          string
	Name                string
	Command             []string
	KubernetesNamespace string
	Volume              string
}

var runOpt = &RunOptions{}

func RunRun(options *RunOptions) error {

	// ------------------------ Run in Kubernetes ------------------------
	if err := kubernetes.NewKubernetesRunner(&kubernetes.Options{
		Options: runner.Options{
			Command:    options.Command,
			ImageQuery: options.ImageQuery,
			Name:       options.Name,
		},
		Namespace: options.KubernetesNamespace,
	}).Run(); err != nil {
		return err
	}
	return nil
}

const UsageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "<image>:(tag) <name> [flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}
Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}
Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
