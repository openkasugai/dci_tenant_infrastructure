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
	Services      []Service `json:"services,omitempty"`
	InputSubsets  []string  `json:"inputSubsets,omitempty"`
	OutputSubsets []string  `json:"outputSubsets,omitempty"`
}

// Service defines the desired state of AppConnectControlInfo.
type Service struct {
	Label        string        `json:"label"`
	Destinations []Destination `json:"destinations,omitempty"`
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
