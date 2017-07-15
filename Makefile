PKGS=$(shell go list ./... | grep -v /vendor)
SHELL_IMAGE=golang:1.8.3
GIT_SHA=$(shell git rev-parse --verify HEAD)
VERSION=$(shell cat VERSION)

default: bindata compile
compile:
	@go build -o ${GOPATH}/bin/kubicorn -ldflags "-X github.com/kris-nova/kubicorn/cmd.GitSha=${GIT_SHA} -X github.com/kris-nova/kubicorn/cmd.Version=${VERSION}" main.go

bindata:
	rm -rf bootstrap/bootstrap.go
	go-bindata -pkg bootstrap -o bootstrap/bootstrap.go bootstrap/


build: clean build-linux-amd64 build-darwin-amd64 build-freebsd-amd64 build-windows-amd64

clean:
	rm -rf bin/*

gofmt:
	gofmt -w ./cmd

# Because of https://github.com/golang/go/issues/6376 We actually have to build this in a container
build-linux-amd64:
	docker run \
	-it \
	-w /go/src/github.com/kris-nova/kubicorn \
	-v ${GOPATH}/src/github.com/kris-nova/klone:/go/src/github.com/kris-nova/kubicorn \
	-e GOPATH=/go \
	--rm golang:1.8.1 make docker-build-linux-amd64

docker-build-linux-amd64:
	go build -v -o bin/linux-amd64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -v -o bin/darwin-amd64 &

build-freebsd-amd64:
	GOOS=freebsd GOARCH=amd64 go build -v -o bin/freebsd-amd64 &

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -v -o bin/windows-amd64 &

linux: shell
shell:
	docker run \
	-i -t \
	-w /go/src/github.com/kris-nova/kubicorn \
	-v ${GOPATH}/src/github.com/Nivenly/kubicorn:/go/src/github.com/kris-nova/kubicorn \
	--rm ${SHELL_IMAGE} /bin/bash

test:
	@go test $(PKGS)

.PHONY: apimachinery
apimachinery:
	go get k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
