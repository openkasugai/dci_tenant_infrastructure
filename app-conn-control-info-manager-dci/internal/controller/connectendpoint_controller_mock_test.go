// Copyright 2026 NTT, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"syscall"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"encoding/json"

	"github.com/golang/mock/gomock"
	"k8s.io/klog/v2/textlogger"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConnectEndpointClientGet(k8sClient client.Client) func(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
		return k8sClient.Get(ctx, key, obj, opts...)
	}
}
func ConnectEndpointClientPatch(k8sClient client.Client) func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
		return k8sClient.Patch(ctx, obj, patch, opts...)
	}
}

type typeMetaMatcher struct {
	typ       interface{}
	name      string
	namespace string
}

func (m *typeMetaMatcher) Matches(x interface{}) bool {
	if reflect.TypeOf(x) != reflect.TypeOf(m.typ) {
		return false
	}
	obj, ok := x.(client.Object)
	if !ok {
		return false
	}
	return obj.GetName() == m.name && obj.GetNamespace() == m.namespace
}
func (m *typeMetaMatcher) String() string {
	return "matches resource with specific Type, Name, and Namespace"
}

func TypeMetaEq(typ interface{}, name, namespace string) gomock.Matcher {
	return &typeMetaMatcher{typ: typ, name: name, namespace: namespace}
}

var _ = Describe("ConnectEndpoint Controller", func() {
	var mgr ctrl.Manager
	var err error
	ctx := context.Background()
	var writer = bytes.Buffer{}

	Context("Test for ConnectEndpointReconciler", Ordered, func() {

		var reconciler ConnectEndpointReconciler

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
			ConnectEndpointCR201AToStatus0[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus0[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus0[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus0[1].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus1[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus1[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus1[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatus1[1].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE[1].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE[1].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE[1].ToPod[3].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus2[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR101AToStatus2[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE2[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE2[0].ToPod[1].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE2[0].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE2[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE3[0].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE3[1].ToPod[0].AppStatus.Raw = appStatusData01.Raw
			ConnectEndpointCR201AToStatusE3[1].ToPod[2].AppStatus.Raw = appStatusData01.Raw
			DCIAppStateCR101.Status.AppStatus.Raw = appStatusData01.Raw
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

			gomockCtrl := gomock.NewController(GinkgoT())
			defer gomockCtrl.Finish()

			mockClient = NewMockClient(gomockCtrl)
			reconciler = ConnectEndpointReconciler{
				Client: mockClient,
				Scheme: testScheme,
				Log:    log.WithName("controllers").WithName("ConnectEndpoint"),
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

		It("ConnectEndpoint Update MockTest01A", func() {

			By("MockTest01A: Reconciling the update process Client.Get() error(ConnectEndpoinrCR)")
			fmt.Println("ConnectEndpoint Update MockTest01A: Reconciling the update process Client.Get() error(ConnectEndpoinrCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-101", Namespace: cConnectEndpointNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
				Return(errors.New("ConnectEndpointCR Get Error Test(not NotFound)"))

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile201A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest02A", func() {

			By("MockTest02A: Reconciling the update process Client.Get() error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest02A: Reconciling the update process Client.Get() error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-205", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				Return(errors.New("DCIAppStateCR Get Error Test(not NotFound)")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile201A))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest03A", func() {

			By("MockTest03A: Reconciling the update process Client.Patch() Labels Add error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest03A: Reconciling the update process Client.Patch() Labels Add error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.DCIAppState{}, "dciappstate-205", cDCIAppStateNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppStateCR dciappstate-205 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus1))
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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest04A-1", func() {

			By("MockTest04A-1: Reconciling the update process Client.Patch() Labels Delete error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest04A-1: Reconciling the update process Client.Patch() Labels Delete error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.DCIAppState{}, "dciappstate-202", cDCIAppStateNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppStateCR dciappstate-202 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatusE))
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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest04A-2", func() {

			By("MockTest04A-2: Reconciling the update process Client.Patch() Labels Delete error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest04A-2: Reconciling the update process Client.Patch() Labels Delete error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			pcreateConnectEndpoint := ConnectEndpointCR201A.DeepCopy()
			pcreateConnectEndpoint.Spec.To = ConnectEndpointCR201To
			err = createConnectEndpoint(ctx, pcreateConnectEndpoint)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
			}
			Expect(err).NotTo(HaveOccurred())

			// Create DCIAppStateCR 901
			createDCIAppStateCR := DCIAppStateCR901.DeepCopy()
			createDCIAppStateCR.ObjectMeta.Labels = cLabels
			err = createDCIAppState(ctx, createDCIAppStateCR)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.DCIAppState{}, "dciappstate-901", cDCIAppStateNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppStateCR dciappstate-202 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatusE2))
			Expect(*(connectEndpointCR.Status).InputEndpoint).To(Equal(externalInputEndpoint))
			Expect(*(connectEndpointCR.Status).OutputEndpoint).To(Equal(externalOutputEndpoint))

			// Get DCIAppState CR
			var dciAppStateCR901 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-901",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR901)
			Expect(err).NotTo(HaveOccurred())

			err = showDCIAppState(ctx, "dciappstate-901", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			// Check DCIAppState CR
			Expect(dciAppStateCR901.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR901.Labels).To(Equal(cLabels))

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

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Delete MockTest05A", func() {

			By("MockTest05A: Reconciling the Delete process delAppStateCRList() Client.Get() error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest05A: Reconciling the Delete process delAppStateCRList() Client.Get() error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-202", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				Return(errors.New("DCIAppStateCR Get dciappstate-202 Error Test(not NotFound)")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR204.Labels).To(BeNil())

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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Delete MockTest06A", func() {

			By("MockTest06A: Reconciling the Delete process delAppStateCRList() Client.Patch() error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Delete MockTest06A: Reconciling the Delete process delAppStateCRList() Client.Patch() error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.DCIAppState{}, "dciappstate-202", cDCIAppStateNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppStateCR dciappstate-202 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR204.Labels).To(BeNil())

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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Create MockTest07A", func() {

			By("MockTest07A: Reconciling the create process reconcileCreate() Client.Patch() error")
			fmt.Println("ConnectEndpoint Create MockTest07A: Reconciling the create process reconcileCreate() Client.Patch() error")

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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.ConnectEndpoint{}, "dciappstate-101", cConnectEndpointNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppConnection dciappstate-101 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Finalizers).To(BeNil())
			Expect(connectEndpointCR.Labels).To(BeNil())
			Expect(connectEndpointCR.Status.Condition).To(Equal(""))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())

			// Get DCIAppState CR
			var dciAppStateCR102 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-102",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR102)
			Expect(err).NotTo(HaveOccurred())

			// Check DCIAppState CR
			err = showDCIAppState(ctx, "dciappstate-102", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}

			Expect(dciAppStateCR102.Finalizers).To(BeNil())
			Expect(dciAppStateCR102.Labels).To(BeNil())

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			// Check DCIAppState CR
			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			Expect(dciAppStateCR103.Finalizers).To(BeNil())
			Expect(dciAppStateCR103.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
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
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest08A", func() {

			By("MockTest08A: Reconciling the update process reconcileUpdate() Cilent.Patch() error(ConnectEndpointCR)")
			fmt.Println("ConnectEndpoint Update MockTest08A: Reconciling the update process reconcileUpdate() Cilent.Patch() error(ConnectEndpointCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.ConnectEndpoint{}, "dciappstate-101", cConnectEndpointNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("DCIAppConnection dciappstate-101 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(configFileData).To(Equal(ExpectedConfigFile201A))

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

		It("ConnectEndpoint Delete MockTest09A", func() {

			By("MockTest09A: Reconciling the Delete process reconcileDelete() Client.Patch() error(ConnectEndpointCR)")
			fmt.Println("ConnectEndpoint Update MockTest09A: Reconciling the Delete process reconcileDelete() Client.Patch() error(ConnectEndpointCR)")

			// Create ConnectEndpointCR
			err = createConnectEndpoint(ctx, &ConnectEndpointCR201A)
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

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					TypeMetaEq(&examplecomv1.ConnectEndpoint{}, "dciappstate-101", cConnectEndpointNamespace),
					gomock.Any(),
					gomock.Any()).
				Return(errors.New("ConnectEndpointCR dciappstate-101 Patch Error Test")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR204.Labels).To(BeNil())

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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(false).To(Equal(checkTestProcess(checkPID)))

			err = showappConnectControlInfo()
			if err != nil {
				fmt.Println("error showappConnectControlInfo.", err)
			}

			if !checkTestProcess(checkPID) {
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest10A", func() {

			By("MockTest10A: Reconciling the create process Client.Get() error DCIAppStateCR")
			fmt.Println("ConnectEndpoint Create MockTest10A: Reconciling the create process Client.Get() error DCIAppStateCR")

			// Create ConnectEndpointCR
			pconnectEndpointCR := ConnectEndpointCR101A.DeepCopy()
			controllerutil.AddFinalizer(pconnectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			pconnectEndpointCR.Status.Condition = examplecomv1.NotConfigured
			err = createConnectEndpoint(ctx, pconnectEndpointCR)
			if err != nil {
				fmt.Println("There is a problem in createing ConnectEndpointCR ", err)
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

			err = showConnectEndpoint(ctx, cConnectEndpointNamespace)
			if err != nil {
				fmt.Println("error showConnectEndpoint.", err)
			}

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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-103", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				Return(errors.New("DCIAppStateCR Get dciappstate-103 Error Test(not NotFound)")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101FromStatus0))
			Expect(connectEndpointCR.Status.To).To(BeNil())
			Expect(connectEndpointCR.Status.InputEndpoint).To(BeNil())
			Expect(connectEndpointCR.Status.OutputEndpoint).To(BeNil())

			// Get DCIAppState CR
			var dciAppStateCR102 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-102",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR102)
			Expect(err).NotTo(HaveOccurred())

			// Check DCIAppState CR
			err = showDCIAppState(ctx, "dciappstate-102", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			Expect(dciAppStateCR102.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR102.Labels).To(BeNil())

			// Get DCIAppState CR
			var dciAppStateCR103 examplecomv1.DCIAppState
			err = k8sClient.Get(ctx, client.ObjectKey{
				Name:      "dciappstate-103",
				Namespace: cDCIAppStateNamespace,
			},
				&dciAppStateCR103)
			Expect(err).NotTo(HaveOccurred())

			// Check DCIAppState CR
			err = showDCIAppState(ctx, "dciappstate-103", cDCIAppStateNamespace)
			if err != nil {
				fmt.Println("error showDCIAppState.", err)
			}
			Expect(dciAppStateCR103.Finalizers).To(BeEmpty())
			Expect(dciAppStateCR103.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
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
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest11A", func() {

			By("MockTest11A: Reconciling the update process Client.Get() error(DCIAppStateCR)")
			fmt.Println("ConnectEndpoint Update MockTest11A: Reconciling the update process Client.Get() error(DCIAppStateCR)")

			// Create ConnectEndpointCR
			pconnectEndpointCR := ConnectEndpointCR201A.DeepCopy()
			pconnectEndpointCR.Status.Condition = examplecomv1.NotConfigured
			pconnectEndpointCR.Status.To[1].ToPod[1].AppStatus.Raw = nil
			err = createConnectEndpoint(ctx, pconnectEndpointCR)
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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-203", Namespace: cDCIAppStateNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.DCIAppState{})).
				Return(errors.New("DCIAppStateCR Get dciappstate-203 Error Test(not NotFound)")).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

			// Run Reconcile
			got, errReconcile := reconciler.Reconcile(ctx,
				ctrl.Request{NamespacedName: types.NamespacedName{
					Name:      "dciappstate-101",
					Namespace: cConnectEndpointNamespace,
				}})
			Expect(got).To(Equal(ctrl.Result{}))
			Expect(errReconcile).To(HaveOccurred())

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
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.NotConfigured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatusE3))
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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
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
				Expect(configFileData).To(Equal(ExpectedConfigFile101Del))
			}

			// Update RemoveFinalizer to connectEndpoint CR
			controllerutil.RemoveFinalizer(&connectEndpointCR, examplecomv1.ConnectEndpointFinalizer)
			err = k8sClient.Update(ctx, &connectEndpointCR)
			if err != nil {
				fmt.Println("error update RemoveFinalizer to connectEndpointCR.", err)
			}
		})

		It("ConnectEndpoint Update MockTest12A", func() {

			By("MockTest12A: Reconciling the not action process Client.Get() error(ConnectEndpoinrCR)")
			fmt.Println("ConnectEndpoint Update MockTest12A: Reconciling the not action process Client.Get() error(ConnectEndpoinrCR)")

			// Create ConnectEndpointCR
			pcreateConnectEndpoint := ConnectEndpointCR201A.DeepCopy()
			controllerutil.RemoveFinalizer(pcreateConnectEndpoint, examplecomv1.ConnectEndpointFinalizer)
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

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					client.ObjectKey{Name: "dciappstate-101", Namespace: cConnectEndpointNamespace},
					gomock.AssignableToTypeOf(&examplecomv1.ConnectEndpoint{})).
				DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					err := ConnectEndpointClientGet(k8sClient)(ctx, key, obj, opts...)
					now := metav1.NewTime(metav1.Now().Time)
					stConnectEndpoint, ok := obj.(*examplecomv1.ConnectEndpoint)
					if ok {
						stConnectEndpoint.ObjectMeta.DeletionTimestamp = &now
					}
					return err
				}).Times(1)

			mockClient.EXPECT().
				Get(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientGet(k8sClient)).AnyTimes()

			mockClient.EXPECT().
				Patch(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
				DoAndReturn(ConnectEndpointClientPatch(k8sClient)).AnyTimes()

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
			Expect(connectEndpointCR.Finalizers).To(BeEmpty())
			Expect(connectEndpointCR.Status.Condition).To(Equal(examplecomv1.Configured))
			Expect(connectEndpointCR.Status.From).To(Equal(ConnectEndpointCR101From))
			Expect(connectEndpointCR.Status.To).To(Equal(ConnectEndpointCR201AToStatus0))
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
			Expect(dciAppStateCR205.Labels).To(BeNil())

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
			Expect(dciAppStateCR206.Labels).To(BeNil())

			// Check SIGHUP receive pid
			Expect(true).To(Equal(checkTestProcess(checkPID)))

			if !checkTestProcess(checkPID) {
				err = showappConnectControlInfo()
				if err != nil {
					fmt.Println("error showappConnectControlInfo.", err)
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
				Expect(configFileData).To(Equal(ExpectedConfigFile201A))
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
