package resources

import (
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/digitalocean/godoSdk"
)

var Sdk *godoSdk.Sdk

type Shared struct {
	CloudID        string
	Name           string
	TagResource    cloud.Resource
	Tags           []string
	CachedActual   cloud.Resource
	CachedExpected cloud.Resource
}
