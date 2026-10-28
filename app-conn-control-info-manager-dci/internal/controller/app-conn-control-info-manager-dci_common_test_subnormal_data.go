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
