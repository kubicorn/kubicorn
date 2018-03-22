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
	"os"
	"runtime"
	"time"

	"github.com/kubicorn/kubicorn/pkg/cli"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	versionFile = "/src/github.com/kubicorn/kubicorn/VERSION"

	// The GitSha of the current commit (automatically set at compile time)
	GitSha string

	// The Version of the program from the VERSION file (automatically set at compile time)
	Version string
)

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
	options.Version = Version
	options.GitCommit = GitSha
	options.BuildDate = time.Now().UTC().String()
	options.GOVersion = runtime.Version()
	options.GOARCH = runtime.GOARCH
	options.GOOS = runtime.GOOS
	voBytes, err := json.Marshal(options)
	if err != nil {
		return err
	}
	// Keep this true json so we can query via jq
	fmt.Printf("%s\n", string(voBytes))
	return nil
}
