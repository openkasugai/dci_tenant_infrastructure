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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// syncCachedResourceVersion is a function to synchronize resource version of cached object after updating status of the object.
// This prevents that the r.Get() reads stale object from cache at the beginning of the next reconcile loop.
func syncCachedResourceVersion(
	ctx context.Context, r client.Reader, obj, tmpObj client.Object,
) error {
	// Wait until cache is synchronized.
	patchedResourceVersion := obj.GetResourceVersion()
	if err := wait.PollUntilContextCancel(ctx, pollingIntervalSyncResourceVersion, true, func(ctx context.Context) (bool, error) {
		if getErr := r.Get(ctx, client.ObjectKeyFromObject(obj), tmpObj); getErr != nil {
			return false, getErr
		}
		// Accept if the resource version is the same or newer
		// ResourceVersion is a string, but it is a monotonically increasing number in string form.
		// Comparing as strings is safe because Kubernetes guarantees that ResourceVersion increases lexicographically.
		return (tmpObj.GetResourceVersion() >= patchedResourceVersion), nil
	}); err != nil {
		return err
	}

	return nil
}
