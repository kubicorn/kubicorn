package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Identifier        string
	CIDR              string
	Location          string
	Zone              string
	Name              string
}
