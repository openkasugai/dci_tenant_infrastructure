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
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"
)

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.

//nolint:lll // kubebuilder marker cannot be split into multiple lines.
// +kubebuilder:webhook:path=/validate-cluster-my-domain-v1alpha1-cardscalingrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=cluster.my.domain,resources=cardscalingrequests,verbs=create;update,versions=v1alpha1,name=vcardscalingrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// CardScalingRequestCustomValidator struct is responsible for validating the CardScalingRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type CardScalingRequestCustomValidator struct {
	// ClientHelper for workload cluster and management cluster
	ClientHelper helperclient.ClientHelper
}

var _ webhook.CustomValidator = &CardScalingRequestCustomValidator{}

// SetupWithManager registers the webhook for CardScalingRequest in the manager.
func (v *CardScalingRequestCustomValidator) SetupWithManager(
	mgr ctrl.Manager,
) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&scalev1.CardScalingRequest{}).
		WithValidator(v).
		Complete()
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type CardScalingRequest.
func (v *CardScalingRequestCustomValidator) ValidateCreate(
	ctx context.Context, obj runtime.Object,
) (admission.Warnings, error) {
	csr, ok := obj.(*scalev1.CardScalingRequest)
	if !ok {
		// This case is only for failsafe
		return nil, fmt.Errorf("expected object to be CardScalingRequest but got %T", obj)
	}
	if csr == nil {
		// This case is only for failsafe
		return nil, errors.New("expected CardScalingRequest not to be nil, but got nil")
	}

	logger := log.FromContext(ctx).WithName("cardscalingrequest-validator")
	logger.Info("Validation for CardScalingRequest upon creation")

	var errList field.ErrorList

	// Get LogicalServer associated with the target Node
	logicalServer, err := v.getValidLogicalServer(ctx, csr.Spec.NodeName)
	if err != nil {
		errList = append(errList,
			field.Invalid(field.NewPath("spec").Child("nodeName"), csr.Spec.NodeName, err.Error()),
		)
	} else {
		// Validate GPU specifications after merging with CardScalingRequest
		if err = v.validateScalingParameter(&csr.Spec.ScalingParameter, logicalServer); err != nil {
			errList = append(errList,
				field.Invalid(field.NewPath("spec").Child("scalingParameter"), csr.Spec.ScalingParameter, err.Error()),
			)
		}
	}

	if len(errList) > 0 {
		return nil, apierrors.NewInvalid(schema.GroupKind{
			Group: scalev1.GroupVersion.Group,
			Kind:  "CardScalingRequest",
		}, csr.Name, errList)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type CardScalingRequest.
func (*CardScalingRequestCustomValidator) ValidateUpdate(
	ctx context.Context, _, newObj runtime.Object,
) (admission.Warnings, error) {
	_, ok := newObj.(*scalev1.CardScalingRequest)
	if !ok {
		return nil, fmt.Errorf("expected a CardScalingRequest object for the newObj but got %T", newObj)
	}

	logger := log.FromContext(ctx).WithName("cardscalingrequest-validator")
	logger.Info("Validation for CardScalingRequest upon update")

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type CardScalingRequest.
// This function will never be called because we set the webhook to ignore delete operations in the kubebuilder marker.
// However, we implement this method to satisfy the webhook.CustomValidator interface.
func (*CardScalingRequestCustomValidator) ValidateDelete(
	ctx context.Context, obj runtime.Object,
) (admission.Warnings, error) {
	_, ok := obj.(*scalev1.CardScalingRequest)
	if !ok {
		return nil, fmt.Errorf("expected a CardScalingRequest object but got %T", obj)
	}

	logger := log.FromContext(ctx).WithName("cardscalingrequest-validator")
	logger.Info("Validation for CardScalingRequest upon deletion")

	return nil, nil
}

func (v *CardScalingRequestCustomValidator) getValidLogicalServer(
	ctx context.Context, nodeName string,
) (*rpmv1.LogicalServer, error) {
	// Get Node resource
	node, err := v.ClientHelper.GetNode(ctx, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Node(%s): %w", nodeName, err)
	}

	// Get cluster name and namespace from Node annotations
	_, clusterNamespace, err := v.ClientHelper.GetClusterNameAndNamespaceFromNode(node)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info from node(%s): %w", nodeName, err)
	}

	// Get Machine resource from Node
	machine, err := v.ClientHelper.GetMachineFromNode(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to get Machine from node(%s): %w", nodeName, err)
	}

	// Get MachineDeployment resource from Machine
	machineDeployment, err := v.ClientHelper.GetMachineDeploymentFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get MachineDeployment from Machine(%s/%s): %w", clusterNamespace, machine.Name, err)
	}

	// Check if the MachineDeployment has ClusterAutoscaler annotations.
	// If it exists, CardScalingRequest is not allowed for this MachineDeployment.
	if v.ClientHelper.IsAutoScalingEnabledMachineDeployment(machineDeployment) {
		return nil, fmt.Errorf("expected MachineDeployment(%s/%s) has no annotations for ClusterAutoscaler, but it has them",
			machineDeployment.Namespace, machineDeployment.Name)
	}

	// Get DCIMachine resource from Machine
	dciMachine, err := v.ClientHelper.GetDCIMachineFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get DCIMachine from Machine(%s/%s): %w", clusterNamespace, machine.Name, err)
	}

	// Get LogicalServer resource from DCIMachine
	logicalServer, err := v.ClientHelper.GetLogicalServerFromDCIMachine(ctx, dciMachine)
	if err != nil {
		return nil, fmt.Errorf("failed to get LogicalServer from DCIMachine(%s/%s): %w", dciMachine.Namespace, dciMachine.Name, err)
	}

	// Ensure LogicalServer is of type ComposedServer
	if logicalServer.Status.AppliedSpec == nil || logicalServer.Status.AppliedSpec.Type != rpmv1.AppliedServerTypeComposed {
		return nil, fmt.Errorf("expected applied spec type of LogicalServer(%s/%s) to be '%s', but got spec '%v'",
			logicalServer.Namespace, logicalServer.Name, rpmv1.AppliedServerTypeComposed, logicalServer.Status.AppliedSpec)
	}

	// Ensure LogicalServer is in Provisioned state
	if logicalServer.Status.State != rpmv1.LogicalServerStateProvisioned {
		return nil, fmt.Errorf("expected LogicalServer(%s/%s) status.state to be '%s', but got '%s'",
			logicalServer.Namespace, logicalServer.Name, rpmv1.LogicalServerStateProvisioned, logicalServer.Status.State)
	}

	return logicalServer, nil
}

// validateScalingParameter is a function to validate GPU specifications in CardScalingRequest with the existing LogicalServer.
func (v *CardScalingRequestCustomValidator) validateScalingParameter(
	scalingParameter *scalev1.ScalingParameter, logicalServer *rpmv1.LogicalServer,
) error {
	// Generate the expected ServerSpec after scaling, if error occurs, the scaling parameters are invalid
	// This validation ignores GPU assignments and UUIDs in arguments.
	if _, err := utils.GenerateServerSpecAfterScaling(&logicalServer.Spec.ServerSpec, scalingParameter); err != nil {
		return fmt.Errorf("failed to generate expected serverSpec after scaling: %w", err)
	}

	// Validate GPU UUIDs if the operation is delete or replace
	if err := v.validateScalingParameterUUIDs(scalingParameter.UUIDs, logicalServer); err != nil {
		return fmt.Errorf("failed to validate GPU UUIDs in scaling parameters: %w", err)
	}

	return nil
}

// validateScalingParameterUUIDs validates that the GPU UUIDs specified in the CardScalingRequest
// exist in the GPU assignments of the LogicalServer.
func (*CardScalingRequestCustomValidator) validateScalingParameterUUIDs(
	uuids []string, logicalServer *rpmv1.LogicalServer,
) error {
	// Check if GPU UUIDs are specified in CardScalingRequest (only when delete or replace)
	if len(uuids) == 0 {
		return nil
	}

	// Get the GPU UUID list from LogicalServer.Spec.ServerSpec.GPU.Assignments and convert assignments to string slice
	if !utils.HasAllUUIDsInLogicalServer(uuids, logicalServer.Spec.ServerSpec.GPU) {
		return fmt.Errorf("expected all UUIDs in CardScalingRequest to exist in LogicalServer(%s/%s), but some do not. GPUSpec: %v",
			logicalServer.Namespace, logicalServer.Name, logicalServer.Spec.ServerSpec.GPU)
	}
	return nil
}
