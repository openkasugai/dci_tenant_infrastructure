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
	"bytes"
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	examplecomv1 "app-conn-info-maker-dci/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func pod_controller_Test16_1_1(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_1_1 Pod Create
	//  X-A (X=Replicas 1to2)
	//  Y-None
	//  In-None,Out-X
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_1_1")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels

	fmt.Println("Change Reolicas 1to2 DeploymentCR FromX.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX

	fmt.Println("Change Reolicas 1to2 ReplicaSetCR FromX.")
	stReplicaSetCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR := Pod01FromX1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stPodCR corev1.Pod
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01FromX2.DeepCopy()
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR = Pod01ToA1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01ToA2.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection16.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint16FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_1_1 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint16FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint16FromX2.Spec))

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(stDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_1_1")
	fmt.Println("<<<<<")
}

func pod_controller_Test16_2_1(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_2_1 Pod Delete
	//  X-A (X=Replicas 2to1)
	//  Y-None
	//  In-None,Out-X
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_2_1")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels

	fmt.Println("Change Reolicas 2to1 DeploymentCR ToA.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01ToA2to1

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR := Pod01FromX1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Delete Test
	stPodCR = Pod01FromX2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing PodCR ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection16.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint16FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX2
	stConnectEndpointCR = ConnectEndpoint16FromX2
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_2_1 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint16FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if errors.IsNotFound(err) {
		fmt.Println("Success to deleted ConnectEndpointCR FromX2.")
		Expect(err).To(HaveOccurred())
	} else if err != nil {
		fmt.Println("Failed to get Failed ConnectEndpointCR FromX2.")
		fmt.Println(err)
		Expect(err).NotTo(HaveOccurred())
	} else {
		fmt.Println("Success to get Success ConnectEndpointCR FromX2.")
		Expect(stConnectEndpointCR.DeletionTimestamp.IsZero()).To(BeTrue())
		Expect(err).To(HaveOccurred())
	}

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR FromX.")
	}
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR ToA.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR FromX1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Success to deleted PodCR FromX2.")
		} else {
			fmt.Println("Failed to get PodCR FromX2.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).To(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA2.")
	}

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)

	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update status DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update status DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_2_1")
	fmt.Println("<<<<<")
}

func pod_controller_Test16_1_2(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_1_2 Pod Create
	//  X-A (X=Replicas 1to2)
	//  Y-None
	//  In-X,Out-None
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_1_2")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels

	fmt.Println("Change Reolicas 1to2 DeploymentCR FromX.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX

	fmt.Println("Change Reolicas 1to2 ReplicaSetCR FromX.")
	stReplicaSetCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR := Pod01FromX1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stPodCR corev1.Pod
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01FromX2.DeepCopy()
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR = Pod01ToA1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01ToA2.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection17.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint17FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_1_2 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint17FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint17FromX2.Spec))

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(stDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_1_2")
	fmt.Println("<<<<<")
}

func pod_controller_Test16_2_2(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_2_2 Pod Delete
	//  X-A (X=Replicas 2to1)
	//  Y-None
	//  In-X,Out-None
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_2_2")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels

	fmt.Println("Change Reolicas 2to1 DeploymentCR ToA.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01ToA2to1

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR := Pod01FromX1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Delete Test
	stPodCR = Pod01FromX2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing PodCR ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection17.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint17FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX2
	stConnectEndpointCR = ConnectEndpoint17FromX2
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_2_2 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint17FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if errors.IsNotFound(err) {
		fmt.Println("Success to deleted ConnectEndpointCR FromX2.")
		Expect(err).To(HaveOccurred())
	} else if err != nil {
		fmt.Println("Failed to get Failed ConnectEndpointCR FromX2.")
		fmt.Println(err)
		Expect(err).NotTo(HaveOccurred())
	} else {
		fmt.Println("Success to get Success ConnectEndpointCR FromX2.")
		Expect(stConnectEndpointCR.DeletionTimestamp.IsZero()).To(BeTrue())
		Expect(err).To(HaveOccurred())
	}

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR FromX.")
	}
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR ToA.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR FromX1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Success to deleted PodCR FromX2.")
		} else {
			fmt.Println("Failed to get PodCR FromX2.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).To(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA2.")
	}

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)

	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update status DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update status DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_2_2")
	fmt.Println("<<<<<")
}
func pod_controller_Test16_1_3(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_1_3 Pod Create
	//  X-A (X=Replicas 1to2)
	//  Y-None
	//  In-None,Out-None
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_1_3")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels

	fmt.Println("Change Reolicas 1to2 DeploymentCR FromX.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX

	fmt.Println("Change Reolicas 1to2 ReplicaSetCR FromX.")
	stReplicaSetCR.Spec.Replicas = &iReplicas01FromX1to2

	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR := Pod01FromX1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stPodCR corev1.Pod
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01FromX2.DeepCopy()
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	pstPodCR = Pod01ToA1.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Add Test
	pstPodCR = Pod01ToA2.DeepCopy()
	controllerutil.AddFinalizer(pstPodCR, cDCIAppConnectionFinalizer)
	pstPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, pstPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection18.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint18FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_1_3 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint18FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint18FromX2.Spec))

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(stDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_1_3")
	fmt.Println("<<<<<")
}

func pod_controller_Test16_2_3(writer bytes.Buffer, reconciler PodReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//	Test16_2_3 Pod Delete
	//  X-A (X=Replicas 2to1)
	//  Y-None
	//  In-None,Out-None
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_Test16_test.go")
	fmt.Println(">>>>> Start Test16_2_3")
	fmt.Println(">>>>>")

	// Create Deployment
	pstDeploymentCR := Deployment01FromX.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01FromXLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToA.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToALabels

	fmt.Println("Change Reolicas 2to1 DeploymentCR ToA.")
	pstDeploymentCR.Spec.Replicas = &iReplicas01ToA2to1

	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR := Pod01FromX1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod Delete Test
	stPodCR = Pod01FromX2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01FromXLabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR FromX2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA2
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToALabels
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing PodCR ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection18.DeepCopy()
	pstDCIAppConnectionCR.Status.Phase = DCIAppConnection01Status.Phase
	controllerutil.AddFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService := Service01ToA
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice := EndpointSlice01ToA
	stEndpointSlice.Labels = EndpointSlice01ToALabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint18FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX2
	stConnectEndpointCR = ConnectEndpoint18FromX2
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//	Test16_2_3 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint18FromX1.Spec))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if errors.IsNotFound(err) {
		fmt.Println("Success to deleted ConnectEndpointCR FromX2.")
		Expect(err).To(HaveOccurred())
	} else if err != nil {
		fmt.Println("Failed to get Failed ConnectEndpointCR FromX2.")
		fmt.Println(err)
		Expect(err).NotTo(HaveOccurred())
	} else {
		fmt.Println("Success to get Success ConnectEndpointCR FromX2.")
		Expect(stConnectEndpointCR.DeletionTimestamp.IsZero()).To(BeTrue())
		Expect(err).To(HaveOccurred())
	}

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToA.Spec.Ports))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	var stDeploymentCR appsv1.Deployment
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR FromX.")
	}
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(stDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update DeploymentCR ToA.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01FromXLabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR FromX1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Success to deleted PodCR FromX2.")
		} else {
			fmt.Println("Failed to get PodCR FromX2.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get PodCR FromX2.")
	}
	Expect(err).To(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA1.")
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stPodCR.Labels).To(Equal(Pod01ToALabels))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update PodCR ToA2.")
	}

	var stDCIAppConnectionCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)

	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update status DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to update status DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< pod_controller_Test16_test.go")
	fmt.Println("<<<<< End Test16_2_3")
	fmt.Println("<<<<<")
}
