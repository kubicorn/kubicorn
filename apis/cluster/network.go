package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
)

const (
	NetworkType_Local   = "local"
	NetworkType_Public  = "public"
	NetworkType_Private = "private"
)

type NetworkCidr struct {
	String string
	IPNet  *net.IPNet
}

type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	NetworkType       string
	NetworkCidr       *NetworkCidr
}
