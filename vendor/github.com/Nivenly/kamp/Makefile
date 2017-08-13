PKGS=$(shell go list ./... | grep -v /vendor)
SHELL_IMAGE=golang:1.8.3
GIT_SHA=$(shell git rev-parse --verify HEAD)
VERSION=$(shell cat VERSION)

default: compile
compile:
	@go build -o ${GOPATH}/bin/kamp -ldflags "-X github.com/Nivenly/kamp/cmd.GitSha=${GIT_SHA} -X github.com/Nivenly/kamp/cmd.Version=${VERSION}" main.go

build: clean build-linux-amd64 build-darwin-amd64 build-freebsd-amd64 build-windows-amd64

clean:
	rm -rf bin/*

gofmt:
	gofmt -w ./cmd

# Because of https://github.com/golang/go/issues/6376 We actually have to build this in a container
build-linux-amd64:
	docker run \
	-it \
	-w /go/src/github.com/Nivenly/kamp \
	-v ${GOPATH}/src/github.com/kris-nova/klone:/go/src/github.com/Nivenly/kamp \
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
	-w /go/src/github.com/Nivenly/kamp \
	-v ${GOPATH}/src/github.com/Nivenly/kamp:/go/src/github.com/Nivenly/kamp \
	--rm ${SHELL_IMAGE} /bin/bash

test:
	@go test $(PKGS)

push:
	docker build -f kiaora/Dockerfile.kiaora -t nivenly/kiaora:latest .
	docker push nivenly/kiaora:latest
