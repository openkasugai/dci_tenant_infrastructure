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
		It("DCIAppConnection Create", func() {
			By("Test26_1: Reconciling the created resource")
			// ------------------------------
			//  Test26_1 Create
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X(Merge),Out-Y(Merge)
			// ------------------------------
			dciappconnection_controller_Test26_1(writer, reconciler)
		})
		It("DCIAppConnection Delete", func() {
			By("Test26_2: Reconciling the delete resource")
			// ------------------------------
			//  Test26_2 Delete
			//  X-A,Y-C X-Y
			//  Y-None
			//  In-X(Merge),Out-Y(Merge)
			// ------------------------------
			dciappconnection_controller_Test26_2(writer, reconciler)
		})
		It("DCIAppConnection Create", func() {
			By("Test27_1: Reconciling the created resource")
			// ------------------------------
			//  Test27_1 Update In/Out 1st Add Merge
			//  X-None
			//  Y-None
			//  In-X(Merge),Out-Y(Merge)
			// ------------------------------
			dciappconnection_controller_Test27_1(writer, reconciler)
		})
		It("DCIAppConnection Update Y-C Add", func() {
			By("Test31_1: Reconciling the updated resource")
			// ------------------------------
			//  Test31_1 Update Y-C Add
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test31_1(writer, reconciler)
		})
		It("DCIAppConnection Update Y-C Delete", func() {
			By("Test31_2: Reconciling the updated resource")
			// ------------------------------
			//  Test31_2 Update Y-C Delete
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test31_2(writer, reconciler)
		})
		It("DCIAppConnection Update X-Y Add", func() {
			By("Test32_1: Reconciling the updated resource")
			// ------------------------------
			//  Test32_1 Update X-Y Add
			//  X-None
			//  Y-C
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_Test32_1(writer, reconciler)
		})
		It("DCIAppConnection Update X-Y Delete", func() {
			By("Test32_2: Reconciling the updated resource")
			// ------------------------------
			//  Test32_2 Update X-Y Delete
			//  X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test32_2(writer, reconciler)
		})
		It("DCIAppConnection Update X-C Add", func() {
			By("Test33_1: Reconciling the updated resource")
			// ------------------------------
			//  Test33_1 Update X-C Add
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test33_1(writer, reconciler)
		})
		It("DCIAppConnection Update X-C Delete", func() {
			By("Test33_2: Reconciling the updated resource")
			// ------------------------------
			//  Test33_2 Update X-C Delete
			//  X-A X-C X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test33_2(writer, reconciler)
		})
		It("DCIAppConnection Update In-Y Add", func() {
			By("Test34_1: Reconciling the updated resource")
			// ------------------------------
			//  Test34_1 Update In-Y Add
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test34_1(writer, reconciler)
		})
		It("DCIAppConnection Update In-Y Delete", func() {
			By("Test34_2: Reconciling the updated resource")
			// ------------------------------
			//  Test34_2 Update In-Y Delete
			//  X-A X-Y
			//  Y-C
			//  In-X In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_Test34_2(writer, reconciler)
		})
		It("DCIAppConnection Update Out-X Add", func() {
			By("Test35_1: Reconciling the updated resource")
			// ------------------------------
			//  Test35_1 Update Out-X Add
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test35_1(writer, reconciler)
		})
		It("DCIAppConnection Update Out-X Delete", func() {
			By("Test35_2: Reconciling the updated resource")
			// ------------------------------
			//  Test35_2 Update Out-X Delete
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-X Out-Y
			// ------------------------------
			dciappconnection_controller_Test35_2(writer, reconciler)
		})
		It("DCIAppConnection Create All-None", func() {
			By("Test36_1: Reconciling the created resource")
			// ------------------------------
			//  Test36_1 Create All-None
			//  X-None
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			dciappconnection_controller_Test36_1(writer, reconciler)
		})
		It("DCIAppConnection Delete All-None", func() {
			By("Test36_2: Reconciling the deleted resource")
			// ------------------------------
			//  Test36_2 Delete All-None
			//  X-None
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			dciappconnection_controller_Test36_2(writer, reconciler)
		})
		It("DCIAppConnection Update NotChange Created", func() {
			By("Test37_1: Reconciling the updated resource")
			// ------------------------------
			//  Test37_1 Update NotChange Created
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y Created
			// ------------------------------
			dciappconnection_controller_Test37_1(writer, reconciler)
		})
		It("DCIAppConnection Update NotChange Pending", func() {
			By("Test37_2: Reconciling the updated resource")
			// ------------------------------
			//  Test37_2 Update NotChange Pending
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y Pending
			// ------------------------------
			dciappconnection_controller_Test37_2(writer, reconciler)
		})
		It("DCIAppConnection Update From None In-X Out-Y Add", func() {
			By("Test38_1: Reconciling the updated resource")
			// ------------------------------
			//  Test38_1 Update From None In-X Out-Y Add
			//  X-None
			//  Y-None
			//  In-None,Out-None
			// ------------------------------
			dciappconnection_controller_Test38_1(writer, reconciler)
		})
		It("DCIAppConnection Update From None In-X Out-Y Delete", func() {
			By("Test38_2: Reconciling the updated resource")
			// ------------------------------
			//  Test38_2 Update From None In-X Out-Y Delete
			//  X-None
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test38_2(writer, reconciler)
		})
		It("DCIAppConnection Update From None Out-Y Change", func() {
			By("Test39_1: Reconciling the updated resource")
			// ------------------------------
			//  Test39_1 Update From None Out-Y Change
			//  X-None
			//  Y-None
			//  In-X,Out-X
			// ------------------------------
			dciappconnection_controller_Test39_1(writer, reconciler)
		})
		It("DCIAppConnection Update From None Out-X URL Change", func() {
			By("Test39_2: Reconciling the updated resource")
			// ------------------------------
			//  Test39_2 Update From None Out-Y URL Change
			//  X-None
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test39_2(writer, reconciler)
		})
		It("DCIAppConnection Update From None In-Y Change", func() {
			By("Test40_1: Reconciling the updated resource")
			// ------------------------------
			//  Test40_1 Update From None In-Y Change
			//  X-None
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test40_1(writer, reconciler)
		})
		It("DCIAppConnection Update From None In-X URL Change", func() {
			By("Test40_2: Reconciling the updated resource")
			// ------------------------------
			//  Test40_2 Update From None In-X URL Change
			//  X-None
			//  Y-None
			//  In-Y,Out-Y
			// ------------------------------
			dciappconnection_controller_Test40_2(writer, reconciler)
		})
		It("DCIAppConnection Update From X-A->Y-A Change", func() {
			By("Test41_1: Reconciling the updated resource")
			// ------------------------------
			//  Test41_1 Update From X-A->Y-A Change
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test41_1(writer, reconciler)
		})
		It("DCIAppConnection Update To X-A->X-C Change", func() {
			By("Test41_2: Reconciling the updated resource")
			// ------------------------------
			//  Test41_2 Update To X-A->X-C Change
			//  X-A X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test41_2(writer, reconciler)
		})
		It("DCIAppConnection Update To X-C Delete", func() {
			By("Test41_3: Reconciling the updated resource")
			// ------------------------------
			//  Test41_3 Update To X-C Y-C Delete
			//  X-C X-Y
			//  Y-C
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test41_3(writer, reconciler)
		})
		It("DCIAppConnection Update", func() {
			By("Test42_1: Reconciling the updated resource")
			// ------------------------------
			//  Test42_1 Update Status In/Out 2nd Add Merge ExternalSubsets
			//  X-None
			//  Y-None
			//  In-X(Merge),Out-Y(Merge)
			// ------------------------------
			dciappconnection_controller_Test42_1(writer, reconciler)
		})
		It("DCIAppConnection Update", func() {
			By("Test42_2: Reconciling the updated resource")
			// ------------------------------
			//  Test42_2 Update Status In/Out 1st Add Merge ExternalSubsets
			//  X-None
			//  Y-None
			//  In-X(Merge),Out-Y(Merge)
			// ------------------------------
			dciappconnection_controller_Test42_2(writer, reconciler)
		})
		It("DCIAppConnection Update", func() {
			By("Test43_1: Reconciling the updeted resource")
			// ------------------------------
			//  Test43_1 Update From Y Out-Y Delete
			//  X-A X-Y
			//  Y-None
			//  In-X,Out-Y
			// ------------------------------
			dciappconnection_controller_Test43_1(writer, reconciler)
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
