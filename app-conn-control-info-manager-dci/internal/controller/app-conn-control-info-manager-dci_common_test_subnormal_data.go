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
	"k8s.io/apimachinery/pkg/runtime"
)

var appStatusData03mapFailed = map[string]interface{}{
	"metrics":  appStatusData03Metrics,
	"status":   appStatusData03Status,
	"testtest": appStatusData01Usages,
}

var appStatusData03Failed = runtime.RawExtension{}

var ConnectEndpointCR101AToStatus1Failed = []examplecomv1.ToDeploymentStatus{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef901,
				Address:   "192.168.0.91",
				AppStatus: appStatusData01,
			},
		},
	}, {
		DeploymentRef: DeploymentRef11,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef102,
				Address:   "192.168.0.12",
				AppStatus: appStatusData01,
			}, {
				PodRef:    PodRef103,
				Address:   "192.168.0.13",
				AppStatus: appStatusData03Failed,
			},
		},
	},
}

var DestinationPodStatus103Failed = examplecomv1.Destination{
	Name:      "dciappstate-103",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.13",
	Status:    appStatusData03Status,
	Metrics:   appStatusData03Metrics,
	Usages:    runtime.RawExtension{},
	Metas:     runtime.RawExtension{},
}

var ExpectedConfigFile101AFailed = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-90",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus901,
			},
		}, {
			Label: "dciappstate-11",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus102,
				DestinationPodStatus103Failed,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile209ANoneInput = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-90",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus901,
			},
		}, {
			Label: "dciappstate-20",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus2021,
				DestinationPodStatus203,
			},
		},
	},
	OutputSubsets: externalOutputEndpoint2.ExternalSubsets,
}
