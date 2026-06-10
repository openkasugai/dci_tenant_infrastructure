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
			fmt.Println("> dciappconnection_common_Subnormal_test.go")
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
			fmt.Println("< dciappconnection_common_Subnormal_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal_test.go")
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
			fmt.Println("<<<<< dciappconnection_common_Subnormal_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Common Subnormal Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_common_Subnormal_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_common_Subnormal_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_common_Subnormal_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Delete Deployment Labels None", func() {
			By("SubnormalTest01_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest01_2 Delete Deployment X Labels None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest01_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment Labels Char None", func() {
			By("SubnormalTest02_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest02_1 Create Deployment X Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest02_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment Labels Char None", func() {
			By("SubnormalTest02_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest02_2 Delete Deployment X Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest02_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment Labels Name Unmatch", func() {
			By("SubnormalTest03_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest03_1 Create Deployment X Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest03_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment Labels Name Unmatch", func() {
			By("SubnormalTest03_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest03_2 Delete Deployment X Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest03_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment Labels Namespace Unmatch", func() {
			By("SubnormalTest04_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest04_1 Create Deployment X Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest04_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment Labels Namespace Unmatch", func() {
			By("SubnormalTest04_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest04_2 Delete Deployment X Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest04_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment Labels Namespace None", func() {
			By("SubnormalTest05_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest05_1 Create Deployment X Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest05_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment Labels Namespace None", func() {
			By("SubnormalTest05_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest05_2 Delete Deployment X Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest05_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment DeletionTimestamp Set", func() {
			By("SubnormalTest06_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest06_1 Create Deployment X DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest06_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment DeletionTimestamp Set", func() {
			By("SubnormalTest06_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest06_2 Delete Deployment X DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest06_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 Status.Phase Not Rounning", func() {
			By("SubnormalTest07_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest07_1 Create Pod X1 Status.Phase Not Rounning
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest07_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 Status.Phase Not Rounning", func() {
			By("SubnormalTest07_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest07_2 Delete Pod X1 Status.Phase Not Rounning
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest07_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 2ndNIC Not Found", func() {
			By("SubnormalTest08_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest08_1 Create Pod X1 2ndNIC Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest08_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 2ndNIC Not Found", func() {
			By("SubnormalTest08_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest08_2 Delete Pod X1 2ndNIC Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest08_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 Labels None", func() {
			By("SubnormalTest09_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest09_2 Delete Pod X1 Labels None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest09_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod Labels Char None", func() {
			By("SubnormalTest10_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest10_1 Create Pod X Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest10_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod Labels Char None", func() {
			By("SubnormalTest10_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest10_2 Delete Pod X Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest10_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 Labels Name Unmatch", func() {
			By("SubnormalTest11_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest11_1 Create Pod X1 Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest11_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 Labels Name Unmatch", func() {
			By("SubnormalTest11_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest11_2 Delete Pod X1 Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest11_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 Labels Namespace Unmatch", func() {
			By("SubnormalTest12_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest12_1 Create Pod X1 Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest12_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 Labels Namespace Unmatch", func() {
			By("SubnormalTest12_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest12_2 Delete Pod X1 Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest12_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 Labels Namespace None", func() {
			By("SubnormalTest13_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest13_1 Create Pod X1 Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest13_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 Labels Namespace None", func() {
			By("SubnormalTest13_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest13_2 Delete Pod X1 Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest13_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod X1 DeletionTimestamp Set", func() {
			By("SubnormalTest14_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest14_1 Create Pod X1 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest14_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod X1 DeletionTimestamp Set", func() {
			By("SubnormalTest14_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest14_2 Delete Pod X1 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest14_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment X Not Found", func() {
			By("SubnormalTest15_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest15_1 Create Deployment X Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest15_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment X Not Found", func() {
			By("SubnormalTest15_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest15_2 Delete Deployment X Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest15_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ConnectEndpoint X1 DeletionTimestamp Set", func() {
			By("SubnormalTest16_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest16_1 Create ConnectEndpoint X1 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest16_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ConnectEndpoint X1 DeletionTimestamp Set", func() {
			By("SubnormalTest16_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest16_2 Delete ConnectEndpoint X1 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest16_2(writer, reconciler, mockClient)
		})
	})
})
