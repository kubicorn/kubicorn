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

package boot

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var glideInstallCmd = &cobra.Command{
	Use:   "glide-install",
	Short: "Runs glide install and flatten vendored directories",
	Long:  `Runs glide install and flatten vendored directories`,
	Run:   RunGlideInstall,
}

func AddGlideInstall(cmd *cobra.Command) {
	cmd.AddCommand(glideInstallCmd)
}

func RunGlideInstall(cmd *cobra.Command, args []string) {
	c := exec.Command("glide", "install", "--strip-vendor")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run glide install\n%v\n", err)
		os.Exit(-1)
	}
}
