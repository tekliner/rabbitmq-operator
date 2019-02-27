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

// RabbitmqSpec defines the desired state of Rabbitmq
// +k8s:openapi-gen=true
type RabbitmqSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	RabbitmqReplicas            int32               `json:"replicas"`
	RabbitmqUsername            string              `json:"rabbitmqUsername"`
	RabbitmqVhost               string              `json:"rabbitmqVhost,omitempty"`
	RabbitmqMemoryHighWatermark string              `json:"rabbitmqMemoryHighWatermark,omitempty"`
	RabbitmqEpmdPort            int32               `json:"rabbitmqEpmdPort,omitempty"`
	RabbitmqNodePort            int32               `json:"rabbitmqNodePort,omitempty"`
	RabbitmqManagerPort         int32               `json:"rabbitmqManagerPort,omitempty"`
	RabbitmqHipeCompile         bool                `json:"rabbitmqHipeCompile,omitempty"`
	Image                       RabbitmqImage       `json:"image"`
	SSL                         RabbitmqSSL         `json:"rabbitmqCert,omitempty"`
	RabbitmqAuth                RabbitmqAuth        `json:"rabbitmqAuth,omitempty"`
	ENV                         []corev1.EnvVar     `json:"env,omitempty"`
	RabbitmqVolumeSize          resource.Quantity   `json:"volumeSize"`
	RabbitmqPodRequests         corev1.ResourceList `json:"podRequests,omitempty"`
	RabbitmqPodLimits           corev1.ResourceList `json:"podLimits,omitempty"`
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
