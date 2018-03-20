# Getting started

This document covers building your first apiserver from scratch:

- Setup your environment by installing the necessary binaries using `go get`
- Initialize your package
- Create an API group, version, resource
- Build your API server
- Write an automated test for your API server

## Create your Go project

Create a Go project under GOPATH/src/

For example

> GOPATH/src/github.com/my-org/my-project

## Download and install the code generators

Make sure the GOPATH/bin directory is on your path, and then use
 `go get` to download and compile the code generators:

```sh
go get github.com/kubernetes-incubator/apiserver-builder/cmd/apiregister-boot
go get k8s.io/kubernetes/cmd/libs/go2idl/client-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/lister-gen
go get k8s.io/kubernetes/cmd/libs/go2idl/informer-gen
go get github.com/kubernetes-incubator/apiserver-builder/cmd/apiregister-gen
go get github.com/kubernetes-incubator/reference-docs/gen-apidocs
```

Verify the downloaded code generators can be found on the path by running
`apiserver-boot`

## Create your copyright header

Create a file called `boilerplate.go.txt` that contains the copyright
you want to appear at the top of generated files.

e.g.

```go
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
```

## Initialize your package

At the root of your go package under your GOPATH run the following command.

This will setup the initial file structure for your apiserver, including:

- pkg/apis/doc.go
- pkg/openapi/doc.go
- docs/...
- main.go
- glide.yaml

Flags:

- your-domain: unique namespace for your API groups

```sh
apiserver-boot init --domain <your-domain>
```

## Create an API group

An API group contains one or more related API versions.  It is similar to
a package in go or Java.

This will create:

- pkg/apis/your-group/doc.go

Flags:

- your-group: name of the API group e.g. `cicd` or `apps`

```sh
apiserver-boot create-group --domain <your-domain> --group <your-group>
```

This will create a new API group under pkg/apis/<your-group>

## Create an API version

An API version contains one or more APIs.  The version is used
to support introducing changes to APIs without breaking backwards
compatibility.

This will create:

- pkg/apis/your-group/your-version/doc.go

Flags:

- your-version: name of the API version e.g. `v1beta1` or `v1`

```sh
apiserver-boot create-group create-version --domain <your-domain> --group <your-group> --version <your-version>
```

This will create a new API version under pkg/apis/<your-group>/<your-version>

## Create an API resource

An API resource provides REST endpoints for CRUD operations on a resource
type.  This is what will be used by clients to read and store instances
of the resource kind.

This will create:

- pkg/apis/your-group/your-version/your-kind_types.go
- pkg/apis/your-group/your-version/your-kind_types_test.go

Flags:

- your-kind: camelcase name of the type e.g. `MyKind`
- your-resource: lowercase pluralization of the kind e.g. `mykinds`

```sh
apiserver-boot create-resource --domain <your-domain> --group <your-group> --version <your-version> --kind <your-kind> --resource <your-resource>
```

## Fetch the go dependencies

The following command will run `glide install --strip-vendor` so that vendored dependencies work across vendored packages.

This will take a while.

```sh
apiserver-boot glide-install
```

## Generate the code

The following command will generate the wiring to register your API resources.

**Note:** It must be rerun anytype new fields are added to your resources

- api-versions: comma seperated list of the API group/version packages to generate code for

```sh
apiserver-boot generate --api-versions "your-group/your-version"
```

## Build the apiserver

```sh
go build main.go -o apiserver
```

## Run a test

A placehold test was created for your resource.  The test will
start your apiserver in memory, and allow you to create, read, and write
your resource types.

This is a good way to test validation and defaulting of your types.

```sh
go test pkg/apis/your-group/your-version/your-kind_types_test.go
```