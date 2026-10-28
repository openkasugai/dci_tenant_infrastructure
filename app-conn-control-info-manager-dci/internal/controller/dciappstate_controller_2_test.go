package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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
			ConnectEndpointCR402A.Status.To[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR402A.Status.To[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR207AToStatus0[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus1[1].ToPod[0].AppStatus.Raw = appStatusData02.Raw
			ConnectEndpointCR401AToStatus1[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus2[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus2[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus3[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR401AToStatus3[1].ToPod[0].AppStatus.Raw = appStatusData03.Raw
			ConnectEndpointCR401AToStatus3[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
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

		It("DCIAppState Update ErrorTest01", func() {

			By("ErrorTest01: Reconciling the update process createConfigFile() os.WriteFile() error")
			fmt.Println("DCIAppState Update ErrorTest01: Reconciling the update process createConfigFile() os.WriteFile() error")

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

			writeFileFunc = func(name string, data []byte, perm os.FileMode) error {
				err = errors.New("app-conn-control-info-manager-dci_common.go: os.WriteFile() error test")
				return err
			}

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
			writeFileFunc = os.WriteFile

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

		It("DCIAppState Update ErrorTest02", func() {

			By("ErrorTest02: Reconciling the update process createConfigFile() Signal() error")
			fmt.Println("DCIAppState Update ErrorTest02: Reconciling the update process createConfigFile() Signal() error")

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

			signalFunc = func(p *os.Process, sig syscall.Signal) (err error) {
				err = errors.New("app-conn-control-info-manager-dci_common.go: Signal() error test")
				return err
			}

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-202",
					Namespace: cDCIAppStateNamespace,
				}})
			signalFunc = signalSend

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

		It("DCIAppState Update Test01A", func() {

			By("Test01A: Reconciling the update process")
			fmt.Println("DCIAppState Update Test01A: Reconciling the update process")

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401AToStatus1))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile401A))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test02A", func() {

			By("Test02A: DCIAppStateCR.Status.AppStatus none")
			fmt.Println("DCIAppState Update Test02A: DCIAppStateCR.Status.AppStatus none")

			// Create ConnectEndpointCR
			connectEndpointCR401T01 := ConnectEndpointCR401A.DeepCopy()
			connectEndpointCR401T01.Spec.To[1].ToPod[0].PodRef.Name = dciappstate401nm
			connectEndpointCR401T01.Spec.To[1].ToPod[0].Address = dciappstate401ip
			connectEndpointCR401T01.Status.To[1].ToPod[0].PodRef.Name = dciappstate401nm
			connectEndpointCR401T01.Status.To[1].ToPod[0].Address = dciappstate401ip
			err = createConnectEndpoint(ctx, connectEndpointCR401T01)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401AToStatus2))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile401ANon))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test03A", func() {

			By("Test03A: ConnectEndpoint.Status.To[0].ToPod[0].AppStatus none")
			fmt.Println("DCIAppState Update Test03A: ConnectEndpoint.Status.To[0].ToPod[0].AppStatus none")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR402A)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR207AToStatus0))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile402A))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("DCIAppState Update Test04A", func() {

			By("Test04A: DCIAppState.Status.AppStatus.OperationState none")
			fmt.Println("DCIAppState Update Test04A: DCIAppState.Status.AppStatus.OperationState none")

			// Create ConnectEndpointCR
			ConnectEndpointCR401T04 := ConnectEndpointCR401A.DeepCopy()
			ConnectEndpointCR401T04.Spec.To[1].ToPod[0].PodRef.Name = dciappstate402nm
			ConnectEndpointCR401T04.Spec.To[1].ToPod[0].Address = dciappstate402ip
			ConnectEndpointCR401T04.Status.To[1].ToPod[0].PodRef.Name = dciappstate402nm
			ConnectEndpointCR401T04.Status.To[1].ToPod[0].Address = dciappstate402ip
			err = createConnectEndpoint(ctx, ConnectEndpointCR401T04)
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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR401AToStatus3))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile401A2))

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})
	})
})
