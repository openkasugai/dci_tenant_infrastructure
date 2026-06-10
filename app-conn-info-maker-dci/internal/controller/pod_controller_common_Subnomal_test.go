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
			By("SubNomalTest01-1: Annotation No Set Ips")
			// ------------------------------
			//	Test01-1 Annotation No Set Ips
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest01_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest01-2: Annotation No Set Ips Delete")
			// ------------------------------
			//	Test01-1 Annotation No Set Ips Delete
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest01_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest02-1: bFuncFlag is false for getFromConnectionInformation")
			// ------------------------------
			//	Test02-1 Annotation No Set Ips
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest02_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest02-2: bFuncFlag is false for getFromConnectionInformation Delete")
			// ------------------------------
			//	Test02-1 bFuncFlag is false for Delete
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest02_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest03-1: bFuncFlag is false for getToConnectionInformation")
			// ------------------------------
			//	Test03-1 bFuncFlag is false for getToConnectionInformation
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest03_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest03-2: bFuncFlag is false for getToConnectionInformation Delete")
			// ------------------------------
			//	Test03-2 bFuncFlag is false for getToConnectionInformation Delete
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest03_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest04-1: From Deployment is nil")
			// ------------------------------
			//	Test04-1 From Deployment is nil
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest04_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest04-2: From Deployment is nil Delete")
			// ------------------------------
			//	Test04-2 From Deployment is nil Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest04_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest05-1: From Deployment Label is NotFound dciappconnection")
			// ------------------------------
			//	Test05-1 From Deployment Label is NotFound dciappconnection
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest05_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest05-2: From Deployment Label is NotFound dciappconnection Delete")
			// ------------------------------
			//	Test05-2 From Deployment Label is NotFound dciappconnection Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest05_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest06-1: From Deployment Label dciappconnection len 0")
			// ------------------------------
			//	Test06-1 From Deployment Label dciappconnection len 0
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest06_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest06-2: From Deployment Label dciappconnection len 0 Delete")
			// ------------------------------
			//	Test06-2 From Deployment Label dciappconnection len 0 Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest06_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest07-1: From Deployment Label name/namespace No match")
			// ------------------------------
			//	Test07-1 From Deployment Label name/namespace No match
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest07_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest07-2: From Deployment Label name/namespace No match Delete")
			// ------------------------------
			//	Test07-1 From Deployment Label name/namespace No matc Deleteh
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest07_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest08-1: From Pod Label name/namespace No match")
			// ------------------------------
			//	Test08-1 From Pod Label name/namespace No match
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest08_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest08-2: From Pod Label name/namespace No match Delete")
			// ------------------------------
			//	Test08-2 From Pod Label name/namespace No match Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest08_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest09-1:  From ConnectEndpoint set DeletiontimeStamp")
			// ------------------------------
			//	Test09-1 From ConnectEndpoint set DeletiontimeStamp
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest09_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest09-2:  From ConnectEndpoint set DeletiontimeStamp Delete")
			// ------------------------------
			//	Test09-2 From ConnectEndpoint set DeletiontimeStamp Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest09_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest10-1: To Deployment is nil")
			// ------------------------------
			//	Test10-1 To Deployment is nil
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest10_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest10-2: To Deployment is nil Delete")
			// ------------------------------
			//	Test10-2 To Deployment is nil Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest10_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest11-1: To Deployment Label dciappconnection len 0")
			// ------------------------------
			//	Test11-1 To Deployment Label dciappconnection len 0
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest11_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest11-2: To Deployment Label dciappconnection len 0 Delete")
			// ------------------------------
			//	Test11-2 To Deployment Label dciappconnection len 0 Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest11_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest12-1: To Pod Label dciappconnection Unmatch")
			// ------------------------------
			//	Test12-1 To Pod Label dciappconnection Unmatch
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest12_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest12-2: To Pod Label dciappconnection Unmatch Delete")
			// ------------------------------
			//	Test12-2 To Pod Label dciappconnection Unmatch Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest12_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest13-1: To Deployment Label dciappconnection Not found")
			// ------------------------------
			//	Test13-1 To Deployment Label dciappconnection Not found
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest13_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest13-2: To Deployment Label dciappconnection Not found Delete")
			// ------------------------------
			//	Test13-2 To Deployment Label dciappconnection Not found Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest13_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest14-1: To Deployment Label set DeletionTimestamp")
			// ------------------------------
			//	Test14-1 To Deployment Label set DeletionTimestamp
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest14_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest14-2: To Deployment Label set DeletionTimestamp Delete")
			// ------------------------------
			//	Test14-2 To Deployment Label set DeletionTimestamp Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest14_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest15-1: To Pod Label dciappconnection Not found")
			// ------------------------------
			//	Test15-1 To Pod Label dciappconnection Not found
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest15_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest15-2: To Pod Label dciappconnection Not found Delete")
			// ------------------------------
			//	Test15-2 To Pod Label dciappconnection Not found Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest15_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest16-1: To Replicaset set DeletionTimestamp")
			// ------------------------------
			//	Test16-1 To ReplicaSet set DeletionTimestamp
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest16_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest16_2: To Replicaset set DeletionTimestamp Delete")
			// ------------------------------
			//	Test16-2 To ReplicaSet set DeletionTimestamp Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest16_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest17-1: From Deployment set DeletionTimestamp")
			// ------------------------------
			//	Test17-1 From Deployment set DeletionTimestamp
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest17_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest17-2: From Deployment set DeletionTimestamp Delete")
			// ------------------------------
			//	Test17-2 From Deployment set DeletionTimestamp Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest17_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest18-1: From Deployment Label Name Only")
			// ------------------------------
			//	Test18-1 From Deployment Label Name Only
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest18_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest18-2: From Deployment Label Name Only Delete")
			// ------------------------------
			//	Test18-2 From Deployment Label Name Only Delete
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest18_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest19-1: From Pod Annotations no set IPs")
			// ------------------------------
			//	Test19-1 From Pod Annotations no set IPs
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest19_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest19-2: From Pod Annotations no set IPs Delete")
			// ------------------------------
			//	Test19-2 From Pod Annotations no set IPs Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest19_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest20-1: From Pod Label dciappconnection NotFound !AdditionalProcFlag")
			// ------------------------------
			//	Test20-1 From Pod Label dciappconnection NotFound !AdditionalProcFlag
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest20_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest20-2: From Pod Label dciappconnection NotFound AdditionalProcFlag")
			// ------------------------------
			//	Test20-2 From Pod Label dciappconnection NotFound AdditionalProcFlag
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest20_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest21-1: From Pod Label dciappconnection len 0 !AdditionalProcFlag")
			// ------------------------------
			//	Test21-1 From Pod Label dciappconnection NotFound !AdditionalProcFlag
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest21_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest21-2: From Pod Label dciappconnection len 0 AdditionalProcFlag")
			// ------------------------------
			//	Test21-2 From Pod Label dciappconnection NotFound AdditionalProcFlag
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest21_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest22-1: From Pod Label dciappconnection NameOnly")
			// ------------------------------
			//	Test22-1 From Pod Label dciappconnection NameOnly
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest22_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest22-2: From Pod Label dciappconnection NameOnly Delete")
			// ------------------------------
			//	Test22-2 From Pod Label dciappconnection NameOnly Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest22_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest23-1: To Deployment Label dciappconnection Name Only")
			// ------------------------------
			//	Test23-1 To Deployment Label dciappconnection Name Only
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest23_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest23-2: To Deployment Label dciappconnection Name Only Delete")
			// ------------------------------
			//	Test23-2 To Deployment Label dciappconnection Name Only Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest23_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest24-1: To EndpointSlice Not Found")
			// ------------------------------
			//	Test24-1 To EndpointSlice Not Found
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest24_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest24-2: To EndpointSlice Not Found Delete")
			// ------------------------------
			//	Test24-2 To EndpointSlice Not Found Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest24_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest25-1: To Pod No 2ndNIC")
			// ------------------------------
			//	Test25-1  To Pod No 2ndNIC
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest25_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest25-2: To Pod No 2ndNIC Delete")
			// ------------------------------
			//	Test25-2  To Pod No 2ndNIC Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest25_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest26-1: To Pod Label dciappconnection NotFound !AdditionalProcFlag")
			// ------------------------------
			//	Test26-1 To Pod Label dciappconnection NotFound !AdditionalProcFlag
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest26_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest26-2: To Pod Label dciappconnection NotFound AdditionalProcFlag")
			// ------------------------------
			//	Test26-2 To Pod Label dciappconnection NotFound AdditionalProcFlag
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest26_2(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest27-1: To Pod Label dciappconnection len 0 !AdditionalProcFlag")
			// ------------------------------
			//	Test27-1 To Pod Label dciappconnection len 0 !AdditionalProcFlag
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest27_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest27-2: To Pod Label dciappconnection len 0 AdditionalProcFlag")
			// ------------------------------
			//	Test27-2 To Pod Label dciappconnection len 0 AdditionalProcFlag
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest27_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest28-1: To Pod Labels Only Name")
			// ------------------------------
			//	Test28-1 To Pod Labels Only Name
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest28_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest28-2: To Pod Labels Only Name Delete")
			// ------------------------------
			//	Test28-2 To Pod Labels Only Name Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest28_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest29-1: ReplicaSet Item len 0")
			// ------------------------------
			//	Test29-1 ReplicaSet Item len 0
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest29_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest29-2: ReplicaSet Item len 0 Delete")
			// ------------------------------
			//	Test29-2 ReplicaSet Item len 0 Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest29_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest30-1: ReplicaSet Labels not found pod-template-hash")
			// ------------------------------
			//	Test30-1 ReplicaSet Labels not found pod-template-hash
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest30_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest30-2: ReplicaSet Labels not found pod-template-hash Delete")
			// ------------------------------
			//	Test30-2 ReplicaSet Labels not found pod-template-hash Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest30_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest31-1: Pod Item len 0")
			// ------------------------------
			//	Test31-1 Pod Item len 0
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest31_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest31-2: Pod Item len 0 Delete")
			// ------------------------------
			//	Test31-2 Pod Item len 0 Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest31_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest32-1: Pod Item len 0")
			// ------------------------------
			//	Test32-1 cAnnotationJSON len 0
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest32_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest32-2: Pod Item len 0 Delete")
			// ------------------------------
			//	Test32-2 cAnnotationJSON len 0 Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest32_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest33-1: Ips len 0")
			// ------------------------------
			//	Test33-1 Ips len 0
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest33_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest33-2: Ips len 0 Delete")
			// ------------------------------
			//	Test33-2 Ips len 0 Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest33_2(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("SubNomalTest34-1: EndpointSlice set DeletionTimestamp")
			// ------------------------------
			//	Test34-1 EndpointSlice set DeletionTimestamp
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest34_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("SubNomalTest34-2: EndpointSlice set DeletionTimestamp Delete")
			// ------------------------------
			//	Test34-2 EndpointSlice set DeletionTimestamp Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_common_SubNomalTest34_2(writer, reconciler, mockClient)
		})
	})
})
