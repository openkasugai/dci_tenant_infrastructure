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
