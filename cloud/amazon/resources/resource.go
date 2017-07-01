package resources

import (
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
)

type Resource struct {
	Type   string
	Label  string
	Name   string
	AwsSdk *awsSdkGo.Sdk
	Known  *cluster.Cluster
}

func S(format string, a ...interface{}) *string {
	str := fmt.Sprintf(format, a...)
	return &str
}
