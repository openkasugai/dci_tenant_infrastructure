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

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"
)

const logicalServerKind = "LogicalServer"

// GetLogicalServer gets LogicalServer resource from management cluster.
func (h *clientHelperStruct) GetLogicalServer(
	ctx context.Context, name, namespace string,
) (*rpmv1.LogicalServer, error) {
	logicalServer := &rpmv1.LogicalServer{}
	if err := h.getManagementResource(ctx, logicalServer, name, namespace, logicalServerKind); err != nil {
		return nil, err
	}
	return logicalServer, nil
}

// GetLogicalServerNameAndNamespaceFromDCIMachine gets LogicalServer name and namespace from DCIMachine status.
func (*clientHelperStruct) GetLogicalServerNameAndNamespaceFromDCIMachine(
	dciMachine *capdv1.DCIMachine,
) (name, namespace string, err error) {
	// Check DCIMachine resource and LogicalServerRef
	if dciMachine == nil || dciMachine.Status.LogicalServerRef == nil {
		return "", "", errors.New("expected DCIMachine is not nil, but got nil")
	}

	return dciMachine.Status.LogicalServerRef.Name, dciMachine.Status.LogicalServerRef.Namespace, nil
}

// GetLogicalServerFromDCIMachine retrieves LogicalServer from DCIMachine.
func (h *clientHelperStruct) GetLogicalServerFromDCIMachine(
	ctx context.Context, dciMachine *capdv1.DCIMachine,
) (*rpmv1.LogicalServer, error) {
	// Get LogicalServer name and namespace from DCIMachine resource
	name, namespace, err := h.GetLogicalServerNameAndNamespaceFromDCIMachine(dciMachine)
	if err != nil {
		return nil, fmt.Errorf("failed to get name and namespace of LogicalServer from dciMachine: %w", err)
	}

	// Get LogicalServer resource from management cluster
	logicalServer, err := h.GetLogicalServer(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get LogicalServer resource: %w", err)
	}

	return logicalServer, nil
}
