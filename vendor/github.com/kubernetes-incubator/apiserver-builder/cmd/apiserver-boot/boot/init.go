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
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a directory",
	Long:  `Initialize a directory`,
	Run:   RunInit,
}

func AddInit(cmd *cobra.Command) {
	cmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&domain, "domain", "", "domain the api groups live under")
	initCmd.Flags().StringVar(&copyright, "copyright", "", "path to copyright file.  defaults to boilerplate.go.txt")
}

func RunInit(cmd *cobra.Command, args []string) {
	if len(domain) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot init requires the --domain flag\n")
		os.Exit(-1)
	}
	cr := getCopyright()

	createGlide()
	createMain(cr)
	createAPIs(cr)
	createOpenAPI(cr)
	createDocs()
}

type glideTemplateArguments struct {
	Repo string
}

var glideTemplate = `
package: {{.Repo}}
import:
- package: github.com/go-openapi/spec
- package: github.com/go-openapi/loads
- package: github.com/golang/glog
- package: github.com/pkg/errors
- package: github.com/spf13/cobra
- package: github.com/spf13/pflag
  version: d90f37a48761fe767528f31db1955e4f795d652f
- package: k8s.io/apimachinery
- package: k8s.io/apiserver
- package: k8s.io/client-go
- package: k8s.io/gengo
- package: k8s.io/kubernetes
  subpackages:
  - pkg/api
- package: k8s.io/apimachinery
  subpackages:
  - pkg/apis/meta/v1
  - pkg/apis/meta
ignore:
- {{.Repo}}
`

func createGlide() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "glide.yaml")
	writeIfNotFound(path, "glide-template", glideTemplate, glideTemplateArguments{Repo})
}

type mainTemplateArguments struct {
	BoilerPlate string
	Repo        string
}

var mainTemplate = `
{{.BoilerPlate}}

package main

import (
	// Make sure glide gets these dependencies
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/kubernetes/pkg/api"
	_ "github.com/go-openapi/loads"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/cmd/server"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Enable cloud provider auth

	"{{.Repo}}/pkg/apis"
	"{{.Repo}}/pkg/openapi"
)

func main() {
	server.StartApiServer("/registry/sample.kubernetes.io", apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
}
`

func createMain(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "main.go")
	writeIfNotFound(path, "main-template", mainTemplate, mainTemplateArguments{boilerplate, Repo})

}

func createAPIs(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "pkg", "apis", "doc.go")
	writeIfNotFound(path, "apis-template", apisDocTemplate, apisDocTemplateArguments{boilerplate, domain})
}

type apisDocTemplateArguments struct {
	BoilerPlate string
	Domain      string
}

var apisDocTemplate = `
{{.BoilerPlate}}


//
// +domain={{.Domain}}

package apis

`

func createOpenAPI(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "pkg", "openapi", "doc.go")
	writeIfNotFound(path, "openapi-template", openAPIDoc, openAPITemplateArguments{boilerplate})
}

type openAPITemplateArguments struct {
	BoilerPlate string
}

var openAPIDoc = `
{{.BoilerPlate}}


// Package openapi exists to hold generated openapi code
package openapi

`

func createDocs() {
	exec.Command("mkdir", "-p", filepath.Join("docs", "openapi-spec")).CombinedOutput()
	exec.Command("mkdir", "-p", filepath.Join("docs", "static_includes")).CombinedOutput()
	exec.Command("mkdir", "-p", filepath.Join("docs", "examples")).CombinedOutput()
}
