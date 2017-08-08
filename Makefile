ifndef VERBOSE
	MAKEFLAGS += --silent
endif

PKGS=$(shell go list ./... | grep -v /vendor)
CI_PKGS=$(shell go list ./... | grep -v /vendor | grep -v test)
FMT_PKGS=$(shell go list -f {{.Dir}} ./... | grep -v vendor | grep -v test | tail -n +2)
SHELL_IMAGE=golang:1.8.3
GIT_SHA=$(shell git rev-parse --verify HEAD)
VERSION=$(shell cat VERSION)
PWD=$(shell pwd)

GOIMPORTS := $(shell command -v goimports 2> /dev/null)

default: authorsfile bindata compile

all: default install

compile:
	go build -o bin/kubicorn -ldflags "-X github.com/kris-nova/kubicorn/cmd.GitSha=${GIT_SHA} -X github.com/kris-nova/kubicorn/cmd.Version=${VERSION}" main.go

install:
	install -m 0755 bin/kubicorn ${GOPATH}/bin/kubicorn

bindata:
	which go-bindata > /dev/null || go get -u github.com/jteeuwen/go-bindata/...
	rm -rf bootstrap/bootstrap.go
	go-bindata -pkg bootstrap -o bootstrap/bootstrap.go bootstrap/ bootstrap/vpn

build: authors clean build-linux-amd64 build-darwin-amd64 build-freebsd-amd64 build-windows-amd64

authorsfile:
	git log --all --format='%aN <%cE>' | sort -u | egrep -v "noreply|mailchimp|@Kris" > AUTHORS

clean:
	rm -rf bin/*
	rm -rf bootstrap/bootstrap.go

gofmt:
ifndef GOIMPORTS
	echo "Installing goimports..."
	go get golang.org/x/tools/cmd/goimports
endif
	echo "Fixing format of go files..."; \
	for package in $(FMT_PKGS); \
	do \
		gofmt -w $$package ; \
		goimports -l -w $$package ; \
	done

# Because of https://github.com/golang/go/issues/6376 We actually have to build this in a container
build-linux-amd64:
	mkdir -p bin
	docker run \
	-it \
	-w /go/src/github.com/kris-nova/kubicorn \
	-v ${PWD}:/go/src/github.com/kris-nova/kubicorn \
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
	-v ${PWD}:/go/src/github.com/kris-nova/kubicorn \
	--rm ${SHELL_IMAGE} /bin/bash

lint:
	which golint > /dev/null || go get -u github.com/golang/lint/golint
	golint $(PKGS)


.PHONY: test
test:
	go test -timeout 20m -v $(PKGS)


.PHONY: ci
ci:
	go test -timeout 20m -v $(CI_PKGS)


vet:
	@go vet $(PKGS)

check-header:
	./scripts/check-header.sh

headers:
	./scripts/headers.sh

.PHONY: apimachinery
apimachinery:
	go get k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
