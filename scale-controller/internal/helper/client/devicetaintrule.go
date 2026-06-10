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

	resourcev1 "k8s.io/api/resource/v1alpha3"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDeviceTaintRule lists DeviceTaintRule resources from workload cluster.
func (h *clientHelperStruct) ListDeviceTaintRule(
	ctx context.Context, namespace string,
) (*resourcev1.DeviceTaintRuleList, error) {
	// List DeviceTaintRule resources from workload cluster
	dtrList := &resourcev1.DeviceTaintRuleList{}
	listOptions := &client.ListOptions{
		Namespace: namespace,
	}
	if err := h.List(ctx, dtrList, listOptions); err != nil {
		return nil, fmt.Errorf("failed to list DeviceTaintRules: %w", err)
	}

	return dtrList, nil
}

// CreateDeviceTaintRule creates a DeviceTaintRule resource in the workload cluster.
func (h *clientHelperStruct) CreateDeviceTaintRule(
	ctx context.Context, dtr *resourcev1.DeviceTaintRule,
) error {
	// Create DeviceTaintRule resource in the workload cluster
	if err := h.Create(ctx, dtr); err != nil {
		return fmt.Errorf("failed to create DeviceTaintRule: %w", err)
	}

	return nil
}
