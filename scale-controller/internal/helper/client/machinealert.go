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

	rpmv1 "resource-pool-manager/api/v1"
)

const machineAlertKind = "MachineAlert"

// GetMachineAlert gets MachineAlert resource from management cluster.
func (h *clientHelperStruct) GetMachineAlert(
	ctx context.Context, name, namespace string,
) (*rpmv1.MachineAlert, error) {
	machineAlert := &rpmv1.MachineAlert{}
	if err := h.getManagementResource(ctx, machineAlert, name, namespace, machineAlertKind); err != nil {
		return nil, err
	}
	return machineAlert, nil
}

// ListMachineAlert lists MachineAlert resources from management cluster.
func (h *clientHelperStruct) ListMachineAlert(
	ctx context.Context, namespace string,
) (*rpmv1.MachineAlertList, error) {
	machineAlertList := &rpmv1.MachineAlertList{}
	if err := h.listManagementResources(ctx, machineAlertList, namespace, machineAlertKind); err != nil {
		return nil, err
	}
	return machineAlertList, nil
}

// CreateMachineAlert creates a MachineAlert resource in the management cluster.
func (h *clientHelperStruct) CreateMachineAlert(
	ctx context.Context, machineAlert *rpmv1.MachineAlert,
) error {
	return h.createManagementResource(ctx, machineAlert, machineAlertKind)
}

// DeleteMachineAlert deletes a MachineAlert resource in the management cluster.
func (h *clientHelperStruct) DeleteMachineAlert(
	ctx context.Context, machineAlert *rpmv1.MachineAlert,
) error {
	return h.deleteManagementResource(ctx, machineAlert, machineAlertKind)
}
