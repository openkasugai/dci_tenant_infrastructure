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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
)

// ScalingParameter defines the parameters for scaling a card.
// +kubebuilder:validation:XValidation:rule="self.op != 'add' || (has(self.num) && !has(self.uuids))",message="When op is 'add', only 'num' must be set."
// +kubebuilder:validation:XValidation:rule="self.op != 'delete' || ((has(self.num) && !has(self.uuids)) || (!has(self.num) && has(self.uuids)))",message="When op is 'delete', either 'num' or 'uuids' must be set, but not both."
// +kubebuilder:validation:XValidation:rule="self.op != 'replace' || (!has(self.num) && has(self.uuids))",message="When op is 'replace', only 'uuids' must be set."
//
//nolint:lll // kubebuilder markers cannot be split into multiple lines
type ScalingParameter struct {
	// Op defines the operation to be performed.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=add;delete;replace
	Op capdv1.DCIMachineCardScalingOp `json:"op"`

	// Model defines the model of the GPU card.
	// +kubebuilder:validation:Required
	Model string `json:"model"`

	// Num defines the number of GPU cards to be added or removed.
	// When Num is set with "delete", UUIDs must not be set.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Num int32 `json:"num,omitempty"`

	// UUIDs defines the UUIDs of the GPU cards to be removed or replaced.
	// When UUIDs is set with "delete", Num must not be set.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	UUIDs []string `json:"uuids,omitempty"`
}

// CardScalingRequestSpec defines the desired state of CardScalingRequest.
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="`spec` may not be modified after creation."
type CardScalingRequestSpec struct {
	// +kubebuilder:validation:Required
	NodeName string `json:"nodeName"`

	// +kubebuilder:validation:Required
	ScalingParameter ScalingParameter `json:"scalingParameter"`
}

// DCIMachineInfo defines the DCI machine information associated with the CardScalingRequest.
type DCIMachineInfo struct {
	// +kubebuilder:validation:Optional
	DCIMachineName string `json:"dciMachineName,omitempty"`

	// +kubebuilder:validation:Optional
	DCIMachineTemplateName string `json:"dciMachineTemplateName,omitempty"`

	// +kubebuilder:validation:Optional
	MachineDeploymentName string `json:"machineDeploymentName,omitempty"`

	// +kubebuilder:validation:Optional
	DCIMachineUID *types.UID `json:"dciMachineUID,omitempty"`

	// +kubebuilder:validation:Optional
	DCIMachineTemplateUID *types.UID `json:"dciMachineTemplateUID,omitempty"`

	// +kubebuilder:validation:Optional
	MachineDeploymentUID *types.UID `json:"machineDeploymentUID,omitempty"`
}

// CardScalingRequestStatus defines the observed state of CardScalingRequest.
type CardScalingRequestStatus struct {
	// Phase defines the current phase of the CardScalingRequest.
	// +kubebuilder:validation:Required
	Phase CardScalingRequestPhase `json:"phase"`

	// ClusterNamespace defines the namespace of the cluster where the node is located.
	// +kubebuilder:validation:Optional
	ClusterNamespace string `json:"clusterNamespace,omitempty"`

	// ClusterName defines the name of the cluster where the node is located.
	// +kubebuilder:validation:Optional
	ClusterName string `json:"clusterName,omitempty"`

	// OldDCIMachineInfo defines the DCI machine information before scaling.
	// +kubebuilder:validation:Optional
	OldDCIMachineInfo *DCIMachineInfo `json:"oldDCIMachineInfo,omitempty"`

	// NewDCIMachineInfo defines the DCI machine information after scaling.
	// +kubebuilder:validation:Optional
	NewDCIMachineInfo *DCIMachineInfo `json:"newDCIMachineInfo,omitempty"`

	// LastTransitionTime is the last time the phase transitioned to the completed state.
	// +kubebuilder:validation:Optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CardScalingRequest is the Schema for the cardscalingrequests API.
// +kubebuilder:resource:shortName=csr
// +kubebuilder:printcolumn:name="TargetNode",type=string,JSONPath=".spec.nodeName"
// +kubebuilder:printcolumn:name="Operation",type=string,JSONPath=".spec.scalingParameter.op"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"
type CardScalingRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CardScalingRequestSpec   `json:"spec,omitempty"`
	Status CardScalingRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CardScalingRequestList contains a list of CardScalingRequest.
type CardScalingRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CardScalingRequest `json:"items"`
}

// CardScalingRequestPhase defines the phase of the CardScalingRequest.
// +kubebuilder:validation:Enum="";CardScaleInitializing;DCIMachineTemplateCreating;MachineDeploymentCreating;DCIMachineAnnotating;CardScaleInitialized;NodeScalingDown;NodeScaleDownDeleting;NodeScaleDownCompleted;NodeScalingUp;CardScaleCompleted;NodeScaleDownFailed;NodeScaleUpFailed;DCIMachineTemplateCleanup;MachineDeploymentCleanup;DCIMachineCleanup;CardScaleFinalized;Pending
//
//nolint:lll // kubebuilder markers cannot be split into multiple lines
type CardScalingRequestPhase string

const (
	PhaseCardScaleInitial           CardScalingRequestPhase = ""
	PhaseCardScaleInitializing      CardScalingRequestPhase = "CardScaleInitializing"
	PhaseDCIMachineTemplateCreating CardScalingRequestPhase = "DCIMachineTemplateCreating"
	PhaseMachineDeploymentCreating  CardScalingRequestPhase = "MachineDeploymentCreating"
	PhaseDCIMachineAnnotating       CardScalingRequestPhase = "DCIMachineAnnotating"
	PhaseCardScaleInitialized       CardScalingRequestPhase = "CardScaleInitialized"
	PhaseNodeScalingDown            CardScalingRequestPhase = "NodeScalingDown"
	PhaseNodeScaleDownDeleting      CardScalingRequestPhase = "NodeScaleDownDeleting"
	PhaseNodeScaleDownCompleted     CardScalingRequestPhase = "NodeScaleDownCompleted"
	PhaseNodeScalingUp              CardScalingRequestPhase = "NodeScalingUp"
	PhaseCardScaleCompleted         CardScalingRequestPhase = "CardScaleCompleted"
	PhaseNodeScaleDownFailed        CardScalingRequestPhase = "NodeScaleDownFailed"
	PhaseNodeScaleUpFailed          CardScalingRequestPhase = "NodeScaleUpFailed"
	PhaseDCIMachineTemplateCleanup  CardScalingRequestPhase = "DCIMachineTemplateCleanup"
	PhaseMachineDeploymentCleanup   CardScalingRequestPhase = "MachineDeploymentCleanup"
	PhaseDCIMachineCleanup          CardScalingRequestPhase = "DCIMachineCleanup"
	PhaseCardScaleFinalized         CardScalingRequestPhase = "CardScaleFinalized"
	PhaseCardScalePending           CardScalingRequestPhase = "Pending"
)

const (
	// annotations set on NodeScalingRequest to handle target MachineDeployment to be deleted or not
	CardScalingRequestInProgressUIDAnnotation = "scaling.my.domain/csr-in-progress-uid"
	// finalizer name for CardScalingRequest
	CardScalingRequestFinalizerName = "scaling.my.domain/cardscalingrequest"
)

//nolint:gochecknoinits // this function is automatically generated by kubebuilder
func init() {
	SchemeBuilder.Register(&CardScalingRequest{}, &CardScalingRequestList{})
}
