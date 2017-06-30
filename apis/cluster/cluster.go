package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Name              string
	ServerPools       []*ServerPool
	Cloud             string
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Name: name,
	}
}
