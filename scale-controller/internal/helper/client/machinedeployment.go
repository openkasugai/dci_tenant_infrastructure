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

package client

import (
	"context"
	"errors"
	"fmt"
	"slices"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/internal/helper/utils"
)

const machineDeploymentKind = "MachineDeployment"

// GetMachineDeployment gets MachineDeployment resource from management cluster.
func (h *clientHelperStruct) GetMachineDeployment(
	ctx context.Context, name, namespace string,
) (*capiv1.MachineDeployment, error) {
	machineDeployment := &capiv1.MachineDeployment{}
	if err := h.getManagementResource(ctx, machineDeployment, name, namespace, machineDeploymentKind); err != nil {
		return nil, err
	}
	return machineDeployment, nil
}

// getMachineDeploymentNameFromMachine gets MachineDeployment name from Machine labels.
func (*clientHelperStruct) getMachineDeploymentNameFromMachine(
	machine *capiv1.Machine,
) (string, error) {
	// Get MachineDeployment name from Machine labels
	if machine == nil {
		return "", errors.New("expected machine is not nil, but got nil")
	}
	if machine.Labels == nil {
		return "", fmt.Errorf("expected Machine(%s) labels is not nil, but got nil", machine.Name)
	}

	// Get MachineDeployment name from Machine labels
	machineDeploymentName, isFound := machine.Labels[capiv1.MachineDeploymentNameLabel]
	if !isFound {
		return "", fmt.Errorf("failed to get MachineDeployment name from Machine(%s) labels", machine.Name)
	}

	return machineDeploymentName, nil
}

// GetMachineDeploymentFromMachine gets MachineDeployment resource from management cluster.
func (h *clientHelperStruct) GetMachineDeploymentFromMachine(
	ctx context.Context, machine *capiv1.Machine, namespace string,
) (*capiv1.MachineDeployment, error) {
	// Get MachineDeployment name from Machine resource
	machineDeploymentName, err := h.getMachineDeploymentNameFromMachine(machine)
	if err != nil {
		return nil, fmt.Errorf("failed to get MachineDeployment name from Machine: %w", err)
	}

	// Get MachineDeployment resource from management cluster
	machineDeployment, err := h.GetMachineDeployment(ctx, machineDeploymentName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get MachineDeployment resource from management cluster: %w", err)
	}

	return machineDeployment, nil
}

// listMachineDeployments lists MachineDeployment resources from management cluster.
func (h *clientHelperStruct) listMachineDeployments(
	ctx context.Context, namespace string,
) (*capiv1.MachineDeploymentList, error) {
	mdList := &capiv1.MachineDeploymentList{}
	if err := h.listManagementResources(ctx, mdList, namespace, machineDeploymentKind); err != nil {
		return nil, err
	}
	return mdList, nil
}

// filterMachineDeploymentListByCallback filters MachineDeployment list by callback function.
func (*clientHelperStruct) filterMachineDeploymentListByCallback(
	mdList *capiv1.MachineDeploymentList,
	clbFiltering func(*capiv1.MachineDeployment) bool,
) *capiv1.MachineDeploymentList {
	var mdListFiltered capiv1.MachineDeploymentList
	for idx := range mdList.Items {
		md := &mdList.Items[idx]
		if !clbFiltering(md) {
			continue
		}
		mdListFiltered.Items = append(mdListFiltered.Items, *md)
	}
	return &mdListFiltered
}

// filterMachineDeploymentsInCluster is a function to get MachineDeployment resource and filter them by some conditions.
func (h *clientHelperStruct) filterMachineDeploymentsInCluster(
	mdList *capiv1.MachineDeploymentList, machineTemplateKind, clusterName string,
) []capiv1.MachineDeployment {
	// Prepare callback function for filtering MachineDeployment resources
	clbFilteringInCluster := func(md *capiv1.MachineDeployment) bool {
		if md.Spec.ClusterName != clusterName {
			return false
		}
		if md.Spec.Template.Spec.InfrastructureRef.Kind != machineTemplateKind {
			return false
		}
		if h.IsAutoScalingEnabledMachineDeployment(md) {
			return false
		}
		return true
	}

	// Filter MachineDeployment resources by in-cluster condition
	mdListInCluster := h.filterMachineDeploymentListByCallback(mdList, clbFilteringInCluster)

	return mdListInCluster.Items
}

// GetMachineDeploymentsAllocatable is a function to get MachineDeployment resource and filter them by some conditions.
//
//nolint:gocritic // ignore named returns
func (h *clientHelperStruct) GetMachineDeploymentsAllocatable(
	ctx context.Context, namespace, machineTemplateKind, clusterName string, machineTemplates []client.Object,
) ([]capiv1.MachineDeployment, string, *capiv1.MachineDeployment, error) {
	// List MachineDeployment resources from management cluster
	mdList, err := h.listMachineDeployments(ctx, namespace)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list MachineDeployment resources to get allocatable MachineDeployments: %w", err)
	}

	// List MachineDeployment resources filtered by in-cluster condition from management cluster
	mdsInCluster := h.filterMachineDeploymentsInCluster(mdList, machineTemplateKind, clusterName)
	if len(mdsInCluster) == 0 {
		// Return error if no MachineDeployment exists in the target cluster
		return nil, "", nil, fmt.Errorf("expected at least one MachineDeployment in the target cluster(%s/%s), but got none",
			namespace, clusterName)
	}
	mdListInCluster := &capiv1.MachineDeploymentList{Items: mdsInCluster}

	// Prepare callback function for filtering MachineDeployment resources
	clbFiltering := func(md *capiv1.MachineDeployment) bool {
		if !slices.ContainsFunc(machineTemplates, func(mt client.Object) bool {
			return md.Spec.Template.Spec.InfrastructureRef.Name == mt.GetName() // match MachineTemplate name
		}) {
			return false
		}
		return true
	}

	// Filter MachineDeployment resources by allocatable condition
	mdListFiltered := h.filterMachineDeploymentListByCallback(mdListInCluster, clbFiltering)

	// Return allocatable MachineDeployments if exists
	if len(mdListFiltered.Items) != 0 {
		return mdListFiltered.Items, "", nil, nil
	}

	// Create new unique name for MachineDeployment
	// should check with all MachineDeployments, but not only allocatable ones
	name, err := utils.CreateRandomName("worker-md-", mdList)
	if err != nil {
		err = fmt.Errorf("failed to create name for MachineDeployment: %w", err)
	}

	return nil, name, &mdsInCluster[0], err
}

// GetMachineDeploymentsFromTemplate gets MachineDeployment list that are used by NodeScalingRequest.
func (h *clientHelperStruct) GetMachineDeploymentsFromTemplate(
	ctx context.Context, namespace string, md *capiv1.MachineDeployment, template client.Object,
) ([]capiv1.MachineDeployment, error) {
	// List MachineDeployment resources from management cluster
	mdList, err := h.listMachineDeployments(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list MachineDeployment resources to get used MachineDeployments: %w", err)
	}

	// Prepare callback function for filtering MachineDeployment resources
	clbFilteringUsed := func(m *capiv1.MachineDeployment) bool {
		if md != nil && (m.Name == md.Name && m.Namespace == md.Namespace) {
			return false // skip the target MachineDeployment
		}
		if m.Namespace != template.GetNamespace() {
			return false
		}
		if m.Spec.Template.Spec.InfrastructureRef.Name != template.GetName() {
			return false
		}
		if h.IsAutoScalingEnabledMachineDeployment(m) {
			return false
		}
		return true // exist MachineDeployment using the MachineTemplate
	}

	// Filter MachineDeployment resources
	mdListFiltered := h.filterMachineDeploymentListByCallback(mdList, clbFilteringUsed)

	return mdListFiltered.Items, nil
}

// CreateMachineDeployment creates MachineDeployment resource in management cluster.
func (h *clientHelperStruct) CreateMachineDeployment(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	return h.createManagementResource(ctx, md, machineDeploymentKind)
}

// CreateMachineDeploymentFromNodeScalingRequest creates MachineDeployment resource from NodeScalingRequest.
func (h *clientHelperStruct) CreateMachineDeploymentFromNodeScalingRequest(
	ctx context.Context, name, namespace, dmtName string, nsr *scalev1.NodeScalingRequest, original *capiv1.MachineDeployment,
) error {
	md, err := utils.NewMachineDeploymentFromNodeScalingRequest(name, namespace, dmtName, nsr, original)
	if err != nil {
		// This case is only for failsafe
		err = fmt.Errorf("failed to create MachineDeployment spec from NodeScalingRequest: %w", err)
		return err
	}
	md.Labels = utils.AddNodeScalingRequestInProgressUIDLabel(md.Labels, nsr.GetUID())

	// create a new MachineDeployment
	if err := h.CreateMachineDeployment(ctx, md); err != nil {
		err = fmt.Errorf("failed to create MachineDeployment(%s): %w", md.ObjectMeta.String(), err)
		return err
	}

	return nil
}

// PatchMachineDeployment patches MachineDeployment resource in management cluster.
func (h *clientHelperStruct) PatchMachineDeployment(
	ctx context.Context, md, lastMD *capiv1.MachineDeployment,
) error {
	return h.patchManagementResource(ctx, md, lastMD, machineDeploymentKind)
}

// DeleteMachineDeployment deletes MachineDeployment resource from management cluster.
func (h *clientHelperStruct) DeleteMachineDeployment(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	return h.deleteManagementResource(ctx, md, machineDeploymentKind)
}

// IsLastMachineDeployment is a function to check if another valid machineDeployment exist or not
func (h *clientHelperStruct) IsLastMachineDeployment(
	ctx context.Context, md *capiv1.MachineDeployment, namespace, clusterName string,
) (bool, error) {
	// List MachineDeployment resources from management cluster
	mdList, err := h.listMachineDeployments(ctx, namespace)
	if err != nil {
		return false, fmt.Errorf("failed to list MachineDeployment resources to check last MachineDeployment: %w", err)
	}

	// get valid MachineDeployment List
	machineTemplateKind := md.Spec.Template.Spec.InfrastructureRef.Kind
	mdsInCluster := h.filterMachineDeploymentsInCluster(mdList, machineTemplateKind, clusterName)

	// check if there is the md in the list in case, should always pass this condition
	if !slices.ContainsFunc(mdsInCluster, func(m capiv1.MachineDeployment) bool {
		return md.Name == m.Name && md.Namespace == m.Namespace
	}) {
		return false, fmt.Errorf("failed to find the target MachineDeployment in the List(%v)", mdsInCluster)
	}

	// check if the target MachineDeployment is the last valid MachineDeployment or not
	return len(mdsInCluster) == 1, nil
}

// IsAutoScalingEnabledMachineDeployment is a function to check if MachineDeployment has auto-scaling annotations.
func (*clientHelperStruct) IsAutoScalingEnabledMachineDeployment(
	md *capiv1.MachineDeployment,
) bool {
	if md == nil || md.Annotations == nil {
		return false
	}
	_, isFoundMinAnnotation := md.Annotations[capiv1.AutoscalerMinSizeAnnotation]
	_, isFoundMaxAnnotation := md.Annotations[capiv1.AutoscalerMaxSizeAnnotation]
	if isFoundMinAnnotation || isFoundMaxAnnotation {
		return true
	}
	return false
}
