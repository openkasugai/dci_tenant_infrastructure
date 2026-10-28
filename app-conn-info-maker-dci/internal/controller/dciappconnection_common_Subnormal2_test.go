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
			fmt.Println("> dciappconnection_common_Subnormal2_test.go")
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
			fmt.Println("< dciappconnection_common_Subnormal2_test.go")
			fmt.Println("< End BeforeEach")
			fmt.Println("<")
		})

		BeforeEach(func() {
			writer.Reset()
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal2_test.go")
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
			fmt.Println("<<<<< dciappconnection_common_Subnormal2_test.go")
			fmt.Println("<<<<< End BeforeEach")
			fmt.Println("<<<<<")
		})

		AfterEach(func() {
			By("DCIAppConnection Common Subnormal Test End")
			fmt.Println(">>>>>")
			fmt.Println(">>>>> dciappconnection_common_Subnormal2_test.go")
			fmt.Println(">>>>> Start AfterEach")
			fmt.Println(">>>>>")
			fmt.Println("<<<<<")
			fmt.Println("<<<<< dciappconnection_common_Subnormal2_test.go")
			fmt.Println("<<<<< End AfterEach")
			fmt.Println("<<<<<")
		})

		AfterAll(func() {
			fmt.Println(">>")
			fmt.Println(">> dciappconnection_common_Subnormal2_test.go")
			fmt.Println(">> Start AfterAll")
			fmt.Println(">>")
			fmt.Println("<<")
			fmt.Println("<< dciappconnection_common_Subnormal2_test.go")
			fmt.Println("<< End AfterAll")
			fmt.Println("<<")
		})
		It("DCIAppConnection Delete Deployment A Labels None", func() {
			By("SubnormalTest21_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest21_2 Delete Deployment A Labels None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest21_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment A Labels Char None", func() {
			By("SubnormalTest22_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest22_1 Create Deployment A Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest22_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment A Labels Char None", func() {
			By("SubnormalTest22_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest22_2 Delete Deployment A Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest22_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment A Labels Name Unmatch", func() {
			By("SubnormalTest23_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest23_1 Create Deployment A Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest23_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment A Labels Name Unmatch", func() {
			By("SubnormalTest23_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest23_2 Delete Deployment A Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest23_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment Labels Namespace Unmatch", func() {
			By("SubnormalTest24_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest24_1 Create Deployment A Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest24_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment Labels Namespace Unmatch", func() {
			By("SubnormalTest24_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest24_2 Delete Deployment A Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest24_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment A Labels Namespace None", func() {
			By("SubnormalTest25_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest25_1 Create Deployment A Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest25_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment A Labels Namespace None", func() {
			By("SubnormalTest25_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest25_2 Delete Deployment A Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest25_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment A DeletionTimestamp Set", func() {
			By("SubnormalTest26_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest26_1 Create Deployment A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest26_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment A DeletionTimestamp Set", func() {
			By("SubnormalTest26_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest26_2 Delete Deployment A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest26_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Status.Conditions Not Ready", func() {
			By("SubnormalTest27_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest27_1 Create Pod A2 Status.Conditions Not Ready
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest27_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Status.Conditions Not Ready", func() {
			By("SubnormalTest27_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest27_2 Delete Pod A2 Status.Conditions Not Ready
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest27_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 2ndNIC Not Found", func() {
			By("SubnormalTest28_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest28_1 Create Pod A2 2ndNIC Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest28_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 2ndNIC Not Found", func() {
			By("SubnormalTest28_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest28_2 Delete Pod A2 2ndNIC Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest28_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Labels None", func() {
			By("SubnormalTest29_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest29_2 Delete Pod A2 Labels None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest29_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Labels Char None", func() {
			By("SubnormalTest30_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest30_1 Create Pod A2 Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest30_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Labels Char None", func() {
			By("SubnormalTest30_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest30_2 Delete Pod A2 Labels ""
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest30_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Labels Name Unmatch", func() {
			By("SubnormalTest31_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest31_1 Create Pod A2 Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest31_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Labels Name Unmatch", func() {
			By("SubnormalTest31_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest31_2 Delete Pod A2 Labels Name Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest31_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Labels Namespace Unmatch", func() {
			By("SubnormalTest32_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest32_1 Create Pod A2 Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest32_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Labels Namespace Unmatch", func() {
			By("SubnormalTest32_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest32_2 Delete Pod A2 Labels Namespace Unmatch
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest32_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 Labels Namespace None", func() {
			By("SubnormalTest33_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest33_1 Create Pod A2 Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest33_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 Labels Namespace None", func() {
			By("SubnormalTest33_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest33_2 Delete Pod A2 Labels "_" None
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest33_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Pod A2 DeletionTimestamp Set", func() {
			By("SubnormalTest34_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest34_1 Create Pod A2 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest34_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Pod A2 DeletionTimestamp Set", func() {
			By("SubnormalTest34_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest34_2 Delete Pod A2 DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest34_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Deployment A Not Found", func() {
			By("SubnormalTest35_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest35_1 Create Deployment A Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest35_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Deployment A Not Found", func() {
			By("SubnormalTest35_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest35_2 Delete Deployment A Not Found
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest35_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create Service A DeletionTimestamp Set", func() {
			By("SubnormalTest36_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest36_1 Create Service A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest36_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete Service A DeletionTimestamp Set", func() {
			By("SubnormalTest36_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest36_2 Delete Service A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest36_2(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Create EndpointSlice A DeletionTimestamp Set", func() {
			By("SubnormalTest37_1: Reconciling the create resource")
			// ------------------------------
			//  SubnormalTest37_1 Create EndpointSlice A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest37_1(writer, reconciler, mockClient)
		})
		It("DCIAppConnection Delete EndpointSlice A DeletionTimestamp Set", func() {
			By("SubnormalTest37_2: Reconciling the delete resource")
			// ------------------------------
			//  SubnormalTest37_2 Delete EndpointSlice A DeletionTimestamp Set
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_common_SubnormalTest37_2(writer, reconciler, mockClient)
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
	})
})
