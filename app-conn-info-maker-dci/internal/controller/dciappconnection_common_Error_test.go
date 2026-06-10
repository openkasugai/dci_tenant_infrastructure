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
			fmt.Println("> dciappconnection_common_Error_test.go")
			fmt.Println("> Start BeforeEach")
			fmt.Println(">")

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			// err = reconciler.SetupWithManager(mgr)
			// if err != nil {
			//	fmt.Println("An error occur during setupwithManager: ", err)
			// }
			// Expect(err).NotTo(HaveOccurred())

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
			fmt.Println("< dciappconnection_common_Error_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Error_test.go")
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
			fmt.Println("<<<<< dciappconnection_common_Error_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Common Error Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Error_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_common_Error_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_common_Error_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_common_Error_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Create getDeploymentList().Get() Deployment error", func() {
			By("ErrorTest01_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest01_1 Create X Deployment Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest01_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() Deployment error", func() {
			By("ErrorTest01_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest01_2 Delete X Deployment Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest01_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() ConnectEndpoint error", func() {
			By("ErrorTest02_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest02_1 Create X1 ConnectEndpoint Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest02_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() ConnectEndpoint error", func() {
			By("ErrorTest02_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest02_2 Delete X1 ConnectEndpoint error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest02_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() Deployment error", func() {
			By("ErrorTest03_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest03_1 Create ToA Deployment Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest03_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() Deployment error", func() {
			By("ErrorTest03_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest03_2 Delete ToA Deployment Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest03_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() Service error", func() {
			By("ErrorTest04_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest04_1 Create A Service Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest04_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() Service error", func() {
			By("ErrorTest04_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest04_2 Delete A Service Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest04_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() EndpointSlice error", func() {
			By("ErrorTest05_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest05_1 Create A EndpointSlice Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest05_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() EndpointSlice error", func() {
			By("ErrorTest05_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest05_2 Delete A EndpointSlice Get error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest05_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() ReplicaSet List error", func() {
			By("ErrorTest06_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest06_1 Create A ReplicaSet List error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest06_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() ReplicaSet List error", func() {
			By("ErrorTest06_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest06_2 Delete A ReplicaSet List error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest06_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create getDeploymentList().Get() Pod List error", func() {
			By("ErrorTest07_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest07_1 Create A Pod List error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest07_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete getDeploymentList().Get() Pod List error", func() {
			By("ErrorTest07_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest07_2 Delete A Pod List error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest07_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create FromX1 makeConnectEndpoint().Create() error", func() {
			By("ErrorTest08_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest08_1 Create FromX1 makeConnectEndpoint().Create() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest08_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create FromX1 makeConnectEndpoint().Patch() error", func() {
			By("ErrorTest09_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest09_1 Create FromX1 makeConnectEndpoint().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest09_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ToA2 setDciPodConnectionRelationship().Patch() error", func() {
			By("ErrorTest10_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest10_1 Create ToA2 setDciPodConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest10_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ToA2 unsetDciPodConnectionRelationship().Patch() error", func() {
			By("ErrorTest10_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest10_2 Delete ToA2 unsetDciPodConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest10_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ToA setDciDeploymentConnectionRelationship().Patch() error", func() {
			By("ErrorTest11_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest11_1 Create ToA setDciDeploymentConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest11_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ToA unsetDciDeploymentConnectionRelationship().Patch() error", func() {
			By("ErrorTest11_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest11_2 Delete ToA unsetDciDeploymentConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest11_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod ToA2 Annotation network-status json.Unmarshal() error", func() {
			By("ErrorTest12_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest12_1 Create Pod ToA2 Annotation network-status json.Unmarshal() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest12_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod ToA2 Annotation network-status json.Unmarshal() error", func() {
			By("ErrorTest12_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest12_2 Delete Pod ToA2 Annotation network-status json.Unmarshal() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest12_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ConnectEndpoint FromX1 ctrl.SetControllerReference() error", func() {
			By("ErrorTest13_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest13_1 Create ConnectEndpoint FromX1 ctrl.SetControllerReference() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest13_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create FromX1 setDciPodConnectionRelationship().Patch() error", func() {
			By("ErrorTest14_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest14_1 Create FromX1 setDciPodConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest14_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete FromX1 unsetDciPodConnectionRelationship().Patch() error", func() {
			By("ErrorTest14_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest14_2 Delete FromX1 unsetDciPodConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest14_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create FromX setDciDeploymentConnectionRelationship().Patch() error", func() {
			By("ErrorTest15_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest15_1 Create FromX setDciDeploymentConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest15_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete FromX unsetDciDeploymentConnectionRelationship().Patch() error", func() {
			By("ErrorTest15_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest15_2 Delete FromX unsetDciDeploymentConnectionRelationship().Patch() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest15_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create HostDeviceNetwork FromX1 Get() error", func() {
			By("ErrorTest16_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest16_1 Create HostDeviceNetwork FromX1 Get() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest16_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete HostDeviceNetwork FromX1 Get() error", func() {
			By("ErrorTest16_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest16_2 Delete HostDeviceNetwork FromX1 Get()) error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest16_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create HostDeviceNetwork ToA1 Get() error", func() {
			By("ErrorTest17_1: Reconciling the created resource")
			// ------------------------------
			//  ErrorTest17_1 Create HostDeviceNetwork ToA1 Get() error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest17_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete HostDeviceNetwork ToA1 Get() error", func() {
			By("ErrorTest17_2: Reconciling the delete resource")
			// ------------------------------
			//  ErrorTest17_2 Delete HostDeviceNetwork ToA1 Get()) error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_ErrorTest17_2(writer, reconciler, mockClient)
		})
	})
})
