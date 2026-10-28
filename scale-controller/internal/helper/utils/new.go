package utils

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
)

// NewDCIMachineTemplateFromCardScalingRequest is a function to create a DCIMachineTemplate resource
// from CardScalingRequest and DCIMachineSpec.
func NewDCIMachineTemplateFromCardScalingRequest(
	csr *scalev1.CardScalingRequest, spec *capdv1.DCIMachineSpec,
) (*capdv1.DCIMachineTemplate, error) {
	s := spec.DeepCopy()
	if csr == nil || s == nil {
		// this case is only for failsafe
		return nil, errors.New("expected CardScalingRequest and DCIMachineSpec are not nil, but got nil")
	}

	return &capdv1.DCIMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csr.Status.NewDCIMachineInfo.DCIMachineTemplateName,
			Namespace: csr.Status.ClusterNamespace,
			Labels: map[string]string{
				scalev1.CreatedByNamespaceLabel: csr.Namespace,
				scalev1.CreatedByNameLabel:      csr.Name,
			},
		},
		Spec: capdv1.DCIMachineTemplateSpec{
			Template: capdv1.DCIMachineTemplateResource{
				Spec: *s,
			},
		},
	}, nil
}

// NewDCIMachineTemplateFromNodeScalingRequest is a function to create a DCIMachineTemplate resource
// from NodeScalingRequest.
func NewDCIMachineTemplateFromNodeScalingRequest(
	name, namespace string, nsr *scalev1.NodeScalingRequest,
) (*capdv1.DCIMachineTemplate, error) {
	if nsr == nil || nsr.Spec.ServerSpec == nil {
		// this case is only for failsafe
		return nil, errors.New("expected NodeScalingRequest and ServerSpec is not nil, but got nil")
	}
	return &capdv1.DCIMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				scalev1.CreatedByNamespaceLabel: nsr.Namespace,
				scalev1.CreatedByNameLabel:      nsr.Name,
			},
		},
		Spec: capdv1.DCIMachineTemplateSpec{
			Template: capdv1.DCIMachineTemplateResource{
				Spec: capdv1.DCIMachineSpec{
					Location:   nsr.Spec.Location,
					ServerSpec: *nsr.Spec.ServerSpec,
					ProviderID: "",
				},
			},
		},
	}, nil
}

// NewDockerMachineTemplateFromNodeScalingRequest is a function to create a DockerMachineTemplate resource
// from NodeScalingRequest.
//
//nolint:staticcheck // capdockerv1.DockerMachineTemplate is deprecated but still used because this is only for test purpose
func NewDockerMachineTemplateFromNodeScalingRequest(
	_, namespace string, nsr *scalev1.NodeScalingRequest,
) (*capdockerv1.DockerMachineTemplate, error) {
	if nsr == nil || nsr.Spec.Location == nil {
		// this case is only for failsafe
		return nil, errors.New("expected NodeScalingRequest and Location is not nil, but got nil")
	}
	return &capdockerv1.DockerMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *nsr.Spec.Location,
			Namespace: namespace,
			Labels: map[string]string{
				scalev1.CreatedByNamespaceLabel: nsr.Namespace,
				scalev1.CreatedByNameLabel:      nsr.Name,
			},
		},
		Spec: capdockerv1.DockerMachineTemplateSpec{
			Template: capdockerv1.DockerMachineTemplateResource{
				Spec: capdockerv1.DockerMachineSpec{
					PreLoadImages: []string{},
				},
			},
		},
	}, nil
}

// newMachineDeployment is a function to create a MachineDeployment resource.
func newMachineDeployment(
	name, namespace, machineTemplateName, createdByName, createdByNamespace string, replicas *int32, original *capiv1.MachineDeployment,
) (*capiv1.MachineDeployment, error) {
	if original == nil ||
		!original.Spec.Template.Spec.Bootstrap.ConfigRef.IsDefined() ||
		original.Spec.Selector.MatchLabels == nil {
		// this case is only for failsafe
		return nil, fmt.Errorf("expected MachineDeployment and its parameters are not nil, but got nil: %v", original)
	}

	// Deep copy original MachineDeployment template and update only InfrastructureRef name
	newTemplate := original.Spec.Template.DeepCopy()
	if newTemplate == nil {
		// this case is only for failsafe
		return nil, errors.New("failed to deep copy MachineDeployment template")
	}
	newTemplate.Spec.InfrastructureRef.Name = machineTemplateName

	// Return new MachineDeployment with updated parameters
	return &capiv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: original.APIVersion,
			Kind:       original.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				scalev1.CreatedByNamespaceLabel: createdByNamespace,
				scalev1.CreatedByNameLabel:      createdByName,
			},
		},
		Spec: capiv1.MachineDeploymentSpec{
			ClusterName: original.Spec.ClusterName,
			Replicas:    replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					capiv1.ClusterNameLabel: original.Spec.Selector.MatchLabels[capiv1.ClusterNameLabel],
				},
			},
			Template: *newTemplate,
		},
	}, nil
}

// NewMachineDeploymentFromCardScalingRequest is a function to create a MachineDeployment resource
// from CardScalingRequest.
func NewMachineDeploymentFromCardScalingRequest(
	csr *scalev1.CardScalingRequest, original *capiv1.MachineDeployment,
) (*capiv1.MachineDeployment, error) {
	if csr == nil || csr.Status.NewDCIMachineInfo == nil {
		// this case is only for failsafe
		return nil, fmt.Errorf("expected CardScalingRequest and its NewDCIMachineInfo are not nil, but got nil: %v", csr)
	}
	return newMachineDeployment(
		csr.Status.NewDCIMachineInfo.MachineDeploymentName,
		csr.Status.ClusterNamespace,
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateName,
		csr.Name,
		csr.Namespace,
		ptr.To(int32(0)), // set 0 initially
		original,
	)
}

// NewMachineDeploymentFromNodeScalingRequest is a function to create a MachineDeployment resource
// from NodeScalingRequest.
func NewMachineDeploymentFromNodeScalingRequest(
	name, namespace, machineTemplateName string, nsr *scalev1.NodeScalingRequest, original *capiv1.MachineDeployment,
) (*capiv1.MachineDeployment, error) {
	if nsr == nil {
		// this case is only for failsafe
		return nil, errors.New("expected NodeScalingRequest is not nil, but got nil")
	}
	return newMachineDeployment(name, namespace, machineTemplateName, nsr.Name, nsr.Namespace, ptr.To(int32(1)), original)
}

// GenerateServerSpecAfterScaling generates ServerSpec after scaling based on original ServerSpec and CardScalingRequest.
// This function ignores GPU assignments and UUIDs in arguments.
func GenerateServerSpecAfterScaling(
	original *rpmv1.ServerSpec, param *scalev1.ScalingParameter,
) (*rpmv1.ServerSpec, error) {
	// Check arguments
	if original == nil || param == nil {
		return nil, errors.New("expected ServerSpec and ScalingParameter are not nil, but got nil")
	}

	// Validate scaling parameter, allow nil GPU in original spec
	if original.GPU != nil && original.GPU.Model != param.Model {
		return nil, fmt.Errorf("expected scaling parameter model(%s) is the same as ServerSpec GPU(%v), but different",
			param.Model, original.GPU)
	}

	// Calculate the number of GPUs after scaling
	var diff, gpuNum int32
	if param.Num != 0 {
		diff = param.Num
	} else if len(param.UUIDs) != 0 {
		diff = int32(len(param.UUIDs)) //nolint:gosec // ignore G115: conversion from len() to int32 is safe here
	} else {
		// this case is only for failsafe
		return nil, errors.New("expected either Num or UUIDs is specified in scaling parameter, but both are not specified")
	}
	if original.GPU == nil {
		gpuNum = 0
	} else {
		gpuNum = original.GPU.Num
	}
	switch param.Op {
	case capdv1.DCIMachineAnnotationCardScalingInfoOpAdd:
		gpuNum += diff
	case capdv1.DCIMachineAnnotationCardScalingInfoOpDelete:
		gpuNum -= diff
	case capdv1.DCIMachineAnnotationCardScalingInfoOpReplace:
		if diff > gpuNum {
			return nil, fmt.Errorf("expected the number of GPUs to replace(%d) is less than or equal to the current number of GPUs(%d), but greater",
				diff, gpuNum)
		}
	default:
		// This case is only for failsafe
		return nil, fmt.Errorf("unsupported scaling parameter operation(%s)", param.Op)
	}

	// Check if the number of GPUs is valid
	var gpu *rpmv1.GPUSpec = nil
	if gpuNum < 0 {
		return nil, fmt.Errorf("expected the number of GPUs after scaling is greater than or equal to 0, but got %d", gpuNum)
	} else if gpuNum > 0 {
		gpu = &rpmv1.GPUSpec{
			Model: param.Model, // get from scaling parameter, not original spec because original spec may be nil
			Num:   gpuNum,
		}
	} // else gpuNum == 0 -> gpu remains nil

	// Generate DCIMachineSpec after scaling
	specAfterScaling := original.DeepCopy()
	specAfterScaling.GPU = gpu

	return specAfterScaling, nil
}

// GenerateDCIMachineSpecAfterScaling generates DCIMachineSpec after scaling based on original DCIMachineSpec and CardScalingRequest.
func GenerateDCIMachineSpecAfterScaling(
	original *capdv1.DCIMachineSpec, csr *scalev1.CardScalingRequest,
) (*capdv1.DCIMachineSpec, error) {
	// Check arguments
	if original == nil {
		return nil, errors.New("expected DCIMachineSpec is not nil, but got nil")
	}

	// Generate ServerSpec after scaling
	serverSpecAfterScaling, err := GenerateServerSpecAfterScaling(&original.ServerSpec, &csr.Spec.ScalingParameter)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ServerSpec after scaling: %w", err)
	}

	// Generate DCIMachineSpec after scaling
	specAfterScaling := original.DeepCopy()
	specAfterScaling.ServerSpec = *serverSpecAfterScaling

	return specAfterScaling, nil
}
