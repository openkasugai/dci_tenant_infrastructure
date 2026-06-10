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

	corev1 "k8s.io/api/core/v1"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNode gets Node resource from workload cluster.
func (h *clientHelperStruct) GetNode(
	ctx context.Context, nodeName string,
) (*corev1.Node, error) {
	// Get Node resource from workload cluster
	node := &corev1.Node{}
	nodeKey := client.ObjectKey{
		Name: nodeName,
	}
	if err := h.Get(ctx, nodeKey, node); err != nil {
		return nil, fmt.Errorf("failed to get Node resource: %s: %w", nodeName, err)
	}
	return node, nil
}

// GetClusterNameAndNamespaceFromNode gets ClusterName and ClusterNamespace from Node labels.
//
//nolint:gocritic // ignore named returns
func (*clientHelperStruct) GetClusterNameAndNamespaceFromNode(
	node *corev1.Node,
) (string, string, error) {
	// check Node resource
	if node == nil {
		return "", "", errors.New("expected node is not nil, but got nil")
	}
	if node.Annotations == nil {
		return "", "", fmt.Errorf("expected Node(%s) annotations is not nil, but got nil", node.Name)
	}

	// Get ClusterNamespace and ClusterName from annotations
	clusterNamespace, isFoundNamespace := node.Annotations[capiv1.ClusterNamespaceAnnotation]
	clusterName, isFoundName := node.Annotations[capiv1.ClusterNameAnnotation]
	if !isFoundNamespace || !isFoundName {
		return "", "", fmt.Errorf("failed to get Cluster info from Node(%s) annotations: Namespace(%s), Name(%s)",
			node.Name, clusterNamespace, clusterName)
	}

	return clusterName, clusterNamespace, nil
}

// listNode lists Node resources from workload cluster.
func (h *clientHelperStruct) listNode(
	ctx context.Context,
) (*corev1.NodeList, error) {
	// List Node resources from workload cluster
	nodeList := &corev1.NodeList{}
	if err := h.List(ctx, nodeList); err != nil {
		return nil, fmt.Errorf("failed to list Nodes: %w", err)
	}

	return nodeList, nil
}

// GetClusterNameAndNamespaceFromNode gets ClusterName and ClusterNamespace from Node labels.
//
//nolint:gocritic // ignore named returns
func (h *clientHelperStruct) GetClusterNameAndNamespace(
	ctx context.Context,
) (string, string, error) {
	// List Node resources from workload cluster
	nodeList, err := h.listNode(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to list Node resources to get Cluster info: %w", err)
	}

	for idx := range nodeList.Items {
		node := &nodeList.Items[idx]
		clusterName, clusterNamespace, err := h.GetClusterNameAndNamespaceFromNode(node)
		if err == nil {
			return clusterName, clusterNamespace, nil
		}
	}

	return "", "", errors.New("failed to get Cluster info from any Node annotations")
}
