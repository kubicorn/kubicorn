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
	"strings"

	"github.com/spf13/cobra"
)

var createResourceCmd = &cobra.Command{
	Use:   "create-resource",
	Short: "Creates an API resource",
	Long:  `Creates an API resource`,
	Run:   RunCreateResource,
}

func AddCreateResource(cmd *cobra.Command) {
	createResourceCmd.Flags().StringVar(&groupName, "group", "", "name of the API group")
	createResourceCmd.Flags().StringVar(&versionName, "version", "", "name of the API version")
	createResourceCmd.Flags().StringVar(&kindName, "kind", "", "name of the API kind to create")
	createResourceCmd.Flags().StringVar(&resourceName, "resource", "", "name of the API resource to create, plural name of the kind")
	createResourceCmd.Flags().StringVar(&copyright, "copyright", "", "path to copyright file.  defaults to boilerplate.go.txt")
	createResourceCmd.Flags().StringVar(&domain, "domain", "", "domain the api group lives under")
	cmd.AddCommand(createResourceCmd)
}

func RunCreateResource(cmd *cobra.Command, args []string) {
	if len(domain) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-resource requires the --domain flag\n")
		os.Exit(-1)
	}
	if len(groupName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-resource requires the --group flag\n")
		os.Exit(-1)
	}
	if len(versionName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-resource requires the --version flag\n")
		os.Exit(-1)
	}
	if len(kindName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-resource requires the --kind flag\n")
		os.Exit(-1)
	}
	if len(resourceName) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot create-resource requires the --resource flag\n")
		os.Exit(-1)
	}

	cr := getCopyright()
	createResource(cr)
}

func createResource(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	typesFileName := fmt.Sprintf("%s_types.go", strings.ToLower(kindName))
	path := filepath.Join(dir, "pkg", "apis", groupName, versionName, typesFileName)
	a := resourceTemplateArgs{
		boilerplate,
		domain,
		groupName,
		versionName,
		kindName,
		resourceName,
		Repo,
	}
	created := writeIfNotFound(path, "resource-template", resourceTemplate, a)
	if !created {
		fmt.Fprintf(os.Stderr,
			"API group version kind %s/%s/%s already exists.\n", groupName, versionName, kindName)
		os.Exit(-1)
	}

	typesFileName = fmt.Sprintf("%s_types_test.go", strings.ToLower(kindName))
	path = filepath.Join(dir, "pkg", "apis", groupName, versionName, typesFileName)
	created = writeIfNotFound(path, "resource-test-template", resourceTestTemplate, a)
	if !created {
		fmt.Fprintf(os.Stderr,
			"API group version kind %s/%s/%s already exists.\n", groupName, versionName, kindName)
		os.Exit(-1)
	}
}

type resourceTemplateArgs struct {
	BoilerPlate string
	Domain      string
	Group       string
	Version     string
	Kind        string
	Resource    string
	Repo        string
}

var resourceTemplate = `
{{.BoilerPlate}}

package {{.Version}}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient=true

// +k8s:openapi-gen=true
// +resource={{.Resource}}
// {{.Kind}}
type {{.Kind}} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	Spec   {{.Kind}}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{.Kind}}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// {{.Kind}}Spec defines the desired state of {{.Kind}}
type {{.Kind}}Spec struct {
}

// {{.Kind}}Status defines the observed state of {{.Kind}}
type {{.Kind}}Status struct {
}

`

var resourceTestTemplate = `
{{.BoilerPlate}}

package {{.Version}}

import (
	"os"
	"testing"

	"k8s.io/client-go/rest"
	"github.com/kubernetes-incubator/apiserver-builder/pkg/test"

	"{{.Repo}}/pkg/apis"
	"{{.Repo}}/pkg/client/clientset_generated/clientset"
	"{{.Repo}}/pkg/openapi"
)

var testenv *test.TestEnvironment
var config *rest.Config
var client *clientset.Clientset

// Do Test Suite setup / teardown
func TestMain(m *testing.M) {
	testenv = test.NewTestEnvironment()
	config = testenv.Start(apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
	client = clientset.NewForConfigOrDie(config)
	retCode := m.Run()
	testenv.Stop()
	os.Exit(retCode)
}

func TestCreateDelete{{.Kind}}(t *testing.T) {
}
`
