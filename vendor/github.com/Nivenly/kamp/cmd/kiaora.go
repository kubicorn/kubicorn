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
	"github.com/spf13/cobra"
)

// kiaoraCmd represents the kiaora command
var kiaoraCmd = &cobra.Command{
	Use:   "kiaora",
	Short: "A volume broker for Kubernetes.",
	Long: `Kiaora is a volume broker that runs within the context of a Kubernetes cluster.
The tool has many different components that run harmoniously to allow easy volume mapping
from external resources to internal Kubernetes volumes that pods can easily mount.`,
}

func init() {
	RootCmd.AddCommand(kiaoraCmd)
}
