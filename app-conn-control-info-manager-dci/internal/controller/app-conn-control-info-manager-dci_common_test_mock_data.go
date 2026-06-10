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
)

var ConnectEndpointCR201AToStatusE = []examplecomv1.ToDeploymentStatus{
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
			}, {
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR201AToStatusE2 = []examplecomv1.ToDeploymentStatus{
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
	}, {
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef901,
				Address:   "192.168.0.91",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR201AToStatusE3 = []examplecomv1.ToDeploymentStatus{
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
		DeploymentRef: DeploymentRef20,
		ToPod: []examplecomv1.ToPodStatus{
			{
				PodRef:    PodRef202,
				Address:   "192.168.0.22",
				AppStatus: appStatusData01,
			}, {
				PodRef:  PodRef203,
				Address: "192.168.0.23",
			}, {
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}
