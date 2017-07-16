package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Cloud             string
	Name              string
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Name: name,
	}
}
