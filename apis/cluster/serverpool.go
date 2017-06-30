package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServerPoolType_Master     = "master"
	ServerPoolType_Node       = "node"
	ServerPoolType_Hybrid     = "hybrid"
	Cloud_Amazon    = "amazon"
	Cloud_Azure     = "azure"
	Cloud_Google    = "google"
	Cloud_Baremetal = "baremetal"
)

type ServerPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Count             int
	Type              string
	Name              string
	PoolType          string
	Networks          []*Network
}
