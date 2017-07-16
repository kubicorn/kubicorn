package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Cloud_Amazon    = "amazon"
	Cloud_Azure     = "azure"
	Cloud_Google    = "google"
	Cloud_Baremetal = "baremetal"
)

type Cluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Name              string
	ServerPools       []*ServerPool
	Cloud             string
	Location          string
	Ssh               *Ssh
	Network           *Network
	Values            *Values
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Name: name,
	}
}
