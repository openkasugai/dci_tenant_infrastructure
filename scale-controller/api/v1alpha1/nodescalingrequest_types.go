// Copyright 2026 NTT, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	rpmv1 "resource-pool-manager/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeScalingRequestType defines the operation type for NodeScalingRequest.
// +kubebuilder:validation:Enum=create;delete
type NodeScalingRequestType string

// RequestType for .spec.nodeRequestType
const (
	TypeCreate NodeScalingRequestType = "create"
	TypeDelete NodeScalingRequestType = "delete"
)

// NodeScalingRequestPhase defines the phase of the NodeScalingRequest.
// +kubebuilder:validation:Enum="";"Creating";"Created";"CreateFailed";"Deleting";"Deleted";"DeleteFailed";"Pending"
type NodeScalingRequestPhase string

// Phase for .status.phase
const (
	PhaseNodeScaleInitial NodeScalingRequestPhase = ""
	PhaseNodeCreating     NodeScalingRequestPhase = "Creating"
	PhaseNodeCreated      NodeScalingRequestPhase = "Created"
	PhaseNodeCreateFailed NodeScalingRequestPhase = "CreateFailed"
	PhaseNodeDeleting     NodeScalingRequestPhase = "Deleting"
	PhaseNodeDeleted      NodeScalingRequestPhase = "Deleted"
	PhaseNodeDeleteFailed NodeScalingRequestPhase = "DeleteFailed"
	PhaseNodeScalePending NodeScalingRequestPhase = "Pending"
)

const (
	// labels set on resources created by NodeScalingRequest controller
	CreatedByNamespaceLabel = "scaling.my.domain/created-by-namespace"
	CreatedByNameLabel      = "scaling.my.domain/created-by-name"
	// labels set on MachineDeployment being processed by NodeScalingRequest controller
	InProgressUIDLabel = "scaling.my.domain/in-progress-uid"
	// finalizer name for NodeScalingRequest
	NodeScalingRequestFinalizerName = "scaling.my.domain/nodescalingrequest"
)

// NodeScalingRequestSpec defines the desired state of NodeScalingRequest.
// +kubebuilder:validation:XValidation:rule="self.nodeRequestType == 'create' ? !has(self.node) : true",message="nodeRequestType is create, but spec.node is set"
// +kubebuilder:validation:XValidation:rule="self.nodeRequestType == 'create' ? has(self.serverSpec) : true",message="nodeRequestType is create, but spec.serverSpec is not set"
// +kubebuilder:validation:XValidation:rule="self.nodeRequestType == 'delete' ? has(self.node) : true",message="nodeRequestType is delete, but spec.node is not set"
// +kubebuilder:validation:XValidation:rule="self.nodeRequestType == 'delete' ? !has(self.serverSpec) : true",message="nodeRequestType is delete, but spec.serverSpec is set"
// +kubebuilder:validation:XValidation:rule="self.nodeRequestType == 'delete' ? !has(self.location) : true",message="nodeRequestType is delete, but spec.location is set"
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="`spec` may not be modified after creation."
//
//nolint:lll // kubebuilder markers cannot be split into multiple lines
type NodeScalingRequestSpec struct {
	// NodeRequestType is the type of scaling request (create or delete).
	// +kubebuilder:validation:Required
	NodeRequestType NodeScalingRequestType `json:"nodeRequestType"`

	// Node is the name of the node to delete (required for delete requests).
	// +optional
	Node *string `json:"node,omitempty"`

	// Location is the location for the logical server base (required for create requests).
	// +optional
	Location *string `json:"location,omitempty"`

	// ServerSpec contains information for creating a new node (required for create requests).
	// +optional
	ServerSpec *rpmv1.ServerSpec `json:"serverSpec,omitempty"`
}

// NodeScalingRequestStatus defines the observed state of NodeScalingRequest.
type NodeScalingRequestStatus struct {
	// Phase is the current phase of the scaling request
	// +kubebuilder:validation:Required
	// +default:phase=""
	Phase NodeScalingRequestPhase `json:"phase"`

	// ClusterNamespace is the namespace where the Cluster resource is located.
	// +optional
	ClusterNamespace string `json:"clusterNamespace,omitempty"`

	// ClusterName is the name of the Cluster resource.
	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// MachineDeploymentName is the name of the MachineDeployment being scaled.
	// +optional
	MachineDeploymentName string `json:"machineDeploymentName,omitempty"`

	// MachineDeploymentUID is the uid of the MachineDeployment being scaled.
	// +optional
	MachineDeploymentUID *types.UID `json:"machineDeploymentUID,omitempty"`

	// DesiredReplicas is the desired number of replicas after scaling.
	// +optional
	DesiredReplicas *int32 `json:"desiredReplicas,omitempty"`

	// Conditions represent the state of the NodeScalingRequest.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeScalingRequest is the Schema for the nodescalingrequests API.
// +kubebuilder:resource:shortName=nsr
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=".spec.nodeRequestType"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"
type NodeScalingRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeScalingRequestSpec   `json:"spec,omitempty"`
	Status NodeScalingRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeScalingRequestList contains a list of NodeScalingRequest.
type NodeScalingRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeScalingRequest `json:"items"`
}

//nolint:gochecknoinits // this function is automatically generated by kubebuilder
func init() {
	SchemeBuilder.Register(&NodeScalingRequest{}, &NodeScalingRequestList{})
}
