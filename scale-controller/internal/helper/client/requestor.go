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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getManagementResource gets a Kubernetes resource of any type from the management cluster.
func (h *clientHelperStruct) getManagementResource(
	ctx context.Context, obj client.Object, name, namespace, kind string,
) error {
	if obj == nil {
		return errors.New("expected object to get resource is not nil, but got nil")
	}
	logger := log.FromContext(ctx).WithValues("name", name, "namespace", namespace, "kind", kind)
	logger.Info("Attempting to get resource")
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}
	if err := h.managementClusterClient.Get(ctx, key, obj); err != nil {
		return fmt.Errorf("failed to get resource(%s) of kind %s: %w", key.String(), kind, err)
	}
	logger.Info("Successfully got resource")
	return nil
}

// listManagementResources lists Kubernetes resources of any type from the management cluster.
func (h *clientHelperStruct) listManagementResources(
	ctx context.Context, objList client.ObjectList, namespace, kind string,
) error {
	if objList == nil {
		return errors.New("expected object list to list resources is not nil, but got nil")
	}
	logger := log.FromContext(ctx).WithValues("namespace", namespace, "kind", kind)
	logger.Info("Attempting to list resources")
	listOptions := &client.ListOptions{
		Namespace: namespace,
	}
	if err := h.managementClusterClient.List(ctx, objList, listOptions); err != nil {
		return fmt.Errorf("failed to list resources of kind %s in namespace %s: %w", kind, namespace, err)
	}
	logger.Info("Successfully listed resources")
	return nil
}

// createManagementResource creates a Kubernetes resource of any type in the management cluster.
//
//nolint:dupl // Similar code exists in deleteManagementResource.
func (h *clientHelperStruct) createManagementResource(
	ctx context.Context, obj client.Object, kind string,
) error {
	if obj == nil {
		return errors.New("expected object to create resource is not nil, but got nil")
	}
	logger := log.FromContext(ctx).WithValues("name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", kind)
	logger.Info("Attempting to create resource")
	if err := h.managementClusterClient.Create(ctx, obj); err != nil {
		return fmt.Errorf("failed to create resource(%s/%s) of kind %s: %w", obj.GetNamespace(), obj.GetName(), kind, err)
	}
	logger.Info("Successfully created resource")
	return nil
}

// deleteManagementResource deletes a Kubernetes resource of any type from the management cluster.
//
//nolint:dupl // Similar code exists in createManagementResource.
func (h *clientHelperStruct) deleteManagementResource(
	ctx context.Context, obj client.Object, kind string,
) error {
	if obj == nil {
		return errors.New("expected object to delete resource is not nil, but got nil")
	}
	logger := log.FromContext(ctx).WithValues("name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", kind)
	logger.Info("Attempting to delete resource")
	if err := h.managementClusterClient.Delete(ctx, obj); err != nil {
		return fmt.Errorf("failed to delete resource(%s/%s) of kind %s: %w", obj.GetNamespace(), obj.GetName(), kind, err)
	}
	logger.Info("Successfully deleted resource")
	return nil
}

// patchManagementResource patches a Kubernetes resource of any type in the management cluster.
func (h *clientHelperStruct) patchManagementResource(
	ctx context.Context, obj, lastObj client.Object, kind string,
) error {
	if obj == nil || lastObj == nil {
		return errors.New("expected object to patch resource is not nil, but got nil")
	}
	logger := log.FromContext(ctx).WithValues("name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", kind)
	logger.Info("Attempting to patch resource")
	if err := h.managementClusterClient.Patch(ctx, obj, client.MergeFrom(lastObj)); err != nil {
		return fmt.Errorf("failed to patch resource(%s/%s) of kind %s: %w", obj.GetNamespace(), obj.GetName(), kind, err)
	}
	logger.Info("Successfully patched resource")
	return nil
}
