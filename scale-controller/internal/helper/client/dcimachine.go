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

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
)

const dciMachineKind = "DCIMachine"

// GetDCIMachine gets DCIMachine resource from management cluster.
func (h *clientHelperStruct) GetDCIMachine(
	ctx context.Context, name, namespace string,
) (*capdv1.DCIMachine, error) {
	dciMachine := &capdv1.DCIMachine{}
	if err := h.getManagementResource(ctx, dciMachine, name, namespace, dciMachineKind); err != nil {
		return nil, err
	}
	return dciMachine, nil
}

// getDCIMachineNameFromMachine gets DCIMachine name from Machine spec.
func (*clientHelperStruct) getDCIMachineNameFromMachine(
	machine *capiv1.Machine,
) (string, error) {
	// Check Machine resource
	if machine == nil {
		return "", errors.New("expected machine is not nil, but got nil")
	}

	// Get DCIMachine name from Machine spec
	templateRef := machine.Spec.InfrastructureRef
	if templateRef.Kind != dciMachineKind {
		return "", fmt.Errorf("expected InfrastructureRef kind for Machine(%s/%s) is %s, but got %s",
			machine.Namespace, machine.Name, dciMachineKind, templateRef.Kind)
	}

	return templateRef.Name, nil
}

// GetDCIMachineFromMachine gets DCIMachine resource from management cluster.
func (h *clientHelperStruct) GetDCIMachineFromMachine(
	ctx context.Context, machine *capiv1.Machine, namespace string,
) (*capdv1.DCIMachine, error) {
	// Get DCIMachine name from Machine resource
	name, err := h.getDCIMachineNameFromMachine(machine)
	if err != nil {
		return nil, fmt.Errorf("failed to get DCIMachine name from Machine: %w", err)
	}

	// Get DCIMachine resource from management cluster
	dciMachine, err := h.GetDCIMachine(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get DCIMachine resource: %w", err)
	}

	return dciMachine, nil
}

// PatchDCIMachine patches DCIMachine resource in management cluster.
func (h *clientHelperStruct) PatchDCIMachine(
	ctx context.Context, dciMachine, lastDCIMachine *capdv1.DCIMachine,
) error {
	return h.patchManagementResource(ctx, dciMachine, lastDCIMachine, dciMachineKind)
}
