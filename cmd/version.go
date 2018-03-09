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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

var versionFile = "/src/github.com/kubicorn/kubicorn/VERSION"

// VersionCmd represents the version command
func VersionCmd() *cobra.Command {
	var vo = &cli.VersionOptions{}
	return &cobra.Command{
		Use:   "version",
		Short: "Verify Kubicorn version",
		Long: `Use this command to check the version of Kubicorn.
	
	This command will return the version of the Kubicorn binary.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runVersion(vo); err != nil {
				logger.Critical(err.Error())
				os.Exit(1)
			}
		},
	}
}

func runVersion(options *cli.VersionOptions) error {
	options.Version = getVersion()
	options.GitCommit = getGitCommit()
	options.BuildDate = time.Now().UTC().String()
	options.GOVersion = runtime.Version()
	options.GOARCH = runtime.GOARCH
	options.GOOS = runtime.GOOS
	voBytes, err := json.Marshal(options)
	if err != nil {
		return err
	}
	fmt.Println("Kubicorn version: ", string(voBytes))
	return nil
}

func getVersion() string {
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return string(vBytes)
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return string(output)
}
