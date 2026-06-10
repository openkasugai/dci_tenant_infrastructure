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
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

const machineSetKind = "MachineSet"

// listMachineSet is a function to get the MachineSet owned by a MachineDeployment from its uid.
func (h *clientHelperStruct) listMachineSet(
	ctx context.Context, namespace string,
) (*capiv1.MachineSetList, error) {
	machineSetList := &capiv1.MachineSetList{}
	if err := h.listManagementResources(ctx, machineSetList, namespace, machineSetKind); err != nil {
		return nil, err
	}
	return machineSetList, nil
}

// GetMachineSetFromUID is a function to get the MachineSet owned by a MachineDeployment from its uid.
func (r *clientHelperStruct) GetMachineSetFromUID(
	ctx context.Context, namespace string, mdUID types.UID,
) (*capiv1.MachineSet, error) {
	msList, err := r.listMachineSet(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get MachineSet owned by MachineDeployment with UID %s: %w", string(mdUID), err)
	}

	for idx := range msList.Items {
		m := &msList.Items[idx]
		if slices.ContainsFunc(m.OwnerReferences, func(ownerRef metav1.OwnerReference) bool {
			return ownerRef.UID == mdUID
		}) {
			return m, nil
		}
	}

	return nil, nil
}

// PatchMachineSet patches the MachineSet resource in management cluster.
func (h *clientHelperStruct) PatchMachineSet(
	ctx context.Context, ms *capiv1.MachineSet, lastMS *capiv1.MachineSet,
) error {
	return h.patchManagementResource(ctx, ms, lastMS, machineSetKind)
}
