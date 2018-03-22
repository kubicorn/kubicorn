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

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var targets = []string{}
var output string
var dir string
var dobuild bool
var dofetch bool
var version string

var DefaultTargets = []string{"linux:amd64", "darwin:amd64", "windows:amd64"}

func main() {
	buildCmd.Flags().StringSliceVar(&targets, "targets",
		DefaultTargets, "GOOS:GOARCH pair.  maybe specified multiple times.")
	buildCmd.Flags().StringVar(&dir, "dir", "",
		"if specified, use the build directory instead of creating a tmp directory.")
	buildCmd.Flags().StringVar(&output, "output", "apiserver-builder",
		"value name of the tar file to build")
	buildCmd.Flags().StringVar(&version, "version", "", "version name")

	buildCmd.Flags().BoolVar(&dobuild, "build", true, "if true, build the go packages")
	buildCmd.Flags().BoolVar(&dofetch, "fetch", true, "if true, fetch the go packages")

	cmd.AddCommand(buildCmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(-1)
	}
}

var cmd = &cobra.Command{
	Use:   "apiserver-builder-release",
	Short: "apiserver-builder-release builds a .tar.gz release package",
	Long:  `apiserver-builder-release builds a .tar.gz release package`,
	Run:   RunMain,
}

func RunMain(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build the binaries",
	Long:  `build the binaries`,
	Run:   RunBuild,
}

func RunBuild(cmd *cobra.Command, args []string) {
	if len(version) == 0 {
		fmt.Fprintf(os.Stderr, "must specify the --version flag")
		os.Exit(-1)
	}

	// Create a temporary build directory
	if len(dir) == 0 {
		var err error
		dir, err = ioutil.TempDir(os.TempDir(), "apiserver-builder-release")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create temp directory %s %v\n", dir, err)
			os.Exit(-1)
		}
		fmt.Printf("build directory: %s.  to rerun with cached go fetch use `--dir %s`\n", dir, dir)

		err = os.Mkdir(filepath.Join(dir, "src"), 0700)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create directory %s %v\n", filepath.Join(dir, "src"), err)
			os.Exit(-1)
		}

		err = os.Mkdir(filepath.Join(dir, "bin"), 0700)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create directory %s %v\n", filepath.Join(dir, "bin"), err)
			os.Exit(-1)
		}
	}

	if dofetch {
		for _, pkg := range BuildPackages {
			Fetch(pkg)
		}
	}

	if dobuild {
		if len(targets) == 0 {
			for _, pkg := range BuildPackages {
				Build(filepath.Join("src", pkg, "main.go"),
					filepath.Join("bin", filepath.Base(pkg)),
					"", "",
				)
			}
			PackageTar("", "")
		}
		for _, target := range targets {
			parts := strings.Split(target, ":")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "--targets flags must be GOOS:GOARCH pairs [%s]\n", target)
				os.Exit(-1)
			}
			goos := parts[0]
			goarch := parts[1]
			for _, pkg := range BuildPackages {
				Build(filepath.Join("src", pkg, "main.go"),
					filepath.Join("bin", filepath.Base(pkg)),
					goos, goarch,
				)
			}
			PackageTar(goos, goarch)
		}

	}
}

func RunCmd(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", dir))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	fmt.Printf("%s\n", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(-1)
	}
}

func Build(input, output, goos, goarch string) {
	cmd := exec.Command("go", "build", "-o", output, input)

	// CGO_ENABLED=0 for statically compile binaries
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	if len(goos) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", goos))
	}
	if len(goarch) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", goarch))
	}
	RunCmd(cmd)
}

func Fetch(pkg string) {
	cmd := exec.Command("go", "get", "-d", pkg)
	RunCmd(cmd)
}

var BuildPackages = []string{
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiregister-gen",
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot",
	"github.com/kubernetes-incubator/reference-docs/gen-apidocs",
	"k8s.io/kubernetes/cmd/libs/go2idl/client-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/informer-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/lister-gen",
	"k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen",
}

func PackageTar(goos, goarch string) {
	// create the new file
	fw, err := os.Create(fmt.Sprintf("%s-%s-%s-%s.tar.gz", output, version, goos, goarch))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output file %s %v\n", output, err)
		os.Exit(-1)
	}
	defer fw.Close()

	// setup gzip of tar
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// setup tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add some files to the archive.
	for _, pkg := range BuildPackages {
		name := filepath.Base(pkg)
		path := filepath.Join(dir, "bin", name)
		body, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read file %s %v\n", path, err)
			os.Exit(-1)
		}

		hdr := &tar.Header{
			Name: filepath.Join("apiserver-builder", name),
			Mode: 0500,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output for %s %v\n", path, err)
			os.Exit(-1)
		}
		if _, err := tw.Write(body); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output for %s %v\n", path, err)
			os.Exit(-1)
		}
	}
}
