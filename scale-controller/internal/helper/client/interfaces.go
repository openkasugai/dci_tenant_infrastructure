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

package client

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/types"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
)

// ClientHelper defines helper functions that use clients.
type ClientHelper interface {
	ClientNodeHelper
	ClientDeviceTaintRuleHelper
	ClientMachineHelper
	ClientMachineSetHelper
	ClientDCIMachineHelper
	ClientMachineDeploymentHelper
	ClientDCIMachineTemplateHelper
	ClientDockerMachineTemplateHelper
	ClientLogicalServerHelper
	ClientMachineAlertHelper
}

// ClientDCIMachineHelper defines helper functions that use clients.
type ClientDCIMachineHelper interface {
	// GetDCIMachine gets DCIMachine resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachineName: Name of the DCIMachine resource
	//   - namespace: Namespace of the DCIMachine resource
	// Returns:
	//   - dciMachine: DCIMachine resource
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachine(ctx context.Context, dciMachineName, namespace string,
	) (dciMachine *capdv1.DCIMachine, err error)

	// GetDCIMachineFromMachine gets DCIMachine resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machine: Machine resource
	//   - namespace: Namespace of the management cluster to get the DCIMachine resource
	// Returns:
	//   - dciMachine: DCIMachine resource
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachineFromMachine(ctx context.Context, machine *capiv1.Machine, namespace string,
	) (dciMachine *capdv1.DCIMachine, err error)

	// PatchDCIMachine patches DCIMachine resource in management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachine: DCIMachine resource
	//   - lastDCIMachine: Last DCIMachine resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	PatchDCIMachine(ctx context.Context, dciMachine *capdv1.DCIMachine, lastDCIMachine *capdv1.DCIMachine,
	) (err error)
}

// ClientDCIMachineTemplateHelper defines helper functions that use clients.
type ClientDCIMachineTemplateHelper interface {
	// GetDCIMachineTemplate gets DCIMachineTemplate resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachineTemplateName: Name of the DCIMachineTemplate resource
	//   - namespace: Namespace of the DCIMachineTemplate resource
	// Returns:
	//   - dciMachineTemplate: DCIMachineTemplate resource
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachineTemplate(ctx context.Context, dciMachineTemplateName, namespace string,
	) (dciMachineTemplate *capdv1.DCIMachineTemplate, err error)

	// GetDCIMachineTemplateFromMachineDeployment gets DCIMachineTemplate resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineDeployment: MachineDeployment resource
	//   - namespace: Namespace of the management cluster to get the DCIMachineTemplate resource
	// Returns:
	//   - dciMachineTemplate: DCIMachineTemplate resource
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachineTemplateFromMachineDeployment(ctx context.Context, machineDeployment *capiv1.MachineDeployment, namespace string,
	) (dciMachineTemplate *capdv1.DCIMachineTemplate, err error)

	// GetDCIMachineTemplatesAllocatable gets allocatable DCIMachineTemplate list or new unique name.
	// Parameters:
	//   - ctx: Context for the request
	//   - location: Location of the DCIMachineTemplate resource
	//   - serverSpec: ServerSpec of the DCIMachineTemplate resource
	//   - namespace: Namespace of the management cluster to get the DCIMachineTemplate resource
	// Returns:
	//   - templates: DCIMachineTemplate list
	//   - uniqueName: New unique name
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachineTemplatesAllocatable(ctx context.Context, location *string, serverSpec *rpmv1.ServerSpec, namespace string,
	) (templates []client.Object, uniqueName string, err error)

	// GetDCIMachineTemplatesValid gets valid DCIMachineTemplate list or new unique name.
	// Parameters:
	//   - ctx: Context for the request
	//   - location: Location of the DCIMachineTemplate resource
	//   - serverSpec: ServerSpec of the DCIMachineTemplate resource
	//   - namespace: Namespace of the management cluster to get the DCIMachineTemplate resource
	// Returns:
	//   - templates: DCIMachineTemplate list
	//   - uniqueName: New unique name
	//   - err: Error object if an error occurred, otherwise nil
	GetDCIMachineTemplatesValid(ctx context.Context, location *string, serverSpec *rpmv1.ServerSpec, namespace string,
	) (templates []client.Object, uniqueName string, err error)

	// CreateDCIMachineTemplate creates DCIMachineTemplate resource to management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachineTemplate: DCIMachineTemplate resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateDCIMachineTemplate(ctx context.Context, dciMachineTemplate client.Object,
	) (err error)

	// CreateDCIMachineTemplateFromNodeScalingRequest creates DCIMachineTemplate resource from NodeScalingRequest.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the DCIMachineTemplate resource
	//   - namespace: Namespace of the management cluster to create the DCIMachineTemplate resource
	//   - nsr: NodeScalingRequest resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateDCIMachineTemplateFromNodeScalingRequest(ctx context.Context, name, namespace string, nsr *scalev1.NodeScalingRequest,
	) (err error)

	// DeleteDCIMachineTemplate deletes DCIMachineTemplate resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachineTemplate: DCIMachineTemplate resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteDCIMachineTemplate(ctx context.Context, dciMachineTemplate *capdv1.DCIMachineTemplate,
	) (err error)

	// DeleteDCIMachineTemplateUnused is a function to delete DCIMachineTemplate if it is not referred by other MachineDeployment.
	// Parameters:
	//   - ctx: Context for the request
	//   - md: MachineDeployment resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteDCIMachineTemplateUnused(ctx context.Context, md *capiv1.MachineDeployment,
	) (err error)

	// IsReferredDCIMachineTemplate checks if the DCIMachineTemplate is referred by any MachineDeployment.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachineTemplate: DCIMachineTemplate resource
	//   - machineDeployment: MachineDeployment resource
	//   - namespace: Namespace of the DCIMachineTemplate resource
	// Returns:
	//   - isReferred: true if the DCIMachineTemplate is referred by any MachineDeployment, otherwise false
	//   - err: Error object if an error occurred, otherwise nil
	IsReferredDCIMachineTemplate(ctx context.Context, dciMachineTemplate *capdv1.DCIMachineTemplate,
		machineDeployment *capiv1.MachineDeployment, namespace string,
	) (isReferred bool, err error)
}

// ClientDockerMachineTemplateHelper defines helper functions that use clients.
type ClientDockerMachineTemplateHelper interface {
	// GetDockerMachineTemplate gets DockerMachineTemplate resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dockerMachineTemplateName: Name of the DockerMachineTemplate resource
	//   - namespace: Namespace of the DockerMachineTemplate resource
	// Returns:
	//   - dockerMachineTemplate: DockerMachineTemplate resource
	//   - err: Error object if an error occurred, otherwise nil
	GetDockerMachineTemplate(ctx context.Context, dockerMachineTemplateName, namespace string,
	) (dockerMachineTemplate *capdockerv1.DockerMachineTemplate, err error)

	// GetDockerMachineTemplatesAllocatable gets allocatable DockerMachineTemplate list or new unique name.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the DockerMachineTemplate resource
	//   - namespace: Namespace of the management cluster to get the DockerMachineTemplate resource
	// Returns:
	//   - templates: DockerMachineTemplate list
	//   - uniqueName: New unique name
	//   - err: Error object if an error occurred, otherwise nil
	GetDockerMachineTemplatesAllocatable(ctx context.Context, name, namespace string,
	) (templates []client.Object, uniqueName string, err error)

	// CreateDockerMachineTemplate creates DockerMachineTemplate resource to management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dockerMachineTemplate: DockerMachineTemplate resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateDockerMachineTemplate(ctx context.Context, dockerMachineTemplate client.Object,
	) (err error)

	// CreateDockerMachineTemplateFromNodeScalingRequest creates DockerMachineTemplate resource from NodeScalingRequest.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the DockerMachineTemplate resource
	//   - namespace: Namespace of the management cluster to create the DockerMachineTemplate resource
	//   - nsr: NodeScalingRequest resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateDockerMachineTemplateFromNodeScalingRequest(ctx context.Context, name, namespace string, nsr *scalev1.NodeScalingRequest,
	) (err error)

	// DeleteDockerMachineTemplate deletes DockerMachineTemplate resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dockerMachineTemplate: DockerMachineTemplate resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteDockerMachineTemplate(ctx context.Context, dockerMachineTemplate *capdockerv1.DockerMachineTemplate,
	) (err error)

	// DeleteDockerMachineTemplateUnused is a function to delete DockerMachineTemplate if it is not referred by other MachineDeployment.
	// Parameters:
	//   - ctx: Context for the request
	//   - md: MachineDeployment resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteDockerMachineTemplateUnused(ctx context.Context, md *capiv1.MachineDeployment,
	) (err error)

	// IsReferredDockerMachineTemplate checks if the DockerMachineTemplate is referred by any MachineDeployment.
	// Parameters:
	//   - ctx: Context for the request
	//   - dockerMachineTemplate: DockerMachineTemplate resource
	//   - machineDeployment: MachineDeployment resource
	//   - namespace: Namespace of the management cluster to get the DockerMachineTemplate resource
	// Returns:
	//   - isReferred: true if the DockerMachineTemplate is referred by any MachineDeployment, otherwise false
	//   - err: Error object if an error occurred, otherwise nil
	IsReferredDockerMachineTemplate(ctx context.Context, dockerMachineTemplate *capdockerv1.DockerMachineTemplate,
		machineDeployment *capiv1.MachineDeployment, namespace string,
	) (isReferred bool, err error)
}

// ClientMachineHelper defines helper functions that use clients.
type ClientMachineHelper interface {
	// GetMachine gets Machine resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the Machine resource
	//   - namespace: Namespace of the Machine resource
	// Returns:
	//   - machine: Machine resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachine(ctx context.Context, name, namespace string,
	) (machine *capiv1.Machine, err error)

	// GetMachineFromNode gets Machine resource from Node annotations.
	// Parameters:
	//   - ctx: Context for the request
	//   - node: Node resource
	// Returns:
	//   - machine: Machine resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineFromNode(ctx context.Context, node *corev1.Node,
	) (machine *capiv1.Machine, err error)

	// DeleteMachine deletes Machine resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machine: Machine resource to be deleted
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteMachine(ctx context.Context, machine *capiv1.Machine,
	) (err error)
}

// ClientMachineDeploymentHelper defines helper functions that use clients.
type ClientMachineDeploymentHelper interface {
	// GetMachineDeployment gets MachineDeployment resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the MachineDeployment resource
	//   - namespace: Namespace of the MachineDeployment resource
	// Returns:
	//   - machineDeployment: MachineDeployment resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineDeployment(ctx context.Context, name, namespace string,
	) (machineDeployment *capiv1.MachineDeployment, err error)

	// GetMachineDeploymentFromMachine gets MachineDeployment resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machine: Machine resource
	//   - namespace: Namespace of the management cluster to get the MachineDeployment resource
	// Returns:
	//   - machineDeployment: MachineDeployment resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineDeploymentFromMachine(ctx context.Context, machine *capiv1.Machine, namespace string,
	) (machineDeployment *capiv1.MachineDeployment, err error)

	// GetMachineDeploymentsAllocatable gets allocatable MachineDeployment list or new unique name.
	// Parameters:
	//   - ctx: Context for the request
	//   - namespace: Namespace of the management cluster to get the MachineDeployment resource
	//   - machineTemplateType: MachineTemplate type of the MachineDeployment resource
	//   - clusterName: Cluster name of the MachineDeployment resource
	//   - machineTemplates: MachineTemplate list of the MachineDeployment resource
	// Returns:
	//   - machineDeployments: MachineDeployment list
	//   - uniqueName: New unique name
	//   - originalMachineDeployment: Original MachineDeployment resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineDeploymentsAllocatable(ctx context.Context, namespace, machineTemplateType, clusterName string, machineTemplates []client.Object,
	) (machineDeployments []capiv1.MachineDeployment, uniqueName string, originalMachineDeployment *capiv1.MachineDeployment, err error)

	// GetMachineDeploymentsFromTemplate gets MachineDeployment list that are used by NodeScalingRequest.
	// Parameters:
	//   - ctx: Context for the request
	//   - namespace: Namespace of the management cluster to get the MachineDeployment resource
	//   - md: MachineDeployment resource
	//   - machineTemplate: MachineTemplate resource
	// Returns:
	//   - machineDeployments: MachineDeployment list
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineDeploymentsFromTemplate(ctx context.Context, namespace string, md *capiv1.MachineDeployment, machineTemplate client.Object,
	) (machineDeployments []capiv1.MachineDeployment, err error)

	// CreateMachineDeployment creates MachineDeployment resource to management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineDeployment: MachineDeployment resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateMachineDeployment(ctx context.Context, machineDeployment *capiv1.MachineDeployment,
	) (err error)

	// CreateMachineDeploymentFromNodeScalingRequest creates MachineDeployment resource from NodeScalingRequest.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the MachineDeployment resource
	//   - namespace: Namespace of the management cluster to create the MachineDeployment resource
	//   - machineTemplateName: Name of the MachineTemplate resource
	//   - nodeScalingRequest: NodeScalingRequest resource
	//   - originalMachineDeployment: Original MachineDeployment resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateMachineDeploymentFromNodeScalingRequest(ctx context.Context, name, namespace, machineTemplateName string,
		nodeScalingRequest *scalev1.NodeScalingRequest, originalMachineDeployment *capiv1.MachineDeployment,
	) (err error)

	// PatchMachineDeployment patches MachineDeployment resource in management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineDeployment: MachineDeployment resource
	//   - lastMachineDeployment: Last MachineDeployment resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	PatchMachineDeployment(ctx context.Context, machineDeployment *capiv1.MachineDeployment, lastMachineDeployment *capiv1.MachineDeployment,
	) (err error)

	// DeleteMachineDeployment deletes MachineDeployment resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineDeployment: MachineDeployment resource to be deleted
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteMachineDeployment(ctx context.Context, machineDeployment *capiv1.MachineDeployment,
	) (err error)

	// IsLastMachineDeployment checks if the MachineDeployment is the last one in the cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineDeployment: MachineDeployment resource
	//   - namespace: Namespace of the management cluster to get the MachineDeployment resource
	//   - clusterName: Cluster name of the MachineDeployment resource
	// Returns:
	//   - isLast: true if the MachineDeployment is the last one, otherwise false
	//   - err: Error object if an error occurred, otherwise nil
	IsLastMachineDeployment(ctx context.Context, machineDeployment *capiv1.MachineDeployment, namespace, clusterName string,
	) (isLast bool, err error)

	// IsAutoScalingEnabledMachineDeployment checks if the MachineDeployment has auto-scaling enabled.
	// Parameters:
	//    - md: MachineDeployment resource
	// Returns:
	//    - isAutoScalingEnabled: true if the MachineDeployment has auto-scaling enabled, otherwise false
	IsAutoScalingEnabledMachineDeployment(md *capiv1.MachineDeployment,
	) (isAutoScalingEnabled bool)
}

// ClientMachineSetHelper defines helper functions that use clients.
type ClientMachineSetHelper interface {
	// GetMachineSetFromUID gets the MachineSet owned by a MachineDeployment from its UID.
	// Parameters:
	//   - ctx: Context for the request
	//   - namespace: Namespace of the MachineSet resource
	//   - machineDeploymentUID: UID of the owner MachineDeployment resource
	// Returns:
	//   - machineSet: MachineSet resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineSetFromUID(ctx context.Context, namespace string, machineDeploymentUID types.UID,
	) (machineSet *capiv1.MachineSet, err error)

	// PatchMachineSet patches the MachineSet resource in management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineSet: MachineSet resource to be patched
	//   - lastMachineSet: Last MachineSet resource
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	PatchMachineSet(ctx context.Context, machineSet *capiv1.MachineSet, lastMachineSet *capiv1.MachineSet,
	) (err error)
}

// ClientNodeHelper defines helper functions that use clients.
type ClientNodeHelper interface {
	// GetNode gets Node resource from workload cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - nodeName: Name of the Node resource
	// Returns:
	//   - node: Node resource
	//   - err: Error object if an error occurred, otherwise nil
	GetNode(ctx context.Context, nodeName string,
	) (node *corev1.Node, err error)

	// GetClusterNameAndNamespace gets ClusterName and ClusterNamespace from Node labels.
	// Parameters:
	//   - ctx: Context for the request
	// Returns:
	//   - clusterName: Cluster name
	//   - clusterNamespace: Cluster namespace
	//   - err: Error object if an error occurred, otherwise nil
	GetClusterNameAndNamespace(ctx context.Context,
	) (clusterName string, clusterNamespace string, err error)

	// GetClusterNameAndNamespaceFromNode gets ClusterName and ClusterNamespace from Node labels.
	// Parameters:
	//   - node: Node resource
	// Returns:
	//   - clusterName: Cluster name
	//   - clusterNamespace: Cluster namespace
	//   - err: Error object if an error occurred, otherwise nil
	GetClusterNameAndNamespaceFromNode(node *corev1.Node,
	) (clusterName string, clusterNamespace string, err error)
}

// ClientDeviceTaintRuleHelper defines helper functions that use clients.
type ClientDeviceTaintRuleHelper interface {
	// ListDeviceTaintRule lists DeviceTaintRule resources from workload cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - namespace: Namespace of the DeviceTaintRule resources
	// Returns:
	//   - deviceTaintRuleList: List of DeviceTaintRule resources
	//   - err: Error object if an error occurred, otherwise nil
	ListDeviceTaintRule(ctx context.Context, namespace string,
	) (deviceTaintRuleList *resourcev1.DeviceTaintRuleList, err error)

	// CreateDeviceTaintRule creates a DeviceTaintRule resource in the workload cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - deviceTaintRule: DeviceTaintRule resource to be created
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateDeviceTaintRule(ctx context.Context, deviceTaintRule *resourcev1.DeviceTaintRule,
	) (err error)
}

// ClientLogicalServerHelper defines helper functions that use clients.
type ClientLogicalServerHelper interface {
	// GetLogicalServer gets LogicalServer resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the LogicalServer resource
	//   - namespace: Namespace of the management cluster to get the LogicalServer resource
	// Returns:
	//   - logicalServer: LogicalServer resource
	//   - err: Error object if an error occurred, otherwise nil
	GetLogicalServer(ctx context.Context, name, namespace string,
	) (*rpmv1.LogicalServer, error)

	// GetLogicalServerNameAndNamespaceFromDCIMachine gets LogicalServer name and namespace from DCIMachine status.
	// Parameters:
	//   - dciMachine: DCIMachine resource
	// Returns:
	//   - name: Name of the LogicalServer resource
	//   - namespace: Namespace of the LogicalServer resource
	//   - err: Error object if an error occurred, otherwise nil
	GetLogicalServerNameAndNamespaceFromDCIMachine(dciMachine *capdv1.DCIMachine,
	) (name, namespace string, err error)

	// GetLogicalServerFromDCIMachine gets LogicalServer resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - dciMachine: DCIMachine resource
	//   - namespace: Namespace of the management cluster to get the LogicalServer resource
	// Returns:
	//   - logicalServer: LogicalServer resource
	//   - err: Error object if an error occurred, otherwise nil
	GetLogicalServerFromDCIMachine(ctx context.Context, dciMachine *capdv1.DCIMachine,
	) (*rpmv1.LogicalServer, error)
}

// ClientMachineAlertHelper defines helper functions that use clients.
type ClientMachineAlertHelper interface {
	// GetMachineAlert gets MachineAlert resource from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - name: Name of the MachineAlert resource
	//   - namespace: Namespace of the MachineAlert resource
	// Returns:
	//   - machineAlert: MachineAlert resource
	//   - err: Error object if an error occurred, otherwise nil
	GetMachineAlert(ctx context.Context, name, namespace string,
	) (machineAlert *rpmv1.MachineAlert, err error)

	// ListMachineAlert lists MachineAlert resources from management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - namespace: Namespace of the MachineAlert resources
	// Returns:
	//   - machineAlertList: List of MachineAlert resources
	//   - err: Error object if an error occurred, otherwise nil
	ListMachineAlert(ctx context.Context, namespace string,
	) (machineAlertList *rpmv1.MachineAlertList, err error)

	// CreateMachineAlert creates a MachineAlert resource in the management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineAlert: MachineAlert resource to be created
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	CreateMachineAlert(ctx context.Context, machineAlert *rpmv1.MachineAlert,
	) (err error)

	// DeleteMachineAlert deletes a MachineAlert resource in the management cluster.
	// Parameters:
	//   - ctx: Context for the request
	//   - machineAlert: MachineAlert resource to be deleted
	// Returns:
	//   - err: Error object if an error occurred, otherwise nil
	DeleteMachineAlert(ctx context.Context, machineAlert *rpmv1.MachineAlert,
	) (err error)
}
