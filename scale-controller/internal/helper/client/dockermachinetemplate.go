package client

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/internal/helper/utils"
)

const dockerMachineTemplateKind = "DockerMachineTemplate"

// GetDockerMachineTemplate gets DockerMachineTemplate resource from management cluster.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func (h *clientHelperStruct) GetDockerMachineTemplate(
	ctx context.Context, name, namespace string,
) (*capdockerv1.DockerMachineTemplate, error) {
	dockerMachineTemplate := &capdockerv1.DockerMachineTemplate{}
	if err := h.getManagementResource(ctx, dockerMachineTemplate, name, namespace, dockerMachineTemplateKind); err != nil {
		return nil, err
	}
	return dockerMachineTemplate, nil
}

// listDockerMachineTemplate lists DockerMachineTemplate resources from management cluster.
func (h *clientHelperStruct) listDockerMachineTemplate(
	ctx context.Context, namespace string,
) (*capdockerv1.DockerMachineTemplateList, error) {
	dockerMachineTemplateList := &capdockerv1.DockerMachineTemplateList{}
	if err := h.listManagementResources(ctx, dockerMachineTemplateList, namespace, dockerMachineTemplateKind); err != nil {
		return nil, err
	}
	return dockerMachineTemplateList, nil
}

// GetDockerMachineTemplatesAllocatable gets allocatable DockerMachineTemplate list or new unique name.
func (h *clientHelperStruct) GetDockerMachineTemplatesAllocatable(
	ctx context.Context, name, namespace string,
) ([]client.Object, string, error) {
	// List DockerMachineTemplate resources from management cluster
	templateList, err := h.listDockerMachineTemplate(ctx, namespace)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list DockerMachineTemplate resources to get allocatable DockerMachineTemplates: %w", err)
	}

	// Find allocatable DockerMachineTemplates
	var allocatableTemplates []client.Object
	for idx := range templateList.Items {
		dmt := &templateList.Items[idx]
		if dmt.Name != name {
			continue // not allocatable
		}
		// Target DockerMachineTemplate found
		allocatableTemplates = append(allocatableTemplates, dmt)
		break
	}

	// Return allocatable DockerMachineTemplates if exists
	if len(allocatableTemplates) != 0 {
		return allocatableTemplates, "", nil
	}

	return nil, name, err
}

// CreateDockerMachineTemplate creates DockerMachineTemplate resource in management cluster.
func (h *clientHelperStruct) CreateDockerMachineTemplate(
	ctx context.Context, dockerMachineTemplate client.Object,
) error {
	return h.createManagementResource(ctx, dockerMachineTemplate, dockerMachineTemplateKind)
}

// CreateDockerMachineTemplateFromNodeScalingRequest creates DockerMachineTemplate resource from NodeScalingRequest.
func (h *clientHelperStruct) CreateDockerMachineTemplateFromNodeScalingRequest(
	ctx context.Context, name, namespace string, nsr *scalev1.NodeScalingRequest,
) error {
	dmt, err := utils.NewDockerMachineTemplateFromNodeScalingRequest(name, namespace, nsr)
	if err != nil {
		// This case is only for failsafe
		err = fmt.Errorf("failed to create DockerMachineTemplate spec from NodeScalingRequest: %w", err)
		return err
	}
	if err := h.CreateDockerMachineTemplate(ctx, dmt); err != nil {
		err = fmt.Errorf("failed to create DockerMachineTemplate(%s) in clusterNamespace(%s): %w", name, namespace, err)
		return err
	}
	return nil
}

// DeleteDockerMachineTemplate deletes DockerMachineTemplate resource from management cluster.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func (h *clientHelperStruct) DeleteDockerMachineTemplate(
	ctx context.Context, dockerMachineTemplate *capdockerv1.DockerMachineTemplate,
) error {
	return h.deleteManagementResource(ctx, dockerMachineTemplate, dockerMachineTemplateKind)
}

// DeleteDCIMachineTemplateUnused is a function to delete DCIMachineTemplate if it is not referred by other MachineDeployment.
//
//nolint:dupl // similar code is used in DeleteDCIMachineTemplateUnused
func (h *clientHelperStruct) DeleteDockerMachineTemplateUnused(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	templateNamespace := md.Namespace
	templateName := md.Spec.Template.Spec.InfrastructureRef.Name

	// Get DockerMachineTemplate
	template, err := h.GetDockerMachineTemplate(ctx, templateName, templateNamespace)
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get DockerMachineTemplate(%s) in Namespace(%s): %w", templateName, templateNamespace, err)
	}

	// Check if the DockerMachineTemplate is referred by other MachineDeployment
	isReferred, err := h.IsReferredDockerMachineTemplate(ctx, template, md, templateNamespace)
	if err != nil {
		return fmt.Errorf("failed to check whether DockerMachineTemplate(%s) is referred by other MachineDeployment: %w", templateName, err)
	}
	if isReferred {
		return nil // other MachineDeployment is using the MachineTemplate
	}

	// Delete DockerMachineTemplate
	if err := h.DeleteDockerMachineTemplate(ctx, template); apierrors.IsNotFound(err) {
		return nil // not exist already
	} else if err != nil {
		return fmt.Errorf("failed to delete DockerMachineTemplate(%s) in Namespace(%s): %w", templateName, templateNamespace, err)
	}
	return nil // succeed to delete
}

// IsReferredDockerMachineTemplate checks whether the DockerMachineTemplate is free or not.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func (h *clientHelperStruct) IsReferredDockerMachineTemplate(
	ctx context.Context, template *capdockerv1.DockerMachineTemplate, md *capiv1.MachineDeployment, namespace string,
) (bool, error) {
	// Get all MachineDeployment in the namespace with the template
	mds, err := h.GetMachineDeploymentsFromTemplate(ctx, namespace, md, template)
	if err != nil {
		return false, fmt.Errorf("failed to request MachineDeployment List through client: %w", err)
	}

	if len(mds) == 0 {
		return false, nil // not referred by any MachineDeployment
	}
	return true, nil // referred by other MachineDeployment
}
