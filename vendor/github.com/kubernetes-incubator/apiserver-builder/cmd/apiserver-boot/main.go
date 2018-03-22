/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot"
	"github.com/spf13/cobra"
)

var gopath string
var wd string

func main() {
	gopath = os.Getenv("GOPATH")
	if len(gopath) == 0 {
		fmt.Fprintf(os.Stderr, "GOPATH not defined\n")
		os.Exit(-1)
	}
	boot.GoSrc = filepath.Join(gopath, "src")

	var err error
	wd, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	if !strings.HasPrefix(filepath.Dir(wd), boot.GoSrc) {
		fmt.Fprintf(os.Stderr,
			"apiserver-boot must be run from the directory containing the go package to "+
				"bootstrap. This must be under $GOPATH/src/<package>. "+
				"\nCurrent GOPATH=%s.  \nCurrent directory=%s\n", gopath, wd)
		os.Exit(-1)
	}
	boot.Repo = strings.Replace(wd, boot.GoSrc+"/", "", 1)
	boot.AddCreateGroup(cmd)
	boot.AddCreateResource(cmd)
	boot.AddCreateVersion(cmd)
	boot.AddGenerate(cmd)
	boot.AddInit(cmd)
	boot.AddGlideInstall(cmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(-1)
	}
}

var cmd = &cobra.Command{
	Use:   "apiserver-boot",
	Short: "apiserver-boot bootstraps building Kubernetes extensions",
	Long:  `apiserver-boot bootstraps building Kubernetes extensions`,
	Run:   RunMain,
}

func RunMain(cmd *cobra.Command, args []string) {
	cmd.Help()
}
