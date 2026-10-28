package client

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

const machineKind = "Machine"

// GetMachine gets Machine resource from management cluster.
func (h *clientHelperStruct) GetMachine(
	ctx context.Context, name, namespace string,
) (*capiv1.Machine, error) {
	machine := &capiv1.Machine{}
	if err := h.getManagementResource(ctx, machine, name, namespace, machineKind); err != nil {
		return nil, err
	}
	return machine, nil
}

// getMachineNameFromNodeName gets Machine name from Node annotations.
func (*clientHelperStruct) getMachineNameFromNode(
	node *corev1.Node,
) (string, error) {
	// Check Node resource
	if node == nil {
		return "", errors.New("expected node is not nil, but got nil")
	}
	if node.Annotations == nil {
		return "", fmt.Errorf("expected Node(%s) annotations is not nil, but got nil", node.ObjectMeta.String())
	}

	// Get Machine name from Node annotations
	machineName, isFoundMachineName := node.Annotations[capiv1.MachineAnnotation]
	if !isFoundMachineName {
		return "", fmt.Errorf("failed to get Machine name from Node(%s) annotations", node.ObjectMeta.String())
	}

	return machineName, nil
}

// GetMachineFromNode gets Machine resource from Node annotations.
func (h *clientHelperStruct) GetMachineFromNode(
	ctx context.Context, node *corev1.Node,
) (*capiv1.Machine, error) {
	// Get Machine name from Node resource
	name, err := h.getMachineNameFromNode(node)
	if err != nil {
		return nil, fmt.Errorf("failed to get Machine name from Node: %w", err)
	}

	// Get namespace from Node resource
	_, namespace, err := h.GetClusterNameAndNamespaceFromNode(node)
	if err != nil {
		return nil, fmt.Errorf("failed to get Cluster info from Node: %w", err)
	}

	// Get Machine from Node resource
	machine, err := h.GetMachine(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get Machine resource: %w", err)
	}

	return machine, nil
}

// DeleteMachineFromMachine deletes Machine resource from management cluster.
func (h *clientHelperStruct) DeleteMachine(
	ctx context.Context, machine *capiv1.Machine,
) error {
	return h.deleteManagementResource(ctx, machine, machineKind)
}
