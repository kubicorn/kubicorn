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
	"fmt"
	"github.com/Nivenly/kamp/local"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kamp",
	Short: "Rapidly develop, run, and build containers directly in Kubernetes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

type Options struct {
	Verbosity  int
	Dockerfile string
}

var O = &Options{}

func init() {
	RootCmd.PersistentFlags().IntVarP(&O.Verbosity, "verbosity", "v", 2, "Verbosity [0 - 4]")
	RootCmd.PersistentFlags().StringVarP(&O.Dockerfile, "dockerfile", "f", "./Dockerfile", "The dockerfile to check for and use.")
}

var Version string
var GitSha string

func sttySize() (string, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func termDim() (int, int) {
	stty, err := sttySize()
	if err != nil {
		fmt.Println(1)
		return 100, 100
	}
	spl := strings.Split(stty, " ")
	if len(spl) != 2 {
		fmt.Println(2)
		return 100, 100
	}
	w := strings.TrimSpace(spl[1])
	l := strings.TrimSpace(spl[0])
	width, err := strconv.Atoi(w)
	if err != nil {
		fmt.Println(3)
		return 100, 100
	}
	length, err := strconv.Atoi(l)
	if err != nil {
		fmt.Println(4)
		return 100, 100
	}
	return length, width
}

func KampBannerMessage(msg string) string {
	var banner string
	banner = `Copyright 2017 - The Kamp Authors
 _
| | ____ _ _ __ ___  _ __
| |/ / _\ | '_  _  \| '_ \   v%s
|   < (_| | | | | | | |_) |
|_|\_\__,_|_| |_| |_| .__/   [%s]
                    |_|
%s

`
	_, w := termDim()
	s := ""
	for w > 0 {
		s = fmt.Sprintf("%s—", s)
		w--
	}
	return fmt.Sprintf(banner, Version, GitSha, msg)
}

func Check(err error) {
	if err != nil {
		local.Critical("Fatal Error: %v", err)
		os.Exit(1)
	}
}
