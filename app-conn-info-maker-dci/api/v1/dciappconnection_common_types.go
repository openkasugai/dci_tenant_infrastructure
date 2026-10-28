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
	InputSubsets      []string
	OutputSubsets     []string
	IOMode            int
	MstPodFlag        bool
	DeleteFlag        bool
}

// FromDeploymentInfos defines the desired state of getStatusRelationship.
type FromDeploymentInfos struct {
	FromDeployments []FromDeploymentInfo
}

type FromDeploymentInfo struct {
	DeploymentRef    types.NamespacedName
	DeploymentR      *appsv1.Deployment
	DeploymentIndexs []int
	InputSubsets     []string
	OutputSubsets    []string
	IOMode           int
	DeleteFlag       bool
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

// ToDeploymentInfos defines the desired state of GetToConnectionInformation.
type ToDeploymentInfos struct {
	ToDeploymentInfos map[int]ToDeployments
}

type ToDeployments struct {
	DeploymentRef types.NamespacedName
	DeleteFlag    bool
}

type NetworkStatus struct {
	Name      string   `json:"name"`
	Interface string   `json:"interface,omitempty"`
	Ips       []string `json:"ips,omitempty"`
}
