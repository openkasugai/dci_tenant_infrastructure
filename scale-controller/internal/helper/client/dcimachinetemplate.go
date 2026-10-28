package client

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/internal/helper/utils"
)

const dciMachineTemplateKind = "DCIMachineTemplate"

// GetDCIMachineTemplate gets DCIMachineTemplate resource from management cluster.
func (h *clientHelperStruct) GetDCIMachineTemplate(
	ctx context.Context, name, namespace string,
) (*capdv1.DCIMachineTemplate, error) {
	dciMachineTemplate := &capdv1.DCIMachineTemplate{}
	if err := h.getManagementResource(ctx, dciMachineTemplate, name, namespace, dciMachineTemplateKind); err != nil {
		return nil, err
	}
	return dciMachineTemplate, nil
}

// GetDCIMachineTemplateNameFromMachineDeployment gets DCIMachineTemplate name from MachineDeployment spec.
func (*clientHelperStruct) GetDCIMachineTemplateNameFromMachineDeployment(
	md *capiv1.MachineDeployment,
) (string, error) {
	// Check MachineDeployment resource
	if md == nil {
		return "", errors.New("expected MachineDeployment is not nil, but got nil")
	}

	// Get DCIMachineTemplate name from MachineDeployment spec
	templateRef := md.Spec.Template.Spec.InfrastructureRef
	if templateRef.Kind != "DCIMachineTemplate" {
		return "", fmt.Errorf("failed to get DCIMachineTemplate name from MachineDeployment(%s) spec", md.Name)
	}

	return templateRef.Name, nil
}

// GetDCIMachineTemplateFromMachineDeployment gets DCIMachineTemplate resource from management cluster.
func (h *clientHelperStruct) GetDCIMachineTemplateFromMachineDeployment(
	ctx context.Context, md *capiv1.MachineDeployment, namespace string,
) (*capdv1.DCIMachineTemplate, error) {
	// Get DCIMachineTemplate name from MachineDeployment resource
	name, err := h.GetDCIMachineTemplateNameFromMachineDeployment(md)
	if err != nil {
		return nil, fmt.Errorf("failed to get DCIMachineTemplate name from MachineDeployment: %w", err)
	}

	// Get Machine resource from management cluster
	dciMachineTemplate, err := h.GetDCIMachineTemplate(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get DCIMachineTemplate resource: %w", err)
	}

	return dciMachineTemplate, nil
}

// listDCIMachineTemplate lists DCIMachineTemplate resources from management cluster.
func (h *clientHelperStruct) listDCIMachineTemplate(
	ctx context.Context, namespace string,
) (*capdv1.DCIMachineTemplateList, error) {
	dciMachineTemplateList := &capdv1.DCIMachineTemplateList{}
	if err := h.listManagementResources(ctx, dciMachineTemplateList, namespace, dciMachineTemplateKind); err != nil {
		return nil, err
	}
	return dciMachineTemplateList, nil
}

// filterDCIMachineTemplateListByCallback filters DCIMachineTemplate list by callback function.
func (*clientHelperStruct) filterDCIMachineTemplateListByCallback(
	templateList *capdv1.DCIMachineTemplateList, clbFiltering func(*capdv1.DCIMachineTemplate) bool,
) *capdv1.DCIMachineTemplateList {
	// Find allocatable DCIMachineTemplates
	var allocatableTemplates capdv1.DCIMachineTemplateList
	for idx := range templateList.Items {
		dmt := &templateList.Items[idx]
		if !clbFiltering(dmt) {
			continue
		}
		allocatableTemplates.Items = append(allocatableTemplates.Items, *dmt)
	}
	return &allocatableTemplates
}

// GetDCIMachineTemplatesAllocatable gets allocatable DCIMachineTemplate list or new unique name.
func (h *clientHelperStruct) GetDCIMachineTemplatesAllocatable(
	ctx context.Context, location *string, serverSpec *rpmv1.ServerSpec, namespace string,
) ([]client.Object, string, error) {
	// List DCIMachineTemplate resources from management cluster
	templateList, err := h.listDCIMachineTemplate(ctx, namespace)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list DCIMachineTemplate resources to get allocatable DCIMachineTemplates: %w", err)
	}

	// Prepare filtering callback function
	clbFilteringValid := func(template *capdv1.DCIMachineTemplate) bool {
		if !utils.IsSameServerSpec(serverSpec, &template.Spec.Template.Spec.ServerSpec) {
			return false
		}
		if !utils.IsSameLocation(location, template.Spec.Template.Spec.Location) {
			return false
		}
		return true
	}
	clbFilteringAllocatable := func(template *capdv1.DCIMachineTemplate) bool {
		if !utils.IsAllocatableServerSpec(serverSpec, &template.Spec.Template.Spec.ServerSpec) {
			return false
		}
		if !utils.IsAllocatableLocation(location, template.Spec.Template.Spec.Location) {
			return false
		}
		return true
	}

	var templateListFiltered *capdv1.DCIMachineTemplateList
	templateListFiltered = h.filterDCIMachineTemplateListByCallback(templateList, clbFilteringValid)
	// Return valid DCIMachineTemplates if exists
	if len(templateListFiltered.Items) != 0 {
		return h.convertDCIMachineTemplateListToClientObject(templateListFiltered), "", nil
	}
	templateListFiltered = h.filterDCIMachineTemplateListByCallback(templateList, clbFilteringAllocatable)
	// Return allocatable DCIMachineTemplates if exists
	if len(templateListFiltered.Items) != 0 {
		return h.convertDCIMachineTemplateListToClientObject(templateListFiltered), "", nil
	}

	// Create new unique name for DCIMachineTemplate
	name, err := utils.CreateRandomName("worker-dmt-", templateList)
	if err != nil {
		// This case is only for failsafe
		err = fmt.Errorf("failed to create name for DCIMachineTemplate: %w", err)
	}

	return nil, name, err
}

// GetDCIMachineTemplatesValid gets valid DCIMachineTemplate list or new unique name.
func (h *clientHelperStruct) GetDCIMachineTemplatesValid(
	ctx context.Context, location *string, serverSpec *rpmv1.ServerSpec, namespace string,
) ([]client.Object, string, error) {
	// List DCIMachineTemplate resources from management cluster
	templateList, err := h.listDCIMachineTemplate(ctx, namespace)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list DCIMachineTemplate resources to get allocatable DCIMachineTemplates: %w", err)
	}

	// Prepare filtering callback function
	clbFiltering := func(template *capdv1.DCIMachineTemplate) bool {
		if !utils.IsSameServerSpec(serverSpec, &template.Spec.Template.Spec.ServerSpec) {
			return false
		}
		if !utils.IsSameLocation(location, template.Spec.Template.Spec.Location) {
			return false
		}
		return true
	}

	templateListFiltered := h.filterDCIMachineTemplateListByCallback(templateList, clbFiltering)

	// Return allocatable DCIMachineTemplates if exists
	if len(templateListFiltered.Items) != 0 {
		return h.convertDCIMachineTemplateListToClientObject(templateListFiltered), "", nil
	}

	// Create new unique name for DCIMachineTemplate
	name, err := utils.CreateRandomName("worker-dmt-", templateList)
	if err != nil {
		// This case is only for failsafe
		err = fmt.Errorf("failed to create name for DCIMachineTemplate: %w", err)
	}

	return nil, name, err
}

// CreateDCIMachineTemplate creates DCIMachineTemplate resource in management cluster.
func (h *clientHelperStruct) CreateDCIMachineTemplate(
	ctx context.Context, dciMachineTemplate client.Object,
) error {
	return h.createManagementResource(ctx, dciMachineTemplate, dciMachineTemplateKind)
}

// CreateDCIMachineTemplateFromNodeScalingRequest creates DCIMachineTemplate resource from NodeScalingRequest.
func (h *clientHelperStruct) CreateDCIMachineTemplateFromNodeScalingRequest(
	ctx context.Context, name, namespace string, nsr *scalev1.NodeScalingRequest,
) error {
	dmt, err := utils.NewDCIMachineTemplateFromNodeScalingRequest(name, namespace, nsr)
	if err != nil {
		// This case is only for failsafe
		err = fmt.Errorf("failed to create DCIMachineTemplate spec from NodeScalingRequest: %w", err)
		return err
	}
	if err := h.CreateDCIMachineTemplate(ctx, dmt); err != nil {
		err = fmt.Errorf("failed to create DCIMachineTemplate(%s) in clusterNamespace(%s): %w", name, namespace, err)
		return err
	}
	return nil
}

// DeleteDCIMachineTemplate deletes DCIMachineTemplate resource from management cluster.
func (h *clientHelperStruct) DeleteDCIMachineTemplate(
	ctx context.Context, dciMachineTemplate *capdv1.DCIMachineTemplate,
) error {
	return h.deleteManagementResource(ctx, dciMachineTemplate, dciMachineTemplateKind)
}

// DeleteDCIMachineTemplateUnused is a function to delete DCIMachineTemplate if it is not referred by other MachineDeployment.
//
//nolint:dupl // similar code is used in DeleteDockerMachineTemplateUnused
func (h *clientHelperStruct) DeleteDCIMachineTemplateUnused(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	templateNamespace := md.Namespace
	templateName := md.Spec.Template.Spec.InfrastructureRef.Name

	// Get DCIMachineTemplate
	template, err := h.GetDCIMachineTemplate(ctx, templateName, templateNamespace)
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get DCIMachineTemplate(%s) in Namespace(%s): %w", templateName, templateNamespace, err)
	}

	// Check whether the DCIMachineTemplate is referred by other MachineDeployment
	isReferred, err := h.IsReferredDCIMachineTemplate(ctx, template, md, templateNamespace)
	if err != nil {
		return fmt.Errorf("failed to check whether the DCIMachineTemplate(%s) is free or not: %w", templateName, err)
	}
	if isReferred {
		return nil // other MachineDeployment is referring the DCIMachineTemplate
	}

	// Delete DCIMachineTemplate
	if err := h.DeleteDCIMachineTemplate(ctx, template); apierrors.IsNotFound(err) {
		return nil // not exist already
	} else if err != nil {
		return fmt.Errorf("failed to delete DCIMachineTemplate(%s) in Namespace(%s): %w", templateName, templateNamespace, err)
	}
	return nil // succeed to delete
}

// IsReferredDCIMachineTemplate checks whether the DCIMachineTemplate is free (not referred by any MachineDeployment) or not.
func (h *clientHelperStruct) IsReferredDCIMachineTemplate(
	ctx context.Context, dciMachineTemplate *capdv1.DCIMachineTemplate, md *capiv1.MachineDeployment, namespace string,
) (bool, error) {
	// List MachineDeployment referring to the DCIMachineTemplate
	mds, err := h.GetMachineDeploymentsFromTemplate(ctx, namespace, md, dciMachineTemplate)
	if err != nil {
		return false, fmt.Errorf("failed to get MachineDeployment List through client: %w", err)
	}

	// Check whether the DCIMachineTemplate is referred by any MachineDeployment
	if len(mds) == 0 {
		return false, nil // not referred by any MachineDeployment
	}
	return true, nil // referred by other MachineDeployment
}

// convertDCIMachineTemplateListToClientObject converts DCIMachineTemplateList to []client.Object.
func (*clientHelperStruct) convertDCIMachineTemplateListToClientObject(
	list *capdv1.DCIMachineTemplateList,
) []client.Object {
	var objList []client.Object
	if list == nil {
		return objList
	}
	for idx := range list.Items {
		objList = append(objList, &list.Items[idx])
	}
	return objList
}
