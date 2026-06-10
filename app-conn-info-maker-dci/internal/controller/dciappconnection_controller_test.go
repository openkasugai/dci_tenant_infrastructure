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

		BeforeAll(func() {
			fmt.Println(">")
			fmt.Println("> dciappconnection_controller_test.go")
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

			err = reconciler.SetupWithManager(mgr)
			if err != nil {
				fmt.Println("An error occur during setupwithManager: ", err)
			}
			Expect(err).NotTo(HaveOccurred())

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
			fmt.Println("< dciappconnection_controller_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_test.go")
			fmt.Println(">>>>> Start BeforeEach")
			fmt.Println(">>>>>")
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
			fmt.Println("<<<<< dciappconnection_controller_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_controller_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_controller_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_controller_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Create", func() {
			By("Test01_1: Reconciling the created resource")
			// ------------------------------
			//  Test01_1 Create
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test01_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test01_2: Reconciling the delete resource")
			// ------------------------------
			//  Test01_2 Delete
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test01_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test02_1: Reconciling the created resource")
			// ------------------------------
			//  Test02-1 Create
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test02_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test02-2: Reconciling the delete resource")
			// ------------------------------
			//  Test02_2 Delete
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test02_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test03_1: Reconciling the created resource")
			// ------------------------------
			//  Test03-1 Create
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test03_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test03-2: Reconciling the delete resource")
			// ------------------------------
			//  Test03_2 Delete
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test03_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test04_1: Reconciling the created resource")
			// ------------------------------
			//  Test04_1 Create HostDeviceNetwork None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test04_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test04_2: Reconciling the delete resource")
			// ------------------------------
			//  Test04_2 Delete Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test04_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test05_1: Reconciling the created resource")
			// ------------------------------
			//  Test05_1 Create Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test05_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test05_2: Reconciling the delete resource")
			// ------------------------------
			//  Test05_2 Delete Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test05_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test06_1: Reconciling the created resource")
			// ------------------------------
			//  Test06_1 Create Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test06_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test06_2: Reconciling the delete resource")
			// ------------------------------
			//  Test06_2 Delete Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test06_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test07_1: Reconciling the created resource")
			// ------------------------------
			//  Test07_1 Create Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test07_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test07_2: Reconciling the delete resource")
			// ------------------------------
			//  Test07_2 Delete Network None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test07_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test09-1: Reconciling the created resource")
			// ------------------------------
			//  Test09_1 Create
			//  X-None
			//  Y-A,Y-C,Y-X
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test09_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test09-2: Reconciling the created resource")
			// ------------------------------
			//  Test09_2 Delete
			//  X-None
			//  Y-A,Y-C,Y-X
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test09_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test10_1: Reconciling the created resource")
			// ------------------------------
			//  Test10_1 Create
			//  X-A,X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test10_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test10_2: Reconciling the created resource")
			// ------------------------------
			//  Test10_2 Delete
			//  X-A,X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test10_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test11_1: Reconciling the created resource")
			// ------------------------------
			//  Test11_1 Create
			//  X-A,X-Y
			//  Y-C
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_Test11_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test11_2: Reconciling the created resource")
			// ------------------------------
			//  Test11_2 Delete
			//  X-A,X-Y
			//  Y-C
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_Test11_2(writer, reconciler)
		})

		It("DCIAppConnection Create", func() {
			By("Test12_1: Reconciling the created resource")
			// ------------------------------
			//  Test12_1 Create
			//  X-None
			//  In-X,Out-X
			// ------------------------------
			dciappconnection_controller_Test12_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test12_2: Reconciling the created resource")
			// ------------------------------
			//  Test12_2 Delete
			//  X-None
			//  In-X,Out-X
			// ------------------------------
			dciappconnection_controller_Test12_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test13_1: Reconciling the created resource")
			// ------------------------------
			//  Test13_1 Create
			//  X-A
			//  In-X,Out-X
			// ------------------------------
			dciappconnection_controller_Test13_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test13_2: Reconciling the created resource")
			// ------------------------------
			//  Test13_2 Delete
			//  X-A
			//  In-X,Out-X
			// ------------------------------
			dciappconnection_controller_Test13_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test14_1: Reconciling the created resource")
			// ------------------------------
			//  Test14_1 Create
			//  X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test14_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test14_2: Reconciling the created resource")
			// ------------------------------
			//  Test14_2 Delete
			//  X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test14_2(writer, reconciler)
		})

		It("DCIAppConnection Create", func() {
			By("Test15_1: Reconciling the created resource")
			// ------------------------------
			//  Test15_1 Create
			//  X-A,Y-C,X-Y
			//  Y-C,Y-A
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test15_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test15_2: Reconciling the created resource")
			// ------------------------------
			//  Test15_2 Delete
			//  X-A,Y-C,X-Y
			//  Y-C,Y-A
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test15_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test16_1: Reconciling the created resource")
			// ------------------------------
			//  Test16_1 Create
			//  X-A
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			dciappconnection_controller_Test16_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test16_2: Reconciling the created resource")
			// ------------------------------
			//  Test16_2 Delete
			//  X-A
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			dciappconnection_controller_Test16_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test17_1: Reconciling the created resource")
			// ------------------------------
			//  Test17_1 Create
			//  X-A
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			dciappconnection_controller_Test17_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test17_2: Reconciling the created resource")
			// ------------------------------
			//  Test17_2 Delete
			//  X-A
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			dciappconnection_controller_Test17_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test18_1: Reconciling the created resource")
			// ------------------------------
			//  Test18_1 Create
			//  X-A
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			dciappconnection_controller_Test18_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test18_2: Reconciling the created resource")
			// ------------------------------
			//  Test18_2 Delete
			//  X-A
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			dciappconnection_controller_Test18_2(writer, reconciler)
		})

		It("DCIAppConnection Create", func() {
			By("Test21_1: Reconciling the created resource")
			// ------------------------------
			//  Test21_1 Create
			//  Base:Test01_1
			//  X-A,Y-C,X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test21_1(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test22_1: Reconciling the created resource")
			// ------------------------------
			//  Test22_1 Create
			//  Base:Test01_1
			//  X-A,Y-C,X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test22_1(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test23_1: Reconciling the created resource")
			// ------------------------------
			//  Test23_1 Create A2 Pod Deleting
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test23_1(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test24_1: Reconciling the created resource")
			// ------------------------------
			//  Test24_1 Create Y0 ReplicaSet
			//  X-A,X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test24_1(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test24_2: Reconciling the deleted resource")
			// ------------------------------
			//  Test24_2 Delete Y0 ReplicaSet
			//  X-A,X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test24_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test25_1: Reconciling the created resource")
			// ------------------------------
			//  Test25_1 Create Y0 ReplicaSet
			//  X-A,X-Y Y1Y2 Pod Replicas 1
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test25_1(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test25_2: Reconciling the delete resource")
			// ------------------------------
			//  Test25_2 Delete Y0 ReplicaSet
			//  X-A,X-Y Y1Y2 Pod Replicas 1
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test25_2(writer, reconciler)
		})
		// EventPredicate Tests
		It("EventPredicate Update Test01", func() {
			By("EventPredicate_Update_Test01: ObjectOld is nil")
			dciappconnection_controller_EventPredicate_Update_Test01(writer, reconciler)
		})
		It("EventPredicate Update Test02", func() {
			By("EventPredicate_Update_Test02: ObjectNew is nil")
			dciappconnection_controller_EventPredicate_Update_Test02(writer, reconciler)
		})
		It("EventPredicate Update Test03", func() {
			By("EventPredicate_Update_Test03: ObjectNew is not DCIAppConnection type")
			dciappconnection_controller_EventPredicate_Update_Test03(writer, reconciler)
		})
		It("EventPredicate Update Test04", func() {
			By("EventPredicate_Update_Test04: DeletionTimestamp is not zero")
			dciappconnection_controller_EventPredicate_Update_Test04(writer, reconciler)
		})
		It("EventPredicate Update Test05", func() {
			By("EventPredicate_Update_Test05: Normal update (all conditions pass)")
			dciappconnection_controller_EventPredicate_Update_Test05(writer, reconciler)
		})
		It("EventPredicate Create Test01", func() {
			By("EventPredicate_Create_Test01: Create event always returns true")
			dciappconnection_controller_EventPredicate_Create_Test01(writer, reconciler)
		})
		It("EventPredicate Create Test02", func() {
			By("EventPredicate_Create_Test02: Create event with nil object")
			dciappconnection_controller_EventPredicate_Create_Test02(writer, reconciler)
		})
		It("EventPredicate Create Test03", func() {
			By("EventPredicate_Create_Test03: Create event with DeletionTimestamp set")
			dciappconnection_controller_EventPredicate_Create_Test03(writer, reconciler)
		})
		It("EventPredicate Create Test04", func() {
			By("EventPredicate_Create_Test04: Create event with non-DCIAppConnection type")
			dciappconnection_controller_EventPredicate_Create_Test04(writer, reconciler)
		})
	})
})
