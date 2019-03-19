package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RabbitmqImage Sets image url and tag
type RabbitmqImage struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// RabbitmqSSL sets SSL parameters
type RabbitmqSSL struct {
	Enabled       bool   `json:"enabled"`
	ExitingSecret string `json:"exitingSecret,omitempty"`
	Cacertfile    string `json:"cacertfile,omitempty"`
	Certfile      string `json:"certfile,omitempty"`
	Keyfile       string `json:"keyfile,omitempty"`
}

// RabbitmqAuth auth config
type RabbitmqAuth struct {
	Enabled bool `json:"enabled"`
	// +kubebuilder:validation:UniqueItems=true
	Config []string `json:"mechanisms,omitempty"`
}

// RabbitmqManagementPlugin admin panel
type RabbitmqManagementPlugin struct {
	Enabled bool `json:"enabled"`
}

// RabbitmqPolicy type
type RabbitmqPolicy struct {
	Vhost      string            `json:"vhost"`
	Name       string            `json:"name"`
	Pattern    string            `json:"pattern"`
	Definition map[string]string `json:"definition"`
	Priority   int64             `json:"priority"`
	ApplyTo    string            `json:"apply-to"`
}

// // RabbitmqCredentials sets credentials
// type RabbitmqCredentials struct {
// 	User     string `json:"username"`
// 	Password string `json:"password"`
// }

// RabbitmqSpec defines the desired state of Rabbitmq
// +k8s:openapi-gen=true
type RabbitmqSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	RabbitmqReplicas                    int32                  `json:"replicas"`
	RabbitmqVhost                       string                 `json:"default_vhost,omitempty"`
	RabbitmqSecretCredentials           string                 `json:"secret_credentials,omitempty"`
	RabbitmqSecretServiceAccount        string                 `json:"secret_service_account,omitempty"`
	RabbitmqMemoryHighWatermark         string                 `json:"memory_high_watermark,omitempty"`
	RabbitmqHipeCompile                 bool                   `json:"hipe_compile,omitempty"`
	K8SImage                            RabbitmqImage          `json:"image"`
	RabbitmqSSL                         RabbitmqSSL            `json:"cert,omitempty"`
	RabbitmqAuth                        RabbitmqAuth           `json:"auth,omitempty"`
	K8SENV                              []corev1.EnvVar        `json:"env,omitempty"`
	K8SLabels                           []metav1.LabelSelector `json:"k8s_labels"`
	RabbitmqVolumeSize                  resource.Quantity      `json:"volume_size"`
	RabbitmqPodRequests                 corev1.ResourceList    `json:"pod_requests,omitempty"`
	RabbitmqPodLimits                   corev1.ResourceList    `json:"pod_limits,omitempty"`
	RabbitmqPolicies                    []RabbitmqPolicy       `json:"policies"`
	RabbitmqK8SHost                     string                 `json:"k8s_host"`
	RabbitmqK8SAddrType                 string                 `json:"k8s_addrtype"`
	RabbitmqK8SPeerDiscoveryBackend     string                 `json:"k8s_peer_discovery_backend"`
	RabbitmqClusterFormationNodeCleanup int64                  `json:"cluster_node_cleanup_interval"`
	RabbitmqClusterPartitionHandling    string                 `json:"cluster_partition_handling"`
	RabbitmqPlugins                     []string               `json:"plugins"`
	RabbitmqCredentials                 map[string][]byte      `json:"credentials"`
}

// RabbitmqStatus defines the observed state of Rabbitmq
// +k8s:openapi-gen=true
type RabbitmqStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rabbitmq is the Schema for the rabbitmqs API
// +k8s:openapi-gen=true
type Rabbitmq struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitmqSpec   `json:"spec,omitempty"`
	Status RabbitmqStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqList contains a list of Rabbitmq
type RabbitmqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rabbitmq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rabbitmq{}, &RabbitmqList{})
}
