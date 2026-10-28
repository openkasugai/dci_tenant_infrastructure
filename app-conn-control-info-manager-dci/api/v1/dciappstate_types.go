package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DCIAppStateSpec defines the desired state of DCIAppState.
type DCIAppStateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum="computation";"connection"
	Type string `json:"type"`
}

// DCIAppStateStatus defines the observed state of DCIAppState.
type DCIAppStateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:pruning:PreserveUnknownFields
	AppStatus runtime.RawExtension `json:"appStatus,omitempty"`
}

// +kubebuilder:object:root=true
//  kubebuilder:subresource:status

// DCIAppState is the Schema for the dciappstates API.
type DCIAppState struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DCIAppState
	// +required
	Spec DCIAppStateSpec `json:"spec"`

	// status defines the observed state of DCIAppState
	// +optional
	Status DCIAppStateStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DCIAppStateList contains a list of DCIAppState.
type DCIAppStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DCIAppState `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &DCIAppState{}, &DCIAppStateList{})
		return nil
	})
}
