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
			fmt.Println("> dciappconnection_controller_Subnormal_test.go")
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
			fmt.Println("< dciappconnection_controller_Subnormal_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_Subnormal_test.go")
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
			fmt.Println("<<<<< dciappconnection_controller_Subnormal_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_controller_Subnormal_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_controller_Subnormal_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_controller_Subnormal_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_controller_Subnormal_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection DCIAppConnection Reconciling the not action process error", func() {
			By("SubnormalTest01_1: Reconciling the created resource")
			// ------------------------------
			//  SubnormalTest01_1 DCIAppConnection Reconciling the not action process error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest01_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection DCIAppConnection Reconciling the not update process error", func() {
			By("SubnormalTest01_2: Reconciling the update resource")
			// ------------------------------
			//  SubnormalTest01_2 DCIAppConnection Reconciling the not update process error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			// dciappconnection_controller_SubnormalTest01_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create InputEndpoint Get Label Not Found Error", func() {
			By("SubnormalTest02_1: Reconciling the created resource")
			// ------------------------------
			//  SubnormalTest02_1 Create DCIAppConnection InputEndpoint Get Label Not Found Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest02_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete InputEndpoint Get Label Not Found Error", func() {
			By("SubnormalTest02_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest02_2 Delete DCIAppConnection InputEndpoint Get Label Not Found Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest02_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create OutputEndpoint Get Label Not Found Error", func() {
			By("SubnormalTest03_1: Reconciling the created resource")
			// ------------------------------
			//  SubnormalTest03_1 Create DCIAppConnection OutputEndpoint Get Label Not Found Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest03_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete OutputEndpoint Get Label Not Found Error", func() {
			By("SubnormalTest03_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest03_2 Delete DCIAppConnection OutputEndpoint Get Label Not Found Error
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest03_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Service and EndpointSlice Patch", func() {
			By("SubnormalTest04_1: Reconciling the created resource")
			// ------------------------------
			//  SubnormalTest04_1 DCIAppConnection Create Service and EndpointSlice Patch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest04_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Update Y-C Add Replicas Unmatch Error", func() {
			By("SubnormalTest05_1: Reconciling the updated resource")
			// ------------------------------
			//  SubnormalTest05_1 DCIAppConnection Update Y-C Add Replicas Unmatch Error
			//  X-A X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest05_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Update In-Y Chg FromX1 Del ConnectEndpoint DeletionTimestamp True", func() {
			By("SubnormalTest06_1: Reconciling the updated resource")
			// ------------------------------
			//  SubnormalTest06_1 DCIAppConnection Update In-Y Chg FromX1 Del ConnectEndpoint DeletionTimestamp True
			//  X-None
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest06_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Update Y-C Del Service NotFound Error", func() {
			By("SubnormalTest07_1: Reconciling the updated resource")
			// ------------------------------
			//  SubnormalTest07_1 DCIAppConnection Update Y-C Del Service NotFound Error
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest07_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Update Y-C Del Service DeletionTimestamp Error", func() {
			By("SubnormalTest07_2: Reconciling the updated resource")
			// ------------------------------
			//  SubnormalTest07_2 DCIAppConnection Update Y-C Del Service DeletionTimestamp Error
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest07_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Update Y-C Del Service OwnerRef Error", func() {
			By("SubnormalTest07_3: Reconciling the updated resource")
			// ------------------------------
			//  SubnormalTest07_3 DCIAppConnection Update Y-C Del Service OwnerRef Error
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_SubnormalTest07_3(writer, reconciler, mockClient)
		})
	})
})
