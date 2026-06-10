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

func dciappconnection_controller_Test18_1(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  Test18-1 Create
	//  X-A
	//  In-None,Out-None
	// ------------------------------

	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_test.go")
	fmt.Println(">>>>> Start Test18-1")
	fmt.Println(">>>>>")

	// Create Deployment
	stDeploymentCR := Deployment01FromX
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	stDeploymentCR = Deployment01ToA
	err = k8sClient.Create(ctx, &stDeploymentCR)
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
	stPodCR := Pod01ToA1
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
	err = k8sClient.Create(ctx, &stPodCR)
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

	// Create Pod
	stPodCR = Pod01FromX1
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

	// Create DCIAppConnectionCR
	stDCIAppConnectionCR := DCIAppConnection18
	err = k8sClient.Create(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection18")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//  Test18-1 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	var stDCIAppConnCR examplecomv1.DCIAppConnection
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDCIAppConnCR.Finalizers).NotTo(BeEmpty())
	Expect(stDCIAppConnCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(stDCIAppConnCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(&stDCIAppConnCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDCIAppConnCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	var stConnectEndpointCR examplecomv1.ConnectEndpoint
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
	Expect(stServiceCR.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
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
	Expect(stEndpointSliceCR.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.OwnerReferences[0].Name).To(Equal(cService01NameToA))
	Expect(stEndpointSliceCR.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

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
	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR FromX..")
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
	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_test.go")
	fmt.Println("<<<<< End Test18-1")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_Test18_2(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  Test18-2 Delete
	//  X-A
	//  In-None,Out-None
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_test.go")
	fmt.Println(">>>>> Start Test18-2")
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
	stPodCR := Pod01ToA1
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

	// Create Pod
	stPodCR = Pod01FromX1
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

	// Create DCIAppConnectionCR
	stDCIAppConnectionCR := DCIAppConnection18
	controllerutil.AddFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Create(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection18")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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

	// Delete DCIAppConnectionCR
	err = k8sClient.Delete(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in deleting DCIAppConnection18")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//  Test18-2 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false}))
	Expect(got).To(Equal(ctrl.Result{RequeueAfter: 0}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Success to deleted DCIAppConnectionCR.")
			Expect(err).To(HaveOccurred())
		} else {
			fmt.Println("Failed to get DCIAppConnectionCR.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
		Expect(err).To(HaveOccurred())
	}

	var stConnectEndpointCR examplecomv1.ConnectEndpoint
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Success to deleted ConnectEndpointCR FromX1.")
			Expect(err).To(HaveOccurred())
		} else {
			fmt.Println("Failed to get ConnectEndpointCR FromX1.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
		Expect(stConnectEndpointCR.DeletionTimestamp.IsZero()).To(BeTrue())
		fmt.Println("Success to DeletionTimestamp IsZero ConnectEndpointCR FromX1.")
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(ctx, &stConnectEndpointCR)
		if err != nil {
			fmt.Println("Failed to delete ConnectEndpointCR FromX1.")
			fmt.Println(err)
		}
		Expect(err).NotTo(HaveOccurred())
	}

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

	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01FromX.Labels))

	err = k8sClient.Delete(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to delete DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete DeploymentCR FromX.")
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
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.Labels).To(Equal(Deployment01ToA.Labels))

	err = k8sClient.Delete(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to delete DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete DeploymentCR ToA.")
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

	err = k8sClient.Delete(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to delete ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())

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

	err = k8sClient.Delete(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to delete ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete ReplicaSetCR ToA.")
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
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete PodCR FromX1.")
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
	Expect(stPodCR.Finalizers).To(BeEmpty())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete PodCR ToA1.")
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
	Expect(stPodCR.Finalizers).To(BeEmpty())

	err = k8sClient.Delete(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to delete PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to delete PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_test.go")
	fmt.Println("<<<<< End Test18-2")
	fmt.Println("<<<<<")
}
