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
	"errors"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	examplecomv1 "app-conn-info-maker-dci/api/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func dciappconnection_common_ErrorTest08_1(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest08_1 Create ConnectEndpoint Create Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_common_ErrorTest08_test.go")
	fmt.Println(">>>>> Start ErrorTest08_1 Create ConnectEndpoint Create Error")
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

	// Create Deployment
	stDeploymentCR = Deployment01ToC
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	stDeploymentCR = Deployment01ToY
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToY")
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

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToC
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToY
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToY")
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

	// Create Pod
	stPodCR = Pod01ToA1
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
	stPodCR = Pod01ToC1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToC1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToC1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToC1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToC1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToY1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	stDCIAppConnectionCR := DCIAppConnection01
	err = k8sClient.Create(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Create(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.ConnectEndpoint{}, cPod01NameFromX1, cConnectEndpointNamespace),
			gomock.Any()).
		Return(errors.New("ConnectEndpoint FromX1 Create Error Test")).AnyTimes()

	mockClient.EXPECT().
		Get(
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientGet(k8sClient)).AnyTimes()

	mockClient.EXPECT().
		List(
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientList(k8sClient)).AnyTimes()

	mockClient.EXPECT().
		Create(
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientCreate(k8sClient)).AnyTimes()

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientPatch(k8sClient)).AnyTimes()

	mockClient.EXPECT().
		Delete(
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientDelete(k8sClient)).AnyTimes()

	// ------------------------------
	//  ErrorTest08_1 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).To(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: false, RequeueAfter: 0}))

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
	Expect(stDCIAppConnectionCR.Finalizers).To(BeEmpty())
	Expect(stDCIAppConnectionCR.Status.Phase).To(Equal(""))

	controllerutil.RemoveFinalizer(&stDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDCIAppConnectionCR)
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
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound ConnectEndpointCR FromX1.")
	} else if err != nil {
		fmt.Println("Failed to get failed ConnectEndpointCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success ConnectEndpointCR FromX1.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound ConnectEndpointCR ToY1.")
	} else if err != nil {
		fmt.Println("Failed to get failed ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success ConnectEndpointCR ToY1.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound ServiceCR ToA.")
	} else if err != nil {
		fmt.Println("Failed to get failed ServiceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success ServiceCR ToA.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound ServiceCR ToC.")
	} else if err != nil {
		fmt.Println("Failed to get failed ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success ServiceCR ToC.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound ServiceCR ToY.")
	} else if err != nil {
		fmt.Println("Failed to get failed ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success ServiceCR ToY.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound EndpointSliceCR ToA.")
	} else if err != nil {
		fmt.Println("Failed to get failed EndpointSliceCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success EndpointSliceCR ToA.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound EndpointSliceCR ToC.")
	} else if err != nil {
		fmt.Println("Failed to get failed EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success EndpointSliceCR ToC.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if apierrors.IsNotFound(err) {
		fmt.Println("Success to get notfound EndpointSliceCR ToY.")
	} else if err != nil {
		fmt.Println("Failed to get failedEndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Failed to get success EndpointSliceCR ToY.")
	}
	Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

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
	Expect(stDeploymentCR.ObjectMeta.Labels).To(Equal(stDeployment01FromXLabels0))

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
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).To(Equal(stDeployment01ToALabels0))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToA.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToC,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).To(Equal(stDeployment01ToCLabels0))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToC.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToY,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).To(Equal(stDeployment01ToYLabels0))

	controllerutil.RemoveFinalizer(&stDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to update DeploymentCR ToY.")
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
			Name:      cReplicaSet01NameToC,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToY,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToY.")
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
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01FromXLabels0))

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
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToALabels0))

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
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToALabels0))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToC1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToC1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToC1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToCLabels0))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToYLabels0))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_common_ErrorTest08_test.go")
	fmt.Println("<<<<< End ErrorTest08_1")
	fmt.Println("<<<<<")
}
