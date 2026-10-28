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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"

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

	Context("Test for ConnectEndpointReconciler", Ordered, func() {

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

			err = reconciler.SetupWithManager(mgr)
			if err != nil {
				fmt.Println("Error in SetupWithManager", err)
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
			DCIAppStateCR101.Status.AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101ToStatus2[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101ToStatus3[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus2[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus2[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus3[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus3[0].ToPod[1].AppStatus.Raw = appStatusData03.Raw
			ConnectEndpointCR103ToStatus4[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR103ToStatus4[0].ToPod[1].AppStatus.Raw = appStatusData02.Raw
			ConnectEndpointCR109ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR120ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201ToStatus1[0].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203ToStatus0[0].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR203ToStatus1[0].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR204ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR205ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR205ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR206ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR206ToStatus0[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207ToStatus0[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR102.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR103.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR202.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR203.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR204.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR205.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR206.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR207.Status.AppStatus.Raw = appStatusData01.Raw
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

		})

		AfterEach(func() {
			By("ConnectEndpoint Test End")
			writer.Reset()
		})

		AfterAll(func() {
			// Delete for SIGHUP process
			deleteProcess()

			// Delete for SIGHUP process shell
			deleteShellfile()
		})

		It("ConnectEndpoint Exclusive Test01", func() {

			By("Test01: Reconciling the exclusive process")
			fmt.Println("ConnectEndpoint Exclusive Test01: Reconciling the exclusive process")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create exclusive file
			filepath := examplecomv1.ExclCtrlComFilePath + "dciappstate-101" + "_" +
				cDCIAppStateNamespace + examplecomv1.ExclCtrlFileExtension
			fd, fdErr := syscall.Open(filepath, syscall.O_CREAT|syscall.O_RDWR, 0644)
			if fdErr != nil {
				fmt.Println("Requeue for exclusive control ", err)
			}
			fmt.Println("exclusive file: ", filepath)

			defer func() {
				closeErr := syscall.Close(fd)
				if closeErr != nil {
					fmt.Println("close err ", closeErr)
				}

				unlinkErr := syscall.Unlink(filepath)
				if unlinkErr != nil {
					fmt.Println("unlink err ", unlinkErr)
				}
			}()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 100000000}))
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

			Expect(connectEndpointCR.Spec).To(Equal(ConnectEndpointCR101.Spec))
			Expect(connectEndpointCR.Status).To(Equal(ConnectEndpointCR101.Status))
			Expect(connectEndpointCR.Labels).To(Equal(ConnectEndpointCR101.Labels))
			Expect(connectEndpointCR.Finalizers).To(Equal(ConnectEndpointCR101.Finalizers))
		})

		It("ConnectEndpoint Create Test01", func() {

			By("Test01: Reconciling the create process")
			fmt.Println("ConnectEndpoint Create Test01: Reconciling the create process")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile101))

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test02", func() {

			By("Test02: DCIAppState CR not found")
			fmt.Println("ConnectEndpoint Create Test02: DCIAppState CR not found")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus2))
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

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
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

			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile101))

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

		It("ConnectEndpoint Create Test03", func() {

			By("Test03: Pattern with podref difference")
			fmt.Println("ConnectEndpoint Create Test03: Pattern with podref difference")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Status.To[0].ToPod[1].PodRef.Name = createtest03
			connectEndpointCR103Test.Status.To[0].ToPod[1].PodRef.Namespace = createtest03
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test04", func() {

			By("Test04: Pattern with address difference")
			fmt.Println("ConnectEndpoint Create Test04: Pattern with address difference")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Spec.To[0].ToPod[1].Address = "192.168.0.14"
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus2))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR102.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR102.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
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

		It("ConnectEndpoint Create Test05", func() {

			By("Test05: Patterns with different Status.To.AppStatus key numbers")
			fmt.Println("ConnectEndpoint Create Test05: Patterns with different Status.To.AppStatus key numbers")

			// Create ConnectEndpointCR
			connectEndpointCR103CTest := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103CTest.Status.To[0].ToPod[1].AppStatus = appStatusData03
			err = createConnectEndpoint(ctx, connectEndpointCR103CTest)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus3))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus3))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test06", func() {

			By("Test06: Patterns with different Status.To.AppStatus key value")
			fmt.Println("ConnectEndpoint Create Test06: Patterns with different Status.To.AppStatus key value")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Status.To[0].ToPod[1].AppStatus = appStatusData02
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus4))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR103ToStatus4))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test07", func() {

			By("Test07: InputEndpoint none")
			fmt.Println("ConnectEndpoint Create Test07: InputEndpoint none")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR107)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect((connectEndpointCR.Status).InputEndpoint).To(BeNil())
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
			Expect(configFileData).To(Equal(ExpectedConfigFile107))

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

		It("ConnectEndpoint Create Test08", func() {

			By("Test08: OutputEndpoint none")
			fmt.Println("ConnectEndpoint Create Test08: OutputEndpoint none")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR108)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect((connectEndpointCR.Status).OutputEndpoint).To(BeNil())

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
			Expect(configFileData).To(Equal(ExpectedConfigFile108))

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

		It("ConnectEndpoint Create Test09", func() {

			By("Test09: Unupdated Processing")
			fmt.Println("ConnectEndpoint Create Test09: Unupdated Processing")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR109)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			err = os.Remove(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to remove. FileName = " + examplecomv1.ConfigFilePath)
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
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR109ToStatus0))

			// Run Reconcile
			got, errReconcile = reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus2))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Check appConnectControlInfo.json
			_, err = os.Stat(examplecomv1.ConfigFilePath)
			Expect(false).To(Equal(!os.IsNotExist(err)))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test10", func() {

			By("Test10: DCIAppState CR not found except last")
			fmt.Println("ConnectEndpoint Create Test10: DCIAppState CR not found except last")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
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

			// Check ConnectEndpoint CR
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
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus3))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR102)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Run Reconcile
			got, err = reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(err).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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

			var configFileBytes []byte
			var configFileData examplecomv1.AppConnectControlInfo
			configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
			}
			if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
				fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
			}
			Expect(configFileData).To(Equal(ExpectedConfigFile101))

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

		It("ConnectEndpoint Create Test11", func() {

			By("Test11: Patterns with different Status.To.InputEndpoint list numbers")
			fmt.Println("ConnectEndpoint Create Test11: Patterns with different Status.To.InputEndpoint list numbers")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Spec.InputEndpoint = &externalInputEndpoint2
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint2))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test12", func() {

			By("Test12: Patterns with different Status.To.OutputEndpoint list numbers")
			fmt.Println("ConnectEndpoint Create Test12: Patterns with different Status.To.OutputEndpoint list numbers")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Spec.OutputEndpoint = &externalOutputEndpoint2
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint2))

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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test13", func() {

			By("Test13: Patterns with different Status.To.InputEndpoint value")
			fmt.Println("ConnectEndpoint Create Test13: Patterns with different Status.To.InputEndpoint value")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Spec.InputEndpoint = &externalInputEndpoint3
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint3))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test14", func() {

			By("Test14: Patterns with different Status.To.OutputEndpoint value")
			fmt.Println("ConnectEndpoint Create Test14: Patterns with different Status.To.OutputEndpoint value")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Spec.OutputEndpoint = &externalOutputEndpoint3
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint3))

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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test15", func() {

			By("Test15: Patterns in which Status.To.InputEndpoint is added from nil")
			fmt.Println("ConnectEndpoint Create Test15: Patterns in which Status.To.InputEndpoint is added from nil")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Status.InputEndpoint = nil
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test16", func() {

			By("Test16: Patterns in which Status.To.OutputEndpoint is added from nil")
			fmt.Println("ConnectEndpoint Create Test16: Patterns in which Status.To.OutputEndpoint is added from nil")

			// Create ConnectEndpointCR
			connectEndpointCR103Test := ConnectEndpointCR103.DeepCopy()
			connectEndpointCR103Test.Status.OutputEndpoint = nil
			err = createConnectEndpoint(ctx, connectEndpointCR103Test)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check ConnectEndpoint CR
			Expect(connectEndpointCR.Finalizers).To(Equal(cFinalizers))
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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
			Expect(dciAppStateCR102.Labels).To(BeNil())

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
			Expect(dciAppStateCR103.Labels).To(BeNil())

			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test17", func() {

			By("Test17: Signal send error")
			fmt.Println("ConnectEndpoint Create Test17: Signal send error")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(got).To(Equal(ctrl.Result{Requeue: false, RequeueAfter: 0}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create Test18", func() {

			By("Test18: Check labels overwrite")
			fmt.Println("ConnectEndpoint Create Test18: Check labels overwrite")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR101)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			dciAppStateCR102Test := DCIAppStateCR102.DeepCopy()
			dciAppStateCR102Test.Labels = make(map[string]string)
			dciAppStateCR102Test.Labels["testname"] = "testnamespace"
			err = createDCIAppState(ctx, dciAppStateCR102Test)
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

			// Check ConnectEndpoint CR
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
			Expect(got).To(Equal(ctrl.Result{Requeue: false, RequeueAfter: 0}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR101ToStatus1))
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
			var cLabelsTest = map[string]string{
				"testname":        "testnamespace",
				"dciappstate-101": cConnectEndpointNamespace}
			Expect(err).NotTo(HaveOccurred())
			Expect(dciAppStateCR102.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR102.Labels).To(Equal(cLabelsTest))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile101))

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

		It("ConnectEndpoint Create Test19", func() {

			By("Test19: No Pod list")
			fmt.Println("ConnectEndpoint Create Test19: No Pod list")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR119)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			err = os.Remove(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to remove. FileName = " + examplecomv1.ConfigFilePath)
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
			Expect(got).To(Equal(ctrl.Result{Requeue: false, RequeueAfter: 0}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR119ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile119))

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

		It("ConnectEndpoint Create Test20", func() {

			By("Test20: DCIAppStateCR.Status is nil")
			fmt.Println("ConnectEndpoint Create Test20: DCIAppStateCR.Status is nil")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR120)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR120)
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

			err = os.Remove(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to remove. FileName = " + examplecomv1.ConfigFilePath)
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
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR120ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR120 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-120",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR120)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-120", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check DCIAppState CR
			Expect(dciAppStateCR120.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR120.Labels).To(Equal(cLabels))

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

			// Check appConnectControlInfo.json
			_, err = os.Stat(examplecomv1.ConfigFilePath)
			Expect(false).To(Equal(!os.IsNotExist(err)))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update Test01", func() {

			By("Test01: Reconciling the update process")
			fmt.Println("ConnectEndpoint Update Test01: Reconciling the update process")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

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

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR204)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR205)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR206)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201ToStatus1))
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
			Expect(dciAppStateCR202.Labels).To(BeNil())

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
			Expect(dciAppStateCR203.Labels).To(BeNil())

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

			// Check DCIAppState CR
			var dciAppStateCR205 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-205",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR205)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-205", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check DCIAppState CR
			Expect(dciAppStateCR205.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR205.Labels).To(Equal(cLabels))

			// Check DCIAppState CR
			var dciAppStateCR206 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-206",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR206)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-206", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check DCIAppState CR
			Expect(dciAppStateCR206.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR206.Labels).To(Equal(cLabels))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile201))

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

		It("ConnectEndpoint Update Test02", func() {

			By("Test02: Unupdated Processing")
			fmt.Println("ConnectEndpoint Update Test02: Unupdated Processing")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR301)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			err = os.Remove(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to remove. FileName = " + examplecomv1.ConfigFilePath)
			}

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207ToStatus0))
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

			// Check appConnectControlInfo.json
			_, err = os.Stat(examplecomv1.ConfigFilePath)
			Expect(false).To(Equal(!os.IsNotExist(err)))

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			// Update RemoveFinalizer to connectEndpoint CR
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update Test03", func() {

			By("Test03: DCIAppState CR not found")
			fmt.Println("ConnectEndpoint Update Test03: DCIAppState CR not found")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR203)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR203ToStatus0))
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
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

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
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR103)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Run Reconcile
			got, err = reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(err).NotTo(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR203ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
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
			Expect(configFileData).To(Equal(ExpectedConfigFile203))

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

		It("ConnectEndpoint Update Test04", func() {

			By("Test04: Patterns with different number of Status.To lists")
			fmt.Println("ConnectEndpoint Update Test04: Patterns with different number of Status.To lists")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR204)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

		It("ConnectEndpoint Update Test05", func() {

			By("Test05: DCIAppState CR not found delete pattern")
			fmt.Println("ConnectEndpoint Update Test05: DCIAppState CR not found delete pattern")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR205)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR205ToStatus1))
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
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile205))

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

		It("ConnectEndpoint Update Test06", func() {

			By("Test06: No Pod list")
			fmt.Println("ConnectEndpoint Update Test06: No Pod list")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR206)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			err = os.Remove(examplecomv1.ConfigFilePath)
			if err != nil {
				fmt.Println("unable to remove. FileName = " + examplecomv1.ConfigFilePath)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR206ToStatus1))
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
			Expect(dciAppStateCR202.Labels).To(BeNil())

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
			Expect(dciAppStateCR204.Labels).To(BeNil())

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
			Expect(configFileData).To(Equal(ExpectedConfigFile206))

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

		It("ConnectEndpoint Update Test07", func() {

			By("Test07: Patterns with different From Param")
			fmt.Println("ConnectEndpoint Update Test07: Patterns with different From Param")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR207)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR207From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207ToStatus0))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile207))

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

		It("ConnectEndpoint Update Test08", func() {

			By("Test08: Patterns with different InputEndpoint Param")
			fmt.Println("ConnectEndpoint Update Test08: Patterns with different InputEndpoint Param")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR208)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207ToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint2))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile208))

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

		It("ConnectEndpoint Update Test09", func() {

			By("Test09: Patterns with different OutputEndpoint Param")
			fmt.Println("ConnectEndpoint Update Test09: Patterns with different OutputEndpoint Param")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR209)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207ToStatus0))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile209))

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

		It("ConnectEndpoint Delete Test01", func() {

			By("Test01: Reconciling the delete process")
			fmt.Println("ConnectEndpoint Delete Test01: Reconciling the delete process")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR301)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile101Del))

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

			Expect(dciAppStateCR202.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR202.Labels).To(BeNil())

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

			Expect(dciAppStateCR203.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR203.Labels).To(BeNil())

		})

		It("ConnectEndpoint Delete Test02", func() {

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR301)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR203 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-203",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR203)
			Expect(err).NotTo(HaveOccurred())

			// Delete ConnectEndpoint CR
			err = k8sClient.Delete(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("Can not delete ConnectEndpointCR", err)
			}

			// Delete DCIAppState CR
			err = k8sClient.Delete(ctx, &dciAppStateCR202)
			if err != nil {
				fmt.Println("Can not delete DCIAppStateCR", err)
			}

			// Delete DCIAppState CR
			err = k8sClient.Delete(ctx, &dciAppStateCR203)
			if err != nil {
				fmt.Println("Can not delete DCIAppStateCR", err)
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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile101Del))

			// Get DCIAppState CR
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))

			// Get DCIAppState CR
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-203",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR203)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))
		})
	})

	Describe("ConnectEndpointEventPredicate Update Function", func() {
		var predicate ConnectEndpointEventPredicate

		BeforeEach(func() {
			predicate = ConnectEndpointEventPredicate{}
		})

		It("should return true when ObjectOld is nil", func() {
			event := event.UpdateEvent{
				ObjectOld: nil,
				ObjectNew: &examplecomv1.ConnectEndpoint{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectNew is nil", func() {
			event := event.UpdateEvent{
				ObjectOld: &examplecomv1.ConnectEndpoint{},
				ObjectNew: nil,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectOld is not ConnectEndpoint type", func() {
			event := event.UpdateEvent{
				ObjectOld: &corev1.Pod{},
				ObjectNew: &examplecomv1.ConnectEndpoint{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectNew is not ConnectEndpoint type", func() {
			event := event.UpdateEvent{
				ObjectOld: &examplecomv1.ConnectEndpoint{},
				ObjectNew: &corev1.Pod{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when DeletionTimestamp is not zero", func() {
			now := metav1.NewTime(time.Now())
			oldObj := &examplecomv1.ConnectEndpoint{}
			newObj := &examplecomv1.ConnectEndpoint{}
			newObj.DeletionTimestamp = &now

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when Status.Condition has changed", func() {
			oldObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.NotConfigured,
				},
			}
			newObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.Configured,
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return false when conditions are met for no processing", func() {
			oldObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.Configured,
				},
			}
			newObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.Configured,
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeFalse())
		})

		It("should return false when Status.Condition is same and DeletionTimestamp is zero", func() {
			oldObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.NotConfigured,
				},
			}
			newObj := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.NotConfigured,
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeFalse())
		})
	})

	Describe("ConnectEndpointEventPredicate Create Function", func() {
		var predicate ConnectEndpointEventPredicate

		BeforeEach(func() {
			predicate = ConnectEndpointEventPredicate{}
		})

		It("should return true when Object is nil", func() {
			event := event.CreateEvent{
				Object: nil,
			}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when Object is ConnectEndpoint", func() {
			connectEndpoint := &examplecomv1.ConnectEndpoint{}
			event := event.CreateEvent{
				Object: connectEndpoint,
			}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when Object is not ConnectEndpoint type", func() {
			pod := &corev1.Pod{}
			event := event.CreateEvent{
				Object: pod,
			}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when Object has DeletionTimestamp set", func() {
			now := metav1.NewTime(time.Now())
			connectEndpoint := &examplecomv1.ConnectEndpoint{}
			connectEndpoint.DeletionTimestamp = &now
			event := event.CreateEvent{
				Object: connectEndpoint,
			}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when Object has specific Status.Condition", func() {
			connectEndpoint := &examplecomv1.ConnectEndpoint{
				Status: examplecomv1.ConnectEndpointStatus{
					Condition: examplecomv1.Configured,
				},
			}
			event := event.CreateEvent{
				Object: connectEndpoint,
			}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})

		It("should return true for any CreateEvent", func() {
			// Test with empty event
			event := event.CreateEvent{}
			result := predicate.Create(event)
			Expect(result).To(BeTrue())
		})
	})
})
