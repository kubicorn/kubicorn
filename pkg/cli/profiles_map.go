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

package cli

import (
	"sort"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/profiles/amazon"
	"github.com/kubicorn/kubicorn/profiles/azure"
	"github.com/kubicorn/kubicorn/profiles/digitalocean"
	"github.com/kubicorn/kubicorn/profiles/googlecompute"
	"github.com/kubicorn/kubicorn/profiles/openstack/ovh"
	"github.com/kubicorn/kubicorn/profiles/packet"

	"fmt"
	"math"
)

func sortedKeys(profileMap map[string]ProfileMap) []string {
	keys := []string{}
	for k := range profileMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var (
	// P helps format usage details
	P = func() string {
		str := ""
		spaces := ""
		maxLen := 0
		for shorthand := range ProfileMapIndexed {
			l := len(shorthand)
			if l > maxLen {
				maxLen = l
			}
		}
		for _, shorthand := range sortedKeys(ProfileMapIndexed) {
			spaces = ""
			k := math.Abs(float64(maxLen) - float64(len(shorthand)) + 3)
			for i := 0; i < int(k); i++ {
				spaces = fmt.Sprintf("%s%s", spaces, " ")
			}
			str = fmt.Sprintf("%s   %s%s %s\n", str, shorthand, spaces, ProfileMapIndexed[shorthand].Description)
		}
		return str
	}()
	// UsageTemplate is a template for showing usage
	UsageTemplate = fmt.Sprintf(`Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}

Profiles:
%s{{if gt .Aliases 0}}
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
`, P)
)

// ProfileFunc represents function type for a profile
type ProfileFunc func(name string) *cluster.Cluster

// ProfileMap object representing profile function and a description
type ProfileMap struct {
	ProfileFunc ProfileFunc
	Description string
}

// ProfileMapIndexed is a map of possible profiles
var ProfileMapIndexed = map[string]ProfileMap{
	"azure": {
		ProfileFunc: azure.NewUbuntuCluster,
		Description: "Ubuntu on Azure",
	},
	"azure-ubuntu": {
		ProfileFunc: azure.NewUbuntuCluster,
		Description: "Ubuntu on Azure",
	},
	"amazon": {
		ProfileFunc: amazon.NewUbuntuCluster,
		Description: "Ubuntu on Amazon",
	},
	"aws": {
		ProfileFunc: amazon.NewUbuntuCluster,
		Description: "Ubuntu on Amazon",
	},
	"do": {
		ProfileFunc: digitalocean.NewUbuntuCluster,
		Description: "Ubuntu on DigitalOcean",
	},
	"google": {
		ProfileFunc: googlecompute.NewUbuntuCluster,
		Description: "Ubuntu on Google Compute",
	},
	"digitalocean": {
		ProfileFunc: digitalocean.NewUbuntuCluster,
		Description: "Ubuntu on DigitalOcean",
	},
	"do-ubuntu": {
		ProfileFunc: digitalocean.NewUbuntuCluster,
		Description: "Ubuntu on DigitalOcean",
	},
	"aws-ubuntu": {
		ProfileFunc: amazon.NewUbuntuCluster,
		Description: "Ubuntu on Amazon",
	},
	"do-centos": {
		ProfileFunc: digitalocean.NewCentosCluster,
		Description: "CentOS on DigitalOcean",
	},
	"aws-centos": {
		ProfileFunc: amazon.NewCentosCluster,
		Description: "CentOS on Amazon",
	},
	"aws-debian": {
		ProfileFunc: amazon.NewDebianCluster,
		Description: "Debian on Amazon",
	},
	"ovh": {
		ProfileFunc: ovh.NewUbuntuCluster,
		Description: "Ubuntu on OVH",
	},
	"ovh-ubuntu": {
		ProfileFunc: ovh.NewUbuntuCluster,
		Description: "Ubuntu on OVH",
	},
	"packet": {
		ProfileFunc: packet.NewUbuntuCluster,
		Description: "Ubuntu on Packet x86",
	},
	"packet-ubuntu": {
		ProfileFunc: packet.NewUbuntuCluster,
		Description: "Ubuntu on Packet x86",
	},

	// -----------------------------------------------------------------------------------------------------------------
	//
	// Controller profiles
	//
	// -----------------------------------------------------------------------------------------------------------------

	"controller-aws-ubuntu": {
		ProfileFunc: amazon.NewControllerUbuntuCluster,
		Description: "Controller Ubuntu on Amazon",
	},

	"caws": {
		ProfileFunc: amazon.NewControllerUbuntuCluster,
		Description: "Controller Ubuntu on Amazon",
	},
}
