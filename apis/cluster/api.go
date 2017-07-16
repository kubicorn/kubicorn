package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesApi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Endpoint          string
	Port              string
}
