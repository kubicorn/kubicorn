package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Firewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Identifier        string
	Rules             []*Rule
	Name              string
}

type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Identifier        string
	IngressFromPort   int
	IngressToPort     int
	IngressSource     string
	IngressProtocol   string
}
