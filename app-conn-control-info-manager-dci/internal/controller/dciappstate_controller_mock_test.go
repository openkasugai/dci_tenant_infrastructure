package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"encoding/json"

	"github.com/golang/mock/gomock"
	"k8s.io/klog/v2/textlogger"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func DCIAppStateClientGet(k8sClient client.Client) func(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
		return k8sClient.Get(ctx, key, obj, opts...)
	}
}
func DCIAppStateClientPatch(k8sClient client.Client) func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
		return k8sClient.Patch(ctx, obj, patch, opts...)
	}
}

var _ = Describe("DCIAppState Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("Test for Mock DCIAppStateReconciler", Ordered, func() {

		var reconciler DCIAppStateReconciler

		var mockClient *MockClient

		// set up log
		log := textlogger.NewLogger(textlogger.NewConfig())

		BeforeAll(func() {

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			// Defined only one per controller
			// err = reconciler.SetupWithManager(mgr)
			// if err != nil {
			// 	fmt.Println("Error in SetupWithManager", err)
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

			err = os.Setenv("MY_POD_NAME", "dciappstate-101")
			if err != nil {
				fmt.Println("set env Error", err)
			}
			err = os.Setenv("MY_POD_NAMESPACE", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("set env Error", err)
			}
			err = os.Setenv("MAIN_CONTAINER_PROCESS_NAME", ProcessCmdEnv)
			if err != nil {
				fmt.Println("set env Error", err)
			}

			// Create for SIGHUP process shell
			createShellfile()
		})

		BeforeEach(func() {
			// loger initialized
			writer.Reset()

			gomockCtrl := gomock.NewController(GinkgoT())
			defer gomockCtrl.Finish()

			mockClient = NewMockClient(gomockCtrl)
			reconciler = DCIAppStateReconciler{
				Client: mockClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("DCIAppState"),
			}
			mockClient.EXPECT().
				Scheme().
				Return(testScheme).
				AnyTimes()

			err = k8sClient.DeleteAllOf(ctx, &examplecomv1.ConnectEndpoint{}, client.InNamespace(cConnectEndpointNamespace))
			if err != nil {
				fmt.Println("Can not delete ConnectEndpointCR", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.DeleteAllOf(ctx, &examplecomv1.DCIAppState{}, client.InNamespace(cDCIAppStateNamespace))
			if err != nil {
				fmt.Println("Can not delete DCIAppStateCR", err)
			}
			Expect(err).NotTo(HaveOccurred())

			filepath := examplecomv1.ExclCtrlComFilePath + "dciappstate-101" + "_" +
				cDCIAppStateNamespace + examplecomv1.ExclCtrlFileExtension
			_ = syscall.Unlink(filepath)

			// Delete for SIGHUP process
			deleteProcess()

			time.Sleep(100 * time.Millisecond)

			appStatusData01.Raw, err = json.Marshal(appStatusData01map)
			if err != nil {
				fmt.Printf("appStatusData01map=%v\n", appStatusData01map)
				fmt.Println("appStatusData01map Marshal error.", err)
			}
			appStatusData02.Raw, err = json.Marshal(appStatusData02map)
			if err != nil {
				fmt.Printf("appStatusData02map=%v\n", appStatusData02map)
				fmt.Println("appStatusData02map Marshal error.", err)
			}
			appStatusData03.Raw, err = json.Marshal(appStatusData03map)
			if err != nil {
				fmt.Printf("appStatusData03map=%v\n", appStatusData03map)
				fmt.Println("appStatusData03map Marshal error.", err)
			}
			DCIAppStateCR202.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR203.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR402.Status.AppStatus.Raw = appStatusData03.Raw
			DCIAppStateCR901.Status.AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401A.Status.To[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401A.Status.To[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401A.Status.To[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
		})

		AfterEach(func() {
			By("DCIAppState Test End")
			writer.Reset()
		})

		AfterAll(func() {
			// Delete for SIGHUP process
			deleteProcess()

			// Delete for SIGHUP process shell
			deleteShellfile()
		})

		It("DCIAppState MockErrorTest01", func() {

			By("MockErrorTest01: Reconciling the update process Client.Get() error(DCIAppStateCR)")
			fmt.Println("DCIAppState Update MockErrorTest01: Reconciling the update process Client.Get() error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR401A)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR 901
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			DCIAppStateCR202T01 := DCIAppStateCR202.DeepCopy()
			DCIAppStateCR202T01.Status.AppStatus = appStatusData02
			err = createDCIAppState(ctx, DCIAppStateCR202T01)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-202", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				Return(errors.New("DCIAppStateCR Get Error Test(not NotFound)"))

			// Run Reconcile
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				},
			}
			got, errReconcile := reconciler.Reconcile(ctx, req)

			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-202", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get ConnectEndpoint CR
			var connectEndpointCR examplecomv1.ConnectEndpoint
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.Configured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207AToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				// Check appConnectControlInfo.json
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
				}

				var configFileBytes []byte
				var configFileData examplecomv1.AppConnectControlInfo
				configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
				if err != nil {
					fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
				}
				if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
					fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
				}
				Expect(configFileData).To(Equal(ExpectedConfigFile401A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState MockErrorTest02", func() {

			By("MockErrorTest02: Reconciling the update process Client.Get() error(ConnectEndpointCR)")
			fmt.Println("DCIAppState Update MockErrorTest02: Reconciling the update process Client.Get() error(ConnectEndpointCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR401A)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR 901
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			DCIAppStateCR202T01 := DCIAppStateCR202.DeepCopy()
			DCIAppStateCR202T01.Status.AppStatus = appStatusData02
			err = createDCIAppState(ctx, DCIAppStateCR202T01)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-101", Namespace: cConnectEndpointNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
				Return(errors.New("ConnectEndpointCR Get Error Test(not NotFound)"))

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(DCIAppStateClientGet(k8sClient)).AnyTimes()

			// Run Reconcile
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				},
			}
			got, errReconcile := reconciler.Reconcile(ctx, req)

			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-202", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get ConnectEndpoint CR
			var connectEndpointCR examplecomv1.ConnectEndpoint
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.Configured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207AToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				// Check appConnectControlInfo.json
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
				}

				var configFileBytes []byte
				var configFileData examplecomv1.AppConnectControlInfo
				configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
				if err != nil {
					fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
				}
				if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
					fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
				}
				Expect(configFileData).To(Equal(ExpectedConfigFile401A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState MockErrorTest03", func() {

			By("MockErrorTest03: Reconciling the update process Client.Patch() error(ConnectEndpointCR)")
			fmt.Println("DCIAppState Update MockErrorTest02: Reconciling the update process Client.Patch() error(ConnectEndpointCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR401A)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR 901
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			DCIAppStateCR202T01 := DCIAppStateCR202.DeepCopy()
			DCIAppStateCR202T01.Status.AppStatus = appStatusData02
			err = createDCIAppState(ctx, DCIAppStateCR202T01)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dciappstate-101",
							Namespace: cConnectEndpointNamespace,
						},
					}),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("ConnectEndpointCR Patch Error Test"))

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(DCIAppStateClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(DCIAppStateClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				},
			}
			got, errReconcile := reconciler.Reconcile(ctx, req)

			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-202", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get ConnectEndpoint CR
			var connectEndpointCR examplecomv1.ConnectEndpoint
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.Configured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207AToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				// Check appConnectControlInfo.json
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
				}

				var configFileBytes []byte
				var configFileData examplecomv1.AppConnectControlInfo
				configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
				if err != nil {
					fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
				}
				if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
					fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
				}
				Expect(configFileData).To(Equal(ExpectedConfigFile401A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState MockErrorTest04", func() {

			By("MockErrorTest04: Reconciling the not action process error(DCIAppStateCR)")
			fmt.Println("DCIAppState Update MockErrorTest04: Reconciling the not action process error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR401A)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR 901
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			DCIAppStateCR202T01 := DCIAppStateCR202.DeepCopy()
			DCIAppStateCR202T01.Status.AppStatus = appStatusData02
			err = createDCIAppState(ctx, DCIAppStateCR202T01)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-202", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					err := DCIAppStateClientGet(k8sClient)(ctx, key, obj, opts...)
					now := metav1.NewTime(metav1.Now().Time)
					stDCIAppState, ok := obj.(*examplecomv1.DCIAppState)
					if ok {
						stDCIAppState.ObjectMeta.DeletionTimestamp = &now
					}
					return err
				}).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(DCIAppStateClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(DCIAppStateClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				},
			}
			got, errReconcile := reconciler.Reconcile(ctx, req)

			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-202", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get ConnectEndpoint CR
			var connectEndpointCR examplecomv1.ConnectEndpoint
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).NotTo(HaveOccurred())

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.Configured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207AToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				// Check appConnectControlInfo.json
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
				}

				var configFileBytes []byte
				var configFileData examplecomv1.AppConnectControlInfo
				configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
				if err != nil {
					fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
				}
				if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
					fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
				}
				Expect(configFileData).To(Equal(ExpectedConfigFile401A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})
	})
})
