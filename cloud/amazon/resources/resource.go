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
	CachedActual   cloud.Resource
	CachedExpected cloud.Resource
}

func S(format string, a ...interface{}) *string {
	str := fmt.Sprintf(format, a...)
	return &str
}

func I64(i int) *int64 {
	i64 := int64(i)
	return &i64
}

func B(b bool) *bool {
	return &b
}

var Sdk *awsSdkGo.Sdk
