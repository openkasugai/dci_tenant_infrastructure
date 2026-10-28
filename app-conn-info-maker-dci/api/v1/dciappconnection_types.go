package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DCIAppConnectionSpec defines the desired state of DCIAppConnection.
type DCIAppConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Connections []Connection `json:"connections,omitempty"`
	// +optional
	InputEndpoints []DCIAppConnectionExternalEndpoint `json:"inputEndpoints,omitempty"`
	// +optional
	OutputEndpoints []DCIAppConnectionExternalEndpoint `json:"outputEndpoints,omitempty"`
}

// Connection defines the desired state of DCIAppConnectionSpec.
type Connection struct {
	From TargetRef `json:"from"`
	To   TargetRef `json:"to"`
}

// DCIAppConnectionExternalEndpoint defines the desired state of DCIAppConnectionSpec.
// It is currently identical to ExternalEndpoint, keep as an alias to avoid duplicate
// struct definitions and to remain compatible with tests that construct
// examplecomv1.ExternalEndpoint values.
// DCIAppConnectionExternalEndpoint defines the desired state of DCIAppConnectionSpec.
type DCIAppConnectionExternalEndpoint struct {
	Internal TargetRef `json:"internal"`
	// +kubebuilder:validation:MinItems=1
	ExternalSubsets []string `json:"externalSubsets"`
}

// Connection defines the desired state of Connection and ExternalEndpoint.
type TargetRef struct {
	DeploymentRef NamespaceName `json:"deploymentRef"`
}
type NamespaceName struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	Name string `json:"name"`
}

// DCIAppConnectionStatus defines the observed state of DCIAppConnection.
type DCIAppConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Enum="Created";"Pending";"Failed";""
	// +default:phase=""
	Phase string `json:"phase"`
	// +optional
	Connections []Connection `json:"connections,omitempty"`
	// +optional
	InputEndpoints []DCIAppConnectionExternalEndpoint `json:"inputEndpoints,omitempty"`
	// +optional
	OutputEndpoints []DCIAppConnectionExternalEndpoint `json:"outputEndpoints,omitempty"`
}

const (
	StatusPhaseCreated = "Created"
	StatusPhaseFailed  = "Failed"
	StatusPhasePending = "Pending"
)

// +kubebuilder:object:root=true
// kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// DCIAppConnection is the Schema for the dciappconnections API.
type DCIAppConnection struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DCIAppConnection
	// +required
	Spec DCIAppConnectionSpec `json:"spec"`

	// status defines the observed state of DCIAppConnection
	// +optional
	Status DCIAppConnectionStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DCIAppConnectionList contains a list of DCIAppConnection.
type DCIAppConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DCIAppConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &DCIAppConnection{}, &DCIAppConnectionList{})
		return nil
	})
}
