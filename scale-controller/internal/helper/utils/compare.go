package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"

	resourcev1 "k8s.io/api/resource/v1beta2"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"
)

// Character set for random string.
const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// Length for suffix of workload MachineDeployment.
const defaultSuffixLength int = 5

// createRandomString is a function to create a random string for suffix.
func createRandomString(
	length int,
) string {
	lengthCreate := length
	if lengthCreate <= 0 {
		lengthCreate = defaultSuffixLength
	}
	randomByte := make([]byte, lengthCreate)
	for i := range randomByte {
		//nolint:gosec // This is acceptable to use math/rand here because this is not for security purpose.
		randomByte[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomByte)
}

// CreateRandomName is a function to create a random unique name with body string and random suffix.
func CreateRandomName[T *capiv1.MachineDeploymentList | *capdv1.DCIMachineTemplateList](
	body string, listExisting T,
) (string, error) {
	for range 10 {
		newName := body + createRandomString(defaultSuffixLength)
		switch v := any(listExisting).(type) {
		case *capiv1.MachineDeploymentList:
			if !slices.ContainsFunc(v.Items, func(md capiv1.MachineDeployment) bool {
				return md.Name == newName
			}) {
				return newName, nil // this is valid name
			}
		case *capdv1.DCIMachineTemplateList:
			if !slices.ContainsFunc(v.Items, func(dmt capdv1.DCIMachineTemplate) bool {
				return dmt.Name == newName
			}) {
				return newName, nil // this is valid name
			}
		default:
			// This case is only for failsafe, but should not happen because type assertion must pass without error.
			return "", fmt.Errorf("failed to convert to valid type %v", v)
		}
	}
	return "", errors.New("failed to create a valid name")
}

// isAllocatableSpec is a function to check request spec is allocatable in template spec.
// if request is nil, skip checking.
func isAllocatableSpec[T int | int32](
	request, template *T,
) bool {
	if request == nil {
		return true // no need to check
	}
	if template == nil {
		return false // when request is not nil, template must not be nil
	}
	if *request > *template {
		return false // compare value
	}
	return true
}

// isValidSpec is a function to check request spec is same as template spec.
// if request is nil, skip checking.
func isValidSpec[T comparable](
	request, template *T,
) bool {
	if request == nil {
		return true // no need to check
	}
	if template == nil {
		return false // when request is not nil, template must not be nil
	}
	if *request != *template {
		return false // compare value
	}
	return true
}

// isSameSpec is a function to check request spec is same as template spec.
func isSameSpec[T comparable](
	request, template *T,
) bool {
	if request == nil && template == nil {
		return true // both nil
	}
	if request == nil || template == nil {
		return false // must match nil status
	}
	if *request == *template {
		return true
	}
	return false
}

type containedState int

const (
	containedInSubset containedState = 1 << iota
	containedInSuperSet
	containedInBoth = containedInSubset | containedInSuperSet
)

// hasAllItemsInSlice is a function to check two slices contain same items,
// regardless of the order.
func hasAllItemsInSlice[T comparable](
	superset, subset []T,
) bool {
	if len(subset) == 0 {
		return true // no need to check
	}
	if len(superset) == 0 {
		return false // when subset is not empty, superset must not be empty
	}

	// Use a map to check occurrences of each item in subset and superset
	containedStateMap := make(map[T]containedState)
	for _, v := range subset {
		// set at least contained in subset
		containedStateMap[v] |= containedInSubset
	}
	for _, v := range superset {
		// set at least contained in superset
		containedStateMap[v] |= containedInSuperSet
	}
	for _, state := range containedStateMap {
		if state == containedInSubset {
			// return false if there is an item only contained in subset
			return false
		}
	}
	return true
}

// hasAnyItemsInSlice is a function to check if any items in subset are present in superset,
// regardless of the order.
func hasAnyItemsInSlice[T comparable](
	superset, subset []T,
) bool {
	if len(subset) == 0 {
		return false // return false because there is no common item
	}
	if len(superset) == 0 {
		return false // return false because there is no common item
	}

	// Use a map to check occurrences of each item in subset and superset
	containedStateMap := make(map[T]containedState)
	for _, v := range subset {
		// set at least contained in subset
		containedStateMap[v] |= containedInSubset
	}
	for _, v := range superset {
		// set at least contained in superset
		containedStateMap[v] |= containedInSuperSet
	}
	for _, state := range containedStateMap {
		if state == containedInBoth {
			// return true if there is an item contained in both subset and superset
			return true
		}
	}
	return false
}

// isValidGPUSpec is a function to check request GPUSpec is valid in template GPUSpec
// without comparing .assignments.
// if request is nil, skip checking.
func isValidGPUSpec(
	request, template *rpmv1.GPUSpec,
) bool {
	if request == nil {
		return true // no need to check
	}
	if template == nil {
		return false // when request is not nil, template must not be nil
	}
	if request.Model != template.Model {
		return false
	}
	if request.Num > template.Num {
		return false
	}
	return true
}

// isSameGPUSpec is a function to check request GPUSpec is same as template GPUSpec
// without comparing .assignments.
func isSameGPUSpec(
	request, template *rpmv1.GPUSpec,
) bool {
	if request == nil && template == nil {
		return true // both nil
	}
	if request == nil || template == nil {
		return false // must match nil status
	}
	if request.Model != template.Model {
		return false
	}
	if request.Num != template.Num {
		return false
	}
	return true
}

// IsAllocatableLocation is a function to check request location is allocatable in template location.
// if request is nil, skip checking.
func IsAllocatableLocation(
	request, template *string,
) bool {
	return isValidSpec(request, template)
}

// IsAllocatableServerSpec is a function to check request ServerSpec is allocatable in template ServerSpec.
// if each field in request is nil, skip checking it.
func IsAllocatableServerSpec(
	request, template *rpmv1.ServerSpec,
) bool {
	if request == nil {
		return true
	}
	if template == nil {
		return false
	}
	if request.Type != template.Type {
		return false
	}
	if !isValidSpec(request.OS, template.OS) {
		return false
	}
	if !isAllocatableSpec(request.MinCPU, template.MinCPU) {
		return false
	}
	if !isAllocatableSpec(request.MinMemory, template.MinMemory) {
		return false
	}
	if !isAllocatableSpec(request.MinStorage, template.MinStorage) {
		return false
	}
	if !isValidGPUSpec(request.GPU, template.GPU) {
		return false
	}
	if !isValidSpec(request.NIC, template.NIC) {
		return false
	}
	return true
}

// IsSameLocation is a function to check request location is same as template location.
func IsSameLocation(
	request, template *string,
) bool {
	return isSameSpec(request, template)
}

// IsSameServerSpec is a function to check request ServerSpec is same as template ServerSpec.
func IsSameServerSpec(
	request, template *rpmv1.ServerSpec,
) bool {
	if request == nil && template == nil {
		return true
	}
	if request == nil || template == nil {
		return false
	}
	if request.Type != template.Type {
		return false
	}
	if !isSameSpec(request.OS, template.OS) {
		return false
	}
	if !isSameSpec(request.MinCPU, template.MinCPU) {
		return false
	}
	if !isSameSpec(request.MinMemory, template.MinMemory) {
		return false
	}
	if !isSameSpec(request.MinStorage, template.MinStorage) {
		return false
	}
	if !isSameGPUSpec(request.GPU, template.GPU) {
		return false
	}
	if !isSameSpec(request.NIC, template.NIC) {
		return false
	}
	return true
}

// HasAllUUIDsInLogicalServer is a function to check all UUIDs in request are contained in GPU spec of LogicalServer.
func HasAllUUIDsInLogicalServer(
	uuids []string, gpuSpec *rpmv1.GPUSpec,
) bool {
	return hasAllItemsInSlice(convertGPUSpecToUUIDs(gpuSpec), uuids)
}

// HasAnyUUIDsInLogicalServer is a function to check any UUID in request is contained in GPU spec of LogicalServer.
func HasAnyUUIDsInLogicalServer(
	uuids []string, gpuSpec *rpmv1.GPUSpec,
) bool {
	return hasAnyItemsInSlice(convertGPUSpecToUUIDs(gpuSpec), uuids)
}

// convertGPUSpecToUUIDs is a function to convert GPU spec to UUID slice.
func convertGPUSpecToUUIDs(
	gpuSpec *rpmv1.GPUSpec,
) []string {
	if gpuSpec == nil {
		return nil
	}
	// Convert assignments to string slice
	if len(gpuSpec.Assignments) == 0 {
		return nil
	}
	uuids := make([]string, 0, len(gpuSpec.Assignments))
	for _, assignment := range gpuSpec.Assignments {
		uuids = append(uuids, assignment.UUID)
	}
	return uuids
}

// IsSameDeviceTaintSelector is a function to check if the DeviceTaintSelector is the same as the given spec.
func IsSameDeviceTaintSelector(
	s1, s2 *resourcev1.DeviceTaintSelector,
) (isSame bool) {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	return isSameSpec(s1.Driver, s2.Driver) &&
		isSameSpec(s1.Pool, s2.Pool) &&
		isSameSpec(s1.Device, s2.Device)
}
