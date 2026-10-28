package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Event type
const (
	Create = "Create"
	Update = "Update"
	Delete = "Delete"
	None   = "None"
)

const (
	ExclBaseFilePath      = "/tmp/app-conn-control-info-manager-dci"
	ExclCtrlComFilePath   = ExclBaseFilePath + "/.ConnectEndpoint_"
	ExclCtrlFileExtension = ".exclusive"
	ConfigBaseFilePath    = "/app-conn-control-info-manager-dci"
	ConfigFilePath        = ConfigBaseFilePath + "/appConnectControlInfo.json"
)

const (
	MyPodName                = "MY_POD_NAME"
	MyPodNamespace           = "MY_POD_NAMESPACE"
	MainContainerProcessName = "MAIN_CONTAINER_PROCESS_NAME"
)

const (
	NotConfigured = "NotConfigured"
	Configured    = "Configured"
)

const (
	DCIAppStateKind     = "DCIAppState"
	ConnectEndpointKind = "ConnectEndpoint"
)

const (
	Computation = "computation"
	Connection  = "connection"
)

const (
	ExclusiveInterval       = 100
	ConnectEndpointInterval = 1000
	DCIAppStateInterval     = 1000
)

const (
	FailureRateMin = 1
	FailureRateMax = 300
	RateLimit      = 10
	RateBurst      = 100
)

const (
	ConnectEndpointFinalizer = "connectendpoint.finalizers.example.com/v1"
)

// AppConnectControlInfo is the Schema for the AppConnectControlInfoFile.
type AppConnectControlInfo struct {
	Address       string    `json:"address"`
	Services      []Service `json:"services"`
	InputSubsets  []string  `json:"inputSubsets,omitempty"`
	OutputSubsets []string  `json:"outputSubsets,omitempty"`
}

// Service defines the desired state of AppConnectControlInfo.
type Service struct {
	Label        string        `json:"label"`
	Destinations []Destination `json:"destinations"`
}

// Destination defines the desired state of Service[].
type Destination struct {
	Name      string               `json:"name"`
	Namespace string               `json:"namespace"`
	Address   string               `json:"address"`
	Port      string               `json:"port,omitempty"`
	Status    runtime.RawExtension `json:"status"`
	Metrics   runtime.RawExtension `json:"metrics"`
	Usages    runtime.RawExtension `json:"usages"`
	Metas     runtime.RawExtension `json:"metas"`
}
