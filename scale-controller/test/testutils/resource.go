package testutils

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/internal/helper/utils"
)

func cleanupResource(
	ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object,
) error {
	if err := c.Get(ctx, key, obj); apierrors.IsNotFound(err) {
		return nil // no need to delete
	} else if err != nil {
		return err
	}
	return c.Delete(ctx, obj)
}

// CreateNodeResource creates a Node resource in the cluster.
func CreateNodeResource(
	ctx context.Context, c client.Client, name, clusterName, clusterNamespace, machineName string,
) error {
	return c.Create(ctx, NewNode(name, clusterName, clusterNamespace, machineName))
}

// CleanupNodeResource deletes a Node resource from the cluster.
func CleanupNodeResource(
	ctx context.Context, c client.Client, name string,
) error {
	return cleanupResource(ctx, c, types.NamespacedName{Name: name}, &corev1.Node{})
}

// CleanupNodeResources deletes all Node resources in the cluster.
func CleanupNodeResources(
	ctx context.Context, c client.Client,
) error {
	return c.DeleteAllOf(ctx, &corev1.Node{})
}

// CreateDeviceTaintRuleResource creates a DeviceTaintRule resource in the cluster.
func CreateDeviceTaintRuleResource(
	ctx context.Context, c client.Client, name, key, val string, effect resourcev1.DeviceTaintEffect,
	driverName, poolName, deviceName *string,
) error {
	return c.Create(ctx, NewDeviceTaintRule(name, key, val, effect, driverName, poolName, deviceName))
}

// CleanupDeviceTaintRuleResource deletes a DeviceTaintRule resource from the cluster.
func CleanupDeviceTaintRuleResource(
	ctx context.Context, c client.Client, name string,
) error {
	return cleanupResource(ctx, c, types.NamespacedName{Name: name}, &resourcev1.DeviceTaintRule{})
}

// CleanupDeviceTaintRuleResources deletes all DeviceTaintRule resources in the cluster.
func CleanupDeviceTaintRuleResources(
	ctx context.Context, c client.Client,
) error {
	return c.DeleteAllOf(ctx, &resourcev1.DeviceTaintRule{})
}

// CreateResourceSliceResource creates a ResourceSlice resource in the cluster.
func CreateResourceSliceResource(
	ctx context.Context, c client.Client, name, driverName, poolName string, devices map[string]map[string]string,
) error {
	if len(devices) == 0 {
		return errors.New("expected devices to be non-empty, but got empty")
	}
	return c.Create(ctx, NewResourceSlice(name, driverName, poolName, devices))
}

// CleanupResourceSlices deletes all ResourceSlices resources in the specified namespace.
func CleanupResourceSliceResources(
	ctx context.Context, c client.Client,
) error {
	return c.DeleteAllOf(ctx, &resourcev1.ResourceSlice{})
}

// CreateMachineResource creates a Machine resource in the cluster.
func CreateMachineResource(
	ctx context.Context, c client.Client, name, namespace, mdName, clusterName string,
) error {
	return c.Create(ctx, NewMachine(name, namespace, mdName, clusterName))
}

// CleanupMachineResource deletes a Machine resource from the cluster.
func CleanupMachineResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capiv1.Machine{})
}

// CleanupMachineResources deletes all Machine resources in the specified namespace.
func CleanupMachineResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capiv1.Machine{}, client.InNamespace(namespace))
}

// CreateMachineDeploymentResource creates a MachineDeployment resource in the cluster.
func CreateMachineDeploymentResource(
	ctx context.Context, c client.Client, name, namespace, clusterName, apiGroup, kind, dmtName string, replicas *int32,
) error {
	return c.Create(ctx, NewMachineDeployment(name, namespace, clusterName, apiGroup, kind, dmtName, replicas))
}

// CleanupMachineDeploymentResource deletes a MachineDeployment resource from the cluster.
func CleanupMachineDeploymentResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capiv1.MachineDeployment{})
}

// CleanupMachineDeploymentResources deletes all MachineDeployment resources in the specified namespace.
func CleanupMachineDeploymentResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capiv1.MachineDeployment{}, client.InNamespace(namespace))
}

// CreateMachineSetResource creates a MachineSet resource in the cluster.
func CreateMachineSetResource(
	ctx context.Context, c client.Client, name, namespace, clusterName, mdName string, replicas *int32, mdUID types.UID,
) error {
	return c.Create(ctx, NewMachineSet(name, namespace, clusterName, mdName, replicas, mdUID))
}

// CleanupMachineSetResource deletes a MachineSet resource from the cluster.
func CleanupMachineSetResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capiv1.MachineSet{})
}

// CleanupMachineSetResources deletes all MachineSet resources in the specified namespace.
func CleanupMachineSetResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capiv1.MachineSet{}, client.InNamespace(namespace))
}

// CreateDCIMachineTemplateResource creates a DCIMachineTemplate resource in the cluster.
func CreateDCIMachineTemplateResource(
	ctx context.Context, c client.Client, name, namespace string, location *string, serverSpec *rpmv1.ServerSpec,
) error {
	if serverSpec == nil {
		return errors.New("expected serverSpec to be non-nil, but got nil")
	}
	template := NewDCIMachineTemplate(name, namespace)
	template.Spec.Template.Spec.Location = location
	template.Spec.Template.Spec.ServerSpec = *serverSpec
	return c.Create(ctx, template)
}

// CleanupDCIMachineTemplateResource deletes a DCIMachineTemplate resource from the cluster.
func CleanupDCIMachineTemplateResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capdv1.DCIMachineTemplate{})
}

// CleanupDCIMachineTemplateResources deletes all DCIMachineTemplate resources in the specified namespace.
func CleanupDCIMachineTemplateResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capdv1.DCIMachineTemplate{}, client.InNamespace(namespace))
}

// CreateDockerMachineTemplateResource creates a DockerMachineTemplate resource in the cluster.
func CreateDockerMachineTemplateResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return c.Create(ctx, NewDockerMachineTemplate(name, namespace))
}

// CleanupDockerMachineTemplateResource deletes a DockerMachineTemplate resource from the cluster.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func CleanupDockerMachineTemplateResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capdockerv1.DockerMachineTemplate{})
}

// CleanupDockerMachineTemplateResources deletes all DockerMachineTemplate resources in the specified namespace.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func CleanupDockerMachineTemplateResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capdockerv1.DockerMachineTemplate{}, client.InNamespace(namespace))
}

// CreateDCIMachineResource creates a DCIMachine resource in the cluster.
func CreateDCIMachineResource(
	ctx context.Context, c client.Client, name, namespace string, location *string, providerID string, serverSpec *rpmv1.ServerSpec,
) error {
	if serverSpec == nil {
		return errors.New("expected serverSpec to be non-nil, but got nil")
	}
	return c.Create(ctx, NewDCIMachine(name, namespace, location, providerID, serverSpec))
}

// CleanupDCIMachineResource deletes a DCIMachine resource from the cluster.
func CleanupDCIMachineResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &capdv1.DCIMachine{})
}

// CleanupDCIMachineResources deletes all DCIMachine resources in the specified namespace.
func CleanupDCIMachineResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &capdv1.DCIMachine{}, client.InNamespace(namespace))
}

// CreateLogicalServerResource creates a LogicalServer resource in the cluster.
func CreateLogicalServerResource(
	ctx context.Context, c client.Client, name, namespace string, location *string, serverSpec *rpmv1.ServerSpec, role rpmv1.LogicalServerRoleType,
) error {
	if serverSpec == nil {
		return errors.New("expected serverSpec to be non-nil, but got nil")
	}
	return c.Create(ctx, NewLogicalServer(name, namespace, location, serverSpec, role))
}

// CleanupLogicalServerResource deletes a LogicalServer resource from the cluster.
func CleanupLogicalServerResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &rpmv1.LogicalServer{})
}

// CleanupLogicalServerResources deletes all LogicalServer resources in the specified namespace.
func CleanupLogicalServerResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &rpmv1.LogicalServer{}, client.InNamespace(namespace))
}

// PatchMachineDeploymentReplicas patches the .Replicas field of the specified MachineDeployment.
// If 'add' is nil, it sets the replicas to the provided 'replicas' value.
// If 'add' is not nil, it adds the provided 'add' value to the current replicas.
func PatchMachineDeploymentReplicas(
	ctx context.Context, c client.Client, name, namespace string, replicas, add *int32,
) error {
	md := &capiv1.MachineDeployment{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), md); err != nil {
		return err
	}
	mdLast := md.DeepCopy()
	replicasVal := int32(0)
	if add == nil {
		// use provided replicas value
		md.Spec.Replicas = replicas
	} else {
		// use replicas in MachineDeployment and add the provided add value
		if md.Spec.Replicas != nil {
			replicasVal = *md.Spec.Replicas
		}
		md.Spec.Replicas = ptr.To(replicasVal + *add)
	}
	return c.Patch(ctx, md, client.MergeFrom(mdLast))
}

// PatchMachineDeploymentReadyReplicas patches the .status.readyReplicas field of the specified MachineDeployment.
func PatchMachineDeploymentReadyReplicas(
	ctx context.Context, c client.Client, name, namespace string, add int32,
) error {
	md := &capiv1.MachineDeployment{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), md); err != nil {
		return err
	}
	mdLast := md.DeepCopy()
	replicas := int32(0)
	if md.Spec.Replicas != nil {
		replicas = *md.Spec.Replicas
	}
	md.Status.ReadyReplicas = ptr.To(replicas + add)
	return c.Status().Patch(ctx, md, client.MergeFrom(mdLast))
}

// PatchMachineDeploymentAddInProgressLabel patches the InProgress label with a different UID in the specified MachineDeployment.
func PatchMachineDeploymentAddInProgressLabel(
	ctx context.Context, c client.Client, name, namespace, uid string,
) error {
	md := &capiv1.MachineDeployment{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), md); err != nil {
		return err
	}
	mdLast := md.DeepCopy()
	md.Labels = utils.AddNodeScalingRequestInProgressUIDLabel(md.Labels, types.UID(uid))
	return c.Patch(ctx, md, client.MergeFrom(mdLast))
}

// PatchMachineDeploymentRemoveInProgressLabel removes the InProgress label from the specified MachineDeployment.
func PatchMachineDeploymentRemoveInProgressLabel(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	md := &capiv1.MachineDeployment{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), md); err != nil {
		return err
	}
	mdLast := md.DeepCopy()
	utils.RemoveNodeScalingRequestInProgressUIDLabel(md.Labels)
	return c.Patch(ctx, md, client.MergeFrom(mdLast))
}

// PatchDCIMachineLogicalServerRef patches the .status.logicalServerRef field of the specified DCIMachine.
func patchDCIMachineLogicalServerRef(
	ctx context.Context, c client.Client, name, namespace string, lsRef *corev1.ObjectReference,
) error {
	dciMachine := &capdv1.DCIMachine{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), dciMachine); err != nil {
		return err
	}
	dciMachineLast := dciMachine.DeepCopy()
	dciMachine.Status.LogicalServerRef = lsRef
	return c.Status().Patch(ctx, dciMachine, client.MergeFrom(dciMachineLast))
}

// PatchDCIMachineLogicalServerRef patches the .status.logicalServerRef field of the specified DCIMachine.
func PatchDCIMachineLogicalServerRef(
	ctx context.Context, c client.Client, name, namespace, lsName, lsNamespace string,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(lsName, lsNamespace), ls); err != nil {
		return err
	}
	lsRef := &corev1.ObjectReference{
		APIVersion: rpmv1.GroupVersion.String(),
		Kind:       "LogicalServer",
		Name:       ls.Name,
		Namespace:  ls.Namespace,
		UID:        ls.UID,
	}
	return patchDCIMachineLogicalServerRef(ctx, c, name, namespace, lsRef)
}

// RemoveDCIMachineLogicalServerRef removes the .status.logicalServerRef field of the specified DCIMachine.
func RemoveDCIMachineLogicalServerRef(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return patchDCIMachineLogicalServerRef(ctx, c, name, namespace, nil)
}

// PatchMachineInfrastructureRef patches the .spec.infrastructureRef field of the specified Machine.
func PatchMachineInfrastructureRef(
	ctx context.Context, c client.Client, name, namespace, dciMachineName, dciMachineNamespace string,
) error {
	machine := &capiv1.Machine{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), machine); err != nil {
		return err
	}
	dciMachine := &capdv1.DCIMachine{}
	if err := c.Get(ctx, NewNamespacedName(dciMachineName, dciMachineNamespace), dciMachine); err != nil {
		return err
	}
	machineLast := machine.DeepCopy()
	machine.Spec.InfrastructureRef = capiv1.ContractVersionedObjectReference{
		APIGroup: capdv1.GroupVersion.Group,
		Kind:     "DCIMachine",
		Name:     dciMachineName,
	}
	return c.Patch(ctx, machine, client.MergeFrom(machineLast))
}

// PatchLogicalServerAppliedType patches the .status.appliedSpec.type field of the specified LogicalServer.
func PatchLogicalServerAppliedType(
	ctx context.Context, c client.Client, name, namespace string, appliedType rpmv1.AppliedServerType,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), ls); err != nil {
		return err
	}
	lsLast := ls.DeepCopy()
	ls.Status.AppliedSpec = &rpmv1.AppliedSpec{
		Type:       appliedType,
		Role:       ls.Spec.Role,
		ServerSpec: rpmv1.AppliedServerSpec{},
	}
	return c.Status().Patch(ctx, ls, client.MergeFrom(lsLast))
}

// PatchLogicalServerGPUSpec patches the .spec.serverSpec.gpu field of the specified LogicalServer.
func PatchLogicalServerGPUSpec(
	ctx context.Context, c client.Client, name, namespace string, gpuSpec *rpmv1.GPUSpec,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), ls); err != nil {
		return err
	}
	lsLast := ls.DeepCopy()
	ls.Spec.ServerSpec.GPU = gpuSpec
	return c.Patch(ctx, ls, client.MergeFrom(lsLast))
}

// PatchLogicalServerState patches the .status.state field of the specified LogicalServer.
func PatchLogicalServerState(
	ctx context.Context, c client.Client, name, namespace string, state rpmv1.LogicalServerState,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), ls); err != nil {
		return err
	}
	lsLast := ls.DeepCopy()
	ls.Status.State = state
	return c.Status().Patch(ctx, ls, client.MergeFrom(lsLast))
}

func PatchLogicalServerAppliedServerSpec(
	ctx context.Context, c client.Client, name, namespace string, serverSpec *rpmv1.ServerSpec,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), ls); err != nil {
		return err
	}
	lsLast := ls.DeepCopy()
	if ls.Status.AppliedSpec == nil {
		ls.Status.AppliedSpec = &rpmv1.AppliedSpec{
			ServerSpec: rpmv1.AppliedServerSpec{
				GPUs: []rpmv1.AppliedGPUSpec{
					{
						Required:    true,
						Model:       serverSpec.GPU.Model,
						Num:         serverSpec.GPU.Num,
						Assignments: serverSpec.GPU.Assignments,
					},
				},
			},
		}
	} else {
		ls.Status.AppliedSpec.ServerSpec.GPUs = []rpmv1.AppliedGPUSpec{
			{
				Required:    true,
				Model:       serverSpec.GPU.Model,
				Num:         serverSpec.GPU.Num,
				Assignments: serverSpec.GPU.Assignments,
			},
		}
	}
	return c.Status().Patch(ctx, ls, client.MergeFrom(lsLast))
}

// PatchLogicalServerType patches the .status.appliedSpec.type field of the specified LogicalServer.
func PatchLogicalServerType(
	ctx context.Context, c client.Client, name, namespace string, appliedType rpmv1.AppliedServerType,
) error {
	ls := &rpmv1.LogicalServer{}
	if err := c.Get(ctx, NewNamespacedName(name, namespace), ls); err != nil {
		return err
	}
	lsLast := ls.DeepCopy()
	if ls.Status.AppliedSpec == nil {
		ls.Status.AppliedSpec = &rpmv1.AppliedSpec{
			Type: appliedType,
			Role: ls.Spec.Role,
		}
	} else {
		ls.Status.AppliedSpec.Type = appliedType
	}
	return c.Status().Patch(ctx, ls, client.MergeFrom(lsLast))
}

func CleanupNodeScalingRequestResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	nsrList := &scalev1.NodeScalingRequestList{}
	if err := c.List(ctx, nsrList, client.InNamespace(namespace)); err != nil {
		return err
	}
	var allErr error
	for idx := range nsrList.Items {
		nsr := &nsrList.Items[idx]
		nsr.Finalizers = nil
		if err := c.Update(ctx, nsr); err != nil {
			allErr = fmt.Errorf("failed to remove finalizers from NodeScalingRequest %s/%s: %w: %w", nsr.Namespace, nsr.Name, err, allErr)
			continue
		}
		if err := c.Delete(ctx, nsr); apierrors.IsNotFound(err) {
			continue
		} else if err != nil {
			allErr = fmt.Errorf("failed to delete NodeScalingRequest %s/%s: %w: %w", nsr.Namespace, nsr.Name, err, allErr)
			continue
		}
	}
	return allErr
}

func CleanupCardScalingRequestResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	csrList := &scalev1.CardScalingRequestList{}
	if err := c.List(ctx, csrList, client.InNamespace(namespace)); err != nil {
		return err
	}
	var allErr error
	for idx := range csrList.Items {
		csr := &csrList.Items[idx]
		csr.Finalizers = nil
		if err := c.Update(ctx, csr); err != nil {
			allErr = fmt.Errorf("failed to remove finalizers from CardScalingRequest %s/%s: %w: %w", csr.Namespace, csr.Name, err, allErr)
			continue
		}
		if err := c.Delete(ctx, csr); apierrors.IsNotFound(err) {
			continue
		} else if err != nil {
			allErr = fmt.Errorf("failed to delete CardScalingRequest %s/%s: %w: %w", csr.Namespace, csr.Name, err, allErr)
			continue
		}
	}
	return allErr
}

func CleanupCardAlertResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	caList := &scalev1.CardAlertList{}
	if err := c.List(ctx, caList, client.InNamespace(namespace)); err != nil {
		return err
	}
	var allErr error
	for idx := range caList.Items {
		ca := &caList.Items[idx]
		ca.Finalizers = nil
		if err := c.Update(ctx, ca); err != nil {
			allErr = fmt.Errorf("failed to remove finalizers from CardAlert %s/%s: %w: %w", ca.Namespace, ca.Name, err, allErr)
			continue
		}
		if err := c.Delete(ctx, ca); apierrors.IsNotFound(err) {
			continue
		} else if err != nil {
			allErr = fmt.Errorf("failed to delete CardAlert %s/%s: %w: %w", ca.Namespace, ca.Name, err, allErr)
			continue
		}
	}
	return allErr
}

func CleanupCardAlertResourcesWithoutFinalizers(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &scalev1.CardAlert{}, client.InNamespace(namespace))
}

// CleanupMachineAlertResource deletes a MachineAlert resource from the cluster.
func CleanupMachineAlertResource(
	ctx context.Context, c client.Client, name, namespace string,
) error {
	return cleanupResource(ctx, c, NewNamespacedName(name, namespace), &rpmv1.MachineAlert{})
}

// CleanupMachineAlertResources deletes all MachineAlert resources in the specified namespace.
func CleanupMachineAlertResources(
	ctx context.Context, c client.Client, namespace string,
) error {
	return c.DeleteAllOf(ctx, &rpmv1.MachineAlert{}, client.InNamespace(namespace))
}
