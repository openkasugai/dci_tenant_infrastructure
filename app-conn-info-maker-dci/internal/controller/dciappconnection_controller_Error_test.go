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
	discoveryv1 "k8s.io/api/discovery/v1"

	hostdevicenetwork "github.com/Mellanox/network-operator/api/v1alpha1"
	"k8s.io/klog/v2/textlogger"
)

var _ = Describe("DCIAppConnection Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("Test for DCIAppConnectionReconciler", Ordered, func() {

		var reconciler DCIAppConnectionReconciler

		var mockClient *MockClient

		// set up log
		log := textlogger.NewLogger(textlogger.NewConfig())

		BeforeAll(func() {
			fmt.Println(">")
			fmt.Println("> dciappconnection_controller_Error_test.go")
			fmt.Println("> Start BeforeEach")
			fmt.Println(">")
			// set up log
			log := textlogger.NewLogger(textlogger.NewConfig())

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			reconciler = DCIAppConnectionReconciler{
				Client: k8sClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("DCIAppConnection"),
			}

			//err = reconciler.SetupWithManager(mgr)
			//if err != nil {
			//	fmt.Println("An error occur during setupwithManager: ", err)
			//}
			//Expect(err).NotTo(HaveOccurred())

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
			fmt.Println("< dciappconnection_controller_Error_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
			fmt.Println(">>>>> Start BeforeEach")
			fmt.Println(">>>>>")

			gomockCtrl := gomock.NewController(GinkgoT())
			defer gomockCtrl.Finish()

			mockClient = NewMockClient(gomockCtrl)
			reconciler = DCIAppConnectionReconciler{
				Client: mockClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("DCIAppConnection"),
			}
			mockClient.EXPECT().
				Scheme().
				Return(testScheme).
				AnyTimes()

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
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_Error_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_controller_Error_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_controller_Error_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_controller_Error_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Create DCIAppConnection Get Error", func() {
			By("ErrorTest01_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest01_1 Create DCIAppConnection Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest01_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create DCIAppConnection Patch Error", func() {
			By("ErrorTest02_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest02_1 Create DCIAppConnection Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest02_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete DCIAppConnection Patch Error", func() {
			By("ErrorTest02_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest02_2 Delete DCIAppConnection Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest02_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create DCIAppConnection InpueEndpoint Get Error", func() {
			By("ErrorTest03_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest03_1 Create DCIAppConnection InpueEndpoint Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest03_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete DCIAppConnection InpueEndpoint Get Error", func() {
			By("ErrorTest03_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest03_2 Delete DCIAppConnection InpueEndpoint Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest03_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create DCIAppConnection OutpueEndpoint Get Error", func() {
			By("ErrorTest04_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest04_1 Create DCIAppConnection OutpueEndpoint Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest04_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete DCIAppConnection OutpueEndpoint Get Error", func() {
			By("ErrorTest04_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest04_2 Delete DCIAppConnection OutpueEndpoint Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest04_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Service ToA Create Error", func() {
			By("ErrorTest05_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest05_1 Create DCIAppConnection Service ToA Create Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest05_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Service ToA Patch Error", func() {
			By("ErrorTest05_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest05_2 Create DCIAppConnection Service ToA Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest05_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create EndpointSlice ToA Create Error", func() {
			By("ErrorTest06_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest06_1 Create DCIAppConnection EndpointSlice ToA Create Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest06_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create EndpointSlice ToA Patch Error", func() {
			By("ErrorTest06_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest06_2 Create DCIAppConnection EndpointSlice ToA Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest06_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Service ToA SetControllerReference Error", func() {
			By("ErrorTest07_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest07_1 Create DCIAppConnection Service ToA SetControllerReference Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest07_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create EndpointSlice ToA SetControllerReference Error", func() {
			By("ErrorTest08_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest08_1 Create DCIAppConnection EndpointSlice ToA SetControllerReference Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest08_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create DCIAppConnection Create Error", func() {
			By("ErrorTest09_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest09_1 Create DCIAppConnection Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest09_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete DCIAppConnection Patch Error", func() {
			By("ErrorTest09_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest09_2 Delete DCIAppConnection Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest09_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback ConnectEndpoint Get Error", func() {
			By("ErrorTest11_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest11_1 Create Rollback ConnectEndpoint Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest11_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback ConnectEndpoint Delete Error", func() {
			By("ErrorTest11_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest11_2 Create Rollback ConnectEndpoint Delete Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest11_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Pod Y1 Get Error", func() {
			By("ErrorTest12_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest12_1 Create Rollback Pod Y1 Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest12_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Pod Y1 unsetDciPodConnectionRelationship Error", func() {
			By("ErrorTest12_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest12_2 Create Rollback Pod Y1 unsetDciPodConnectionRelationship Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest12_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Deployment Y Get Error", func() {
			By("ErrorTest13_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest13_1 Create Rollback Deployment Y Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest13_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Deployment Y unsetDciPodConnectionRelationship Error", func() {
			By("ErrorTest13_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest13_2 Create Rollback Deployment Y unsetDciPodConnectionRelationship Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest13_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback EndpointSlice Y Get Error", func() {
			By("ErrorTest14_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest14_1 Create Rollback EndpointSlice Y Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest14_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback EndpointSlice Y Patch Error", func() {
			By("ErrorTest14_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest14_2 Create Rollback EndpointSlice Y Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest14_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Service Y Get Error", func() {
			By("ErrorTest15_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest15_1 Create Rollback Service Y Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest15_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Rollback Service Y Patch Error", func() {
			By("ErrorTest15_2: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest15_2 Create Rollback Service Y Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest15_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment X Not Found & Phase Patch Error", func() {
			By("ErrorTest16_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest16_1 Create Deployment X Not Found & Phase Patch Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_ErrorTest16_1(writer, reconciler, mockClient)
		})
	})
})
