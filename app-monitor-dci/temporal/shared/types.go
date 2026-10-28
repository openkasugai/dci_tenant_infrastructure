package shared

import (
	"encoding/json"
)

// PodMetaData is the structure corresponding to the metadata field
type PodMetaData struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// PodSpec is the structure corresponding to the spec field
type PodSpec struct {
	Type string `json:"type"`
}

// AppStatus is the structure corresponding to the status.appStatus field
type AppStatus struct {
	// Status, metrics, and metadata are JSON objects.
	// Use json.RawMessage to store raw JSON data whose content changes dynamically.
	Status  json.RawMessage `json:"status"`
	Metrics json.RawMessage `json:"metrics"`
	Usages  json.RawMessage `json:"usages"` // The fields of this struct are fixed
	Metas   json.RawMessage `json:"metas"`
}

// PodStatus is the structure corresponding to the status field
type PodStatus struct {
	AppStatus AppStatus `json:"appStatus"`
}

// PodUpdateStatus is the structure corresponding to the application status information
type PodUpdateStatus struct {
	Metadata PodMetaData `json:"metadata"`
	Spec     PodSpec     `json:"spec"`
	Status   PodStatus   `json:"status"`
}

// OuterStruct is the structure representing the outermost structure of the HTTP request body.
// The "metadata" field is defined as a string type to receive an escaped JSON string.

type OuterStruct struct {
	Metadata string `json:"metadata"`
}

type InnerMetadata struct {
	Message    string      `json:"message"`
	Metadata   PodMetaData `json:"metadata"` // Nested "metadata" inside "metadata"
	SourceType string      `json:"source_type"`
	Spec       PodSpec     `json:"spec"`
	Status     PodStatus   `json:"status"`
	Timestamp  string      `json:"timestamp"` // Time string is received as string type and converted to time.Time if necessary
}

// PodAlerts represents the structure of alerts received from Alertmanager
type PodAlerts struct {
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Labels AlertLabels `json:"labels"`
}

type AlertLabels struct {
	Namespace         string `json:"namespace"`
	Pod               string `json:"pod"`
	ExportedNamespace string `json:"exported_namespace"`
	ExportedPod       string `json:"exported_pod"`
}

// GetNamespace returns the namespace, preferring the standard field over the exported field
func (l AlertLabels) GetNamespace() string {
	if l.Namespace != "" {
		return l.Namespace
	}
	return l.ExportedNamespace
}

// GetPod returns the pod name, preferring the standard field over the exported field
func (l AlertLabels) GetPod() string {
	if l.Pod != "" {
		return l.Pod
	}
	return l.ExportedPod
}
