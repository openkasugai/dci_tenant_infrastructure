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
		log := textlogger.NewLogger(textlogger.NewConfig())

		BeforeAll(func() {
			fmt.Println(">")
			fmt.Println("> pod_controller_test.go")
			fmt.Println("> Start BeforeAll")
			fmt.Println(">")
			// set up log
			log := textlogger.NewLogger(textlogger.NewConfig())

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
			fmt.Println(">>")
			fmt.Println(">> pod_controller_test.go")
			fmt.Println(">> Start BeforeEach")
			fmt.Println(">>")

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
			By("SubNomalTest01-1: PodResource is NotFound.")
			// ------------------------------
			//	Test01-1 PodResource is NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest01_1(writer, reconciler, mockClient)
		})

		It("Pod Create", func() {
			By("Test02-1: SubNomal getPodRelationshipInfo")
			// ------------------------------
			//	Test02-1 The Pod won't use a 2nd NIC.
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest02_1(writer, reconciler, mockClient)
		})

		It("Pod Create", func() {
			By("Test03-1: CreatePod Conditions is Not Ready")
			// ------------------------------
			//	Test03-1 CreatePod Conditions is Not Ready
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest03_1(writer, reconciler, mockClient)
		})

		It("Pod Update", func() {
			By("Test04-1: EventKind Update")
			// ------------------------------
			//	Test04-1 EventKind Update
			//  X-A,Y-C,X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest04_1(writer, reconciler, mockClient)
		})

		It("Event None", func() {
			By("Test05-1: EventKind None")
			// ------------------------------
			//	Test05-1 Pod None
			//  X-A,Y-C,X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest05_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test06-1: Replicaset set DeletionTimeStamp")
			// ------------------------------
			//	Test06-1 Replicaset set DeletionTimeStamp
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest06_1(writer, reconciler, mockClient)
		})

		It("Pod Create", func() {
			By("Test07-1: DeploymentResource is NotFound")
			// ------------------------------
			//	Test07-1 Pod Create
			//  X-Y,X-A,Y-C (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest07_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test08-1: Deployment set DeletionTimeStamp")
			// ------------------------------
			//	Test08-1 Deployment set DeletionTimeStamp
			//  X-Y,X-A,Y-C (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest08_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test09-1: Deployment Label is nil")
			// ------------------------------
			//	Test09-1 Deployment Label is nil
			//  X-Y,X-A,Y-C (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest09_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test10-1: Deployment Label NotFound")
			// ------------------------------
			//	Test10-1 Deployment Label NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest10_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test011-1: Deployment Label Exists empty")
			// ------------------------------
			//	Test11-1 Deployment Label Exists empty
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest11_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test12-1: Deployment Labels Only Name")
			// ------------------------------
			//	Test12-1 Deployment Labels Only Name
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest12_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test13-1: DCIAppConnection set DeletionTimeStamp")
			// ------------------------------
			//	Test13-1 DCIAppConnection set DeletionTimeStamp
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest13_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test14-1: DCIAppConnection is Not Created")
			// ------------------------------
			//	Test14-1 DCIAppConnection is Not Created
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest14_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test15-1: ConnectEndpoint set DeletionTimeStamp INOUT")
			// ------------------------------
			//	Test15-1 ConnectEndpoint set DeletionTimeStamp
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest15_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test16-1: ConnectEndpoint set DeletionTimeStamp CONN")
			// ------------------------------
			//	Test16-1 ConnectEndpoint set DeletionTimeStamp CONN
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest16_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test17-1: FromPod But Not Inuput/OutputEndpoint")
			// ------------------------------
			//	Test17-1 FromPod But Not Inuput/OutputEndpoint
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None
			//  In-X,Out-X
			// ------------------------------
			pod_controller_SubNomalTest17_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test18-1: EndpointSlice Patch Failed delAddressEndpointSlice")
			// ------------------------------
			//	Test18-1 EndpointSlice Patch Failed delAddressEndpointSlice
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest18_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test19-1: DeletePod Label Len 0")
			// ------------------------------
			//	Test19-1 DeletePod Label Len 0
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest19_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test20-1: Pod Label Name Only")
			// ------------------------------
			//	Test20-1 Pod Label Name Only
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest20_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test21-1: FromPodResource is nil")
			// ------------------------------
			//	Test21-1 PodResource is nil
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest21_1(writer, reconciler, mockClient)
		})
		It("Pod Create", func() {
			By("Test22-1: FromPodResource is nil after makeConnectEndpoint")
			// ------------------------------
			//	Test22-1 PodResource is nil after makeConnectEndpoint
			//  X-A,X-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest22_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test23-1: FromPodResource is nil")
			// ------------------------------
			//	Test23-1 PodResource is nil
			//  X-A,X-A,X-Y (X=Replicas 0to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest23_1(writer, reconciler, mockClient)
		})
		It("Pod Delete", func() {
			By("Test24-1: ToPodResource is nil after makeConnectEndpoint")
			// ------------------------------
			//	Test24-1 PodResource is nil
			//  X-A,X-A,X-Y (X=Replicas 0to1)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_SubNomalTest24_1(writer, reconciler, mockClient)
		})
	})
})
