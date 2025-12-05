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
	// +kubebuilder:default=3
	Replicas int `json:"replicas,omitempty"`

	// Application name
	// +kubebuilder:default="myapp"
	AppName string `json:"appName,omitempty"`

	// Enable debug mode
	// +kubebuilder:default=true
	Debug bool `json:"debug,omitempty"`

	// Service configuration
	Service ServiceConfig `json:"service,omitempty"`

	// List of environment variables
	Env []EnvConfig `json:"env,omitempty"`
}

// ServiceConfig defines the service configuration
type ServiceConfig struct {
	// Service port
	// +kubebuilder:default=8080
	Port int `json:"port,omitempty"`

	// Service type
	// +kubebuilder:default="ClusterIP"
	ServiceType string `json:"serviceType,omitempty"`
}

// EnvConfig defines the environment variable configuration
type EnvConfig struct {
	// Variable name
	// +kubebuilder:default="LOG_LEVEL"
	Name string `json:"name,omitempty"`

	// Variable value
	// +kubebuilder:default="info"
	Value string `json:"value,omitempty"`
}
