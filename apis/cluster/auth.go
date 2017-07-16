package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Ssh struct {
	metav1.TypeMeta      `json:",inline"`
	metav1.ObjectMeta    `json:"metadata,omitempty"`
	Name                 string
	User                 string
	Identifier           string
	PublicKeyPath        string
	PublicKeyData        []byte
	PublicKeyFingerprint string
}
