package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// MyApp is the Schema for the myapps API
type MyApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Application name
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	AppName string `json:"appName,omitempty"`

	// Number of replicas to deploy
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Replicas int `json:"replicas,omitempty"`

	// Enable debug mode for verbose logging
	Debug bool `json:"debug,omitempty"`

	// # Image configuration
	Image ImageConfig `json:"image,omitempty"`

	// # Service configuration
	Service ServiceConfig `json:"service,omitempty"`

	// # Ingress configuration
	Ingress IngressConfig `json:"ingress,omitempty"`

	// # Resource limits and requests
	Resources ResourcesConfig `json:"resources,omitempty"`

	// # Environment variables
	Env []EnvConfig `json:"env,omitempty"`

	// # Environment variables from ConfigMaps/Secrets
	EnvFrom []EnvFromConfig `json:"envFrom,omitempty"`

	// # Security context
	SecurityContext SecurityContextConfig `json:"securityContext,omitempty"`

	// # Liveness probe configuration
	LivenessProbe LivenessProbeConfig `json:"livenessProbe,omitempty"`

	// # Readiness probe configuration
	ReadinessProbe ReadinessProbeConfig `json:"readinessProbe,omitempty"`

	// # Volume mounts
	VolumeMounts []VolumeMountsConfig `json:"volumeMounts,omitempty"`

	// # Volumes
	Volumes []VolumesConfig `json:"volumes,omitempty"`

	// # Node selector for pod assignment
	// +miaka:type: map[string]string
	NodeSelector NodeSelectorConfig `json:"nodeSelector,omitempty"`

	// # Tolerations for pod assignment
	Tolerations []TolerationsConfig `json:"tolerations,omitempty"`

	// # Affinity rules for pod assignment
	Affinity AffinityConfig `json:"affinity,omitempty"`

	// # Pod annotations
	// +miaka:type: map[string]string
	PodAnnotations PodAnnotationsConfig `json:"podAnnotations,omitempty"`

	// # Pod labels
	// +miaka:type: map[string]string
	PodLabels PodLabelsConfig `json:"podLabels,omitempty"`

	// # ServiceAccount configuration
	ServiceAccount ServiceAccountConfig `json:"serviceAccount,omitempty"`

	// # Autoscaling configuration
	Autoscaling AutoscalingConfig `json:"autoscaling,omitempty"`

	// # Monitoring configuration
	Monitoring MonitoringConfig `json:"monitoring,omitempty"`

	// # Database configuration example
	Database DatabaseConfig `json:"database,omitempty"`

	// # Cache/Redis configuration example
	Cache CacheConfig `json:"cache,omitempty"`
}

// ImageConfig defines the image configuration
type ImageConfig struct {
	// Container image repository
	Repository string `json:"repository,omitempty"`

	// Image pull policy
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	PullPolicy string `json:"pullPolicy,omitempty"`

	// Container image tag (immutable tags are recommended)
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9._-]+$
	Tag string `json:"tag,omitempty"`
}

// AnnotationsConfig defines the annotations configuration
type AnnotationsConfig struct {
	ServiceBetaKubernetesIoAwsLoadBalancerType string `json:"service.beta.kubernetes.io/aws-load-balancer-type,omitempty"`
}

// LabelsConfig defines the labels configuration
type LabelsConfig struct {
	Monitoring string `json:"monitoring,omitempty"`
}

// ServiceConfig defines the service configuration
type ServiceConfig struct {
	// Kubernetes service type
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	Type string `json:"type,omitempty"`

	// Service port
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// Service annotations (e.g., for cloud load balancers)
	// +miaka:type: map[string]string
	Annotations AnnotationsConfig `json:"annotations,omitempty"`

	// Service labels
	// +miaka:type: map[string]string
	Labels LabelsConfig `json:"labels,omitempty"`
}

// IngressConfigAnnotationsConfig defines the ingress config annotations configuration
type IngressConfigAnnotationsConfig struct {
	CertManagerIoClusterIssuer string `json:"cert-manager.io/cluster-issuer,omitempty"`
}

// PathsConfig defines the paths configuration
type PathsConfig struct {
	// Path string
	Path string `json:"path,omitempty"`

	// Path type (Prefix, Exact, or ImplementationSpecific)
	// +kubebuilder:validation:Enum=Prefix;Exact;ImplementationSpecific
	PathType string `json:"pathType,omitempty"`
}

// HostsConfig defines the hosts configuration
type HostsConfig struct {
	// Hostname
	Host string `json:"host,omitempty"`

	// Paths for this host
	Paths []PathsConfig `json:"paths,omitempty"`
}

// TlsConfig defines the tls configuration
type TlsConfig struct {
	// Secret name containing TLS certificate
	SecretName string `json:"secretName,omitempty"`

	// Hosts covered by this certificate
	Hosts []string `json:"hosts,omitempty"`
}

// IngressConfig defines the ingress configuration
type IngressConfig struct {
	// Enable ingress resource
	Enabled bool `json:"enabled,omitempty"`

	// Ingress class name (e.g., nginx, traefik)
	ClassName string `json:"className,omitempty"`

	// Ingress annotations
	// +miaka:type: map[string]string
	Annotations IngressConfigAnnotationsConfig `json:"annotations,omitempty"`

	// Ingress hosts configuration
	Hosts []HostsConfig `json:"hosts,omitempty"`

	// TLS configuration for ingress
	Tls []TlsConfig `json:"tls,omitempty"`
}

// LimitsConfig defines the limits configuration
type LimitsConfig struct {
	// CPU limit
	Cpu string `json:"cpu,omitempty"`

	// Memory limit
	Memory string `json:"memory,omitempty"`
}

// RequestsConfig defines the requests configuration
type RequestsConfig struct {
	// CPU request
	Cpu string `json:"cpu,omitempty"`

	// Memory request
	Memory string `json:"memory,omitempty"`
}

// ResourcesConfig defines the resources configuration
type ResourcesConfig struct {
	// Resource limits
	Limits LimitsConfig `json:"limits,omitempty"`

	// Resource requests
	Requests RequestsConfig `json:"requests,omitempty"`
}

// EnvConfig defines the env configuration
type EnvConfig struct {
	// Variable name
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Variable value
	Value string `json:"value,omitempty"`
}

// ConfigMapRefConfig defines the config map ref configuration
type ConfigMapRefConfig struct {
	// ConfigMap name
	Name string `json:"name,omitempty"`
}

// SecretRefConfig defines the secret ref configuration
type SecretRefConfig struct {
	// Secret name
	Name string `json:"name,omitempty"`
}

// EnvFromConfig defines the env from configuration
type EnvFromConfig struct {
	// ConfigMap reference
	ConfigMapRef ConfigMapRefConfig `json:"configMapRef,omitempty"`

	// Secret reference
	SecretRef SecretRefConfig `json:"secretRef,omitempty"`
}

// CapabilitiesConfig defines the capabilities configuration
type CapabilitiesConfig struct {
	Drop []string `json:"drop,omitempty"`
	Add  []string `json:"add,omitempty"`
}

// SecurityContextConfig defines the security context configuration
type SecurityContextConfig struct {
	// Run as non-root user
	RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`

	// User ID to run as
	// +kubebuilder:validation:Minimum=1
	RunAsUser int `json:"runAsUser,omitempty"`

	// Group ID to run as
	// +kubebuilder:validation:Minimum=1
	RunAsGroup int `json:"runAsGroup,omitempty"`

	// FSGroup for volume ownership
	// +kubebuilder:validation:Minimum=1
	FsGroup int `json:"fsGroup,omitempty"`

	// Drop all capabilities and add only required ones
	Capabilities CapabilitiesConfig `json:"capabilities,omitempty"`
}

// HttpGetConfig defines the http get configuration
type HttpGetConfig struct {
	// Path to probe
	Path string `json:"path,omitempty"`

	// Port to probe
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`
}

// LivenessProbeConfig defines the liveness probe configuration
type LivenessProbeConfig struct {
	// HTTP GET liveness probe
	HttpGet HttpGetConfig `json:"httpGet,omitempty"`

	// Initial delay before liveness probe
	// +kubebuilder:validation:Minimum=0
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// Period between liveness probes
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// Timeout for liveness probe
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// Success threshold for liveness probe
	// +kubebuilder:validation:Minimum=1
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// Failure threshold for liveness probe
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int `json:"failureThreshold,omitempty"`
}

// ReadinessProbeConfigHttpGetConfig defines the readiness probe config http get configuration
type ReadinessProbeConfigHttpGetConfig struct {
	// Path to probe
	Path string `json:"path,omitempty"`

	// Port to probe
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`
}

// ReadinessProbeConfig defines the readiness probe configuration
type ReadinessProbeConfig struct {
	// HTTP GET readiness probe
	HttpGet ReadinessProbeConfigHttpGetConfig `json:"httpGet,omitempty"`

	// Initial delay before readiness probe
	// +kubebuilder:validation:Minimum=0
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// Period between readiness probes
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// Timeout for readiness probe
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// Success threshold for readiness probe
	// +kubebuilder:validation:Minimum=1
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// Failure threshold for readiness probe
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int `json:"failureThreshold,omitempty"`
}

// VolumeMountsConfig defines the volume mounts configuration
type VolumeMountsConfig struct {
	// Mount name
	Name string `json:"name,omitempty"`

	// Mount path in container
	MountPath string `json:"mountPath,omitempty"`

	// Mount as read-only
	ReadOnly bool `json:"readOnly,omitempty"`
}

// ConfigMapConfig defines the config map configuration
type ConfigMapConfig struct {
	// ConfigMap name
	Name string `json:"name,omitempty"`
}

// VolumesConfig defines the volumes configuration
type VolumesConfig struct {
	// Volume name
	Name string `json:"name,omitempty"`

	// ConfigMap volume source
	ConfigMap ConfigMapConfig `json:"configMap,omitempty"`
}

// NodeSelectorConfig defines the node selector configuration
type NodeSelectorConfig struct {
	KubernetesIoArch string `json:"kubernetes.io/arch,omitempty"`
	NodeRole         string `json:"node-role,omitempty"`
}

// TolerationsConfig defines the tolerations configuration
type TolerationsConfig struct {
	// Toleration key
	Key string `json:"key,omitempty"`

	// Toleration operator
	// +kubebuilder:validation:Enum=Exists;Equal
	Operator string `json:"operator,omitempty"`

	// Toleration effect
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect,omitempty"`

	// Toleration duration in seconds
	// +kubebuilder:validation:Minimum=0
	TolerationSeconds int `json:"tolerationSeconds,omitempty"`
}

// MatchExpressionsConfig defines the match expressions configuration
type MatchExpressionsConfig struct {
	// Label key
	Key string `json:"key,omitempty"`

	// Match operator
	// +kubebuilder:validation:Enum=In;NotIn;Exists;DoesNotExist;Gt;Lt
	Operator string `json:"operator,omitempty"`

	// Values to match
	Values []string `json:"values,omitempty"`
}

// NodeSelectorTermsConfig defines the node selector terms configuration
type NodeSelectorTermsConfig struct {
	// Match expressions for node selection
	MatchExpressions []MatchExpressionsConfig `json:"matchExpressions,omitempty"`
}

// RequiredDuringSchedulingIgnoredDuringExecutionConfig defines the required during scheduling ignored during execution configuration
type RequiredDuringSchedulingIgnoredDuringExecutionConfig struct {
	// Node selector terms
	NodeSelectorTerms []NodeSelectorTermsConfig `json:"nodeSelectorTerms,omitempty"`
}

// NodeAffinityConfig defines the node affinity configuration
type NodeAffinityConfig struct {
	// Required node affinity
	RequiredDuringSchedulingIgnoredDuringExecution RequiredDuringSchedulingIgnoredDuringExecutionConfig `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// AffinityConfig defines the affinity configuration
type AffinityConfig struct {
	// Node affinity
	NodeAffinity NodeAffinityConfig `json:"nodeAffinity,omitempty"`
}

// PodAnnotationsConfig defines the pod annotations configuration
type PodAnnotationsConfig struct {
	PrometheusIoScrape string `json:"prometheus.io/scrape,omitempty"`
	PrometheusIoPort   string `json:"prometheus.io/port,omitempty"`
	PrometheusIoPath   string `json:"prometheus.io/path,omitempty"`
}

// PodLabelsConfig defines the pod labels configuration
type PodLabelsConfig struct {
	App     string `json:"app,omitempty"`
	Version string `json:"version,omitempty"`
}

// ServiceAccountConfigAnnotationsConfig defines the service account config annotations configuration
type ServiceAccountConfigAnnotationsConfig struct {
	EksAmazonawsComRoleArn string `json:"eks.amazonaws.com/role-arn,omitempty"`
}

// ServiceAccountConfig defines the service account configuration
type ServiceAccountConfig struct {
	// Create a service account
	Create bool `json:"create,omitempty"`

	// Service account name (leave empty to generate)
	Name string `json:"name,omitempty"`

	// Service account annotations
	// +miaka:type: map[string]string
	Annotations ServiceAccountConfigAnnotationsConfig `json:"annotations,omitempty"`
}

// AutoscalingConfig defines the autoscaling configuration
type AutoscalingConfig struct {
	// Enable horizontal pod autoscaling
	Enabled bool `json:"enabled,omitempty"`

	// Minimum number of replicas
	// +kubebuilder:validation:Minimum=1
	MinReplicas int `json:"minReplicas,omitempty"`

	// Maximum number of replicas
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int `json:"maxReplicas,omitempty"`

	// Target CPU utilization percentage
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	TargetCPUUtilizationPercentage int `json:"targetCPUUtilizationPercentage,omitempty"`

	// Target memory utilization percentage
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	TargetMemoryUtilizationPercentage int `json:"targetMemoryUtilizationPercentage,omitempty"`
}

// ServiceMonitorConfigLabelsConfig defines the service monitor config labels configuration
type ServiceMonitorConfigLabelsConfig struct {
	Prometheus string `json:"prometheus,omitempty"`
}

// ServiceMonitorConfig defines the service monitor configuration
type ServiceMonitorConfig struct {
	// Enable ServiceMonitor resource
	Enabled bool `json:"enabled,omitempty"`

	// ServiceMonitor interval
	Interval string `json:"interval,omitempty"`

	// ServiceMonitor labels
	// +miaka:type: map[string]string
	Labels ServiceMonitorConfigLabelsConfig `json:"labels,omitempty"`
}

// MonitoringConfig defines the monitoring configuration
type MonitoringConfig struct {
	// Enable Prometheus monitoring
	Enabled bool `json:"enabled,omitempty"`

	// ServiceMonitor configuration
	ServiceMonitor ServiceMonitorConfig `json:"serviceMonitor,omitempty"`
}

// DatabaseConfig defines the database configuration
type DatabaseConfig struct {
	// Database host
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host,omitempty"`

	// Database port
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// Database name
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Connection pool size
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	PoolSize int `json:"poolSize,omitempty"`

	// SSL mode for database connection
	// +kubebuilder:validation:Enum=disable;allow;prefer;require;verify-ca;verify-full
	SslMode string `json:"sslMode,omitempty"`
}

// CacheConfig defines the cache configuration
type CacheConfig struct {
	// Enable Redis cache
	Enabled bool `json:"enabled,omitempty"`

	// Redis host
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host,omitempty"`

	// Redis port
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// Redis database number
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=15
	Database int `json:"database,omitempty"`

	// Cache TTL in seconds
	// +kubebuilder:validation:Minimum=0
	Ttl int `json:"ttl,omitempty"`
}
