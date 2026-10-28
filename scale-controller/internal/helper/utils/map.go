package utils

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"

	scalev1 "scale-controller/api/v1alpha1"
)

// insertToMap inserts key and value to map and returns the map.
func insertToMap(
	m map[string]string,
	key, value string,
) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m[key] = value
	return m
}

// deleteFromMap deletes key from map.
func deleteFromMap(
	m map[string]string,
	key string,
) {
	delete(m, key)
}

// isMatchInMap checks if key exists in map.
// If value is nil, it only checks if key exists.
// If value is not nil, it checks if key exists and its value matches the given value.
func isMatchInMap(
	m map[string]string,
	key string,
	value *string,
) bool {
	if m == nil {
		return false
	}
	v, exists := m[key]
	return exists && (value == nil || v == *value)
}

// hasKeyInMap checks if key exists in map.
func hasKeyInMap(
	m map[string]string,
	key string,
) bool {
	return isMatchInMap(m, key, nil)
}

// isSameValueInMap compares value of key in map.
func isSameValueInMap(
	m map[string]string,
	key, value string,
) bool {
	return isMatchInMap(m, key, &value)
}

// AddNodeScalingRequestInProgressUIDLabel adds InProgressUIDLabel to labels.
func AddNodeScalingRequestInProgressUIDLabel(
	labels map[string]string,
	nsrUID types.UID,
) map[string]string {
	return insertToMap(labels, scalev1.InProgressUIDLabel, string(nsrUID))
}

// RemoveNodeScalingRequestInProgressUIDLabel removes InProgressUIDLabel from labels.
func RemoveNodeScalingRequestInProgressUIDLabel(
	labels map[string]string,
) {
	deleteFromMap(labels, scalev1.InProgressUIDLabel)
}

// AddCardScalingRequestInProgressUIDAnnotation adds CardScalingRequestInProgressUIDAnnotation to annotations.
func AddCardScalingRequestInProgressUIDAnnotation(
	annotations map[string]string,
	csrUID types.UID,
) map[string]string {
	return insertToMap(annotations, scalev1.CardScalingRequestInProgressUIDAnnotation, string(csrUID))
}

// RemoveCardScalingRequestInProgressUIDAnnotation removes CardScalingRequestInProgressUIDAnnotation from annotations.
func RemoveCardScalingRequestInProgressUIDAnnotation(
	annotations map[string]string,
) {
	deleteFromMap(annotations, scalev1.CardScalingRequestInProgressUIDAnnotation)
}

// HasCardScalingRequestInProgressUIDAnnotation checks if CardScalingRequestInProgressUIDAnnotation exists in annotations.
func HasCardScalingRequestInProgressUIDAnnotation(
	annotations map[string]string,
) bool {
	return hasKeyInMap(annotations, scalev1.CardScalingRequestInProgressUIDAnnotation)
}

// IsSameCardScalingRequestInProgressUIDAnnotation checks if CardScalingRequestInProgressUIDAnnotation matches the given value.
func IsSameCardScalingRequestInProgressUIDAnnotation(
	annotations map[string]string,
	csrUID types.UID,
) bool {
	return isSameValueInMap(annotations, scalev1.CardScalingRequestInProgressUIDAnnotation, string(csrUID))
}

// AddClusterAPIPausedAnnotation adds paused annotation to annotations.
func AddClusterAPIPausedAnnotation(
	annotations map[string]string,
) map[string]string {
	return insertToMap(annotations, capiv1.PausedAnnotation, "true")
}

// RemoveClusterAPIPausedAnnotation removes paused annotation from annotations.
func RemoveClusterAPIPausedAnnotation(
	annotations map[string]string,
) {
	deleteFromMap(annotations, capiv1.PausedAnnotation)
}

// AddDCIMachineCardScalingInfoAnnotation adds DCIMachineCardScalingInfoAnnotation to annotations.
func AddDCIMachineCardScalingInfoAnnotation(
	annotations map[string]string,
	cardScalingInfo string,
) map[string]string {
	return insertToMap(annotations, capdv1.DCIMachineAnnotationCardScalingInfoName, cardScalingInfo)
}

// RemoveDCIMachineCardScalingInfoAnnotation removes DCIMachineCardScalingInfoAnnotation from annotations.
func RemoveDCIMachineCardScalingInfoAnnotation(
	annotations map[string]string,
) {
	deleteFromMap(annotations, capdv1.DCIMachineAnnotationCardScalingInfoName)
}

// HasDCIMachineCardScalingInfoAnnotation checks if DCIMachine has CardScalingInfo annotation.
func HasDCIMachineCardScalingInfoAnnotation(
	annotations map[string]string,
) bool {
	return hasKeyInMap(annotations, capdv1.DCIMachineAnnotationCardScalingInfoName)
}

// IsSameDCIMachineCardScalingInfoAnnotation checks if DCIMachine's CardScalingInfo annotation matches the given value.
func IsSameDCIMachineCardScalingInfoAnnotation(
	annotations map[string]string,
	cardScalingInfo string,
) bool {
	return isSameValueInMap(annotations, capdv1.DCIMachineAnnotationCardScalingInfoName, cardScalingInfo)
}

// IsSameCardAlertCreatedByAnnotations checks if annotations have the same
// CreatedByNameLabel and CreatedByNamespaceLabel values as the given name and namespace.
func IsSameCardAlertCreatedByAnnotations(
	annotations map[string]string,
	name, namespace string,
) bool {
	return isSameValueInMap(annotations, scalev1.CreatedByNameLabel, name) &&
		isSameValueInMap(annotations, scalev1.CreatedByNamespaceLabel, namespace)
}

// GenerateDCIMachineCardScalingInfoAnnotation generates annotation for card scaling info for DCIMachine resource.
func GenerateDCIMachineCardScalingInfoAnnotation(
	name string,
	op capdv1.DCIMachineCardScalingOp,
	uuids []string,
) (string, error) {
	// Generate annotation info
	cardInfo := &capdv1.DCIMachineAnnotationCardScalingInfo{
		NewMachineDeploymentName: name,
		Op:                       op,
		UUIDs:                    uuids,
	}

	// Marshal to JSON string
	b, err := json.Marshal(cardInfo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal annotation info: %w", err)
	}
	return string(b), nil
}
