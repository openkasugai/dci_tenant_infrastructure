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
	examplecomv1 "app-conn-info-maker-dci/api/v1"

	hostdevicenetwork "github.com/Mellanox/network-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	klog.InitFlags(nil)
}

func getMgr(mgr ctrl.Manager) (ctrl.Manager, error) {
	if mgr == nil {
		return ctrl.NewManager(cfg, ctrl.Options{
			Scheme: testScheme,
		})
	}
	return mgr, nil
}

var cDCIAppConnectionKind string = "dciappconnection"
var cDCIAppConnectionAPIVersion string = "examplecom/v1"
var cDeploymentKind string = "deployment"
var cDeploymentAPIVersion string = "apps/v1"
var cReplicaSetKind string = "replicaset"
var cReplicaSetAPIVersion string = "apps/v1"
var cPodKind string = "pod"
var cPodAPIVersion string = "core/v1"
var cServiceKind string = "service"
var cServiceAPIVersion string = "core/v1"
var cEndpointSliceKind string = "endpointslice"
var cEndpointSliceAPIVersion string = "discovery/v1"
var cConnectEndpointKind string = "connectendpoint"
var cConnectEndpointAPIVersion string = "examplecom/v1"

var cDCIAppConnectionNamespace string = "conntest"
var cServiceNamespace string = cDCIAppConnectionNamespace
var cEndpointSliceNamespace string = cDCIAppConnectionNamespace
var cConnectEndpointNamespace string = cDCIAppConnectionNamespace
var cDeploymentNamespace string = "podtest"
var cReplicaSetNamespace string = cDeploymentNamespace
var cPodNamespace string = cReplicaSetNamespace

var cNamespaces []string = []string{
	cDCIAppConnectionNamespace,
	cDeploymentNamespace,
}

var cDCIAppConnectionFinalizer string = "dciappconnection.finalizers.example.com/v1"

var cDCIAppConnection01Name string = "dciappconnection01"
var cDeployment01NameFromX string = "testdeployment01from-x"
var cDeployment01NameToA string = "testdeployment01to-a"
var cDeployment01NameToC string = "testdeployment01to-c"
var cDeployment01NameToY string = "testdeployment01to-y"
var cDeployment01NameFromY string = cDeployment01NameToY
var cDeployment01NameToX string = cDeployment01NameFromX
var cDeployment01NameFromZ string = "testdeployment01from-z"
var cReplicaSet01NameFromX string = "testdeployment01from-x-1"
var cReplicaSet01NameToA string = "testdeployment01to-a-1"
var cReplicaSet01NameTo2A string = "testdeployment01to-a-2"
var cReplicaSet01NameToC string = "testdeployment01to-c-1"
var cReplicaSet01NameToY string = "testdeployment01to-y-1"
var cReplicaSet01NameToY0 string = "testdeployment01to-y-0"
var cPod01NameFromX1 string = "testdeployment01from-x-1-1"
var cPod01NameFromX2 string = "testdeployment01from-x-1-2"
var cPod01NameToA1 string = "testdeployment01to-a-1-1"
var cPod01NameToA2 string = "testdeployment01to-a-1-2"
var cPod01NameTo2A string = "testdeployment01to-a-2-1"
var cPod01NameToC1 string = "testdeployment01to-c-1-1"
var cPod01NameToY1 string = "testdeployment01to-y-1-1"
var cPod01NameToY2 string = "testdeployment01to-y-1-2"
var cService01NameToA string = cDeployment01NameToA
var cService01NameToC string = cDeployment01NameToC
var cService01NameToY string = cDeployment01NameToY
var cEndpointSlice01NameToA string = cDeployment01NameToA
var cEndpointSlice01NameToC string = cDeployment01NameToC
var cEndpointSlice01NameToY string = cDeployment01NameToY
var cConnectEndpoint01NameFromX1 string = "testdeployment01from-x-1-1"
var cConnectEndpoint01NameFromX2 string = "testdeployment01from-x-1-2"
var cConnectEndpoint01NameToY1 string = "testdeployment01to-y-1-1"
var cConnectEndpoint01NameToY2 string = "testdeployment01to-y-1-2"
var cConnectEndpoint01NameFromZ1 string = "testdeployment01to-z-1-1"

var cPod01IpAddressFromX1 string = "10.10.1.11"
var cPod01IpAddressFromX2 string = "10.10.1.12"
var cPod01IpAddressToA1 string = "10.10.10.11"
var cPod01IpAddressToA2 string = "10.10.10.12"
var cPod01IpAddressTo2A string = "10.10.10.112"
var cPod01IpAddressToC1 string = "10.10.11.11"
var cPod01IpAddressToY1 string = "10.10.12.11"
var cPod01IpAddressToY2 string = "10.10.12.12"
var cPodIfName string = "net1"
var cPodIps string = "ips"

// HostDeviceNetwork
var cHostDeviceNetwork string = "HostDeviceNetwork"
var cHostDeviceNetworkAPIVersion string = "mellanox.com/v1alpha1"
var cHostDeviceNetworkNamespace string = cPodNamespace
var cHostDeviceNetwork01Name string = "hostdevicenetwork01"
var cHostDeviceNetworkNoneName string = "hostdevicenetworknone"
var cHostDeviceNetwork01ResourceName string = "hostdevicenetwork01resource"
var cHostDeviceNetwork01IPAM string = "hostdevicenetwork01ipam"
var HostDeviceNetwork01 = hostdevicenetwork.HostDeviceNetwork{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cHostDeviceNetworkAPIVersion,
		Kind:       cHostDeviceNetwork,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cHostDeviceNetwork01Name,
		Namespace: cHostDeviceNetworkNamespace,
	},
	Spec: hostdevicenetwork.HostDeviceNetworkSpec{
		NetworkNamespace: cHostDeviceNetworkNamespace,
		ResourceName:     cHostDeviceNetwork01ResourceName,
		IPAM:             cHostDeviceNetwork01IPAM,
	},
}

// DCIAppConnection
var DCIAppConnection01InputEndpoint = examplecomv1.DCIAppConnectionExternalEndpoint{
	Internal: examplecomv1.TargetRef{
		DeploymentRef: examplecomv1.NamespaceName{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
	},
	ExternalSubsets: []string{
		"rtsp://localhost:8554",
		"rtsp://localhost:8555",
	},
}
var DCIAppConnection01OutputEndpoint = examplecomv1.DCIAppConnectionExternalEndpoint{
	Internal: examplecomv1.TargetRef{
		DeploymentRef: examplecomv1.NamespaceName{
			Name:      cDeployment01NameToY,
			Namespace: cDeploymentNamespace,
		},
	},
	ExternalSubsets: []string{
		"https://localhost:8600",
		"https://localhost:8601",
	},
}

var DCIAppConnection01 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}

var DCIAppConnection01Status = examplecomv1.DCIAppConnectionStatus{
	Phase: "Created",
}

// DCIAppConnection Test05 Change the definition order with X-Y at the top
var DCIAppConnection01topToY = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}

var DCIAppConnectionE01 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}
var DCIAppConnectionCE01OutputEndpoint = examplecomv1.DCIAppConnectionExternalEndpoint{
	Internal: examplecomv1.TargetRef{
		DeploymentRef: examplecomv1.NamespaceName{
			Name:      cDeployment01NameToX,
			Namespace: cDeploymentNamespace,
		},
	},
	ExternalSubsets: []string{
		"https://localhost:8600",
		"https://localhost:8601",
	},
}
var DCIAppConnectionCE01 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnectionCE01OutputEndpoint,
	},
}
var DCIAppConnectionCE02OutputEndpoint = examplecomv1.DCIAppConnectionExternalEndpoint{
	Internal: examplecomv1.TargetRef{
		DeploymentRef: examplecomv1.NamespaceName{
			Name:      cDeployment01NameToX,
			Namespace: cDeploymentNamespace,
		},
	},
	ExternalSubsets: []string{
		"https://localhost:8554",
		"https://localhost:8555",
	},
}
var DCIAppConnectionCE02 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnectionCE02OutputEndpoint,
	},
}

// Service
var Service01ToA = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cServiceAPIVersion,
		Kind:       cServiceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cService01NameToA,
		Namespace: cServiceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDCIAppConnectionAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{
			{
				Protocol: "TCP",
				Port:     80,
				TargetPort: intstr.IntOrString{
					Type:   0,
					IntVal: 80,
					StrVal: "",
				},
			},
		},
	},
}
var Service01ToC = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cServiceAPIVersion,
		Kind:       cServiceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cService01NameToC,
		Namespace: cServiceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDCIAppConnectionAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{
			{
				Protocol: "TCP",
				Port:     80,
				TargetPort: intstr.IntOrString{
					Type:   0,
					IntVal: 80,
					StrVal: "",
				},
			},
		},
	},
}
var Service01ToY = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cServiceAPIVersion,
		Kind:       cServiceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cService01NameToY,
		Namespace: cServiceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDCIAppConnectionAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{
			{
				Protocol: "TCP",
				Port:     80,
				TargetPort: intstr.IntOrString{
					Type:   0,
					IntVal: 80,
					StrVal: "",
				},
			},
		},
	},
}

// EndpointSlice
var EndpointSlice01ToALabels = map[string]string{
	"kubernetes.io/service-name": cService01NameToA,
}
var EndpointSlice01ToA1TergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameToA1,
	Namespace: cPodNamespace,
}
var EndpointSlice01ToA2TergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameToA2,
	Namespace: cPodNamespace,
}
var EndpointSlice01To2ATergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameTo2A,
	Namespace: cPodNamespace,
}
var EndpointSlice01ToA = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToA,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToA,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToA1,
			},
			TargetRef: &EndpointSlice01ToA1TergetRef,
		}, {
			Addresses: []string{
				cPod01IpAddressToA2,
			},
			TargetRef: &EndpointSlice01ToA2TergetRef,
		},
	},
}
var EndpointSlice01ToA2None = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToA,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToA,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToA1,
			},
			TargetRef: &EndpointSlice01ToA1TergetRef,
		},
	},
}
var EndpointSlice01To2A = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToA,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToA,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToA1,
			},
			TargetRef: &EndpointSlice01ToA1TergetRef,
		}, {
			Addresses: []string{
				cPod01IpAddressToA2,
			},
			TargetRef: &EndpointSlice01ToA2TergetRef,
		}, {
			Addresses: []string{
				cPod01IpAddressTo2A,
			},
			TargetRef: &EndpointSlice01To2ATergetRef,
		},
	},
}
var EndpointSlice01ToCLabels = map[string]string{
	"kubernetes.io/service-name": cService01NameToC,
}
var EndpointSlice01ToC1TergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameToC1,
	Namespace: cPodNamespace,
}
var EndpointSlice01ToC = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToC,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToC,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToC1,
			},
			TargetRef: &EndpointSlice01ToC1TergetRef,
		},
	},
}
var EndpointSlice01ToYLabels = map[string]string{
	"kubernetes.io/service-name": cService01NameToY,
}
var EndpointSlice01ToY1TergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameToY1,
	Namespace: cPodNamespace,
}
var EndpointSlice01ToY = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToY,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToY,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToY1,
			},
			TargetRef: &EndpointSlice01ToY1TergetRef,
		},
	},
}
var EndpointSlice01ToY2TergetRef = corev1.ObjectReference{
	Kind:      cPodKind,
	Name:      cPod01NameToY2,
	Namespace: cPodNamespace,
}
var EndpointSlice01ToY2add = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToY,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToY,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToY1,
			},
			TargetRef: &EndpointSlice01ToY1TergetRef,
		}, {
			Addresses: []string{
				cPod01IpAddressToY2,
			},
			TargetRef: &EndpointSlice01ToY2TergetRef,
		},
	},
}
var EndpointSlice01ToY2upd = discoveryv1.EndpointSlice{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cEndpointSliceAPIVersion,
		Kind:       cEndpointSliceKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cEndpointSlice01NameToY,
		Namespace: cEndpointSliceNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cServiceKind,
				Name:       cService01NameToY,
				UID:        "0",
			},
		},
	},
	AddressType: discoveryv1.AddressTypeIPv4,
	Endpoints: []discoveryv1.Endpoint{
		{
			Addresses: []string{
				cPod01IpAddressToY1,
			},
			TargetRef: &EndpointSlice01ToY1TergetRef,
		}, {
			Addresses: []string{
				cPod01IpAddressToA1,
			},
			TargetRef: &EndpointSlice01ToY2TergetRef,
		},
	},
}

// ConnectEndpointSpec Output Format
var ConnectEndpoint01FromX1SpecInputEndpoint = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"rtsp://localhost:8554",
		"rtsp://localhost:8555",
	},
}

var ConnectEndpoint01ToY1SpecOutputEndpoint = examplecomv1.ExternalEndpoint{
	ExternalSubsets: []string{
		"https://localhost:8600",
		"https://localhost:8601",
	},
}

var ConnectEndpoint01FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			},
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: nil,
	},
}

var ConnectEndpoint01FromX1NoneA2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX1NoneASpec = examplecomv1.ConnectEndpointSpec{
	From: examplecomv1.FromDeployment{
		DeploymentRef: examplecomv1.NamespacedName{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		FromPod: examplecomv1.FromPod{
			PodRef: examplecomv1.NamespacedName{
				Name:      cPod01NameFromX1,
				Namespace: cPodNamespace,
			},
			Address: cPod01IpAddressFromX1,
		},
	},
	To: []examplecomv1.ToDeployment{
		{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToA,
				Namespace: cDeploymentNamespace,
			},
			ToPod: nil,
		}, {
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToC,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToC1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToC1,
				},
			},
		}, {
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToY,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToY1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToY1,
				},
			},
		},
	},
	InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
}
var ConnectEndpoint01FromX1add2A = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameTo2A,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressTo2A,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX1addY2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY2,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX1addY2topToY = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX1NoneIpA2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: "",
					},
				},
			},
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			},
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: nil,
	},
}
var ConnectEndpoint01FromX2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX2NoneIp = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: "",
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint01FromX2NoneA2Spec = examplecomv1.ConnectEndpointSpec{
	From: examplecomv1.FromDeployment{
		DeploymentRef: examplecomv1.NamespacedName{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		FromPod: examplecomv1.FromPod{
			PodRef: examplecomv1.NamespacedName{
				Name:      cPod01NameFromX2,
				Namespace: cPodNamespace,
			},
			Address: cPod01IpAddressFromX2,
		},
	},
	To: []examplecomv1.ToDeployment{
		{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToA,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToA1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToA1,
				},
			},
		}, {
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToC,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToC1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToC1,
				},
			},
		}, {
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToY,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToY1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToY1,
				},
			},
		},
	},
	InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
}
var ConnectEndpoint01ToY1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY1,
			},
		},
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

var ConnectEndpoint01ToY2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY2,
			},
		},
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

// Deployment
var iReplicas01FromX int32 = 1
var iReplicas01FromX1to2 int32 = 2
var iReplicas01FromX0 int32 = 0
var stDeploymentSelector01FromX = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app":                    "nginx",
		"app.kubernetes.io/name": cDeployment01NameFromX,
	},
}
var stDeployment01FromXLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
}
var Deployment01FromX = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDeploymentAPIVersion,
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDeployment01NameFromX,
		Namespace: cDeploymentNamespace,
		Labels:    stDeployment01FromXLabels0,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &iReplicas01FromX,
		Selector: &stDeploymentSelector01FromX,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameFromX,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var Deployment01FromXLabels = map[string]string{
	"app":                    "nginx",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
	"app.kubernetes.io/name": cDeployment01NameFromX,
}
var Deployment01FromXLabelsNone = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	cDCIAppConnectionKind:    "",
}
var Deployment01FromXLabelsUnmatchName = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	cDCIAppConnectionKind:    "testtest_conntest",
}
var Deployment01FromXLabelsUnmatchNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	cDCIAppConnectionKind:    "dciappconnection01_test",
}
var Deployment01FromXLabelsNoneNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	cDCIAppConnectionKind:    cDCIAppConnection01Name,
}
var iReplicas01ToA int32 = 2
var iReplicas01ToA2to1 int32 = 1
var stDeploymentSelector01ToA = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app":                    "nginx",
		"app.kubernetes.io/name": cDeployment01NameToA,
	},
}
var stDeployment01ToALabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
}
var Deployment01ToA = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDeploymentAPIVersion,
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDeployment01NameToA,
		Namespace: cDeploymentNamespace,
		Labels:    stDeployment01ToALabels0,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &iReplicas01ToA,
		Selector: &stDeploymentSelector01ToA,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToA,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var Deployment01ToALabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Deployment01ToALabelsNone = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	cDCIAppConnectionKind:    "",
}
var Deployment01ToALabelsUnmatchName = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	cDCIAppConnectionKind:    "testtest_conntest",
}
var Deployment01ToALabelsUnmatchNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	cDCIAppConnectionKind:    "dciappconnection01_test",
}
var Deployment01ToALabelsNoneNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	cDCIAppConnectionKind:    cDCIAppConnection01Name,
}
var iReplicas01ToC int32 = 1
var stDeploymentSelector01ToC = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app":                    "nginx",
		"app.kubernetes.io/name": cDeployment01NameToC,
	},
}
var stDeployment01ToCLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToC,
}
var Deployment01ToC = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDeploymentAPIVersion,
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDeployment01NameToC,
		Namespace: cDeploymentNamespace,
		Labels:    stDeployment01ToCLabels0,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &iReplicas01ToC,
		Selector: &stDeploymentSelector01ToC,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToC,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var Deployment01ToCLabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToC,
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var iReplicas01ToY0 int32 = 0
var iReplicas01ToY int32 = 1
var iReplicas01ToY1to2 int32 = 2
var stDeploymentSelector01ToY = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app":                    "nginx",
		"app.kubernetes.io/name": cDeployment01NameToY,
	},
}
var stDeployment01ToYLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToY,
}
var Deployment01ToY = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDeploymentAPIVersion,
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDeployment01NameToY,
		Namespace: cDeploymentNamespace,
		Labels:    stDeployment01ToYLabels0,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &iReplicas01ToY,
		Selector: &stDeploymentSelector01ToY,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToY,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var Deployment01ToYLabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToY,
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var iReplicas01FromZ int32 = 1
var stDeploymentSelector01FromZ = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app": "nginx",
	},
}
var stDeployment01FromZLabels0 = map[string]string{
	"app": "nginx",
}
var Deployment01FromZ = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDeploymentAPIVersion,
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDeployment01NameFromZ,
		Namespace: cDeploymentNamespace,
		Labels:    stDeployment01FromZLabels0,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: &iReplicas01FromZ,
		Selector: &stDeploymentSelector01FromZ,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "nginx",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}

// ReplicaSet
var stReplicaSetSelector01FromX = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameFromX,
		"pod-template-hash":      "1",
	},
}
var ReplicaSet01FromX = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameFromX,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameFromX,
			"pod-template-hash":      "1",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameFromX,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01FromX,
		Selector: &stReplicaSetSelector01FromX,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameFromX,
					"pod-template-hash":      "1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var stReplicaSetSelector01ToA = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameToA,
		"pod-template-hash":      "1",
	},
}
var stReplicaSet01ToALabelHashNone = map[string]string{
	"app.kubernetes.io/name": cDeployment01NameToA,
}
var ReplicaSet01ToA = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameToA,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameToA,
			"pod-template-hash":      "1",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01ToA,
		Selector: &stReplicaSetSelector01ToA,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToA,
					"pod-template-hash":      "1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var stReplicaSetSelector01To2A = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameToA,
		"pod-template-hash":      "2",
	},
}
var ReplicaSet01To2A = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameTo2A,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameToA,
			"pod-template-hash":      "2",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01ToA,
		Selector: &stReplicaSetSelector01To2A,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToA,
					"pod-template-hash":      "2",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var stReplicaSetSelector01ToC = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameToC,
		"pod-template-hash":      "1",
	},
}
var ReplicaSet01ToC = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameToC,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameToC,
			"pod-template-hash":      "1",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameToC,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01ToC,
		Selector: &stReplicaSetSelector01ToC,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToC,
					"pod-template-hash":      "1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var stReplicaSetSelector01ToY = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameToY,
		"pod-template-hash":      "1",
	},
}
var ReplicaSet01ToY = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameToY,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameToY,
			"pod-template-hash":      "1",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameToY,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01ToY,
		Selector: &stReplicaSetSelector01ToY,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToY,
					"pod-template-hash":      "1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}
var stReplicaSetSelector01ToY0 = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app.kubernetes.io/name": cDeployment01NameToY,
		"pod-template-hash":      "0",
	},
}
var ReplicaSet01ToY0 = appsv1.ReplicaSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cReplicaSetAPIVersion,
		Kind:       "ReplicaSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cReplicaSet01NameToY0,
		Namespace: cReplicaSetNamespace,
		Labels: map[string]string{
			"app.kubernetes.io/name": cDeployment01NameToY,
			"pod-template-hash":      "0",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cDeploymentAPIVersion,
				Kind:       cDeploymentKind,
				Name:       cDeployment01NameToY,
				UID:        "0",
			},
		},
	},
	Spec: appsv1.ReplicaSetSpec{
		Replicas: &iReplicas01ToY0,
		Selector: &stReplicaSetSelector01ToY0,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                    "nginx",
					"app.kubernetes.io/name": cDeployment01NameToY,
					"pod-template-hash":      "0",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "nginx",
						Image:   "nginx",
						Command: []string{"sh", "-c"},
						Args:    []string{"while true;do sleep 2; done"},
					},
				},
			},
		},
	},
}

// Pod
var PodStatusRunning = corev1.PodStatus{
	Phase: corev1.PodRunning,
}
var Pod01FromXLabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Pod01FromXLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
}
var Pod01FromXLabelsNone = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    "",
}
var Pod01FromXLabelsUnmatchName = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    "testtest" + "_" + cDCIAppConnectionNamespace,
}
var Pod01FromXLabelsUnmatchNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_test",
}
var Pod01FromXLabelsNoneNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameFromX,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name,
}
var Pod01FromXAnno2ndNicNone = map[string]string{
	"k8s.v1.cni.cncf.io/networks": cHostDeviceNetwork01Name,
}

var Pod01FromX1 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameFromX1,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01FromX1Annotation,
		},
		Labels: Pod01FromXLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameFromX,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01FromX2 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameFromX2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01FromX2Annotation,
		},
		Labels: Pod01FromXLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameFromX,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01FromX2HostDeviceNetworkNone = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameFromX2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetworkNoneName,
			"k8s.v1.cni.cncf.io/network-status": Pod01FromX2Annotation,
		},
		Labels: map[string]string{
			"app":                    "nginx",
			"app.kubernetes.io/name": cDeployment01NameFromX,
			"pod-template-hash":      "1",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameFromX,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01FromX2NicNone = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameFromX2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/network-status": Pod01FromX2Annotation,
		},
		Labels: Pod01FromXLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameFromX,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToALabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Pod01ToALabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
}
var Pod01ToALabelsNone = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    "",
}
var Pod01ToALabelsUnmatchName = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    "testtest_" + cDCIAppConnectionNamespace,
}
var Pod01ToALabelsUnmatchNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_test",
}
var Pod01ToALabelsNoneNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name,
}
var Pod01ToA1Anno2ndNicNone = map[string]string{
	"k8s.v1.cni.cncf.io/networks": cHostDeviceNetwork01Name,
}
var Pod01ToA1Anno2ndNicCharNone = map[string]string{
	"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
	"k8s.v1.cni.cncf.io/network-status": "",
}
var Pod01ToA1 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToA1,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToA1Annotation,
		},
		Labels: Pod01ToALabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToA2Anno2ndNicNone = map[string]string{
	"k8s.v1.cni.cncf.io/networks": cHostDeviceNetwork01Name,
}
var Pod01ToA2Anno2ndNicCharNone = map[string]string{
	"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
	"k8s.v1.cni.cncf.io/network-status": "",
}
var Pod01ToA2 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToA2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToA2Annotation,
		},
		Labels: Pod01ToALabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToA2HostDeviceNetworkNone = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToA2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetworkNoneName,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToA2Annotation,
		},
		Labels: Pod01ToALabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToA2NicNone = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToA2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/network-status": Pod01ToA2Annotation,
		},
		Labels: Pod01ToALabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01To2ALabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "2",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Pod01To2ALabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToA,
	"pod-template-hash":      "2",
}
var Pod01To2A = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameTo2A,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01To2AAnnotation,
		},
		Labels: Pod01To2ALabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToA,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToCLabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToC,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Pod01ToCLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToC,
	"pod-template-hash":      "1",
}
var Pod01ToC1 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToC1,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToC1Annotation,
		},
		Labels: Pod01ToCLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToC,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToYLabels = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToY,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name + "_" + cDCIAppConnectionNamespace,
}
var Pod01ToYLabels0 = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToY,
	"pod-template-hash":      "1",
}
var Pod01ToYLabelsNoneNamespace = map[string]string{
	"app":                    "nginx",
	"app.kubernetes.io/name": cDeployment01NameToY,
	"pod-template-hash":      "1",
	cDCIAppConnectionKind:    cDCIAppConnection01Name,
}
var Pod01ToY1 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToY1,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToY1Annotation,
		},
		Labels: Pod01ToYLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToY,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}
var Pod01ToY2 = corev1.Pod{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cPodAPIVersion,
		Kind:       cPodKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cPod01NameToY2,
		Namespace: cPodNamespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks":       cHostDeviceNetwork01Name,
			"k8s.v1.cni.cncf.io/network-status": Pod01ToY2Annotation,
		},
		Labels: Pod01ToYLabels0,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cReplicaSetAPIVersion,
				Kind:       cReplicaSetKind,
				Name:       cReplicaSet01NameToY,
				UID:        "0",
			},
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "nginx",
				Image:   "nginx",
				Command: []string{"sh", "-c"},
				Args:    []string{"while true;do sleep 2; done"},
			},
		},
	},
	Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
	},
}

var Pod01FromX1Annotation = "[{ \"namea\":\"testeth0\", \"interface\": \"eth0\", \"" + cPodIps + "\": [\"100.100.100.10\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressFromX1 + "\"], \"dns\": {} }]"

//	"[{ \"name\":\"testeth0\",\
//		\"interface\": \"eth0\",\
//		\"" + cPodIps + "\": [\"100.100.100.10\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressFromX1 + "\"],
//		\"dns\": {}
//	}]"
var Pod01FromX1Anno2ndNICNone = "[{ \"namea\":\"testeth0\", \"interface\": \"eth0\", \"" + cPodIps + "\": [\"100.100.100.10\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"dns\": {} }]"

var Pod01FromX2Annotation = "[{ \"namea\":\"testeth0\", \"interface\": \"eth0\", \"" + cPodIps + "\": [\"100.100.100.10\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressFromX2 + "\"], \"dns\": {} }]"

//	"[{ \"name\":\"testeth0\",\
//		\"interface\": \"eth0\",\
//		\"" + cPodIps + "\": [\"100.100.100.10\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressFromX2 + "\"],
//		\"dns\": {}
//	}]"

var Pod01ToA1Annotation = "[{	\"name\": \"testeth1\", \"interface\": \"eth1\", \"" + cPodIps + "\": [\"100.100.100.11\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressToA1 + "\"], \"dns\": {} }]"

//	"[{	\"name\": \"testeth1\",
//		\"interface\": \"eth1\",
//		\"" + cPodIps + "\": [\"100.100.100.11\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressToA1 + "\"],
//		\"dns\": {}
//	}]"

var Pod01ToA1Anno2ndNICNone = "[{	\"name\": \"testeth1\", \"interface\": \"eth1\", \"" + cPodIps + "\": [\"100.100.100.11\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"dns\": {} }]"

var Pod01ToA2Annotation = "[{	\"name\": \"testeth2\", \"interface\": \"eth2\", \"" + cPodIps + "\": [\"100.100.100.12\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressToA2 + "\"], \"dns\": {} }]"

//	"[{	\"name\": \"testeth2\",
//		\"interface\": \"eth2\",
//		\"" + cPodIps + "\": [\"100.100.100.12\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressToA2 + "\"],
//		\"dns\": {}
//	}]"

var Pod01ToA2Anno2ndNICNone = "[{	\"name\": \"testeth2\", \"interface\": \"eth2\", \"" + cPodIps + "\": [\"100.100.100.12\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"dns\": {} }]"

// "[{	\"name\": \"testeth2\",
var Pod01ToA2AnnotationIPNone = "[{	\"name\": \"testeth2\", \"interface\": \"eth2\", \"" + cPodIps + "\": [\"100.100.100.12\"], \"default\": true, \"dns\": {} }]"

//	"[{	\"name\": \"testeth2\",
//		\"interface\": \"eth2\",
//		\"" + cPodIps + "\": [\"100.100.100.12\"],
//		\"default\": true,
//		\"dns\": {}
//	}]"

var Pod01To2AAnnotation = "[{\t\"name\": \"testeth2\", \"interface\": \"eth2\", \"" + cPodIps + "\": [\"100.100.100.12\"], \"default\": true, \"dns\": {} },{ \"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressTo2A + "\"], \"dns\": {} }]"

var Pod01ToC1Annotation = "[{	\"name\": \"testeth1\", \"interface\": \"eth1\", \"" + cPodIps + "\": [\"100.100.100.11\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressToC1 + "\"], \"dns\": {} }]"

//	"[{	\"name\": \"testeth1\",
//		\"interface\": \"eth1\",
//		\"" + cPodIps + "\": [\"100.100.100.11\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressToC1 + "\"],
//		\"dns\": {}
//	}]"

var Pod01ToY1Annotation = "[{	\"name\": \"testeth1\", \"interface\": \"eth1\", \"" + cPodIps + "\": [\"100.100.100.11\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressToY1 + "\"], \"dns\": {} }]"

//	"[{	\"name\": \"testeth3\",
//		\"interface\": \"eth3\",
//		\"" + cPodIps + "\": [\"100.100.101.10\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressToY1 + "\"],
//		\"dns\": {}
//	}]"

var Pod01ToY2Annotation = "[{	\"name\": \"testeth1\", \"interface\": \"eth1\", \"" + cPodIps + "\": [\"100.100.100.11\"], \"default\": true, \"dns\": {} },{	\"name\": \"default/macvlan-conf\", \"interface\": \"" + cPodIfName + "\", \"" + cPodIps + "\": [\"" + cPod01IpAddressToY2 + "\"], \"dns\": {} }]"

//	"[{	\"name\": \"testeth3\",
//		\"interface\": \"eth3\",
//		\"" + cPodIps + "\": [\"100.100.101.10\"],
//		\"default\": true,
//		\"dns\": {}
//	},{	\"name\": \"default/macvlan-conf\",
//		\"interface\": \"" + cPodIfName + "\",
//		\"" + cPodIps + "\": [\"" + cPod01IpAddressToY2 + "\"],
//		\"dns\": {}
//	}]"

// DCIAppConnection
// X-None,Y-A,Y-C,Y-X,in-X,out-Y
var DCIAppConnection09 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToX,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}
var DCIAppConnection09Status = examplecomv1.DCIAppConnectionStatus{
	Phase: "Created",
}
var ConnectEndpoint09FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint09ToY1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToX,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameFromX1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressFromX1,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

// DCIAppConnection
// X-A,X-Y,Y-C,in-Y,out-Y
var DCIAppConnection10 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}
var DCIAppConnection10Status = examplecomv1.DCIAppConnectionStatus{
	Phase: "Created",
}
var ConnectEndpoint10FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint10ToY1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

// DCIAppConnection
// X-A,X-C,X-Y,Y-C,Y-A,in-X,out-Y
var DCIAppConnection15 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}
var DCIAppConnection15Status = examplecomv1.DCIAppConnectionStatus{
	Phase: "Created",
}
var ConnectEndpoint15FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint15ToY1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

var DCIAppConnection12 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections:    []examplecomv1.Connection{},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnectionCE01OutputEndpoint,
	},
}

var ConnectEndpoint12FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To:             nil,
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: &ConnectEndpoint12ToX1SpecOutputEndpoint,
	},
}

var ConnectEndpoint12ToX1SpecOutputEndpoint = ConnectEndpoint01ToY1SpecOutputEndpoint

var DCIAppConnection13 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnectionCE01OutputEndpoint,
	},
}

var ConnectEndpoint13FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: &ConnectEndpoint12ToX1SpecOutputEndpoint,
	},
}

var DCIAppConnection14 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection01InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}

var ConnectEndpoint14FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}

// DCIAppConnection
// X-A,X-Y,Y-C,in-Y,out-Y
var DCIAppConnection11InputEndpoint = examplecomv1.DCIAppConnectionExternalEndpoint{
	Internal: examplecomv1.TargetRef{
		DeploymentRef: examplecomv1.NamespaceName{
			Name:      cDeployment01NameFromY,
			Namespace: cDeploymentNamespace,
		},
	},
	ExternalSubsets: []string{
		"rtsp://localhost:8554",
		"rtsp://localhost:8555",
	},
}
var DCIAppConnection11 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToY,
						Namespace: cDeploymentNamespace,
					},
				},
			}, {
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromY,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToC,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint:  &DCIAppConnection11InputEndpoint,
		OutputEndpoint: &DCIAppConnection01OutputEndpoint,
	},
}

var DCIAppConnection11Status = examplecomv1.DCIAppConnectionStatus{
	Phase: "Created",
}

var ConnectEndpoint11FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
	},
}

var ConnectEndpoint11ToY1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToC,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToC1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToC1,
					},
				},
			},
		},
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

var ConnectEndpoint11FromX1addY2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY2,
					},
				},
			},
		},
	},
}

var ConnectEndpoint11ToY2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameToY2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromY,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameToY2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressToY2,
			},
		},
		To: []examplecomv1.ToDeployment{{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameToC,
				Namespace: cDeploymentNamespace,
			},
			ToPod: []examplecomv1.ToPod{
				{
					PodRef: examplecomv1.NamespacedName{
						Name:      cPod01NameToC1,
						Namespace: cPodNamespace,
					},
					Address: cPod01IpAddressToC1,
				},
			},
		}},
		InputEndpoint:  &ConnectEndpoint01FromX1SpecInputEndpoint,
		OutputEndpoint: &ConnectEndpoint01ToY1SpecOutputEndpoint,
	},
}

var ConnectEndpoint11FromX2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					},
				},
			},
		},
	},
}

var ConnectEndpoint11FromX2addY2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			}, {
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToY,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY1,
					}, {
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToY2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToY2,
					},
				},
			},
		},
	},
}

var DCIAppConnection16 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		OutputEndpoint: &DCIAppConnectionCE01OutputEndpoint,
	},
}

var ConnectEndpoint16FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint12ToX1SpecOutputEndpoint,
	},
}
var ConnectEndpoint16FromX1NoneA2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint12ToX1SpecOutputEndpoint,
	},
}
var ConnectEndpoint16FromX2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		OutputEndpoint: &ConnectEndpoint12ToX1SpecOutputEndpoint,
	},
}

var DCIAppConnection17 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
		InputEndpoint: &DCIAppConnection01InputEndpoint,
	},
}

var ConnectEndpoint17FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint17FromX1NoneA2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}
var ConnectEndpoint17FromX2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
		InputEndpoint: &ConnectEndpoint01FromX1SpecInputEndpoint,
	},
}

var DCIAppConnection18 = examplecomv1.DCIAppConnection{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cDCIAppConnectionAPIVersion,
		Kind:       cDCIAppConnectionKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cDCIAppConnection01Name,
		Namespace: cDCIAppConnectionNamespace,
	},
	Spec: examplecomv1.DCIAppConnectionSpec{
		Connections: []examplecomv1.Connection{
			{
				From: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameFromX,
						Namespace: cDeploymentNamespace,
					},
				},
				To: examplecomv1.TargetRef{
					DeploymentRef: examplecomv1.NamespaceName{
						Name:      cDeployment01NameToA,
						Namespace: cDeploymentNamespace,
					},
				},
			},
		},
	},
}

var ConnectEndpoint18FromX1 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
	},
}
var ConnectEndpoint18FromX1NoneA2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX1,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX1,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX1,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
				},
			},
		},
	},
}
var ConnectEndpoint18FromX2 = examplecomv1.ConnectEndpoint{
	TypeMeta: metav1.TypeMeta{
		APIVersion: cConnectEndpointAPIVersion,
		Kind:       cConnectEndpointKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      cConnectEndpoint01NameFromX2,
		Namespace: cConnectEndpointNamespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cServiceAPIVersion,
				Kind:       cDCIAppConnectionKind,
				Name:       cDCIAppConnection01Name,
				UID:        "0",
			},
		},
	},
	Spec: examplecomv1.ConnectEndpointSpec{
		From: examplecomv1.FromDeployment{
			DeploymentRef: examplecomv1.NamespacedName{
				Name:      cDeployment01NameFromX,
				Namespace: cDeploymentNamespace,
			},
			FromPod: examplecomv1.FromPod{
				PodRef: examplecomv1.NamespacedName{
					Name:      cPod01NameFromX2,
					Namespace: cPodNamespace,
				},
				Address: cPod01IpAddressFromX2,
			},
		},
		To: []examplecomv1.ToDeployment{
			{
				DeploymentRef: examplecomv1.NamespacedName{
					Name:      cDeployment01NameToA,
					Namespace: cDeploymentNamespace,
				},
				ToPod: []examplecomv1.ToPod{
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA1,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA1,
					},
					{
						PodRef: examplecomv1.NamespacedName{
							Name:      cPod01NameToA2,
							Namespace: cPodNamespace,
						},
						Address: cPod01IpAddressToA2,
					},
				},
			},
		},
	},
}
