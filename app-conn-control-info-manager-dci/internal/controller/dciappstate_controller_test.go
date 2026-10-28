package controller

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"syscall"

	"encoding/json"

	"k8s.io/klog/v2/textlogger"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("DCIAppState Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("Test for DCIAppStateReconciler", Ordered, func() {

		var reconciler DCIAppStateReconciler

		BeforeAll(func() {
			// set up log
			log := textlogger.NewLogger(textlogger.NewConfig())

			// set manager
			mgr, err = getMgr(mgr)
			Expect(err).NotTo(HaveOccurred())

			reconciler = DCIAppStateReconciler{
				Client: k8sClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("DCIAppState"),
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
			DCIAppStateCR202.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR203.Status.AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR402.Status.AppStatus.Raw = appStatusData03.Raw
			ConnectEndpointCR401.Status.To[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401.Status.To[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR402.Status.To[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207ToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207ToStatus0[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401ToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData02.Raw
			ConnectEndpointCR401ToStatus1[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401ToStatus2[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401ToStatus3[0].ToPod[0].AppStatus.Raw = appStatusData03.Raw
			ConnectEndpointCR401ToStatus3[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
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

		It("DCIAppState Exclusive Test01", func() {

			By("Test01: Reconciling the exclusive process")
			fmt.Println("DCIAppState Exclusive Test01: Reconciling the exclusive process")

			Expect(err).NotTo(HaveOccurred())
			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
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
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 100000000}))
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

			// Check DCIAppState CR
			Expect(dciAppStateCR202.Spec).To(Equal(DCIAppStateCR202.Spec))
			Expect(dciAppStateCR202.Status).To(Equal(DCIAppStateCR202.Status))
			Expect(dciAppStateCR202.Labels).To(Equal(DCIAppStateCR202.Labels))
			Expect(dciAppStateCR202.Finalizers).To(BeNil())
		})

		It("DCIAppState Create Test01", func() {

			By("Test01: Reconciling the create process")
			fmt.Println("DCIAppState Create Test01: Reconciling the create process")

			// Create DCIAppStateCR
			createDCIAppStateCR101 := DCIAppStateCR101.DeepCopy()
			createDCIAppStateCR101.Labels = cLabels
			err = createDCIAppState(ctx, createDCIAppStateCR101)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cDCIAppStateNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR101 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-101",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR101)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-101", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check DCIAppState CR
			Expect(dciAppStateCR101.Spec).To(Equal(DCIAppStateCR101.Spec))
			Expect(dciAppStateCR101.Status).To(Equal(DCIAppStateCR101.Status))
			Expect(dciAppStateCR101.Labels).To(Equal(cLabels))
			Expect(dciAppStateCR101.Finalizers).To(BeNil())
		})

		It("DCIAppState Update Test01", func() {

			By("Test01: Reconciling the update process")
			fmt.Println("DCIAppState Update Test01: Reconciling the update process")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR401)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401ToStatus1))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

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
			Expect(configFileData).To(Equal(ExpectedConfigFile401))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test02", func() {

			By("Test02: DCIAppStateCR.Status.AppStatus none")
			fmt.Println("DCIAppState Update Test02: DCIAppStateCR.Status.AppStatus none")

			// Create ConnectEndpointCR
			connectEndpointCR401T01 := ConnectEndpointCR401.DeepCopy()
			connectEndpointCR401T01.Spec.To[0].ToPod[0].PodRef.Name = dciappstate401nm
			connectEndpointCR401T01.Spec.To[0].ToPod[0].Address = dciappstate401ip
			connectEndpointCR401T01.Status.To[0].ToPod[0].PodRef.Name = dciappstate401nm
			connectEndpointCR401T01.Status.To[0].ToPod[0].Address = dciappstate401ip
			err = createConnectEndpoint(ctx, connectEndpointCR401T01)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR401)
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
					Name:      "dciappstate-401",
					Namespace: cDCIAppStateNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get DCIAppState CR
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401ToStatus2))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR401 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-401",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR401)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-401", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

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
			Expect(configFileData).To(Equal(ExpectedConfigFile401Non))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test03", func() {

			By("Test03: ConnectEndpoint.Status.To[0].ToPod[0].AppStatus none")
			fmt.Println("DCIAppState Update Test03: ConnectEndpoint.Status.To[0].ToPod[0].AppStatus none")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR402)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
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
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get DCIAppState CR
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
			Expect(configFileData).To(Equal(ExpectedConfigFile402))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test04", func() {

			By("Test04: DCIAppState.Status.AppStatus.OperationState none")
			fmt.Println("DCIAppState Update Test04: DCIAppState.Status.AppStatus.OperationState none")

			// Create ConnectEndpointCR
			ConnectEndpointCR401T04 := ConnectEndpointCR401.DeepCopy()
			ConnectEndpointCR401T04.Spec.To[0].ToPod[0].PodRef.Name = dciappstate402nm
			ConnectEndpointCR401T04.Spec.To[0].ToPod[0].Address = dciappstate402ip
			ConnectEndpointCR401T04.Status.To[0].ToPod[0].PodRef.Name = dciappstate402nm
			ConnectEndpointCR401T04.Status.To[0].ToPod[0].Address = dciappstate402ip
			err = createConnectEndpoint(ctx, ConnectEndpointCR401T04)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR402)
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
					Name:      "dciappstate-402",
					Namespace: cDCIAppStateNamespace,
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401ToStatus3))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR402 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-402",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR402)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-402", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

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
			Expect(configFileData).To(Equal(ExpectedConfigFile4012))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Delete Test01", func() {

			By("Test01: Reconciling the delete process")
			fmt.Println("DCIAppState Delete Test01: Reconciling the delete process")

			// Create DCIAppStateCR
			err = createDCIAppState(ctx, &DCIAppStateCR202)
			if err != nil {
				fmt.Println("There is a problem in createing DCIAppStateCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Get DCIAppState CR
			var dciAppStateCR202 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).NotTo(HaveOccurred())

			// Delete DCIAppState CR
			err = k8sClient.Delete(ctx, &dciAppStateCR202)
			if err != nil {
				fmt.Println("Can not delete DCIAppStateCR", err)
			}

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).NotTo(HaveOccurred())

			// Get DCIAppState CR
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-202",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR202)
			Expect(err).To(MatchError(apierrors.IsNotFound, "NofFound"))
		})
	})

	Describe("DCIAppStateEventPredicate Update Function", func() {
		var predicate DCIAppStateEventPredicate

		BeforeEach(func() {
			predicate = DCIAppStateEventPredicate{}
		})

		It("should return true when ObjectOld is nil", func() {
			event := event.UpdateEvent{
				ObjectOld: nil,
				ObjectNew: &examplecomv1.DCIAppState{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectNew is nil", func() {
			event := event.UpdateEvent{
				ObjectOld: &examplecomv1.DCIAppState{},
				ObjectNew: nil,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectOld is not DCIAppState type", func() {
			event := event.UpdateEvent{
				ObjectOld: &corev1.Pod{},
				ObjectNew: &examplecomv1.DCIAppState{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when ObjectNew is not DCIAppState type", func() {
			event := event.UpdateEvent{
				ObjectOld: &examplecomv1.DCIAppState{},
				ObjectNew: &corev1.Pod{},
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when AppStatus.Raw are both nil and different", func() {
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: nil,
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"key": "value"}`),
					},
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when AppStatus.Raw are different", func() {
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"old": "value"}`),
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"new": "value"}`),
					},
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return true when old has Raw and new is nil", func() {
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"key": "value"}`),
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: nil,
					},
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeTrue())
		})

		It("should return false when AppStatus.Raw are both nil", func() {
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: nil,
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: nil,
					},
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeFalse())
		})

		It("should return false when AppStatus.Raw are equal", func() {
			rawData := []byte(`{"key": "value"}`)
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: rawData,
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: rawData,
					},
				},
			}

			event := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}
			result := predicate.Update(event)
			Expect(result).To(BeFalse())
		})

		It("should return false when AppStatus.Raw have same content but different byte slices", func() {
			oldObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"key": "value"}`),
					},
				},
			}
			newObj := &examplecomv1.DCIAppState{
				Status: examplecomv1.DCIAppStateStatus{
					AppStatus: runtime.RawExtension{
						Raw: []byte(`{"key": "value"}`),
					},
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
})
