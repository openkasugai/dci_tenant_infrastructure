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
			By("Test01-1: Reconciling the created resource")
			// ------------------------------
			//	Test01-1 Pod Create
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test01_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test01-2: Reconciling the delete resource")
			// ------------------------------
			//	Test01-2 Pod Delete
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test01_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test02-1: Reconciling the created resource")
			// ------------------------------
			//	Test02-1 Pod Create
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test02_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test02-2: Reconciling the delete resource")
			// ------------------------------
			//	Test02-2 Pod Delete
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test02_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test03-1: Reconciling the created resource")
			// ------------------------------
			//	Test03-1 Pod Create
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None  (Y=Replicas 1to2)
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test03_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test03-2: Reconciling the delete resource")
			// ------------------------------
			//	Test03-2 Pod Delete
			//  X-A,Y-C,X-Y (Y=Replicas 2to1)
			//  Y-None (Y=Replicas 1to2)
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test03_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test04-1: Reconciling the created resource")
			// ------------------------------
			//	Test04-1 Pod Create
			//  X-A,Y-C,X-Y (Y=Replicas 1to2)
			//  Y-None (Y=Replicas 1to2 & EndpointSlice Change)
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test04_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test05-1: Reconciling the created resource")
			// ------------------------------
			//	Test05-1 Pod Create
			//  X-Y,X-A,Y-C (Y=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			//  Change the definition order with X-Y at the top
			// ------------------------------
			pod_controller_Test05_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test06-1: Reconciling the created resource")
			// ------------------------------
			//	Test06-1 Pod Create
			//  X-Y,X-A,Y-C (A is two ReplicaSet)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test06_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test06-2: Reconciling the created resource")
			// ------------------------------
			//	Test06-2 Pod Create
			//  X-Y,X-A,Y-C (A is two ReplicaSet to one)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test06_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test07-1: Reconciling the created resource")
			// ------------------------------
			//	Test07-1 Pod Create HostDeviceNetwork None
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test07_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test07-2: Reconciling the delete resource")
			// ------------------------------
			//	Test07-2 Pod Delete HostDeviceNetwork None
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test07_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test08-1: Reconciling the created resource")
			// ------------------------------
			//	Test08-1 Pod Create Network None
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test08_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test08-2: Reconciling the delete resource")
			// ------------------------------
			//	Test08-2 Pod Delete Network None
			//  X-A,Y-C,X-Y (A=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test08_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test09-1: Reconciling the created resource")
			// ------------------------------
			//	Test09-1 Pod Create HostDeviceNetwork None
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test09_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test09-2: Reconciling the delete resource")
			// ------------------------------
			//	Test09-2 Pod Delete HostDeviceNetwork None
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test09_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test10-1: Reconciling the created resource")
			// ------------------------------
			//	Test10-1 Pod Create Network None
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test10_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test10-2: Reconciling the delete resource")
			// ------------------------------
			//	Test10-2 Pod Delete Network None
			//  X-A,Y-C,X-Y (X=Replicas 2to1)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test10_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test12-1: Reconciling the created resource")
			// ------------------------------
			//	Test12-1 Pod Create
			//  X-A,X-Y (X=Replicas 1to2)
			//  Y-C
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test12_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test13-1: Reconciling the created resource")
			// ------------------------------
			//	Test13-1 Pod Create
			//  X-A,X-Y
			//  Y-C (Y=Replicas 1to2)
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test13_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test14-1: Reconciling the created resource")
			// ------------------------------
			//	Test14-1
			//  X-A,X-Y (X=Reploicas 2)
			//  Y-C (Y=Replicas 1to2)
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test14_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test15-1-1: Reconciling the created resource")
			// ------------------------------
			//	Test15-1-1 Pod Create
			//  X-A (A=Replicas 1to2)
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			pod_controller_Test15_1_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test15-2-1: Reconciling the delete resource")
			// ------------------------------
			//	Test15-2-1 Pod Delete
			//  X-A (A=Replicas 2to1)
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			pod_controller_Test15_2_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test15-1-2: Reconciling the created resource")
			// ------------------------------
			//	Test15-1-2 Pod Create
			//  X-A (A=Replicas 1to2)
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			pod_controller_Test15_1_2(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test15-2-2: Reconciling the delete resource")
			// ------------------------------
			//	Test15-2-2 Pod Delete
			//  X-A (A=Replicas 2to1)
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			pod_controller_Test15_2_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test15-1-3: Reconciling the created resource")
			// ------------------------------
			//	Test15-1-3 Pod Create
			//  X-A (A=Replicas 1to2)
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			pod_controller_Test15_1_3(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test15-2-3: Reconciling the delete resource")
			// ------------------------------
			//	Test15-2-3 Pod Delete
			//  X-A (A=Replicas 2to1)
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			pod_controller_Test15_2_3(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test16-1-1: Reconciling the created resource")
			// ------------------------------
			//	Test16-1-1 Pod Create
			//  X-A (X=Replicas 1to2)
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			pod_controller_Test16_1_1(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test16-2-1: Reconciling the delete resource")
			// ------------------------------
			//	Test16-2-1 Pod Delete
			//  X-A (X=Replicas 2to1)
			//  Y-None
			//  In-None,Out-X
			// ------------------------------
			pod_controller_Test16_2_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test16-1-2: Reconciling the created resource")
			// ------------------------------
			//	Test16-1-2 Pod Create
			//  X-A (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			pod_controller_Test16_1_2(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test16-2-2: Reconciling the delete resource")
			// ------------------------------
			//	Test16-2-2 Pod Delete
			//  X-A (X=Replicas 2to1)
			//  Y-None
			//  In-X,Out-None
			// ------------------------------
			pod_controller_Test16_2_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test16-1-3: Reconciling the created resource")
			// ------------------------------
			//	Test16-1-3 Pod Create
			//  X-A (X=Replicas 1to2)
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			pod_controller_Test16_1_3(writer, reconciler)
		})
		It("Pod Delete", func() {
			By("Test16-2-3: Reconciling the delete resource")
			// ------------------------------
			//	Test16-2-3 Pod Delete
			//  X-A (X=Replicas 2to1)
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			pod_controller_Test16_2_3(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test17-1: Reconciling the created resource")
			// ------------------------------
			//	Test17-1 Pod Create In/Out Merge
			//  X-A,X-Y (A=Replicas 1to2)
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test17_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test17-2: Reconciling the created resource")
			// ------------------------------
			//	Test17-2 Pod Create In/Out Merge
			//  X-A,X-Y
			//  Y-C (C=Replicas 1to2)
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test17_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test17-3: Reconciling the created resource")
			// ------------------------------
			//	Test17-3 Pod Create In/Out Merge
			//  X-A,X-Y (X=Replicas 1to2)
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test17_3(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test17-4: Reconciling the created resource")
			// ------------------------------
			//	Test17-4 Pod Create In/Out Merge
			//  X-A,X-Y (Y=Replicas 1to2)
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test17_4(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test21-1: Reconciling the created resource")
			// ------------------------------
			//	Test21-1 Pod Create PodA2 CR NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test21_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test22-1: Reconciling the created resource")
			// ------------------------------
			//	Test22-1 Pod Create PodA2 is 2ndNic NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test22_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test23-1: Reconciling the created resource")
			// ------------------------------
			//	Test23-1 Pod Create DCIAppConnection CR NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test23_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test23-2: Reconciling the created resource")
			// ------------------------------
			//	Test23-2 Pod Create Deployment NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test23_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test23-3: Reconciling the created resource")
			// ------------------------------
			//	Test23-3 Pod Create ReplicaSet NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test23_3(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test24-1: Reconciling the created resource")
			// ------------------------------
			//	Test24-1 Pod Create DCIAppConnectionCR Finalizer NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test24_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test25-1: Reconciling the created resource")
			// ------------------------------
			//	Test25-1 Pod Create Service NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test25_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test25-2: Reconciling the created resource")
			// ------------------------------
			//	Test25-2 Pod Create EndpointSlice NotFound
			//  X-A,Y-C,X-Y (A=Replicas 1to2)
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			pod_controller_Test25_2(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test32-1: Reconciling the created resource")
			// ------------------------------
			//	Test32-1 Pod Create ToY1 Not Ready
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test32_1(writer, reconciler)
		})
		It("Pod Create", func() {
			By("Test33-1: Reconciling the created resource")
			// ------------------------------
			//	Test33-1 Pod Create In Y & FromX1 Not Ready
			//  X-A,Y-C,X-Y (X=Replicas 1to2)
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			pod_controller_Test33_1(writer, reconciler)
		})

		// EventPredicate Tests
		It("EventPredicate Update Test01", func() {
			By("EventPredicate_Update_Test01: ObjectOld is nil")
			pod_controller_EventPredicate_Update_Test01(writer, reconciler)
		})
		It("EventPredicate Update Test02", func() {
			By("EventPredicate_Update_Test02: ObjectNew is nil")
			pod_controller_EventPredicate_Update_Test02(writer, reconciler)
		})
		It("EventPredicate Update Test03", func() {
			By("EventPredicate_Update_Test03: ObjectOld is not Pod type")
			pod_controller_EventPredicate_Update_Test03(writer, reconciler)
		})
		It("EventPredicate Update Test04", func() {
			By("EventPredicate_Update_Test04: ObjectNew is not Pod type")
			pod_controller_EventPredicate_Update_Test04(writer, reconciler)
		})
		It("EventPredicate Update Test05", func() {
			By("EventPredicate_Update_Test05: DeletionTimestamp is not zero")
			pod_controller_EventPredicate_Update_Test05(writer, reconciler)
		})
		It("EventPredicate Update Test06", func() {
			By("EventPredicate_Update_Test06: Annotations are different")
			pod_controller_EventPredicate_Update_Test06(writer, reconciler)
		})
		It("EventPredicate Update Test07", func() {
			By("EventPredicate_Update_Test07: Pod Phase is different")
			pod_controller_EventPredicate_Update_Test07(writer, reconciler)
		})
		It("EventPredicate Update Test08", func() {
			By("EventPredicate_Update_Test08: Normal update (all conditions pass)")
			pod_controller_EventPredicate_Update_Test08(writer, reconciler)
		})
		It("EventPredicate Create Test01", func() {
			By("EventPredicate_Create_Test01: Create event always returns true")
			pod_controller_EventPredicate_Create_Test01(writer, reconciler)
		})
		It("EventPredicate Create Test02", func() {
			By("EventPredicate_Create_Test02: Create event with nil object")
			pod_controller_EventPredicate_Create_Test02(writer, reconciler)
		})
		It("EventPredicate Create Test03", func() {
			By("EventPredicate_Create_Test03: Create event with DeletionTimestamp set")
			pod_controller_EventPredicate_Create_Test03(writer, reconciler)
		})
		It("EventPredicate Create Test04", func() {
			By("EventPredicate_Create_Test04: Create event with non-Pod type")
			pod_controller_EventPredicate_Create_Test04(writer, reconciler)
		})
	})
})
