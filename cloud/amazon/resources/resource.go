package resources

import (
	"fmt"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
)

type Shared struct {
	CloudID        string
	Name           string
	TagResource    cloud.Resource
	Tags           map[string]string
	CachedActual   *Vpc
	CachedExpected *Vpc
}

func S(format string, a ...interface{}) *string {
	str := fmt.Sprintf(format, a...)
	return &str
}

var Sdk *awsSdkGo.Sdk
