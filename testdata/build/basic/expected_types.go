package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// Example is the Schema for the examples API
type Example struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Number of replicas
	// +kubebuilder:validation:Minimum=1
	Replicas int `json:"replicas,omitempty"`

	// Application name
	AppName string `json:"appName,omitempty"`

	// Enable debug mode
	Debug bool `json:"debug,omitempty"`

	// Service configuration
	Service ServiceConfig `json:"service,omitempty"`

	// List of environment variables
	Env []EnvConfig `json:"env,omitempty"`
}

// ServiceConfig defines the service configuration
type ServiceConfig struct {
	// Service port
	Port int `json:"port,omitempty"`

	// Service type
	ServiceType string `json:"serviceType,omitempty"`
}

// EnvConfig defines the environment variable configuration
type EnvConfig struct {
	// Variable name
	Name string `json:"name,omitempty"`

	// Variable value
	Value string `json:"value,omitempty"`
}
