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

var DeploymentRef90 = examplecomv1.NamespacedName{
	Name:      "dciappstate-90",
	Namespace: cDCIAppStateNamespace,
}
var PodRef901 = examplecomv1.NamespacedName{
	Name:      "dciappstate-901",
	Namespace: cDCIAppStateNamespace,
}
var ConnectEndpointCR101ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR101AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR101AToStatus2 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:  PodRef103,
				Address: "192.168.0.13",
			},
		},
	},
}

var ConnectEndpointCR101AToStatus3 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR101A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR103AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
				PodRef: examplecomv1.NamespacedName{
					Name:      "createtest03",
					Namespace: "createtest03",
				},
				Address:   "192.168.0.13",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR103AToStatus2 = []examplecomv1.ToDeploymentStatus{
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
				Address:   "192.168.0.14",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR103AToStatus3 = []examplecomv1.ToDeploymentStatus{
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
				AppStatus: appStatusData03,
			},
		},
	},
}

var ConnectEndpointCR103AToStatus4 = []examplecomv1.ToDeploymentStatus{
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
				AppStatus: appStatusData02,
			},
		},
	},
}

var ConnectEndpointCR103A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.NotConfigured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101AToStatus1,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR107A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101ATo,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR108A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:          ConnectEndpointCR101From,
		To:            ConnectEndpointCR101ATo,
		InputEndpoint: &externalInputEndpoint,
	},
}

var ConnectEndpointCR109AToStatus0 = []examplecomv1.ToDeploymentStatus{
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
			},
		},
	},
}

var ConnectEndpointCR109A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR101ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.NotConfigured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR109AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR119ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
		DeploymentRef: DeploymentRef11,
	},
}

var ConnectEndpointCR119AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
	},
}

var ConnectEndpointCR119A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR119ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR120ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR120AToStatus1 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR120A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR101ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR120ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR201ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR201AToStatus0 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR201AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
			},
		},
	},
}

var ConnectEndpointCR201A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR201ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR201AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR203ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR203AToStatus0 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR203AToStatus1 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR203A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR204AToStatus0 = []examplecomv1.ToDeploymentStatus{
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
			},
		},
	},
}

var ConnectEndpointCR204AToStatus1 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR204A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR204AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR205ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR205AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR205A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR205ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR203AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR206ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
		DeploymentRef: DeploymentRef20,
	},
}

var ConnectEndpointCR206AToStatus0 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:    PodRef204,
				Address:   "192.168.0.24",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR206AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
	},
}

var ConnectEndpointCR206A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR206ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR206AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR207ATo = []examplecomv1.ToDeployment{
	{
		DeploymentRef: DeploymentRef90,
		ToPod: []examplecomv1.ToPod{
			{
				PodRef:  PodRef901,
				Address: "192.168.0.91",
			},
		},
	}, {
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

var ConnectEndpointCR207AToStatus0 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR207A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR207From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR208A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint2,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR209A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint2,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR301A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR401AToStatus1 = []examplecomv1.ToDeploymentStatus{
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
				AppStatus: appStatusData02,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR401AToStatus2 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:  PodRef401,
				Address: "192.168.0.41",
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR401AToStatus3 = []examplecomv1.ToDeploymentStatus{
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
				PodRef:    PodRef402,
				Address:   "192.168.0.42",
				AppStatus: appStatusData03,
			}, {
				PodRef:    PodRef203,
				Address:   "192.168.0.23",
				AppStatus: appStatusData01,
			},
		},
	},
}

var ConnectEndpointCR401A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var ConnectEndpointCR402AToStatus0 = []examplecomv1.ToDeploymentStatus{
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

var ConnectEndpointCR402A = examplecomv1.ConnectEndpoint{
	ObjectMeta: ConnectEndpointCR201ObjectMeta,
	Spec: examplecomv1.ConnectEndpointSpec{
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR207ATo,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
	Status: examplecomv1.ConnectEndpointStatus{
		Condition:      examplecomv1.Configured,
		From:           ConnectEndpointCR101From,
		To:             ConnectEndpointCR402AToStatus0,
		InputEndpoint:  &externalInputEndpoint,
		OutputEndpoint: &externalOutputEndpoint,
	},
}

var DCIAppStateCR901 = examplecomv1.DCIAppState{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dciappstate-901",
		Namespace: cDCIAppStateNamespace,
	},
	Spec: examplecomv1.DCIAppStateSpec{
		Type: "connection",
	},
	Status: examplecomv1.DCIAppStateStatus{
		AppStatus: appStatusData01,
	},
}

var DestinationPodStatus901 = examplecomv1.Destination{
	Name:      "dciappstate-901",
	Namespace: cDCIAppStateNamespace,
	Address:   "192.168.0.91",
	Status:    appStatusData01Status,
	Metrics:   appStatusData01Metrics,
	Usages:    appStatusData01Usages,
	Metas:     runtime.RawExtension{},
}

var ExpectedConfigFile101A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus103,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile107A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus103,
			},
		},
	},
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile108A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus103,
			},
		},
	},
	InputSubsets: externalInputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile119A = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-90",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus901,
			},
		}, {
			Label: "dciappstate-11",
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile201A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus204,
				DestinationPodStatus205,
				DestinationPodStatus206,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile203A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus103,
				DestinationPodStatus204,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile204A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus103,
				DestinationPodStatus204Non,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile205A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus204,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile206A = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.11",
	Services: []examplecomv1.Service{
		{
			Label: "dciappstate-90",
			Destinations: []examplecomv1.Destination{
				DestinationPodStatus901,
			},
		}, {
			Label: "dciappstate-20",
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile207A = examplecomv1.AppConnectControlInfo{
	Address: "192.168.0.27",
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
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile208A = examplecomv1.AppConnectControlInfo{
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
	InputSubsets:  externalInputEndpoint2.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile209A = examplecomv1.AppConnectControlInfo{
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
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint2.ExternalSubsets,
}

var ExpectedConfigFile401A = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus2022,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile401ANon = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus401Non,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile401A2 = examplecomv1.AppConnectControlInfo{
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
				DestinationPodStatus402,
				DestinationPodStatus203,
			},
		},
	},
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}

var ExpectedConfigFile402A = examplecomv1.AppConnectControlInfo{
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
	InputSubsets:  externalInputEndpoint.ExternalSubsets,
	OutputSubsets: externalOutputEndpoint.ExternalSubsets,
}
