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
			/*
				reconciler = PodReconciler{
					Client: k8sClient,
					Scheme: testScheme,
					Log:    log.WithName("controllers").WithName("Pod"),
				}

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
			By("ErrorTest01-1: PodResource GetFailed(Not a NotFoundError)")
			// ------------------------------
			//	ErrorTest01-1 PodResource GetFailed(Not a NotFoundError)
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest01_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test02-1: Failed to getPodRelationshipInfo")
			// ------------------------------
			//	Test02-1 Pod Create
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest02_1(writer, reconciler, mockClient)
		})

		It("Pod Create", func() {
			By("Test03-1: ReconsileCreate Error")
			// ------------------------------
			//	Test03-1 ReconsileCreate Error
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-YXOut-Y
			// ------------------------------
			pod_controller_ErrorTest03_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test04-1: ReconsileDelete Error")
			// ------------------------------
			//	Test04-1 ReconsileDelete Error
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest04_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test05-1: PatchError")
			// ------------------------------
			//	Test05-1 PatchError
			//  X-A,Y-C,X-Y (Y=Replicas 0to1)
			//  Y-None (Y=Replicas 0to1)
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest05_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test06-1: Failed to Get Deployment")
			// ------------------------------
			//	Test06-1 Failed to Get Deployment
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest06_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test07-1: Failed to Get DCIAppConnection")
			// ------------------------------
			//	Test07-1 Failed to Get DCIAppConnection
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest07_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test08-1: Failed to Get ConnectEndpoint_INOUT")
			// ------------------------------
			//	Test08-1 Failed to Get ConnecteEndpoint
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest08_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test09-1: Failed to Get ConnectEndpoint_CONN")
			// ------------------------------
			//	Test09-1 Failed to Get ConnecteEndpoint
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest09_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test10-1: Failed to Patch EndPointSlice_addAddress")
			// ------------------------------
			//	Test10-1 Failed to Get ConnecteEndpoint
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_ErrorTest10_1(writer, reconciler, mockClient)
		})
	})
})
