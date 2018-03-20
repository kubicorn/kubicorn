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

var createGroupCmd = &cobra.Command{
	Use:   "create-group",
	Short: "Creates an API group",
	Long:  `Creates an API group`,
	Run:   RunCreateGroup,
}

func AddCreateGroup(cmd *cobra.Command) {
	createGroupCmd.Flags().StringVar(&groupName, "group", "", "name of the API group to create")
	createGroupCmd.Flags().StringVar(&copyright, "copyright", "boilerplate.go.txt", "path to copyright file. defaults to boilerplate.go.txt")
	createGroupCmd.Flags().StringVar(&domain, "domain", "", "domain the api group lives under")
	cmd.AddCommand(createGroupCmd)
}

func RunCreateGroup(cmd *cobra.Command, args []string) {
	if len(domain) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-group requires the --domain flag\n")
		os.Exit(-1)
	}
	if len(groupName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-group requires the --groupName flag\n")
		os.Exit(-1)
	}
	cr := getCopyright()
	createGroup(cr)
}

func createGroup(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "pkg", "apis", groupName, "doc.go")
	created := writeIfNotFound(path, "group-template", groupTemplate, groupTemplateArgs{
		boilerplate,
		domain,
		groupName,
	})
	if !created {
		fmt.Fprintf(os.Stderr, "API group %s already exists.\n", groupName)
		os.Exit(-1)
	}
}

type groupTemplateArgs struct {
	BoilerPlate string
	Domain      string
	Name        string
}

var groupTemplate = `
{{.BoilerPlate}}


// +k8s:deepcopy-gen=package,register
// +groupName={{.Name}}.{{.Domain}}

// Package api is the internal version of the API.
package {{.Name}}

`
