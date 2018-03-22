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
	"path/filepath"

	"github.com/spf13/cobra"
)

var createVersionCmd = &cobra.Command{
	Use:   "create-version",
	Short: "Creates an API version",
	Long:  `Creates an API version`,
	Run:   RunCreateVersion,
}

func AddCreateVersion(cmd *cobra.Command) {
	createVersionCmd.Flags().StringVar(&groupName, "group", "", "name of the API group")
	createVersionCmd.Flags().StringVar(&versionName, "version", "", "name of the API version to create")
	createVersionCmd.Flags().StringVar(&copyright, "copyright", "", "path to copyright file. defaults to boilerplate.go.txt")
	createVersionCmd.Flags().StringVar(&domain, "domain", "", "domain the api group lives under")
	cmd.AddCommand(createVersionCmd)
}

func RunCreateVersion(cmd *cobra.Command, args []string) {
	if len(domain) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-version requires the --domain flag\n")
		os.Exit(-1)
	}
	if len(groupName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-version requires the --group flag\n")
		os.Exit(-1)
	}
	if len(versionName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-version requires the --version flag\n")
		os.Exit(-1)
	}

	cr := getCopyright()
	createVersion(cr)
}

func createVersion(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "pkg", "apis", groupName, versionName, "doc.go")
	created := writeIfNotFound(path, "version-template", versionTemplate, versionTemplateArgs{
		boilerplate,
		domain,
		groupName,
		versionName,
		Repo,
	})
	if !created {
		fmt.Fprintf(os.Stderr, "API group version %s/%s already exists.\n", groupName, versionName)
		os.Exit(-1)
	}
}

type versionTemplateArgs struct {
	BoilerPlate string
	Domain      string
	Group       string
	Version     string
	Repo        string
}

var versionTemplate = `
{{.BoilerPlate}}

// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen={{.Repo}}/pkg/apis/{{.Group}}
// +k8s:defaulter-gen=TypeMeta
// +groupName={{.Group}}.{{.Domain}}
package {{.Version}} // import "{{.Repo}}/pkg/apis/{{.Group}}/{{.Version}}"

`
