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
	examplecomv1 "app-conn-control-info-manager-dci/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cConnectEndpointNamespace string = "connectendpointtest"
var cDCIAppStateNamespace string = "dciappstatetest"

var cNamespaces []string = []string{
	cConnectEndpointNamespace,
	cDCIAppStateNamespace,
}

var cFinalizers = []string{
	examplecomv1.ConnectEndpointFinalizer,
}

var cLabels = map[string]string{
	"dciappstate-101": cConnectEndpointNamespace,
}

var externalInputEndpoint = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"rtsp://localhost:8554",
	},
}

var externalInputEndpoint2 = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"rtsp://localhost:8554",
		"rtsp://localhost:8555",
	},
}

var externalInputEndpoint3 = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"rtsp://localhost:8555",
	},
}

var externalOutputEndpoint = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"https://localhost:8600",
		"https://localhost:8601",
	},
}

var externalOutputEndpoint2 = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"https://localhost:8600",
	},
}

var externalOutputEndpoint3 = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"https://localhost:8600",
		"https://localhost:8602",
	},
}

var appStatusStatus01 = []byte(`{"status":"NotReady"}`)
var appStatusMetrics01 = []byte(`{"cpuusedper":11,"gpuusedper":12,"memoryusedper":13}`)
var appStatusUsages01 = []byte(`{"cpu":101,"gpu":102,"memory":103}`)

var appStatusStatus02 = []byte(`{"status":"NotReady"}`)
var appStatusMetrics02 = []byte(`{"cpuusedper":21,"gpuusedper":22,"memoryusedper":23}`)
var appStatusUsages02 = []byte(`{"cpu":201,"gpu":202,"memory":203}`)

var appStatusStatus03 = []byte(`{"status":"Processing"}`)
var appStatusMetrics03 = []byte(`{"cpuusedper":31,"gpuusedper":32,"memoryusedper":33}`)

var appStatusData01Status = runtime.RawExtension{
	Raw: appStatusStatus01,
}
var appStatusData01Metrics = runtime.RawExtension{
	Raw: appStatusMetrics01,
}
var appStatusData01Usages = runtime.RawExtension{
	Raw: appStatusUsages01,
}
var appStatusData02Status = runtime.RawExtension{
	Raw: appStatusStatus02,
}
var appStatusData02Metrics = runtime.RawExtension{
	Raw: appStatusMetrics02,
}
var appStatusData02Usages = runtime.RawExtension{
	Raw: appStatusUsages02,
}
var appStatusData03Status = runtime.RawExtension{
	Raw: appStatusStatus03,
}
var appStatusData03Metrics = runtime.RawExtension{
	Raw: appStatusMetrics03,
}
var appStatusData01map = map[string]interface{}{
	"metrics": appStatusData01Metrics,
	"status":  appStatusData01Status,
	"usages":  appStatusData01Usages,
}
var appStatusData02map = map[string]interface{}{
	"metrics": appStatusData02Metrics,
	"status":  appStatusData02Status,
	"usages":  appStatusData02Usages,
}
var appStatusData03map = map[string]interface{}{
	"metrics": appStatusData03Metrics,
	"status":  appStatusData03Status,
}
var appStatusData01 = runtime.RawExtension{}
var appStatusData02 = runtime.RawExtension{}
var appStatusData03 = runtime.RawExtension{}

var ConnectEndpointCR101ObjectMeta = metav1.ObjectMeta{
	Name:      "dciappstate-101",
	Namespace: cConnectEndpointNamespace,
}

var DeploymentRef10 = examplecomv1.NamespacedName{
	Name:      "dciappstate-10",
	Namespace: cDCIAppStateNamespace,
}
var PodRef101 = examplecomv1.NamespacedName{
	Name:      "dciappstate-101",
	Namespace: cDCIAppStateNamespace,
}
var ConnectEndpointCR101From = examplecomv1.FromDeployment{
	DeploymentRef: DeploymentRef10,
	FromPod: examplecomv1.FromPod{
		PodRef:  PodRef101,
		Address: "192.168.0.11",
	},
}

var ConnectEndpointCR101FromStatus0 = examplecomv1.FromDeployment{}

var DeploymentRef11 = examplecomv1.NamespacedName{
	Name:      "dciappstate-11",
	Namespace: cDCIAppStateNamespace,
}
var PodRef102 = examplecomv1.NamespacedName{
	Name:      "dciappstate-102",
	Namespace: cDCIAppStateNamespace,
}
var PodRef103 = examplecomv1.NamespacedName{
	Name:      "dciappstate-103",
	Namespace: cDCIAppStateNamespace,
}
var ConnectEndpointCR101To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef102,
				Address: "192.168.0.12",
			}, {
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			},
		},
	},
}

var ConnectEndpointCR101ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR101ToStatus2 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			},
		},
	},
}

var ConnectEndpointCR101ToStatus3 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:  PodRef102,
				Address: "192.168.0.12",
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR101 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var createtest03 string = "createtest03"
var dciappstate401nm string = "dciappstate-401"
var dciappstate401ip string = "192.168.0.41"
var dciappstate402nm string = "dciappstate-402"
var dciappstate402ip string = "192.168.0.42"

var ConnectEndpointCR103ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef: examplecomv1.NamespacedName{
					Name:      createtest03,
					Namespace: createtest03,
				},
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR103ToStatus2 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.14",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR103ToStatus3 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData03,
			},
		},
	},
}

var ConnectEndpointCR103ToStatus4 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData02,
			},
		},
	},
}

var ConnectEndpointCR103 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.NotConfigured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101ToStatus1,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR107 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101To,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR108 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:          ConnectEndpointCR101From,
		To:            ConnectEndpointCR101To,
		InputEndpoint: &externalInputEndpoint,
	},
}

var ConnectEndpointCR109ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR109 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.NotConfigured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR109ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR119To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef11,
	},
}

var ConnectEndpointCR119ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
	},
}

var ConnectEndpointCR119 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR119To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var PodRef120 = examplecomv1.NamespacedName{
	Name:      "dciappstate-120",
	Namespace: cDCIAppStateNamespace,
}

var ConnectEndpointCR120To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef120,
				Address: "192.168.0.120",
			}, {
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			},
		},
	},
}

var ConnectEndpointCR120ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:  PodRef120,
				Address: "192.168.0.120",
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR120 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR120To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR201ObjectMeta = metav1.ObjectMeta{
	Name:       "dciappstate-101",
	Namespace:  cConnectEndpointNamespace,
	Finalizers: cFinalizers,
}

var DeploymentRef20 = examplecomv1.NamespacedName{
	Name:      "dciappstate-20",
	Namespace: cDCIAppStateNamespace,
}
var PodRef202 = examplecomv1.NamespacedName{
	Name:      "dciappstate-202",
	Namespace: cDCIAppStateNamespace,
}
var PodRef203 = examplecomv1.NamespacedName{
	Name:      "dciappstate-203",
	Namespace: cDCIAppStateNamespace,
}
var PodRef204 = examplecomv1.NamespacedName{
	Name:      "dciappstate-204",
	Namespace: cDCIAppStateNamespace,
}
var PodRef205 = examplecomv1.NamespacedName{
	Name:      "dciappstate-205",
	Namespace: cDCIAppStateNamespace,
}
var PodRef206 = examplecomv1.NamespacedName{
	Name:      "dciappstate-206",
	Namespace: cDCIAppStateNamespace,
}
var PodRef207 = examplecomv1.NamespacedName{
	Name:      "dciappstate-207",
	Namespace: cDCIAppStateNamespace,
}

var ConnectEndpointCR201To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef204,
				Address: "192.168.0.24",
			}, {
				PodRef:  PodRef205,
				Address: "192.168.0.25",
			}, {
				PodRef:  PodRef206,
				Address: "192.168.0.26",
			},
		},
	},
}

var ConnectEndpointCR201ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR201ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef205,
				Address:   "192.168.0.25",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef206,
				Address:   "192.168.0.26",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR201 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR201To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR201ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR203To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef202,
				Address: "192.168.0.22",
			}, {
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			}, {
				PodRef:  PodRef204,
				Address: "192.168.0.24",
			},
		},
	},
}

var ConnectEndpointCR203ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR203ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR203 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR204ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR204ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			}, {
				PodRef:  PodRef204,
				Address: "192.168.0.24",
			},
		},
	},
}

var ConnectEndpointCR204 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR204ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR205To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef202,
				Address: "192.168.0.22",
			}, {
				PodRef:  PodRef204,
				Address: "192.168.0.24",
			},
		},
	},
}

var ConnectEndpointCR205ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR205 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR205To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR206To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef20,
	},
}

var ConnectEndpointCR206ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR206ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
	},
}

var ConnectEndpointCR206 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR206To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR206ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR207From = examplecomv1.FromDeployment{
	DeploymentRef: DeploymentRef10,
	FromPod: examplecomv1.FromPod{
		PodRef:  PodRef101,
		Address: "192.168.0.27",
	},
}

var ConnectEndpointCR207To = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef202,
				Address: "192.168.0.22",
			}, {
				PodRef:  PodRef203,
				Address: "192.168.0.23",
			},
		},
	},
}

var ConnectEndpointCR207ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR207 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR207From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR208 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint2,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR209 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint2,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR301 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR401ToStatus1 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData02,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var PodRef401 = examplecomv1.NamespacedName{
	Name:      dciappstate401nm,
	Namespace: cDCIAppStateNamespace,
}
var ConnectEndpointCR401ToStatus2 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:  PodRef401,
				Address: dciappstate401ip,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var PodRef402 = examplecomv1.NamespacedName{
	Name:      dciappstate402nm,
	Namespace: cDCIAppStateNamespace,
}
var ConnectEndpointCR401ToStatus3 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef402,
				Address:   dciappstate402ip,
				AppStatus: appStatusData03,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR401 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR402ToStatus0 = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:  PodRef202,
				Address: "192.168.0.22",
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR402 = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207To,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR402ToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var DCIAppStateCR101 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-101",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR102 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-102",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "computation",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR103 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-103",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR120 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-120",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
}

var DCIAppStateCR202 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-202",
		Namespace: cDCIAppStateNamespace,
		Labels:    cLabels,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "computation",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR203 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-203",
		Namespace: cDCIAppStateNamespace,
		Labels:    cLabels,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}
var DCIAppStateCR204 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-204",
		Namespace: cDCIAppStateNamespace,
		Labels:    cLabels,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "computation",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR205 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-205",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "computation",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR206 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-206",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR207 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-207",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "testtype",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DCIAppStateCR401 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      dciappstate401nm,
		Namespace: cDCIAppStateNamespace,
		Labels:    cLabels,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "computation",
	},
}

var DCIAppStateCR402 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      dciappstate402nm,
		Namespace: cDCIAppStateNamespace,
		Labels:    cLabels,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData03,
	},
}

var DestinationPodStatus102 = examplecomv1.Destination{
	Name:      "dciappstate-102",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.12",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus103 = examplecomv1.Destination{
	Name:      "dciappstate-103",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.13",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus2021 = examplecomv1.Destination{
	Name:      "dciappstate-202",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.22",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus2022 = examplecomv1.Destination{
	Name:      "dciappstate-202",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.22",
	Status:    appStatusData02Status,
	Metrics:   appStatusData02Metrics,
	Usages:    appStatusData02Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus203 = examplecomv1.Destination{
	Name:      "dciappstate-203",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.23",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus204 = examplecomv1.Destination{
	Name:      "dciappstate-204",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.24",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus204Non = examplecomv1.Destination{
	Name:      "dciappstate-204",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.24",
}

var DestinationPodStatus205 = examplecomv1.Destination{
	Name:      "dciappstate-205",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.25",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus206 = examplecomv1.Destination{
	Name:      "dciappstate-206",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.26",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus207 = examplecomv1.Destination{
	Name:      "dciappstate-207",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.27",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var DestinationPodStatus401Non = examplecomv1.Destination{
	Name:      dciappstate401nm,
	Namespace: cDCIAppStateNamespace,
	Address:   dciappstate401ip,
}

var DestinationPodStatus402 = examplecomv1.Destination{
	Name:      dciappstate402nm,
	Namespace: cDCIAppStateNamespace,
	Address:   dciappstate402ip,
	Status:    appStatusData03Status,
	Metrics:   appStatusData03Metrics,
}

var ExpectedConfigFile101Del = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
}

var ExpectedConfigFile101 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-11",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus102,
				DestinationPodStatus103,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile107 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-11",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus102,
				DestinationPodStatus103,
			},
		},
	},
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile108 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-11",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus102,
				DestinationPodStatus103,
			},
		},
	},
	InputSubsets: externalInputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile119 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-11",
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile201 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus204,
				DestinationPodStatus205,
				DestinationPodStatus206,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile203 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus103,
				DestinationPodStatus204,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile204 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus103,
				DestinationPodStatus204Non,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile205 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus204,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile206 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile207 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.27",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile208 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint2.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile209 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint2.ExternalSubsets,
}

var ExpectedConfigFile401 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2022,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile401Non = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus401Non,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile4012 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus402,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile402 = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}
