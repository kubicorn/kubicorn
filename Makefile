build:
	@go install .

.PHONY: apimachinery
apimachinery:
	go get k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
	go get k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/conversion-gen --skip-unsafe=true --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.conversion
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
	${GOPATH}/bin/defaulter-gen --input-dirs github.com/kris-nova/kubicorn/apis/cluster/v1alpha1 --v=0  --output-file-base=zz_generated.defaults
