package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Values struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	ItemMap             map[string]string
}


