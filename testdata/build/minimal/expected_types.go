package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// Demo is the Schema for the demos API
type Demo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Port number
	// +kubebuilder:default=8080
	Port int `json:"port,omitempty"`

	// Enable feature
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
}
