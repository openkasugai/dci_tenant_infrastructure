package v1alpha1

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
)

//nolint:lll // kubebuilder marker cannot be split into multiple lines.
// +kubebuilder:webhook:path=/validate-cluster-my-domain-v1alpha1-nodescalingrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=cluster.my.domain,resources=nodescalingrequests,verbs=create;update,versions=v1alpha1,name=vnodescalingrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// NodeScalingRequestCustomValidator struct is responsible for validating the NodeScalingRequest resource
// when it is created, updated, or deleted.
type NodeScalingRequestCustomValidator struct {
	// ClientHelper for workload cluster and management cluster
	ClientHelper helperclient.ClientHelper
}

var _ admission.Validator[*scalev1.NodeScalingRequest] = &NodeScalingRequestCustomValidator{}

// NewNodeScalingRequestWebhookManagedBy is a variable to allow for easier testing of the webhook registration.
// It can be overridden in tests to provide a mock implementation.
var NewNodeScalingRequestWebhookManagedBy = func(mgr ctrl.Manager) *builder.WebhookBuilder[*scalev1.NodeScalingRequest] {
	return ctrl.NewWebhookManagedBy(mgr, &scalev1.NodeScalingRequest{})
}

// SetupWithManager registers the webhook for NodeScalingRequest in the manager.
func (v *NodeScalingRequestCustomValidator) SetupWithManager(
	mgr ctrl.Manager,
) error {
	return NewNodeScalingRequestWebhookManagedBy(mgr).
		WithValidator(v).
		Complete()
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type NodeScalingRequest.
func (v *NodeScalingRequestCustomValidator) ValidateCreate(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (admission.Warnings, error) {
	if nsr == nil {
		// This case is only for failsafe
		return nil, errors.New("expected a NodeScalingRequest is not nil, but got nil")
	}

	logger := log.FromContext(ctx).WithName("nodescalingrequest-validator")
	logger.Info("Validation for NodeScalingRequest upon creation")

	var errList field.ErrorList

	if nsr.Spec.NodeRequestType == scalev1.TypeDelete {
		if err := v.validateNode(ctx, nsr.Spec.Node); err != nil {
			errList = append(errList, field.Invalid(field.NewPath("spec").Child("node"), nsr.Spec.Node, err.Error()))
		}
	}

	if len(errList) > 0 {
		return nil, apierrors.NewInvalid(schema.GroupKind{
			Group: scalev1.GroupVersion.Group,
			Kind:  "NodeScalingRequest",
		}, nsr.Name, errList)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type NodeScalingRequest.
func (*NodeScalingRequestCustomValidator) ValidateUpdate(
	ctx context.Context, _, newNsr *scalev1.NodeScalingRequest,
) (admission.Warnings, error) {
	if newNsr == nil {
		return nil, fmt.Errorf("expected a NodeScalingRequest object for the newObj but got nil")
	}

	logger := log.FromContext(ctx).WithName("nodescalingrequest-validator")
	logger.Info("Validation for NodeScalingRequest upon update")

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type NodeScalingRequest.
// This function will never be called because we set the webhook to ignore delete operations in the kubebuilder marker.
// However, we implement this method to satisfy the webhook.CustomValidator interface.
func (*NodeScalingRequestCustomValidator) ValidateDelete(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (admission.Warnings, error) {
	if nsr == nil {
		return nil, fmt.Errorf("expected a NodeScalingRequest object but got nil")
	}

	logger := log.FromContext(ctx).WithName("nodescalingrequest-validator")
	logger.Info("Validation for NodeScalingRequest upon deletion")

	return nil, nil
}

// validateNode validates that the Node specified in the NodeScalingRequest exists and is associated
// with a MachineDeployment that has auto-scaling disabled.
func (v *NodeScalingRequestCustomValidator) validateNode(
	ctx context.Context, nodeName *string,
) error {
	if nodeName == nil {
		// This case is only for failsafe, but should be validated by CRD schema
		return errors.New("expected node name is not nil, but got nil")
	}

	node, err := v.ClientHelper.GetNode(ctx, *nodeName)
	if err != nil {
		return fmt.Errorf("failed to get Node(%s): %w", *nodeName, err)
	}

	_, clusterNamespace, err := v.ClientHelper.GetClusterNameAndNamespaceFromNode(node)
	if err != nil {
		return fmt.Errorf("failed to get Cluster namespace from Node(%s): %w", *nodeName, err)
	}

	machine, err := v.ClientHelper.GetMachineFromNode(ctx, node)
	if err != nil {
		return fmt.Errorf("failed to get Machine from Node(%s): %w", *nodeName, err)
	}

	md, err := v.ClientHelper.GetMachineDeploymentFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return fmt.Errorf("failed to get MachineDeployment from Machine(%s/%s): %w", machine.Namespace, machine.Name, err)
	}

	if v.ClientHelper.IsAutoScalingEnabledMachineDeployment(md) {
		return fmt.Errorf("expected the MachineDeployment(%s/%s) associated with the Node(%s) has auto-scaling disabled, "+
			"but has auto-scaling enabled", md.Namespace, md.Name, *nodeName)
	}

	return nil
}
