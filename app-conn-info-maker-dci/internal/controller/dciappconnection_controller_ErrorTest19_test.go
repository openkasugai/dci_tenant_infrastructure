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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func dciappconnection_controller_ErrorTest19_1(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest19_1 Update DCIAppConnection Patch Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
	fmt.Println(">>>>> Start ErrorTest19_1 Update Rollback Status Change Patch Error")
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

	// Create Deployment
	pstDeploymentCR = Deployment01ToC.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToCLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToY.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToYLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	stPodCR.Status.Conditions = PodStatusReady
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToCLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToYLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection10.DeepCopy()
	pstDCIAppConnectionCR.Status = DCIAppConnection01Status
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

	// Create Service
	stService = Service01ToC
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService = Service01ToY
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToY")
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

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToC
	stEndpointSlice.Labels = EndpointSlice01ToCLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToY
	stEndpointSlice.Labels = EndpointSlice01ToYLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint01FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointToY
	stConnectEndpointCR = ConnectEndpoint01ToY1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.DCIAppConnection{}, cDCIAppConnection01Name, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("DCIAppConnection Patch Error Test")).Times(1)

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
	//  ErrorTest19_1 Reconcile
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
		pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDCIAppConnectionCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01FromX1.Spec))
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01ToY1.Spec))

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

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToC.Spec.Ports))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToY.Spec.Ports))

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
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToC))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToCLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToC.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToY))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToYLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToY.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToCLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToYLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
	fmt.Println("<<<<< End Test19_1")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_ErrorTest19_2(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest19_2 Update DCIAppConnection Patch Error & Update Roll Back ToY1 Get Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
	fmt.Println(">>>>> Start ErrorTest19_2 Update DCIAppConnection Patch Error & Update Rollback ToY1 Get Error")
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

	// Create Deployment
	pstDeploymentCR = Deployment01ToC.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToCLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToY.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToYLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	stPodCR.Status.Conditions = PodStatusReady
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToCLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToYLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection10.DeepCopy()
	pstDCIAppConnectionCR.Status = DCIAppConnection01Status
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

	// Create Service
	stService = Service01ToC
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService = Service01ToY
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToY")
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

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToC
	stEndpointSlice.Labels = EndpointSlice01ToCLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToY
	stEndpointSlice.Labels = EndpointSlice01ToYLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint01FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointToY
	stConnectEndpointCR = ConnectEndpoint01ToY1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.DCIAppConnection{}, cDCIAppConnection01Name, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("DCIAppConnection Patch Error Test")).Times(1)

	mockClient.EXPECT().
		Get(
			gomock.Any(),
			client.ObjectKey{Name: cPod01NameToY1, Namespace: cDCIAppConnectionNamespace},
			gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
		DoAndReturn(DCIAppConnectionClientGet(k8sClient)).Times(1)

	mockClient.EXPECT().
		Get(
			gomock.Any(),
			client.ObjectKey{Name: cPod01NameToY1, Namespace: cDCIAppConnectionNamespace},
			gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
		Return(errors.New("ConnectEndpoint ToY1 Get Error Test(not NotFound)")).Times(1)

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
	//  ErrorTest19_2 Reconcile
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
		pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDCIAppConnectionCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01FromX1.Spec))
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint10ToY1.Spec))

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

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToC.Spec.Ports))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToY.Spec.Ports))

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
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToC))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToCLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToC.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToY))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToYLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToY.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToCLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToYLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
	fmt.Println("<<<<< End Test19_2")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_ErrorTest19_3(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest19_3 Update DCIAppConnection Patch Error & Update Roll ToY1 Back Patch Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
	fmt.Println(">>>>> Start ErrorTest19_3 Update DCIAppConnection Patch Error & Update Rollback ToY1 Patch Error")
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

	// Create Deployment
	pstDeploymentCR = Deployment01ToC.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToCLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToY.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToYLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	stPodCR.Status.Conditions = PodStatusReady
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToCLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToYLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection10.DeepCopy()
	pstDCIAppConnectionCR.Status = DCIAppConnection01Status
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

	// Create Service
	stService = Service01ToC
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService = Service01ToY
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToY")
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

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToC
	stEndpointSlice.Labels = EndpointSlice01ToCLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToY
	stEndpointSlice.Labels = EndpointSlice01ToYLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint01FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointToY
	stConnectEndpointCR = ConnectEndpoint01ToY1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.DCIAppConnection{}, cDCIAppConnection01Name, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("DCIAppConnection Patch Error Test")).Times(1)

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.ConnectEndpoint{}, cPod01NameToY1, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientPatch(k8sClient)).Times(1)

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.ConnectEndpoint{}, cPod01NameToY1, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("ConnectEndpoint Patch Error Test")).Times(1)

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
	//  ErrorTest19_3 Reconcile
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
		pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDCIAppConnectionCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01FromX1.Spec))
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint10ToY1.Spec))

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

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToC.Spec.Ports))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToY.Spec.Ports))

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
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToC))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToCLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToC.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToY))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToYLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToY.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToCLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToYLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
	fmt.Println("<<<<< End Test19_3")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_ErrorTest19_4(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest19_4 Update DCIAppConnection Patch Error & Update Roll Back FromX1 Get Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
	fmt.Println(">>>>> Start ErrorTest19_4 Update DCIAppConnection Patch Error & Update Rollback FromX1 Get Error")
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

	// Create Deployment
	pstDeploymentCR = Deployment01ToC.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToCLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToY.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToYLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	stPodCR.Status.Conditions = PodStatusReady
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToCLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToYLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection10.DeepCopy()
	pstDCIAppConnectionCR.Status = DCIAppConnection01Status
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

	// Create Service
	stService = Service01ToC
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService = Service01ToY
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToY")
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

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToC
	stEndpointSlice.Labels = EndpointSlice01ToCLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToY
	stEndpointSlice.Labels = EndpointSlice01ToYLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint01FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointToY
	stConnectEndpointCR = ConnectEndpoint01ToY1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.DCIAppConnection{}, cDCIAppConnection01Name, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("DCIAppConnection Patch Error Test")).Times(1)

	mockClient.EXPECT().
		Get(
			gomock.Any(),
			client.ObjectKey{Name: cPod01NameFromX1, Namespace: cDCIAppConnectionNamespace},
			gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
		DoAndReturn(DCIAppConnectionClientGet(k8sClient)).Times(1)

	mockClient.EXPECT().
		Get(
			gomock.Any(),
			client.ObjectKey{Name: cPod01NameFromX1, Namespace: cDCIAppConnectionNamespace},
			gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
		Return(errors.New("ConnectEndpoint ToY1 Get Error Test(not NotFound)")).Times(1)

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
	//  ErrorTest19_4 Reconcile
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
		pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDCIAppConnectionCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint10FromX1.Spec))
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01ToY1.Spec))

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

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToC.Spec.Ports))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToY.Spec.Ports))

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
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToC))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToCLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToC.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToY))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToYLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToY.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToCLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToYLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
	fmt.Println("<<<<< End Test19_4")
	fmt.Println("<<<<<")
}
func dciappconnection_controller_ErrorTest19_5(writer bytes.Buffer, reconciler DCIAppConnectionReconciler, mockClient *MockClient) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  ErrorTest19_5 Update DCIAppConnection Patch Error & Update Roll Back FromX1 Patch Error
	//  X-A,X-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
	fmt.Println(">>>>> Start ErrorTest19_5 Update DCIAppConnection Patch Error & Update Rollback FromX1 Patch Error")
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

	// Create Deployment
	pstDeploymentCR = Deployment01ToC.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToCLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	pstDeploymentCR = Deployment01ToY.DeepCopy()
	controllerutil.AddFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	pstDeploymentCR.Labels = Deployment01ToYLabels
	err = k8sClient.Create(ctx, pstDeploymentCR)
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
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
	stPodCR.Status.Conditions = PodStatusReady
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToCLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	controllerutil.AddFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	stPodCR.Labels = Pod01ToYLabels
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
	stPodCR.Status.Conditions = PodStatusReady
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	pstDCIAppConnectionCR := DCIAppConnection10.DeepCopy()
	pstDCIAppConnectionCR.Status = DCIAppConnection01Status
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

	// Create Service
	stService = Service01ToC
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Service
	stService = Service01ToY
	err = k8sClient.Create(ctx, &stService)
	if err != nil {
		fmt.Println("There is a problem in createing ServiceToY")
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

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToC
	stEndpointSlice.Labels = EndpointSlice01ToCLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create EndpointSlice
	stEndpointSlice = EndpointSlice01ToY
	stEndpointSlice.Labels = EndpointSlice01ToYLabels
	err = k8sClient.Create(ctx, &stEndpointSlice)
	if err != nil {
		fmt.Println("There is a problem in createing EndpointSliceToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointFromX1
	stConnectEndpointCR := ConnectEndpoint01FromX1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ConnectEndpointToY
	stConnectEndpointCR = ConnectEndpoint01ToY1
	err = k8sClient.Create(ctx, &stConnectEndpointCR)
	if err != nil {
		fmt.Println("There is a problem in createing ConnectEndpointFromY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.DCIAppConnection{}, cDCIAppConnection01Name, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("DCIAppConnection Patch Error Test")).Times(1)

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.ConnectEndpoint{}, cPod01NameFromX1, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		DoAndReturn(DCIAppConnectionClientPatch(k8sClient)).Times(1)

	mockClient.EXPECT().
		Patch(
			gomock.Any(),
			TypeMetaEq(&examplecomv1.ConnectEndpoint{}, cPod01NameFromX1, cDCIAppConnectionNamespace),
			gomock.Any(),
			gomock.Any()).
		Return(errors.New("ConnectEndpoint Patch Error Test")).Times(1)

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
	//  ErrorTest19_5 Reconcile
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
		pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDCIAppConnectionCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDCIAppConnectionCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDCIAppConnectionCR.Status).To(Equal(DCIAppConnection01Status))

	controllerutil.RemoveFinalizer(pstDCIAppConnectionCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to update DCIAppConnectionCR.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

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
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint10FromX1.Spec))
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		fmt.Println("Failed to get ConnectEndpointCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stConnectEndpointCR.Spec).To(Equal(ConnectEndpoint01ToY1.Spec))

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

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToC.Spec.Ports))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		fmt.Println("Failed to get ServiceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stServiceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stServiceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cDCIAppConnection01Name))
	Expect(stServiceCR.Spec.Ports).To(Equal(Service01ToY.Spec.Ports))

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
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToALabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToA.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToC))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToCLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToC.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		fmt.Println("Failed to get EndpointSliceCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences).NotTo(BeEmpty())
	Expect(stEndpointSliceCR.ObjectMeta.OwnerReferences[0].Name).To(Equal(cService01NameToY))
	Expect(stEndpointSliceCR.ObjectMeta.Labels).To(Equal(EndpointSlice01ToYLabels))
	Expect(stEndpointSliceCR.Endpoints).To(Equal(EndpointSlice01ToY.Endpoints))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01FromXLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToALabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToCLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
		pstDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(pstDeploymentCR.Finalizers).NotTo(BeEmpty())
	Expect(pstDeploymentCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))
	Expect(pstDeploymentCR.ObjectMeta.Labels).To(Equal(Deployment01ToYLabels))

	controllerutil.RemoveFinalizer(pstDeploymentCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, pstDeploymentCR)
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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

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
	Expect(stPodCR.Finalizers).NotTo(BeEmpty())
	Expect(stPodCR.Finalizers[0]).To(Equal(cDCIAppConnectionFinalizer))

	controllerutil.RemoveFinalizer(&stPodCR, cDCIAppConnectionFinalizer)
	err = k8sClient.Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
	fmt.Println("<<<<< End Test19_5")
	fmt.Println("<<<<<")
}
