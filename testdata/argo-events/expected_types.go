package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//
// EventsChart is the Schema for the eventscharts API
type EventsChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// -- Provide a name in place of `argo-events`
	NameOverride string `json:"nameOverride,omitempty"`

	// -- String to fully override "argo-events.fullname" template
	FullnameOverride string `json:"fullnameOverride,omitempty"`

	// -- Override the namespace
	// @default -- `.Release.Namespace`
	NamespaceOverride string `json:"namespaceOverride,omitempty"`

	// -- Deploy on OpenShift
	Openshift bool `json:"openshift,omitempty"`

	// -- Create clusterroles that extend existing clusterroles to interact with argo-events crds
	// Only applies for cluster-wide installation (`controller.rbac.namespaced: false`)
	// # Ref: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles
	CreateAggregateRoles bool `json:"createAggregateRoles,omitempty"`

	// # Custom resource configuration
	Crds   CrdsConfig   `json:"crds,omitempty"`
	Global GlobalConfig `json:"global,omitempty"`

	// # Event bus configuration
	Configs ConfigsConfig `json:"configs,omitempty"`

	// # Argo Events controller
	Controller ControllerConfig `json:"controller,omitempty"`

	// # Argo Events admission webhook
	Webhook WebhookConfig `json:"webhook,omitempty"`
}

// CrdsConfig defines the crds configuration
type CrdsConfig struct {
	// -- Install and upgrade CRDs
	Install bool `json:"install,omitempty"`

	// -- Keep CRDs on chart uninstall
	Keep bool `json:"keep,omitempty"`

	// -- Annotations to be added to all CRDs
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ImageConfig defines the image configuration
type ImageConfig struct {
	// -- If defined, a repository applied to all Argo Events deployments
	Repository string `json:"repository,omitempty"`

	// -- Overrides the global Argo Events image tag whose default is the chart appVersion
	Tag string `json:"tag,omitempty"`

	// -- If defined, a imagePullPolicy applied to all Argo Events deployments
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// ImagePullSecretsConfig defines the image pull secrets configuration
type ImagePullSecretsConfig struct {
	Name string `json:"name,omitempty"`
}

// SecurityContextConfig defines the security context configuration
type SecurityContextConfig struct {
	RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`
	RunAsUser    int  `json:"runAsUser,omitempty"`
	RunAsGroup   int  `json:"runAsGroup,omitempty"`
	FsGroup      int  `json:"fsGroup,omitempty"`
}

// HostAliasesConfig defines the host aliases configuration
type HostAliasesConfig struct {
	Ip        string   `json:"ip,omitempty"`
	Hostnames []string `json:"hostnames,omitempty"`
}

// GlobalConfig defines the global configuration
type GlobalConfig struct {
	Image ImageConfig `json:"image,omitempty"`

	// -- If defined, uses a Secret to pull an image from a private Docker registry or repository
	ImagePullSecrets []ImagePullSecretsConfig `json:"imagePullSecrets,omitempty"`

	// -- Annotations for the all deployed pods
	// +miaka:type: map[string]string
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// -- Labels for the all deployed pods
	// +miaka:type: map[string]string
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// -- Additional labels to add to all resources
	// +miaka:type: map[string]string
	AdditionalLabels map[string]string `json:"additionalLabels,omitempty"`

	// -- Toggle and define securityContext. See [values.yaml]
	SecurityContext SecurityContextConfig `json:"securityContext,omitempty"`

	// -- Mapping between IP and hostnames that will be injected as entries in the pod's hosts files
	HostAliases []HostAliasesConfig `json:"hostAliases,omitempty"`
}

// VersionsConfig defines the versions configuration
type VersionsConfig struct {
	Version              string `json:"version,omitempty"`
	NatsStreamingImage   string `json:"natsStreamingImage,omitempty"`
	MetricsExporterImage string `json:"metricsExporterImage,omitempty"`
}

// NatsConfig defines the nats configuration
type NatsConfig struct {
	// -- Supported versions of NATS event bus
	// @default -- See [values.yaml]
	Versions []VersionsConfig `json:"versions,omitempty"`
}

// SettingsConfig defines the settings configuration
type SettingsConfig struct {
	// -- Maximum size of the memory storage (e.g. 1G)
	MaxMemoryStore int `json:"maxMemoryStore,omitempty"`

	// -- Maximum size of the file storage (e.g. 20G)
	MaxFileStore int `json:"maxFileStore,omitempty"`
}

// StreamConfig defines the stream configuration
type StreamConfig struct {
	// -- Maximum number of messages before expiring oldest message
	MaxMsgs int `json:"maxMsgs,omitempty"`

	// -- Maximum age of existing messages, i.e. “72h”, “4h35m”
	MaxAge string `json:"maxAge,omitempty"`

	// Total size of messages before expiring oldest message, 0 means unlimited.
	MaxBytes string `json:"maxBytes,omitempty"`

	// -- Number of replicas, defaults to 3 and requires minimal 3
	Replicas int `json:"replicas,omitempty"`

	// -- Not documented at the moment
	Duplicates string `json:"duplicates,omitempty"`

	// -- 0: Limits, 1: Interest, 2: WorkQueue
	Retention int `json:"retention,omitempty"`

	// -- 0: DiscardOld, 1: DiscardNew
	Discard int `json:"discard,omitempty"`
}

// JetstreamConfigVersionsConfig defines the jetstream config versions configuration
type JetstreamConfigVersionsConfig struct {
	Version              string `json:"version,omitempty"`
	NatsImage            string `json:"natsImage,omitempty"`
	MetricsExporterImage string `json:"metricsExporterImage,omitempty"`
	ConfigReloaderImage  string `json:"configReloaderImage,omitempty"`
	StartCommand         string `json:"startCommand,omitempty"`
}

// JetstreamConfig defines the jetstream configuration
type JetstreamConfig struct {
	// Default JetStream settings, could be overridden by EventBus JetStream spec
	// Ref: https://docs.nats.io/running-a-nats-service/configuration#jetstream
	Settings     SettingsConfig `json:"settings,omitempty"`
	StreamConfig StreamConfig   `json:"streamConfig,omitempty"`

	// Supported versions of JetStream eventbus
	Versions []JetstreamConfigVersionsConfig `json:"versions,omitempty"`
}

// ConfigsConfig defines the configs configuration
type ConfigsConfig struct {
	// # NATS event bus
	Nats NatsConfig `json:"nats,omitempty"`

	// # JetStream event bus
	Jetstream JetstreamConfig `json:"jetstream,omitempty"`
}

// RulesConfig defines the rules configuration
type RulesConfig struct {
	ApiGroups []string `json:"apiGroups,omitempty"`
	Resources []string `json:"resources,omitempty"`
	Verbs     []string `json:"verbs,omitempty"`
}

// RbacConfig defines the rbac configuration
type RbacConfig struct {
	// -- Create events controller RBAC
	Enabled bool `json:"enabled,omitempty"`

	// -- Restrict events controller to operate only in a single namespace instead of cluster-wide scope.
	Namespaced bool `json:"namespaced,omitempty"`

	// -- Additional namespace to be monitored by the controller
	ManagedNamespace string `json:"managedNamespace,omitempty"`

	// -- Additional user rules for event controller's rbac
	Rules []RulesConfig `json:"rules,omitempty"`
}

// ControllerConfigImageConfig defines the controller config image configuration
type ControllerConfigImageConfig struct {
	// -- Repository to use for the events controller
	// @default -- `""` (defaults to global.image.repository)
	Repository string `json:"repository,omitempty"`

	// -- Tag to use for the events controller
	// @default -- `""` (defaults to global.image.tag)
	Tag string `json:"tag,omitempty"`

	// -- Image pull policy for the events controller
	// @default -- `""` (defaults to global.image.imagePullPolicy)
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// PdbConfig defines the pdb configuration
type PdbConfig struct {
	// -- Deploy a PodDisruptionBudget for the events controller
	Enabled bool `json:"enabled,omitempty"`

	// minAvailable: 1
	// maxUnavailable: 0
	// -- Labels to be added to events controller pdb
	// +miaka:type: map[string]string
	Labels map[string]string `json:"labels,omitempty"`

	// -- Annotations to be added to events controller pdb
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`
}

// EnvConfig defines the env configuration
type EnvConfig struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ConfigMapRefConfig defines the config map ref configuration
type ConfigMapRefConfig struct {
	Name string `json:"name,omitempty"`
}

// SecretRefConfig defines the secret ref configuration
type SecretRefConfig struct {
	Name string `json:"name,omitempty"`
}

// EnvFromConfig defines the env from configuration
type EnvFromConfig struct {
	ConfigMapRef ConfigMapRefConfig `json:"configMapRef,omitempty"`
	SecretRef    SecretRefConfig    `json:"secretRef,omitempty"`
}

// CapabilitiesConfig defines the capabilities configuration
type CapabilitiesConfig struct {
	Drop                   []string `json:"drop,omitempty"`
	ReadOnlyRootFilesystem bool     `json:"readOnlyRootFilesystem,omitempty"`
	RunAsNonRoot           bool     `json:"runAsNonRoot,omitempty"`
}

// ContainerSecurityContextConfig defines the container security context configuration
type ContainerSecurityContextConfig struct {
	Capabilities CapabilitiesConfig `json:"capabilities,omitempty"`
}

// ReadinessProbeConfig defines the readiness probe configuration
type ReadinessProbeConfig struct {
	// -- Minimum consecutive failures for the [probe] to be considered failed after having succeeded
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// -- Number of seconds after the container has started before [probe] is initiated
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// -- How often (in seconds) to perform the [probe]
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// -- Minimum consecutive successes for the [probe] to be considered successful after having failed
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// -- Number of seconds after which the [probe] times out
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

// LivenessProbeConfig defines the liveness probe configuration
type LivenessProbeConfig struct {
	// -- Minimum consecutive failures for the [probe] to be considered failed after having succeeded
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// -- Number of seconds after the container has started before [probe] is initiated
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// -- How often (in seconds) to perform the [probe]
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// -- Minimum consecutive successes for the [probe] to be considered successful after having failed
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// -- Number of seconds after which the [probe] times out
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

// EmptyDirConfig defines the empty dir configuration
type EmptyDirConfig struct {
}

// VolumesConfig defines the volumes configuration
type VolumesConfig struct {
	Name     string         `json:"name,omitempty"`
	EmptyDir EmptyDirConfig `json:"emptyDir,omitempty"`
}

// VolumeMountsConfig defines the volume mounts configuration
type VolumeMountsConfig struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
}

// NodeSelectorConfig defines the node selector configuration
type NodeSelectorConfig struct {
	KubernetesIoArch string `json:"kubernetes.io/arch,omitempty"`
}

// TolerationsConfig defines the tolerations configuration
type TolerationsConfig struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// MatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig defines the match expressions config node selector term config match expressions configuration
type MatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig struct {
	Key      string   `json:"key,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Values   []string `json:"values,omitempty"`
}

// MatchExpressionsConfigNodeSelectorTermConfig defines the match expressions config node selector term configuration
type MatchExpressionsConfigNodeSelectorTermConfig struct {
	MatchExpressions []MatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig `json:"matchExpressions,omitempty"`
}

// MatchExpressionsConfig defines the match expressions configuration
type MatchExpressionsConfig struct {
	Key              string                                       `json:"key,omitempty"`
	Operator         string                                       `json:"operator,omitempty"`
	Values           []string                                     `json:"values,omitempty"`
	NodeSelectorTerm MatchExpressionsConfigNodeSelectorTermConfig `json:"nodeSelectorTerm,omitempty"`
}

// NodeSelectorTermConfig defines the node selector term configuration
type NodeSelectorTermConfig struct {
	MatchExpressions []MatchExpressionsConfig `json:"matchExpressions,omitempty"`
}

// RequiredConfig defines the required configuration
type RequiredConfig struct {
	NodeSelectorTerm NodeSelectorTermConfig `json:"nodeSelectorTerm,omitempty"`
}

// NodeAffinityConfig defines the node affinity configuration
type NodeAffinityConfig struct {
	Required RequiredConfig `json:"required,omitempty"`
}

// AffinityConfig defines the affinity configuration
type AffinityConfig struct {
	NodeAffinity NodeAffinityConfig `json:"nodeAffinity,omitempty"`
}

// TopologySpreadConstraintsConfig defines the topology spread constraints configuration
type TopologySpreadConstraintsConfig struct {
	MaxSkew           int    `json:"maxSkew,omitempty"`
	TopologyKey       string `json:"topologyKey,omitempty"`
	WhenUnsatisfiable string `json:"whenUnsatisfiable,omitempty"`
}

// LimitsConfig defines the limits configuration
type LimitsConfig struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// RequestsConfig defines the requests configuration
type RequestsConfig struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// ResourcesConfig defines the resources configuration
type ResourcesConfig struct {
	Limits   LimitsConfig   `json:"limits,omitempty"`
	Requests RequestsConfig `json:"requests,omitempty"`
}

// ServiceAccountConfig defines the service account configuration
type ServiceAccountConfig struct {
	// -- Create a service account for the events controller
	Create bool `json:"create,omitempty"`

	// -- Service account name
	Name string `json:"name,omitempty"`

	// -- Annotations applied to created service account
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`

	// -- Automount API credentials for the Service Account
	AutomountServiceAccountToken bool `json:"automountServiceAccountToken,omitempty"`
}

// ServiceConfig defines the service configuration
type ServiceConfig struct {
	// -- Metrics service annotations
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`

	// +miaka:type: map[string]string
	// -- Metrics service labels
	// +miaka:type: map[string]string
	Labels map[string]string `json:"labels,omitempty"`

	// -- Metrics service port
	ServicePort int `json:"servicePort,omitempty"`
}

// RelabelingsConfig defines the relabelings configuration
type RelabelingsConfig struct {
	SourceLabels []string `json:"sourceLabels,omitempty"`
	TargetLabel  string   `json:"targetLabel,omitempty"`
}

// MetricRelabelingsConfig defines the metric relabelings configuration
type MetricRelabelingsConfig struct {
	SourceLabels []string `json:"sourceLabels,omitempty"`
	TargetLabel  string   `json:"targetLabel,omitempty"`
}

// MatchLabelsConfig defines the match labels configuration
type MatchLabelsConfig struct {
	App string `json:"app,omitempty"`
}

// SelectorConfig defines the selector configuration
type SelectorConfig struct {
	MatchLabels MatchLabelsConfig `json:"matchLabels,omitempty"`
}

// ServiceMonitorConfig defines the service monitor configuration
type ServiceMonitorConfig struct {
	// -- Enable a prometheus ServiceMonitor
	Enabled bool `json:"enabled,omitempty"`

	// -- Prometheus ServiceMonitor interval
	Interval string `json:"interval,omitempty"`

	// -- Prometheus [RelabelConfigs] to apply to samples before scraping
	Relabelings []RelabelingsConfig `json:"relabelings,omitempty"`

	// -- Prometheus [MetricRelabelConfigs] to apply to samples before ingestion
	MetricRelabelings []MetricRelabelingsConfig `json:"metricRelabelings,omitempty"`

	// -- Prometheus ServiceMonitor selector
	Selector SelectorConfig `json:"selector,omitempty"`

	// prometheus: kube-prometheus
	// -- Prometheus ServiceMonitor namespace
	Namespace string `json:"namespace,omitempty"`

	// -- Prometheus ServiceMonitor labels
	// +miaka:type: map[string]string
	AdditionalLabels map[string]string `json:"additionalLabels,omitempty"`
}

// MetricsConfig defines the metrics configuration
type MetricsConfig struct {
	// -- Deploy metrics service
	Enabled        bool                 `json:"enabled,omitempty"`
	Service        ServiceConfig        `json:"service,omitempty"`
	ServiceMonitor ServiceMonitorConfig `json:"serviceMonitor,omitempty"`
}

// ControllerConfig defines the controller configuration
type ControllerConfig struct {
	// -- Argo Events controller name string
	Name  string                      `json:"name,omitempty"`
	Rbac  RbacConfig                  `json:"rbac,omitempty"`
	Image ControllerConfigImageConfig `json:"image,omitempty"`

	// -- The number of replicasets history to keep
	RevisionHistoryLimit int `json:"revisionHistoryLimit,omitempty"`

	// -- The number of events controller pods to run.
	Replicas int `json:"replicas,omitempty"`

	// Pod disruption budget
	Pdb PdbConfig `json:"pdb,omitempty"`

	// -- Environment variables to pass to events controller
	Env []EnvConfig `json:"env,omitempty"`

	// -- envFrom to pass to events controller
	// @default -- `[]` (See [values.yaml])
	EnvFrom []EnvFromConfig `json:"envFrom,omitempty"`

	// -- Annotations to be added to events controller pods
	// +miaka:type: map[string]string
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// -- Labels to be added to events controller pods
	// +miaka:type: map[string]string
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// -- Events controller container-level security context
	ContainerSecurityContext ContainerSecurityContextConfig `json:"containerSecurityContext,omitempty"`

	// # Readiness and liveness probes for default backend
	// # Ref: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
	ReadinessProbe ReadinessProbeConfig `json:"readinessProbe,omitempty"`
	LivenessProbe  LivenessProbeConfig  `json:"livenessProbe,omitempty"`

	// -- Additional volumes to the events controller pod
	Volumes []VolumesConfig `json:"volumes,omitempty"`

	// -- Additional volumeMounts to the events controller main container
	VolumeMounts []VolumeMountsConfig `json:"volumeMounts,omitempty"`

	// -- [Node selector]
	NodeSelector NodeSelectorConfig `json:"nodeSelector,omitempty"`

	// -- [Tolerations] for use with node taints
	Tolerations []TolerationsConfig `json:"tolerations,omitempty"`

	// -- Assign custom [affinity] rules to the deployment
	Affinity AffinityConfig `json:"affinity,omitempty"`

	// -- Assign custom [TopologySpreadConstraints] rules to the events controller
	// # Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
	// # If labelSelector is left out, it will default to the labelSelector configuration of the deployment
	TopologySpreadConstraints []TopologySpreadConstraintsConfig `json:"topologySpreadConstraints,omitempty"`

	// -- Priority class for the events controller pods
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// -- Resource limits and requests for the events controller pods
	Resources      ResourcesConfig      `json:"resources,omitempty"`
	ServiceAccount ServiceAccountConfig `json:"serviceAccount,omitempty"`

	// # Events controller metrics configuration
	Metrics MetricsConfig `json:"metrics,omitempty"`
}

// WebhookConfigImageConfig defines the webhook config image configuration
type WebhookConfigImageConfig struct {
	// -- Repository to use for the event controller
	// @default -- `""` (defaults to global.image.repository)
	Repository string `json:"repository,omitempty"`

	// -- Tag to use for the event controller
	// @default -- `""` (defaults to global.image.tag)
	Tag string `json:"tag,omitempty"`

	// -- Image pull policy for the event controller
	// @default -- `""` (defaults to global.image.imagePullPolicy)
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// WebhookConfigPdbConfig defines the webhook config pdb configuration
type WebhookConfigPdbConfig struct {
	// -- Deploy a PodDisruptionBudget for the admission webhook
	Enabled bool `json:"enabled,omitempty"`

	// minAvailable: 1
	// maxUnavailable: 0
	// -- Labels to be added to admission webhook pdb
	// +miaka:type: map[string]string
	Labels map[string]string `json:"labels,omitempty"`

	// -- Annotations to be added to admission webhook pdb
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`
}

// WebhookConfigEnvConfig defines the webhook config env configuration
type WebhookConfigEnvConfig struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// WebhookConfigEnvFromConfigConfigMapRefConfig defines the webhook config env from config config map ref configuration
type WebhookConfigEnvFromConfigConfigMapRefConfig struct {
	Name string `json:"name,omitempty"`
}

// WebhookConfigEnvFromConfigSecretRefConfig defines the webhook config env from config secret ref configuration
type WebhookConfigEnvFromConfigSecretRefConfig struct {
	Name string `json:"name,omitempty"`
}

// WebhookConfigEnvFromConfig defines the webhook config env from configuration
type WebhookConfigEnvFromConfig struct {
	ConfigMapRef WebhookConfigEnvFromConfigConfigMapRefConfig `json:"configMapRef,omitempty"`
	SecretRef    WebhookConfigEnvFromConfigSecretRefConfig    `json:"secretRef,omitempty"`
}

// WebhookConfigContainerSecurityContextConfigCapabilitiesConfig defines the webhook config container security context config capabilities configuration
type WebhookConfigContainerSecurityContextConfigCapabilitiesConfig struct {
	Drop                   []string `json:"drop,omitempty"`
	ReadOnlyRootFilesystem bool     `json:"readOnlyRootFilesystem,omitempty"`
	RunAsNonRoot           bool     `json:"runAsNonRoot,omitempty"`
}

// WebhookConfigContainerSecurityContextConfig defines the webhook config container security context configuration
type WebhookConfigContainerSecurityContextConfig struct {
	Capabilities WebhookConfigContainerSecurityContextConfigCapabilitiesConfig `json:"capabilities,omitempty"`
}

// WebhookConfigReadinessProbeConfig defines the webhook config readiness probe configuration
type WebhookConfigReadinessProbeConfig struct {
	// -- Minimum consecutive failures for the [probe] to be considered failed after having succeeded
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// -- Number of seconds after the container has started before [probe] is initiated
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// -- How often (in seconds) to perform the [probe]
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// -- Minimum consecutive successes for the [probe] to be considered successful after having failed
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// -- Number of seconds after which the [probe] times out
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

// WebhookConfigLivenessProbeConfig defines the webhook config liveness probe configuration
type WebhookConfigLivenessProbeConfig struct {
	// -- Minimum consecutive failures for the [probe] to be considered failed after having succeeded
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// -- Number of seconds after the container has started before [probe] is initiated
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`

	// -- How often (in seconds) to perform the [probe]
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// -- Minimum consecutive successes for the [probe] to be considered successful after having failed
	SuccessThreshold int `json:"successThreshold,omitempty"`

	// -- Number of seconds after which the [probe] times out
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

// WebhookConfigVolumeMountsConfig defines the webhook config volume mounts configuration
type WebhookConfigVolumeMountsConfig struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
}

// WebhookConfigVolumesConfigEmptyDirConfig defines the webhook config volumes config empty dir configuration
type WebhookConfigVolumesConfigEmptyDirConfig struct {
}

// WebhookConfigVolumesConfig defines the webhook config volumes configuration
type WebhookConfigVolumesConfig struct {
	Name     string                                   `json:"name,omitempty"`
	EmptyDir WebhookConfigVolumesConfigEmptyDirConfig `json:"emptyDir,omitempty"`
}

// WebhookConfigNodeSelectorConfig defines the webhook config node selector configuration
type WebhookConfigNodeSelectorConfig struct {
	KubernetesIoArch string `json:"kubernetes.io/arch,omitempty"`
}

// WebhookConfigTolerationsConfig defines the webhook config tolerations configuration
type WebhookConfigTolerationsConfig struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig defines the webhook config affinity config node affinity config required config node selector term config match expressions config node selector term config match expressions configuration
type WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig struct {
	Key      string   `json:"key,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Values   []string `json:"values,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfig defines the webhook config affinity config node affinity config required config node selector term config match expressions config node selector term configuration
type WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfig struct {
	MatchExpressions []WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfigMatchExpressionsConfig `json:"matchExpressions,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfig defines the webhook config affinity config node affinity config required config node selector term config match expressions configuration
type WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfig struct {
	Key              string                                                                                                                        `json:"key,omitempty"`
	Operator         string                                                                                                                        `json:"operator,omitempty"`
	Values           []string                                                                                                                      `json:"values,omitempty"`
	NodeSelectorTerm WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfigNodeSelectorTermConfig `json:"nodeSelectorTerm,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfig defines the webhook config affinity config node affinity config required config node selector term configuration
type WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfig struct {
	MatchExpressions []WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfigMatchExpressionsConfig `json:"matchExpressions,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfig defines the webhook config affinity config node affinity config required configuration
type WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfig struct {
	NodeSelectorTerm WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfigNodeSelectorTermConfig `json:"nodeSelectorTerm,omitempty"`
}

// WebhookConfigAffinityConfigNodeAffinityConfig defines the webhook config affinity config node affinity configuration
type WebhookConfigAffinityConfigNodeAffinityConfig struct {
	Required WebhookConfigAffinityConfigNodeAffinityConfigRequiredConfig `json:"required,omitempty"`
}

// WebhookConfigAffinityConfig defines the webhook config affinity configuration
type WebhookConfigAffinityConfig struct {
	NodeAffinity WebhookConfigAffinityConfigNodeAffinityConfig `json:"nodeAffinity,omitempty"`
}

// WebhookConfigTopologySpreadConstraintsConfig defines the webhook config topology spread constraints configuration
type WebhookConfigTopologySpreadConstraintsConfig struct {
	MaxSkew           int    `json:"maxSkew,omitempty"`
	TopologyKey       string `json:"topologyKey,omitempty"`
	WhenUnsatisfiable string `json:"whenUnsatisfiable,omitempty"`
}

// WebhookConfigResourcesConfigLimitsConfig defines the webhook config resources config limits configuration
type WebhookConfigResourcesConfigLimitsConfig struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// WebhookConfigResourcesConfigRequestsConfig defines the webhook config resources config requests configuration
type WebhookConfigResourcesConfigRequestsConfig struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// WebhookConfigResourcesConfig defines the webhook config resources configuration
type WebhookConfigResourcesConfig struct {
	Limits   WebhookConfigResourcesConfigLimitsConfig   `json:"limits,omitempty"`
	Requests WebhookConfigResourcesConfigRequestsConfig `json:"requests,omitempty"`
}

// WebhookConfigServiceAccountConfig defines the webhook config service account configuration
type WebhookConfigServiceAccountConfig struct {
	// -- Create a service account for the admission webhook
	Create bool `json:"create,omitempty"`

	// -- Service account name
	Name string `json:"name,omitempty"`

	// -- Annotations applied to created service account
	// +miaka:type: map[string]string
	Annotations map[string]string `json:"annotations,omitempty"`

	// -- Automount API credentials for the Service Account
	AutomountServiceAccountToken bool `json:"automountServiceAccountToken,omitempty"`
}

// WebhookConfig defines the webhook configuration
type WebhookConfig struct {
	// -- Enable admission webhook. Applies only for cluster-wide installation
	Enabled bool `json:"enabled,omitempty"`

	// -- Argo Events admission webhook name string
	Name  string                   `json:"name,omitempty"`
	Image WebhookConfigImageConfig `json:"image,omitempty"`

	// -- The number of replicasets history to keep
	RevisionHistoryLimit int `json:"revisionHistoryLimit,omitempty"`

	// -- The number of webhook pods to run.
	Replicas int `json:"replicas,omitempty"`

	// Pod disruption budget
	Pdb WebhookConfigPdbConfig `json:"pdb,omitempty"`

	// -- Environment variables to pass to event controller
	// @default -- `[]` (See [values.yaml])
	Env []WebhookConfigEnvConfig `json:"env,omitempty"`

	// -- envFrom to pass to event controller
	// @default -- `[]` (See [values.yaml])
	EnvFrom []WebhookConfigEnvFromConfig `json:"envFrom,omitempty"`

	// -- Annotations to be added to event controller pods
	// +miaka:type: map[string]string
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// -- Labels to be added to event controller pods
	// +miaka:type: map[string]string
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// -- Port to listen on
	Port int `json:"port,omitempty"`

	// -- Event controller container-level security context
	ContainerSecurityContext WebhookConfigContainerSecurityContextConfig `json:"containerSecurityContext,omitempty"`

	// # Readiness and liveness probes for default backend
	// # Ref: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
	ReadinessProbe WebhookConfigReadinessProbeConfig `json:"readinessProbe,omitempty"`
	LivenessProbe  WebhookConfigLivenessProbeConfig  `json:"livenessProbe,omitempty"`

	// -- Additional volumeMounts to the event controller main container
	VolumeMounts []WebhookConfigVolumeMountsConfig `json:"volumeMounts,omitempty"`

	// -- Additional volumes to the event controller pod
	Volumes []WebhookConfigVolumesConfig `json:"volumes,omitempty"`

	// -- [Node selector]
	NodeSelector WebhookConfigNodeSelectorConfig `json:"nodeSelector,omitempty"`

	// -- [Tolerations] for use with node taints
	Tolerations []WebhookConfigTolerationsConfig `json:"tolerations,omitempty"`

	// -- Assign custom [affinity] rules to the deployment
	Affinity WebhookConfigAffinityConfig `json:"affinity,omitempty"`

	// -- Assign custom [TopologySpreadConstraints] rules to the event controller
	// # Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
	// # If labelSelector is left out, it will default to the labelSelector configuration of the deployment
	TopologySpreadConstraints []WebhookConfigTopologySpreadConstraintsConfig `json:"topologySpreadConstraints,omitempty"`

	// -- Priority class for the event controller pods
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// -- Resource limits and requests for the event controller pods
	Resources      WebhookConfigResourcesConfig      `json:"resources,omitempty"`
	ServiceAccount WebhookConfigServiceAccountConfig `json:"serviceAccount,omitempty"`
}
