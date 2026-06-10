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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
)

// FromPodInfos defines the desired state of GetFromConnectionInformation.
type FromPodInfos struct {
	FromPods []FromPodInfo
}

type FromPodInfo struct {
	DeploymentRef     types.NamespacedName
	DeploymentR       *appsv1.Deployment
	PodR              *corev1.Pod
	ConnectEndpointCR *ConnectEndpoint
	Address           string
	DeploymentIndexs  []int
	IOMode            int
	MstPodFlag        bool
	DeleteFlag        bool
}

const (
	IOModeNONE  = 0
	IOModeIN    = 1
	IOModeOUT   = 2
	IOModeINOUT = IOModeIN + IOModeOUT
)

// ToPodInfos defines the desired state of GetToConnectionInformation.
type ToPodInfos struct {
	ToPodInfos map[int]ToPods
}

type ToPods struct {
	DeploymentRef  types.NamespacedName
	DeploymentR    *appsv1.Deployment
	ServiceR       *corev1.Service
	EndpointSliceR *discoveryv1.EndpointSlice
	ToPods         []ToPodInfo
}

type ToPodInfo struct {
	PodR       *corev1.Pod
	Address    string
	DeleteFlag bool
}

type NetworkStatus struct {
	Name      string   `json:"name"`
	Interface string   `json:"interface,omitempty"`
	Ips       []string `json:"ips,omitempty"`
}
