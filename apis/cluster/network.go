package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NetworkType_Local   = "local"
	NetworkType_Public  = "public"
	NetworkType_Private = "private"
)

type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	CIDR              string
	Identifier        string
	Type              string
}
