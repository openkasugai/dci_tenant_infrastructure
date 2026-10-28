package controller

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"encoding/json"

	"k8s.io/klog/v2/textlogger"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ConnectEndpoint Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("SubnormalTest for ConnectEndpointReconciler", Ordered, func() {

		var reconciler ConnectEndpointReconciler

		BeforeAll(func() {
			// set up log
			log := textlogger.NewLogger(textlogger.NewConfig())

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			reconciler = ConnectEndpointReconciler{
				Client: k8sClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("ConnectEndpoint"),
			}

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

			// Create for SIGHUP process shell
			createShellfile()

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
			appStatusData03Failed.Raw, err = json.Marshal(appStatusData03mapFailed)
			if err != nil {
				fmt.Printf("appStatusData03mapFailed=%v\n", appStatusData03mapFailed)
				fmt.Println("appStatusData03mapFailed Marshal error.", err)
			}
			DCIAppStateCR101.Status.AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1Failed[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1Failed[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus1Failed[1].ToPod[1].AppStatus.Raw = appStatusData03Failed.Raw
			ConnectEndpointCR203AToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203AToStatus0[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203AToStatus0[1].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204AToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204AToStatus1[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204AToStatus1[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR205AToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR205AToStatus1[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR205AToStatus1[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR102.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR103.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR202.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR203.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR204.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR205.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR206.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR207.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR901.Status.AppStatus.Raw = appStatusData01.Raw
		})

		BeforeEach(func() {
			// loger initialized
			writer.Reset()

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

		})

		AfterEach(func() {
			By("ConnectEndpoint SubnormalTest End")
			writer.Reset()
		})

		AfterAll(func() {
			// Delete for SIGHUP process
			deleteProcess()

			// Delete for SIGHUP process shell
			deleteShellfile()
		})

		It("ConnectEndpoint Create SubnormalTest01", func() {

			By("SubnormalTest01: ConnectEndpointCR not found.")
			fmt.Println("ConnectEndpoint Create SubnormalTest01: ConnectEndpointCR not found.")

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())
		})

		It("ConnectEndpoint Create SubnormalTest02", func() {

			By("SubnormalTest02: ConnectEndpointCR unmatch CRName.")
			fmt.Println("ConnectEndpoint Create SubnormalTest02: ConnectEndpointCR unmatch CRName")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("MY_POD_NAME", "testname")
			if err != nil {
				fmt.Println("set env Error", err)
			}

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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

			Expect(connectEndpointCR.Finalizers).To(BeNil())
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(len(connectEndpointCR.Status.Condition)).To(Equal(0))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())
		})

		It("ConnectEndpoint Create SubnormalTest03", func() {

			By("SubnormalTest03: ConnectEndpointCR unmatch Namespace")
			fmt.Println("ConnectEndpoint Create SubnormalTest03: ConnectEndpointCR unmatch Namespace")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("MY_POD_NAMESPACE", "testnamespace")
			if err != nil {
				fmt.Println("set env Error", err)
			}

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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

			Expect(connectEndpointCR.Finalizers).To(BeNil())
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(len(connectEndpointCR.Status.Condition)).To(Equal(0))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())
		})

		It("ConnectEndpoint Update SubnormalTest04", func() {

			By("SubnormalTest04: Patterns with different number of Status.To Deployment lists")
			fmt.Println("ConnectEndpoint Update SubnormalTest04: Patterns with different number of Status.To Deployment lists")

			// Create ConnectEndpointCR
			pConnectEndpointCR := ConnectEndpointCR204A.DeepCopy()
			pConnectEndpointCR.Status.To = ConnectEndpointCR204ToStatus0
			err = createConnectEndpoint(ctx, pConnectEndpointCR)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR204AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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

			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile204A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest05", func() {

			By("SubnormalTest05: Patterns with different number of Status.To Deployment all lists")
			fmt.Println("ConnectEndpoint Update SubnormalTest05: Patterns with different number of Status.To Deployment all lists")

			// Create ConnectEndpointCR
			pConnectEndpointCR := ConnectEndpointCR204.DeepCopy()
			pConnectEndpointCR.Status.To = ConnectEndpointCR101ToStatus1
			err = createConnectEndpoint(ctx, pConnectEndpointCR)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR204ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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

			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile204))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest06", func() {

			By("SubnormalTest06: Patterns with different number of Status.InputEndpoint nil lists")
			fmt.Println("ConnectEndpoint Update SubnormalTest06: Patterns with different number of Status.InputEndpoint nil lists")

			// Create ConnectEndpointCR
			pConnectEndpointCR := ConnectEndpointCR209A.DeepCopy()
			pConnectEndpointCR.Spec.InputEndpoint = nil
			pConnectEndpointCR.Status.InputEndpoint = nil
			err = createConnectEndpoint(ctx, pConnectEndpointCR)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR901)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR203)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: false, RequeueAfter: 0}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.InputEndpoint).To(BeNil())
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint2))

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

			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR203 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-203",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR203)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-203", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR203.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR203.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile209ANoneInput))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest07", func() {

			By("SubnormalTest07: DCIAppState CR Deleting")
			fmt.Println("ConnectEndpoint Update SubnormalTest07: DCIAppState CR Deleting")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR204A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR204 := DCIAppStateCR204.DeepCopy()
			controllerutil.AddFinalizer(pdciAppStateCR204, examplecomv1.ConnectEndpointFinalizer)
			err = createDCIAppState(ctx, pdciAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())
			// Get DCIAppState CR
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				pdciAppStateCR204)
			Expect(err).NotTo(HaveOccurred())
			// Delete DCIAppState CR
			err = k8sClient.Delete(ctx, pdciAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in deleting DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR204AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)

			err = showDCIAppState(ctx, "dciappstate-204", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR204.Finalizers).To(Equal(cFinalizers))
			Expect(dciAppStateCR204.Labels).To(Equal(cLabels))
			Expect(dciAppStateCR204.DeletionTimestamp.IsZero()).NotTo(BeTrue())

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile204A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}

			controllerutil.RemoveFinalizer(&dciAppStateCR204, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to DCIAppState CR
			err = k8sClient.Update(ctx, &dciAppStateCR204)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to  DCIAppStateCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest08", func() {

			By("SubnormalTest08: DCIAppState CR unmatch Labels")
			fmt.Println("ConnectEndpoint Update SubnormalTest08: DCIAppState CR unmatch Labels")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR204A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR204 := DCIAppStateCR204.DeepCopy()
			pdciAppStateCR204.Labels = map[string]string{
				"dciappstate-101": "testtest",
			}
			err = createDCIAppState(ctx, pdciAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR204AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-204", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR204.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR204.Labels["dciappstate-101"]).To(Equal("testtest"))

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile204A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest09", func() {

			By("SubnormalTest09: DCIAppState CR AppStatus None")
			fmt.Println("ConnectEndpoint Update SubnormalTest09: DCIAppState CR AppStatus None")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR204A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR204 := DCIAppStateCR204.DeepCopy()
			pdciAppStateCR204.Status.AppStatus.Raw = nil
			err = createDCIAppState(ctx, pdciAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR204AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-204", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR204.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR204.Labels).To(Equal(cLabels))

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile204A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest10", func() {

			By("SubnormalTest10: DCIAppState CR lists delete pattern")
			fmt.Println("ConnectEndpoint Update SubnormalTest10: DCIAppState CR lists delete pattern")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR205A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR103 := DCIAppStateCR103.DeepCopy()
			pdciAppStateCR103.Labels = map[string]string{
				"dciappstate-101": "testtest",
			}
			err = createDCIAppState(ctx, pdciAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR205AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels["dciappstate-101"]).To(Equal("testtest"))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-204", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR204.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR204.Labels).To(Equal(cLabels))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile205A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update SubnormalTest11", func() {

			By("SubnormalTest10: DCIAppState CR lists delete pattern")
			fmt.Println("ConnectEndpoint Update SubnormalTest10: DCIAppState CR lists delete pattern")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR205A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR103 := DCIAppStateCR103.DeepCopy()
			pdciAppStateCR103.Labels = map[string]string{
				"dciappstate-aaa": "testtest",
			}
			err = createDCIAppState(ctx, pdciAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR205AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			// Check DCIAppState CR
			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(Equal(cLabels))

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels["dciappstate-101"]).To(Equal(""))

			// Get DCIAppState CR
			var dciAppStateCR204 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-204",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR204)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-204", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			// Check DCIAppState CR
			Expect(dciAppStateCR204.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR204.Labels).To(Equal(cLabels))

			// Check appConnectControlInfo.json
			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile205A))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create SubnormalTest12", func() {

			By("SubnormalTest12: Reconciling the create not found process name from the environment variabl")
			fmt.Println("ConnectEndpoint Create SubnormalTest12: Reconciling the create not found process name from the environment variabl")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR102)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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

			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())

			err = os.Setenv("MAIN_CONTAINER_PROCESS_NAME", "")
			if err != nil {
				fmt.Println("set env Error", err)
			}

			// Run Reconcile
			got, errReconcile = reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get ConnectEndpoint CR
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
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101AToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR102 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-102",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR102)

			// Check DCIAppState CR
			Expect(err).NotTo(HaveOccurred())
			Expect(dciAppStateCR102.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR102.Labels).To(Equal(cLabels))

			err = showDCIAppState(ctx, "dciappstate-102", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)

			// Check DCIAppState CR
			Expect(err).NotTo(HaveOccurred())
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

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
			Expect(configFileData).To(Equal(ExpectedConfigFile101A))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create SubnormalTest13", func() {

			By("SubnormalTest13: Reconciling the create process AppStatus reflect failed")
			fmt.Println("ConnectEndpoint Create SubnormalTest13: Reconciling the create process AppStatus reflect failed")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101A)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR102)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			pdciAppStateCR103 := DCIAppStateCR103.DeepCopy()
			pdciAppStateCR103.Status.AppStatus.Raw = appStatusData03Failed.Raw
			err = createDCIAppState(ctx, pdciAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

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

			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())

			// Run Reconcile
			got, errReconcile = reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get ConnectEndpoint CR
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101AToStatus1Failed))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR102 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-102",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR102)

			// Check DCIAppState CR
			Expect(err).NotTo(HaveOccurred())
			Expect(dciAppStateCR102.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR102.Labels).To(Equal(cLabels))

			err = showDCIAppState(ctx, "dciappstate-102", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)

			// Check DCIAppState CR
			Expect(err).NotTo(HaveOccurred())
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(Equal(cLabels))

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

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
			Expect(configFileData).To(Equal(ExpectedConfigFile101AFailed))

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Delete SubnormalTest14", func() {

			By("SubnormalTest14: Reconciling the delete process")
			fmt.Println("ConnectEndpoint Delete SubnormalTest14: Reconciling the delete process ")

			// Create ConnectEndpointCR
			pcreateConnectEndpoint := ConnectEndpointCR301A.DeepCopy()
			pcreateConnectEndpoint.Status.Condition = examplecomv1.NotConfigured
			err = createConnectEndpoint(ctx, pcreateConnectEndpoint)
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
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR203)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Get ConnectEndpoint CR
			var connectEndpointCR examplecomv1.ConnectEndpoint
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).NotTo(HaveOccurred())

			// Delete ConnectEndpoint CR
			err = k8sClient.Delete(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("Can not delete ConnectEndpointCR", err)
			}

			// for SIGHUP process start
			processStart()

			// Get SIGHUP receive pid
			checkPID := getTestProcess()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get ConnectEndpoint CR
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cConnectEndpointNamespace,
			},
				&connectEndpointCR)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			if err == nil {
				err = showDCIAppState(ctx, "dciappstate-202", cDCIAppStateNamespace)
				if err != nil {
					fmt.Println("error showDCIAppState.", err)
				}
				Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
				Expect(dciAppStateCR202.Labels).To(BeNil())
			}

			// Get DCIAppState CR
			var dciAppStateCR203 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-203",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR203)
			if err == nil {
				err = showDCIAppState(ctx, "dciappstate-203", cDCIAppStateNamespace)
				if err != nil {
					fmt.Println("error showDCIAppState.", err)
				}
				Expect(dciAppStateCR203.Finalizers).To(BeEmpty())
				Expect(dciAppStateCR203.Labels).To(BeNil())
			}

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				// Check appConnectControlInfo.json
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
				}

				expectedConfigFile := examplecomv1.AppConnectControlInfo{
					Address: "192.168.0.11",
				}
				var jsonByteData []byte
				var jsonData examplecomv1.AppConnectControlInfo
				jsonByteData, err = json.Marshal(expectedConfigFile)
				if err != nil {
					fmt.Println(err, "unable to marshal.")
				}

				err = json.Unmarshal(jsonByteData, &jsonData)
				if err != nil {
					fmt.Println(err, "unable to unmarshal.")
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
				Expect(configFileData).To(Equal(configFileData))
			}
		})
	})
})
