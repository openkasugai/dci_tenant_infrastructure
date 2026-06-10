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
			fmt.Println("> dciappconnection_common_Subnormal3_test.go")
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
			fmt.Println("< dciappconnection_common_Subnormal3_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal3_test.go")
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
			fmt.Println("<<<<< dciappconnection_common_Subnormal3_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Common Subnormal Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal3_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_common_Subnormal3_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_common_Subnormal3_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_common_Subnormal3_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Create ReplicaSet A List Zero", func() {
			By("SubnormalTest38_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest38_1 Create ReplicaSet A List Zero
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest38_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ReplicaSet A List Zero", func() {
			By("SubnormalTest38_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest38_2 Delete ReplicaSet A List Zero
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest38_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ReplicaSet A DeletionTimestamp Set", func() {
			By("SubnormalTest39_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest39_1 Create ReplicaSet A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest39_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ReplicaSet A DeletionTimestamp Set", func() {
			By("SubnormalTest39_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest39_2 Delete ReplicaSet A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest39_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create ReplicaSet A Hash Label None", func() {
			By("SubnormalTest40_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest40_1 Create ReplicaSet A Hash Label None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest40_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete ReplicaSet A Hash Label None", func() {
			By("SubnormalTest40_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest40_2 Delete ReplicaSet A Hash Label None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest40_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A List Zero", func() {
			By("SubnormalTest41_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest41_1 Create Pod A List Zero
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest41_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A List Zero", func() {
			By("SubnormalTest41_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest41_2 Delete Pod A List Zero
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest41_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A Replicas Not Equal List Num", func() {
			By("SubnormalTest42_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest42_1 Create Pod A Replicas Not Equal List Num
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest42_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A Replicas Not Equal List Num", func() {
			By("SubnormalTest42_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest42_2 Delete Pod A Replicas Not Equal List Num
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest42_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A None Replicas List Num Error", func() {
			By("SubnormalTest43_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest43_1 Create Pod A None Replicas List Num Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest43_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A None Replicas List Num Error", func() {
			By("SubnormalTest43_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest43_2 Delete Pod A None Replicas List Num Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest43_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 HostDeviceNetwork Get Error", func() {
			By("SubnormalTest44_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest44_1 Create Pod A2 HostDeviceNetwork Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest44_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 HostDeviceNetwork Get Error", func() {
			By("SubnormalTest44_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest44_2 Delete Pod A2 HostDeviceNetwork Get Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest44_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Annotation network-status None", func() {
			By("SubnormalTest45_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest45_1 Create Pod A2 Annotation network-status None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest45_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Annotation network-status None", func() {
			By("SubnormalTest45_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest45_2 Delete Pod A2 Annotation network-status None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest45_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Annotation network-status Char None", func() {
			By("SubnormalTest46_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest46_1 Create Pod A2 Annotation network-status Char None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest46_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Annotation network-status Char None", func() {
			By("SubnormalTest46_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest46_2 Delete Pod A2 Annotation network-status Char None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest46_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod Y None Replicas List Num Error", func() {
			By("SubnormalTest47_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest47_1 Create Pod A None Replicas List Num Error
			//  X-A,X-Y Y1Y2 Pod Replicas None
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest47_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod Y None Replicas List Num Error", func() {
			By("SubnormalTest47_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest47_2 Delete Pod A None Replicas List Num Error
			//  X-A,X-Y Y1Y2 Pod Replicas None
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest47_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod Y None Replicas List Num Error", func() {
			By("SubnormalTest48_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest48_1 Create Pod A None Replicas List Num Error
			//  X-A,X-Y Y1 Pod Replicas None
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest48_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod Y None Replicas List Num Error", func() {
			By("SubnormalTest48_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest48_2 Delete Pod A None Replicas List Num Error
			//  X-A,X-Y Y1 Pod Replicas None
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest48_2(writer, reconciler, mockClient)
		})
	})
})
