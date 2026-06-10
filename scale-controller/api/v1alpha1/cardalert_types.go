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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rpmv1 "resource-pool-manager/api/v1"
)

// CardInfo defines information about a card like GPU.
type CardInfo struct {
	// Model defines the model of the card.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	Model string `json:"model"`

	// GPUUUID defines the UUID of the GPU card.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	GPUUUID string `json:"gpuUUID"`

	// MIGUUIDs defines the UUIDs of the MIG (Multi-Instance GPU) cards.
	// +kubebuilder:validation:Optional
	MIGUUIDs []string `json:"migUUIDs,omitempty"`
}

// CardAlertSpec defines the desired state of CardAlert.
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="`spec` may not be modified after creation."
type CardAlertSpec struct {
	// NodeName defines the name of the Node where the alert occurred.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	NodeName string `json:"nodeName"`

	// Cards defines the list of GPU cards that triggered the alert.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Cards []CardInfo `json:"cards"`

	// Source defines the source of the alert.
	// +kubebuilder:validation:Required
	Source CardAlertSourceType `json:"source"`

	// Reason defines the reason for the alert.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`
}

// CardAlertStatus defines the observed state of CardAlert.
type CardAlertStatus struct {
	// State defines the state of the CardAlert.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	State CardAlertState `json:"state"`

	// LogicalServerRef defines the reference to the LogicalServer associated with the Node.
	// +kubebuilder:validation:Optional
	LogicalServerRef corev1.ObjectReference `json:"logicalServerRef,omitempty"`

	// ServerType defines the type of the LogicalServer.
	// +kubebuilder:validation:Optional
	ServerType rpmv1.AppliedServerType `json:"serverType,omitempty"`

	// CardModel defines the model of the card that triggered the alert in the status of LogicalServer.
	// +kubebuilder:validation:Optional
	CardModel string `json:"cardModel,omitempty"`

	// SkipCards defines the list of cards that are skipped for any reasons.
	// +kubebuilder:validation:Optional
	SkipCards []CardInfo `json:"skipCards,omitempty"`

	// TargetCards defines the list of the target cards.
	// +kubebuilder:validation:Optional
	TargetCards []CardInfo `json:"targetCards,omitempty"`

	// ResourceRef defines the reference to the resource created for handling the alert, such as MachineAlert or Taint.
	// +kubebuilder:validation:Optional
	ResourceRef corev1.ObjectReference `json:"resourceRef,omitempty"`

	// CompletedTime defines the timestamp when the alert handling is completed.
	// +kubebuilder:validation:Optional
	CompletedTime metav1.Time `json:"completedTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CardAlert is the Schema for the cardalerts API.
// +kubebuilder:resource:shortName=ca
// +kubebuilder:printcolumn:name="TargetNode",type=string,JSONPath=".spec.nodeName"
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"
type CardAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CardAlertSpec   `json:"spec,omitempty"`
	Status CardAlertStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CardAlertList contains a list of CardAlert.
type CardAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CardAlert `json:"items"`
}

// CardAlertState defines the state of the CardAlert.
// +kubebuilder:validation:Enum="";Initializing;Pending;CreatingReplaceResource;WaitingCardScalingRequest;WaitingMachineAlert;DeletingMachineAlert;CreatingTaintResource;Completed;Failed;Finalized
//
//nolint:lll // kubebuilder markers cannot be split into multiple lines
type CardAlertState string

// State for CardAlert
const (
	CardAlertInitialState                   CardAlertState = ""
	CardAlertInitializingState              CardAlertState = "Initializing"
	CardAlertPendingState                   CardAlertState = "Pending"
	CardAlertCreatingReplaceResourceState   CardAlertState = "CreatingReplaceResource"
	CardAlertWaitingCardScalingRequestState CardAlertState = "WaitingCardScalingRequest"
	CardAlertWaitingMachineAlertState       CardAlertState = "WaitingMachineAlert"
	CardAlertDeletingMachineAlertState      CardAlertState = "DeletingMachineAlert"
	CardAlertCreatingTaintResourceState     CardAlertState = "CreatingTaintResource"
	CardAlertCompletedState                 CardAlertState = "Completed"
	CardAlertFailedState                    CardAlertState = "Failed"
	CardAlertFinalizedState                 CardAlertState = "Finalized"
)

// CardAlertSourceType defines the source of the CardAlert.
// +kubebuilder:validation:Enum=WebhookReceiverOnTenant
type CardAlertSourceType string

const (
	WebhookReceiverOnTenant CardAlertSourceType = "WebhookReceiverOnTenant"
)

const CardAlertFinalizerName = "scaling.my.domain/cardalert"

//nolint:gochecknoinits // this function is automatically generated by kubebuilder
func init() {
	SchemeBuilder.Register(&CardAlert{}, &CardAlertList{})
}
