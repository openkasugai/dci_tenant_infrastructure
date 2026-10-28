package testutils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	// TODO: Switch to "k8s.io/api/resource/v1" once DeviceTaintRule is supported in v1.
	// DeviceTaintRule is currently available only in v1beta2, so we use v1beta2 for now.
	// However, ResourceSlice in v1beta2 is deprecated in v1.36+ and will be removed in v1.39+,
	// so we should migrate to v1 as soon as DeviceTaintRule is available there.
	resourcev1beta "k8s.io/api/resource/v1beta2"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
)

func NewCustomResourceDefinition(
	kind, plural string, gv schema.GroupVersion,
) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: plural + "." + gv.Group,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: gv.Group,
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:   cases.Title(language.Und, cases.NoLower).String(kind),
				Plural: plural,
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    gv.Version,
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": {
									Type:                   "object",
									XPreserveUnknownFields: ptr.To(true),
								},
								"status": {
									Type:                   "object",
									XPreserveUnknownFields: ptr.To(true),
								},
							},
						},
					},
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
				},
			},
		},
	}
}

// NewNamespacedName returns types.NamespacedName by one line.
func NewNamespacedName(
	name, namespace string,
) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

// NewNodeScalingRequest returns *scalev1.NodeScalingRequest by one line.
// Set only Name, Namespace and NodeRequestType, and the other parameters are nil.
func NewNodeScalingRequest(
	name, namespace string, reqType scalev1.NodeScalingRequestType,
) *scalev1.NodeScalingRequest {
	return &scalev1.NodeScalingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: scalev1.NodeScalingRequestSpec{
			NodeRequestType: reqType,
			Location:        nil,
			Node:            nil,
			ServerSpec:      nil,
		},
	}
}

// NewServerSpecTemplate returns *rpmv1.ServerSpec by one line.
// Set only Type, and the other parameters are nil.
func NewServerSpecTemplate(
	serverType rpmv1.ReqServerType,
) *rpmv1.ServerSpec {
	return &rpmv1.ServerSpec{
		Type:       serverType,
		OS:         nil,
		MinCPU:     nil,
		MinMemory:  nil,
		MinStorage: nil,
		GPU:        nil,
		NIC:        nil,
	}
}

// SetServerSpec set parameters by one line.
func SetServerSpec(
	spec *rpmv1.ServerSpec, osType *rpmv1.OSType, cpu, mem, storage *int32, gpu *rpmv1.GPUSpec, nic *rpmv1.NICSpec,
) {
	spec.OS = osType
	spec.MinCPU = cpu
	spec.MinMemory = mem
	spec.MinStorage = storage
	spec.GPU = gpu
	spec.NIC = nic
}

// NewServerSpec returns *rpmv1.ServerSpec by one line.
func NewServerSpec(
	serverType rpmv1.ReqServerType, osType *rpmv1.OSType, cpu, mem, storage *int32, gpu *rpmv1.GPUSpec, nic *rpmv1.NICSpec,
) *rpmv1.ServerSpec {
	serverSpec := NewServerSpecTemplate(serverType)
	SetServerSpec(serverSpec, osType, cpu, mem, storage, gpu, nic)
	return serverSpec
}

// NewNode returns *corev1.Node by one line.
// Annotations are set for emulating cluster-api.
func NewNode(
	name, clusterName, clusterNamespace, machineName string,
) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				capiv1.ClusterNameAnnotation:      clusterName,
				capiv1.ClusterNamespaceAnnotation: clusterNamespace,
				capiv1.MachineAnnotation:          machineName,
			},
		},
	}
}

// NewMachine returns *capiv1.Machine by one line.
// Labels are set for emulating cluster-api.
func NewMachine(
	name, namespace, mdName, clusterName string,
) *capiv1.Machine {
	return &capiv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				capiv1.MachineDeploymentNameLabel: mdName,
			},
		},
		Spec: capiv1.MachineSpec{
			ClusterName: clusterName,
			Bootstrap: capiv1.Bootstrap{
				ConfigRef: capiv1.ContractVersionedObjectReference{
					APIGroup: capiv1.GroupVersion.Group,
					Kind:     "KubeadmConfig",
					Name:     "dummy-config",
				},
			},
			InfrastructureRef: capiv1.ContractVersionedObjectReference{
				APIGroup: capiv1.GroupVersion.Group,
				Kind:     "dummy-infra-kind",
				Name:     "dummy-infra",
			},
		},
	}
}

// NewMachineDeployment returns *capiv1.MachineDeployment by one line.
func NewMachineDeployment(
	name, namespace, clusterName, apiGroup, kind, templateName string, replicas *int32,
) *capiv1.MachineDeployment {
	return &capiv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: capiv1.MachineDeploymentSpec{
			ClusterName: clusterName,
			Replicas:    replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					capiv1.ClusterNameLabel: clusterName,
				},
			},
			Template: capiv1.MachineTemplateSpec{
				Spec: capiv1.MachineSpec{
					Deletion: capiv1.MachineDeletionSpec{
						NodeDrainTimeoutSeconds: ptr.To(int32(180)),
					},
					ClusterName: clusterName,
					Bootstrap: capiv1.Bootstrap{
						ConfigRef: capiv1.ContractVersionedObjectReference{
							APIGroup: capiv1.GroupVersion.Group,
							Kind:     "KubeadmConfigTemplate",
							Name:     "dummy-config-template",
						},
					},
					InfrastructureRef: capiv1.ContractVersionedObjectReference{
						APIGroup: apiGroup,
						Kind:     kind,
						Name:     templateName,
					},
				},
			},
		},
	}
}

// NewMachineSet returns *capiv1.MachineSet by one line.
func NewMachineSet(
	name, namespace, clusterName, mdName string, replicas *int32, uid types.UID,
) *capiv1.MachineSet {
	return &capiv1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: capiv1.GroupVersion.String(),
					Kind:       "MachineDeployment",
					Name:       mdName,
					UID:        uid, // UID of the owner MachineDeployment
				},
			},
		},
		Spec: capiv1.MachineSetSpec{
			ClusterName: clusterName,
			Replicas:    replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					capiv1.MachineDeploymentNameLabel: mdName,
				},
			},
			Template: capiv1.MachineTemplateSpec{
				Spec: capiv1.MachineSpec{
					ClusterName: clusterName,
					Bootstrap: capiv1.Bootstrap{
						ConfigRef: capiv1.ContractVersionedObjectReference{
							APIGroup: capiv1.GroupVersion.Group,
							Kind:     "KubeadmConfigTemplate",
							Name:     "dummy-config-template",
						},
					},
					InfrastructureRef: capiv1.ContractVersionedObjectReference{
						APIGroup: capiv1.GroupVersion.Group,
						Kind:     "dummy-infra-kind",
						Name:     "dummy-infra",
					},
				},
			},
		},
	}
}

// NewDCIMachineTemplate returns *capdv1.DCIMachineTemplate by one line.
// Location and ServerSpec are nil and empty.
func NewDCIMachineTemplate(
	name, namespace string,
) *capdv1.DCIMachineTemplate {
	return &capdv1.DCIMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: capdv1.DCIMachineTemplateSpec{
			Template: capdv1.DCIMachineTemplateResource{
				Spec: capdv1.DCIMachineSpec{
					Location:   nil,
					ServerSpec: rpmv1.ServerSpec{},
				},
			},
		},
	}
}

// NewDockerMachineTemplate returns *capdockerv1.DockerMachineTemplate by one line.
// Location and ServerSpec are nil and empty.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func NewDockerMachineTemplate(
	name, namespace string,
) *capdockerv1.DockerMachineTemplate {
	return &capdockerv1.DockerMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: capdockerv1.DockerMachineTemplateSpec{
			Template: capdockerv1.DockerMachineTemplateResource{
				Spec: capdockerv1.DockerMachineSpec{},
			},
		},
	}
}

// NewDCIMachine returns *capdv1.DCIMachine by one line.
func NewDCIMachine(
	dciMachineName, dciMachineNamespace string, location *string, providerID string, serverSpec *rpmv1.ServerSpec,
) *capdv1.DCIMachine {
	if serverSpec == nil {
		return nil
	}
	return &capdv1.DCIMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dciMachineName,
			Namespace: dciMachineNamespace,
		},
		Spec: capdv1.DCIMachineSpec{
			Location:   location,
			ProviderID: providerID,
			ServerSpec: *serverSpec,
		},
	}
}

// NewLogicalServer returns *rpmv1.LogicalServer by one line.
func NewLogicalServer(
	lsName, lsNamespace string, location *string, serverSpec *rpmv1.ServerSpec, role rpmv1.LogicalServerRoleType,
) *rpmv1.LogicalServer {
	if serverSpec == nil {
		return nil
	}
	return &rpmv1.LogicalServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lsName,
			Namespace: lsNamespace,
		},
		Spec: rpmv1.LogicalServerSpec{
			Location:   location,
			ServerSpec: *serverSpec,
			Role:       role,
			TenantLogicalInfraRef: &corev1.ObjectReference{
				APIVersion: capdv1.GroupVersion.String(),
				Kind:       "DCICluster",
				Name:       "dummy-dci-cluster",
				Namespace:  "dummy-dci-cluster-namespace",
				UID:        "dummy-dci-cluster-uid",
			},
		},
	}
}

// NewDeviceTaintRule returns *resourcev1beta.DeviceTaintRule by one line.
func NewDeviceTaintRule(
	name, key, val string, effect resourcev1beta.DeviceTaintEffect,
	driverName, poolName, deviceName *string,
) *resourcev1beta.DeviceTaintRule {
	return &resourcev1beta.DeviceTaintRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1beta.DeviceTaintRuleSpec{
			Taint: resourcev1beta.DeviceTaint{
				Key:    key,
				Value:  val,
				Effect: effect,
			},
			DeviceSelector: &resourcev1beta.DeviceTaintSelector{
				Driver: driverName,
				Pool:   poolName,
				Device: deviceName,
			},
		},
	}
}

// NewResourceSlice returns *resourcev1.ResourceSlice by one line.
func NewResourceSlice(
	name, driverName, poolName string, devices map[string]map[string]string,
) *resourcev1.ResourceSlice {
	rs := &resourcev1.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1.ResourceSliceSpec{
			Driver:   driverName,
			NodeName: &poolName,
			Pool: resourcev1.ResourcePool{
				Name:               poolName,
				ResourceSliceCount: int64(len(devices)),
			},
			Devices: make([]resourcev1.Device, 0, len(devices)),
		},
	}
	for deviceName, device := range devices {
		attrs := map[resourcev1.QualifiedName]resourcev1.DeviceAttribute{}
		for key, val := range device {
			attrs[resourcev1.QualifiedName(key)] = resourcev1.DeviceAttribute{
				StringValue: &val,
			}
		}
		rs.Spec.Devices = append(rs.Spec.Devices, resourcev1.Device{
			Name:       deviceName,
			Attributes: attrs,
		})
	}
	return rs
}

// NewCardScalingRequest returns *scalev1.CardScalingRequest by one line.
// Set only Name, Namespace and NodeName, and the other parameters are nil.
func NewCardScalingRequest(
	name, namespace, nodeName string,
) *scalev1.CardScalingRequest {
	return &scalev1.CardScalingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: scalev1.CardScalingRequestSpec{
			NodeName: nodeName,
		},
	}
}

// NewScalingParameter returns *scalev1.ScalingParameter by one line.
func NewScalingParameter(
	operation capdv1.DCIMachineCardScalingOp, model string, num int32, uuids []string,
) *scalev1.ScalingParameter {
	return &scalev1.ScalingParameter{
		Op:    operation,
		Model: model,
		Num:   num,
		UUIDs: uuids,
	}
}

// NewCardAlert returns *scalev1.CardAlert by one line.
// Set only Name, Namespace and NodeName, and the other parameters are nil.
func NewCardAlert(
	name, namespace, nodeName string, cards []scalev1.CardInfo, source scalev1.CardAlertSourceType,
) *scalev1.CardAlert {
	return &scalev1.CardAlert{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: scalev1.CardAlertSpec{
			NodeName: nodeName,
			Cards:    cards,
			Source:   source,
		},
	}
}
