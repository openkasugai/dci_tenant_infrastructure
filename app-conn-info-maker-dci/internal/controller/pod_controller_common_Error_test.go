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

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	examplecomv1 "app-conn-info-maker-dci/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	hostdevicenetwork "github.com/Mellanox/network-operator/api/v1alpha1"
	"k8s.io/klog/v2/textlogger"
)

var _ = Describe("Pod Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("Test for PodReconciler", Ordered, func() {

		var reconciler PodReconciler
		var mockClient *MockClient
		// set up log
		log := textlogger.NewLogger(textlogger.NewConfig())

		BeforeAll(func() {
			fmt.Println(">")
			fmt.Println("> pod_controller_test.go")
			fmt.Println("> Start BeforeAll")
			fmt.Println(">")

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			reconciler = PodReconciler{
				Client: k8sClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("Pod"),
			}
			/*
				err = reconciler.SetupWithManager(mgr)
				if err != nil {
					fmt.Println("An error occur during setupwithManager: ", err)
				}
				Expect(err).NotTo(HaveOccurred())
			*/
			// regist namespace
			namespaces := &corev1.NamespaceList{}
			Expect(k8sClient.List(context.Background(), namespaces)).Should(Succeed())

			for _, name := range cNamespaces {
				found := false
				for _, ns := range namespaces.Items {
					if ns.Name == name {
						found = true
						break
					}
				}
				if !found {
					ns := &corev1.Namespace{}
					ns.Name = name
					Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
				}
			}
			fmt.Println("<")
			fmt.Println("< pod_controller_test.go")
			fmt.Println("< End BeforeAll")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()

			gomockCtrl := gomock.NewController(GinkgoT())
			defer gomockCtrl.Finish()

			mockClient = NewMockClient(gomockCtrl)
			reconciler = PodReconciler{
				Client: mockClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("Pod"),
			}
			mockClient.EXPECT().
				Scheme().
				Return(testScheme).
				AnyTimes()

			fmt.Println(">>")
			fmt.Println(">> pod_controller_test.go")
			fmt.Println(">> Start BeforeEach")
			fmt.Println(">>")
			err = k8sClient.DeleteAllOf(ctx, &examplecomv1.DCIAppConnection{}, client.InNamespace(cDCIAppConnectionNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &examplecomv1.ConnectEndpoint{}, client.InNamespace(cConnectEndpointNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &corev1.Service{}, client.InNamespace(cServiceNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &discoveryv1.EndpointSlice{}, client.InNamespace(cEndpointSliceNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace(cDeploymentNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &appsv1.ReplicaSet{}, client.InNamespace(cReplicaSetNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &corev1.Pod{}, client.InNamespace(cPodNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &hostdevicenetwork.HostDeviceNetwork{}, client.InNamespace(cHostDeviceNetworkNamespace))
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("<<")
			fmt.Println("<< pod_controller_test.go")
			fmt.Println("<< End BeforeEach")
			fmt.Println("<<")
		})

		AfterEach(func() {
			By("Pod Test End")
			fmt.Println(">>>")
			fmt.Println(">>> pod_controller_test.go")
			fmt.Println(">>> Start AfterEach")
			fmt.Println(">>>")
			fmt.Println("<<<")
			fmt.Println("<<< pod_controller_test.go")
			fmt.Println("<<< End AfterEach")
			fmt.Println("<<<")
		})

		AfterAll(func() {
			fmt.Println(">>>>")
			fmt.Println(">>>> pod_controller_test.go")
			fmt.Println(">>>> Start AfterAll")
			fmt.Println(">>>>")
			fmt.Println("<<<<")
			fmt.Println("<<<< pod_controller_test.go")
			fmt.Println("<<<< End AfterAll")
			fmt.Println("<<<<")
		})

		It("Pod Create", func() {
			By("CommonErrorTest01-1: Failed makeConnectEndpoint To")
			// ------------------------------
			//	CommonErrorTest01-1: Failed makeConnectEndpoint To"
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest01_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest01-2: Failed makeConnectEndpoint To Delete")
			// ------------------------------
			//	CommonErrorTest01-2: Failed makeConnectEndpoint To Delete"
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest01_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonTest02-1: Failed makeConnectEndpoint From")
			// ------------------------------
			//	Test02-1 Failed makeConnectEndpoint From
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest02_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonTest02-2: Failed makeConnectEndpoint From Delete")
			// ------------------------------
			//	Test02-2 Failed makeConnectEndpoint From Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest02_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test03-1: ReconsileCreateError DeleteConnectEndpoint")
			// ------------------------------
			//	Test03-1 ReconsileCreateError DeleteConnectEndpoint.
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest03_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test03-2: ReconsileCreateError DeleteConnectEndpoint Delete")
			// ------------------------------
			//	Test03-2 ReconsileCreateError DeleteConnectEndpoint Delete
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest03_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test04-1: ReconsileDelete makeConnectEndpoint Error")
			// ------------------------------
			//	Test04-1 ReconsileDelete makeConnectEndpoint Error
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest04_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test04-2: ReconsileCreate makeConnectEndpoint Error")
			// ------------------------------
			//	Test04-2 ReconsileCreate makeConnectEndpoint Error
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest04_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest05-1: ReconcileDelete deleteConnectEndpoint Error")
			// ------------------------------
			//	CommonErrorTest05-1 ReconcileDelete deleteConnectEndpoint Error
			//  X-A
			//  Y-C.Y-A (Y=Replicas 2to1)
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest05_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest05-2: ReconcileCreate deleteConnectEndpoint Error")
			// ------------------------------
			//	CommonErrorTest05-2 ReconcileDelete deleteConnectEndpoint Error
			//  X-A
			//  Y-C.Y-A (Y=Replicas 1to2)
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest05_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest06-1: failed to getFromConnectionInformation")
			// ------------------------------
			//	CommonErrorTest06-1 failed to getFromConnectionInformation
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest06_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest06-2: failed to getFromConnectionInformation Delete")
			// ------------------------------
			//	CommonErrorTest06-2 failed to getFromConnectionInformation Delete
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest06_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest07-1: failed to getToConnectionInformation")
			// ------------------------------
			//	CommonErrorTest07-1 failed to getToConnectionInformation
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest07_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest07-2: failed to getToConnectionInformation Delete")
			// ------------------------------
			//	CommonErrorTest07-2 failed to getToConnectionInformation Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest07_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest08-1: failed to getToConnectionInformationPod")
			// ------------------------------
			//	CommonErrorTest08-1 failed to getToConnectionInformationPod
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest08_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest08-2: failed to getToConnectionInformationPod Delete")
			// ------------------------------
			//	CommonErrorTest08-2 failed to getToConnectionInformationPod Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest08_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest09-1: failed to get Connectionendpoint for getFromConnectionInformationPod")
			// ------------------------------
			//	CommonErrorTest09-1 failed to get Connectionendpoint for getFromConnectionInformationPod
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest09_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest09-2: failed to get Connectionendpoint for getFromConnectionInformationPod Delete")
			// ------------------------------
			//	CommonErrorTest09-2 failed to get Connectionendpoint for getFromConnectionInformationPod Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest09_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest10-1: failed to getToConnectionInformationService")
			// ------------------------------
			//	CommonErrorTest10-1 failed to getToConnectionInformationService
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest10_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest10-2: failed to getToConnectionInformationService Delete")
			// ------------------------------
			//	CommonErrorTest10-2 failed to getToConnectionInformationService Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest10_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest11-1: failed to get EndpointSlice for getToConnectionInformationService")
			// ------------------------------
			//	CommonErrorTest11-1 failed to get EndpointSlice for getToConnectionInformationService
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest11_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest11-2: failed to get EndpointSlice for getToConnectionInformationService Delete")
			// ------------------------------
			//	CommonErrorTest11-2 failed to get EndpointSlice for getToConnectionInformationService Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest11_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest12-1: failed to ReplicasetList Function for getDeploymentList")
			// ------------------------------
			//	CommonErrorTest12-1 failed to ReplicasetList Function for getDeploymentList
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest12_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest12-2: failed to ReplicasetList Function for getDeploymentList Delete")
			// ------------------------------
			//	CommonErrorTest12-2 failed to ReplicasetList Function for getDeploymentList Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest12_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest13-1: failed to PodList Function for getDeploymentList")
			// ------------------------------
			//	CommonErrorTest13-1 failed to PodList Function for getDeploymentList
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest13_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest13-2: failed to PodList Function for getDeploymentList Delete")
			// ------------------------------
			//	CommonErrorTest13-2 failed to PodList Function for getDeploymentList Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest13_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest14-1: failed to setControllerReference")
			// ------------------------------
			//	CommonErrorTest14-1 failed to setControllerReference
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest14_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest14-2: failed to setControllerReference Delete")
			// ------------------------------
			//	CommonErrorTest14-2 failed to setControllerReference Delete
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest14_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest15-1: failed to Patch for unsetDciPodConnectionRelationship")
			// ------------------------------
			//	CommonErrorTest15-1 failed to Patch for unsetDciPodConnectionRelationship
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest15_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest15-2: failed to Patch for unsetDciPodConnectionRelationship Delete")
			// ------------------------------
			//	CommonErrorTest15-2 failed to Patch for unsetDciPodConnectionRelationship Delete
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest15_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest16-1: failed to Get for get2ndNicAddress")
			// ------------------------------
			//	CommonErrorTest16-1 failed to Get for get2ndNicAddress
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest16_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest16-2: failed to Get for get2ndNicAddress Delete")
			// ------------------------------
			//	CommonErrorTest16-2 failed to Get for get2ndNicAddress Delete
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest16_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("CommonErrorTest17-1: failed Unmarshal")
			// ------------------------------
			//	CommonErrorTest17-1 failed Unmarshal
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest17_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("CommonErrorTest17-2: failed Unmarshal Delete")
			// ------------------------------
			//	CommonErrorTest17-2 failed Unmarshal Delete
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_ErrorTest17_2(writer, reconciler, mockClient)
		})
	})
})
