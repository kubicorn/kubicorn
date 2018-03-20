# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# NOTE: Project must live under $GOPATH/src/github.com/kubernetes-incubator/apiserver-builder/example for the code to compile
REPO=github.com/kubernetes-incubator/apiserver-builder/example
BOILER_PLATE="boilerplate.go.txt"
UNVERSIONED="miskatonic/,innsmouth/"
VERSIONED="miskatonic/v1beta1,innsmouth/v1"

REGISTER_GEN="apiregister-gen"
REGISTER_GEN="../cmd/apiregister-gen/main.go"

SRC=$(REPO)/pkg/apis/...
DIR=pkg

GENERIC_API="k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource/,k8s.io/apimachinery/pkg/version/,k8s.io/apimachinery/pkg/runtime/,k8s.io/apimachinery/pkg//util/intstr/"
OUT=$(GOPATH)/src
CLIENT_PKG=$(REPO)/pkg/client
CLIENT_PATH=$(CLIENT_PKG)/clientset_generated
CLIENT=clientset
INTERNAL_CLIENT=internalclientset
LISTERS_PKG="$(CLIENT_PKG)/listers_generated"
INFORMERS_PKG="$(CLIENT_PKG)/informers_generated"

all: build

# Download all of the code generators
# TODO: Package and distribute these in a release tar
deps:
	go get k8s.io/kubernetes/cmd/libs/go2idl/client-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/lister-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/informer-gen
	go get github.com/kubernetes-incubator/apiserver-builder/cmd/apiregister-gen
	go get github.com/kubernetes-incubator/reference-docs/gen-apidocs

# Build the main
build: generate
	go build main.go

# Run code generators
generate: cleangenerated
	go run $(REGISTER_GEN) -i $(REPO)/pkg/apis/...
	conversion-gen -i "$(SRC)"  --extra-peer-dirs="k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime" -o $(OUT) -O zz_generated.conversion --go-header-file $(BOILER_PLATE)
	deepcopy-gen -i "$(SRC)" -o $(OUT) -O zz_generated.deepcopy --go-header-file $(BOILER_PLATE)
	openapi-gen  -i "$(SRC),$(GENERIC_API)" --output-package "$(REPO)/pkg/openapi" --go-header-file $(BOILER_PLATE)
	defaulter-gen -i "$(SRC)" -o $(OUT) -O "zz_generated.defaults" --go-header-file $(BOILER_PLATE) --extra-peer-dirs="k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime"
	client-gen --input-base "$(REPO)/pkg/apis/" --input $(VERSIONED) --clientset-path=$(CLIENT_PATH)  --clientset-name=$(CLIENT) -o $(OUT) --go-header-file $(BOILER_PLATE)
	client-gen --input-base "$(REPO)/pkg/apis/" --input $(UNVERSIONED) --clientset-path=$(CLIENT_PATH) --clientset-name=$(INTERNAL_CLIENT)  -o $(OUT) --go-header-file $(BOILER_PLATE)
	lister-gen -i $(SRC) --output-package=$(LISTERS_PKG) -o $(OUT) --go-header-file $(BOILER_PLATE)
	informer-gen -i $(SRC) -o $(OUT) --go-header-file $(BOILER_PLATE) --internal-clientset-package="$(CLIENT_PATH)/$(INTERNAL_CLIENT)" --versioned-clientset-package="$(CLIENT_PATH)/$(CLIENT)" --listers-package=$(LISTERS_PKG) --output-package=$(INFORMERS_PKG)

# Run integration tests
test: build
	go test pkg/apis/miskatonic/v1beta1/university_types_test.go
	go test pkg/apis/innsmouth/v1/deepone_types_test.go
	bash -c "find $(DIR)/apis/ -name apiserver.local.config | xargs rm -rf"

# Clean up functions

clean: cleanbin cleangenerated cleandocs

cleanbin:
	rm -f main

cleangenerated:
	bash -c "find $(DIR)/apis/ -name zz_generated.api.*.go | xargs rm -f"
	bash -c "find $(DIR)/apis/ -name zz_generated.deepcopy.go | xargs rm -f"
	bash -c "find $(DIR)/apis/ -name zz_generated.defaults.go | xargs rm -f"
	bash -c "find $(DIR)/apis/ -name zz_generated.conversion.go | xargs rm -f"
	bash -c "find . -name apiserver.local.config | xargs rm -rf"
	rm -f $(DIR)/openapi/openapi_generated.go
	rm -rf $(DIR)/client/clientset_generated/
	rm -rf $(DIR)/client/informers_generated/

cleandocs:
	rm -rf $(shell pwd)/docs/build
	rm -rf $(shell pwd)/docs/includes
	rm -rf $(shell pwd)/docs/manifest.json
	rm -rf $(shell pwd)/docs/includes/_generated_*

docs: cleandocs build
	bash -c "mkdir -p docs/openapi-spec; mkdir -p docs/static_includes; echo ''"
	./main --delegated-auth=false --etcd-servers=http://localhost:2379 --secure-port=9443 --print-openapi > ./docs/openapi-spec/swagger.json
	gen-apidocs --allow-errors --use-tags --config-dir docs
	docker run -v $(shell pwd)/docs/includes:/source -v $(shell pwd)/docs/build:/build -v $(shell pwd)/docs/:/manifest pwittrock/brodocs

run: build
	./main -v 10 --authentication-kubeconfig ~/.kube/auth_config --authorization-kubeconfig ~/.kube/auth_config --client-ca-file /var/run/kubernetes/apiserver.crt  --requestheader-client-ca-file /var/run/kubernetes/apiserver.crt --requestheader-username-headers=X-Remote-User --requestheader-group-headers=X-Remote-Group --requestheader-extra-headers-prefix=X-Remote-Extra- --etcd-servers=http://localhost:2379 --secure-port=9443 --tls-ca-file  /var/run/kubernetes/apiserver.crt --print-bearer-token
