package client

import (
	"context"
	"fmt"

	resourcev1 "k8s.io/api/resource/v1beta2"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDeviceTaintRule lists DeviceTaintRule resources from workload cluster.
func (h *clientHelperStruct) ListDeviceTaintRule(
	ctx context.Context,
) (*resourcev1.DeviceTaintRuleList, error) {
	// List DeviceTaintRule resources from workload cluster
	dtrList := &resourcev1.DeviceTaintRuleList{}
	listOptions := &client.ListOptions{}
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
