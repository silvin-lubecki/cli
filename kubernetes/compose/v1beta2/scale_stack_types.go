package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec map[string]int `json:"spec,omitempty"`
	Status map[string]int `json:"status,omitempty"`
}