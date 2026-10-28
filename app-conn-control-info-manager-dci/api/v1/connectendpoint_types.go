package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConnectEndpointSpec defines the desired state of ConnectEndpoint.
type ConnectEndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	From FromDeployment `json:"from"`
	// +optional
	To             []ToDeployment    `json:"to,omitempty"`
	InputEndpoint  *ExternalEndpoint `json:"inputEndpoint,omitempty"`
	OutputEndpoint *ExternalEndpoint `json:"outputEndpoint,omitempty"`
}

// ConnectEndpointStatus defines the observed state of ConnectEndpoint.
type ConnectEndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Condition string         `json:"condition"`
	From      FromDeployment `json:"from"`
	// +optional
	To []ToDeploymentStatus `json:"to,omitempty"`
	// +optional
	InputEndpoint *ExternalEndpoint `json:"inputEndpoint,omitempty"`
	// +optional
	OutputEndpoint *ExternalEndpoint `json:"outputEndpoint,omitempty"`
}

// FromDeployment defines the observed state of ConnectEndpointSpec or ConnectEndpointStatus.
type FromDeployment struct {
	DeploymentRef NamespacedName `json:"deploymentRef"`
	FromPod       FromPod        `json:"fromPod"`
}

// FromPod defines the observed state of FromDeployment
type FromPod struct {
	PodRef NamespacedName `json:"podRef"`
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:Pattern="^$|^(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])){3}$"
	Address string `json:"address"`
}

// ToDeployment defines the observed state of ConnectEndpointSpec.
type ToDeployment struct {
	DeploymentRef NamespacedName `json:"deploymentRef"`
	ToPod         []ToPod        `json:"toPod,omitempty"`
}

// ToPod defines the observed state of ToDeployment.
type ToPod struct {
	PodRef NamespacedName `json:"podRef"`
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:Pattern="^$|^(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])){3}$"
	Address string `json:"address"`
}

// ToDeploymentStatus defines the observed state of ConnectEndpointStatus.
type ToDeploymentStatus struct {
	DeploymentRef NamespacedName `json:"deploymentRef"`
	// +optional
	ToPod []ToPodStatus `json:"toPod,omitempty"`
}

// ToPodStatus defines the observed state of ConnectEndpointStatus.
type ToPodStatus struct {
	PodRef NamespacedName `json:"podRef"`
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:Pattern="^$|^(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])){3}$"
	Address string `json:"address"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	AppStatus runtime.RawExtension `json:"appStatus,omitempty"`
}
type NamespacedName struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern="^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern="^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	Name string `json:"name"`
}

// ExternalEndpoint defines the observed state of ConnectEndpointSpec or ConnectEndpointStatus.
type ExternalEndpoint struct {
	// +kubebuilder:validation:MinItems=0
	ExternalSubsets []string `json:"externalSubsets"`
}

// Note: ExternalEndpoint intentionally contains only external subsets for connect endpoints.

// +kubebuilder:object:root=true
//  kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CONDITION",type="string",JSONPath=".status.condition"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// ConnectEndpoint is the Schema for the connectendpoints API.
type ConnectEndpoint struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ConnectEndpoint
	// +required
	Spec ConnectEndpointSpec `json:"spec"`

	// status defines the observed state of ConnectEndpoint
	// +optional
	Status ConnectEndpointStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ConnectEndpointList contains a list of ConnectEndpoint.
type ConnectEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ConnectEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ConnectEndpoint{}, &ConnectEndpointList{})
		return nil
	})
}
