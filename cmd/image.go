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
	"os"

	lol "github.com/kris-nova/lolgopher"
	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Take an image of a Kubernetes cluster",
	Long: `Use this command to image a Kubernetes cluster.

This command will take an idempotent image of a Kubernetes cluster called a snapshot.
The snapshot can be used to create a copy of your Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("KUBICORN_TRUECOLOR") != "" {
			cmd.SetOutput(&lol.Writer{Output: os.Stdout, ColorMode: lol.ColorModeTrueColor})
		}
		// TODO: Work your own magic here
		fmt.Println("image called")
	},
}

func init() {
	RootCmd.AddCommand(imageCmd)
}
