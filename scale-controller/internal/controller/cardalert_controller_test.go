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
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1alpha3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/test/testutils"
	"scale-controller/test/testutils/mock"
)

// Test for SetupWithManager() method
var _ = Describe("CardAlert Controller Setting Up", Label("ca", "settings", "ttl"), func() {
	Context("When setting up with manager", func() {
		const (
			maNamespace = rpmNamespace
		)
		var (
			newControllerManagedBy     = ctrl.NewControllerManagedBy
			newControllerManagedByMock = func(_ manager.Manager) *ctrl.Builder {
				m, err := ctrl.NewManager(&rest.Config{}, ctrl.Options{
					Scheme: scheme.Scheme,
					Controller: config.Controller{
						SkipNameValidation: ptr.To(true),
					},
				})
				Expect(err).NotTo(HaveOccurred())
				return newControllerManagedBy(m).WithOptions(controller.Options{SkipNameValidation: ptr.To(true)})
			}
			logger = ctrl.Log.WithName("cardalert-controller-test")
		)

		BeforeEach(func() {
			ctrl.NewControllerManagedBy = newControllerManagedByMock
		})
		AfterEach(func() {
			Expect(os.Unsetenv(cardAlertTTLEnvVarName)).To(Succeed())
			Expect(os.Unsetenv(autoReplaceEnabledEnvVarName)).To(Succeed())
			Expect(os.Unsetenv(deviceTaintRulePolicyEnvVarName)).To(Succeed())
			Expect(os.Unsetenv(machineAlertNamespaceEnvVarName)).To(Succeed())
			ctrl.NewControllerManagedBy = newControllerManagedBy
		})

		It("should successfully setup with enabling TTL", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
				settingTTL         = "1"
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())
			Expect(os.Setenv(cardAlertTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with disabling TTL", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with enabling auto replacement", func() {
			// setting variables
			var (
				settingAutoReplace = "true"
				settingMaNamespace = maNamespace
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with disabling auto replacement", func() {
			// setting variables
			var (
				settingAutoReplace = "false"
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with NoExecute DeviceTaintRule policy", func() {
			// setting variables
			var (
				settingAutoReplace = "false"
				settingPolicy      = string(resourcev1.DeviceTaintEffectNoExecute)
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())
			Expect(os.Setenv(deviceTaintRulePolicyEnvVarName, settingPolicy)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with NoSchedule DeviceTaintRule policy", func() {
			// setting variables
			var (
				settingAutoReplace = "false"
				settingPolicy      = string(resourcev1.DeviceTaintEffectNoSchedule)
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())
			Expect(os.Setenv(deviceTaintRulePolicyEnvVarName, settingPolicy)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should unsuccessfully setup with setting invalid empty CARD_ALERT_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
				settingTTL         = ""
				errMsgs            = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds is empty",
				}
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())
			Expect(os.Setenv(cardAlertTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid character CARD_ALERT_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
				settingTTL         = "notNumber"
				errMsgs            = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds cannot convert into int32",
				}
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())
			Expect(os.Setenv(cardAlertTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid minus CARD_ALERT_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
				settingTTL         = "-1"
				errMsgs            = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds is invalid(" + settingTTL + ")",
				}
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())
			Expect(os.Setenv(cardAlertTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with invalid ENV_CARD_ALERT_CARD_TAINT_POLICY", func() {
			// setting variables
			var (
				settingAutoReplace = "false"
				settingPolicy      = ""
				errMsgs            = []string{
					fmt.Sprintf("expected NoExecute or NoSchedule, but got %s for %s", settingPolicy, deviceTaintRulePolicyEnvVarName),
				}
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())
			Expect(os.Setenv(deviceTaintRulePolicyEnvVarName, settingPolicy)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup without setting MACHINE_ALERT_NAMESPACE", func() {
			// setting variables
			var (
				errMsgs = []string{
					fmt.Sprintf("expected %s to be set, but got empty value", machineAlertNamespaceEnvVarName),
				}
			)

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid empty MACHINE_ALERT_NAMESPACE", func() {
			// setting variables
			var (
				settingMaNamespace = ""
				errMsgs            = []string{
					fmt.Sprintf("expected %s to be set, but got empty value", machineAlertNamespaceEnvVarName),
				}
			)

			// setting environment variables
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with invalid ENV_CARD_ALERT_AUTO_REPLACE_ENABLED", func() {
			// setting variables
			var (
				settingAutoReplace = ""
				errMsgs            = []string{
					fmt.Sprintf("expected true or false, but got %s for %s", settingAutoReplace, autoReplaceEnabledEnvVarName),
				}
			)

			// setting environment variables
			Expect(os.Setenv(autoReplaceEnabledEnvVarName, settingAutoReplace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with nil HelperClient", func() {
			// setting variables
			var (
				settingMaNamespace = maNamespace
				settingTTL         = "3600"
				errMsgs            = []string{
					"expected ClientHelper is not nil, but got nil",
				}
			)

			// setting environment variables
			Expect(os.Setenv(cardAlertTTLEnvVarName, settingTTL)).To(Succeed())
			Expect(os.Setenv(machineAlertNamespaceEnvVarName, settingMaNamespace)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: nil})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with nil mgr", func() {
			// setting variables
			var (
				errMsgs = []string{
					"expected mgr is not nil, but got nil",
				}
			)

			// executing SetupWithManager
			r := (&CardAlertReconciler{ClientHelper: nil})
			err := r.SetupWithManager(nil)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})
	})
})

// Test for internal methods with direct calls
var _ = Describe("CardAlert Controller Methods directly calling", Label("ca", "methods"), func() {
	Context("When calling methods directly", func() {
		It("should unsuccessfully call doReconcile with invalid ca.Status", func() {
			By("calling doReconcile", func() {
				const dummyState = "dummy"
				var errMsgs = []string{
					fmt.Sprintf("unexpected state(%s) is detected in doReconcile", dummyState),
					"failed to patch CardAlert status",
				}
				r := newCaReconciler(nil, nil)
				ca := &scalev1.CardAlert{
					Status: scalev1.CardAlertStatus{
						State: dummyState,
					},
				}
				_, err := r.doReconcile(ctx, ca)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call doFinalize with invalid ca.Status", func() {
			By("calling doFinalize", func() {
				const dummyState = "dummy"
				var errMsgs = []string{
					fmt.Sprintf("unexpected state(%s) is detected in doFinalize", dummyState),
					"failed to patch CardAlert status",
				}
				r := newCaReconciler(nil, nil)
				ca := &scalev1.CardAlert{
					Status: scalev1.CardAlertStatus{
						State: dummyState,
					},
				}
				_, err := r.doFinalize(ctx, ca)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call createCELExpression with invalid cards", func() {
			By("calling createCELExpression", func() {
				var (
					cards              = []scalev1.CardInfo{}
					emptyCELExpression = ""
				)
				celExpression := createCELExpression(cards)
				Expect(celExpression).To(Equal(emptyCELExpression))
			})
		})
	})
})

// Test for kubebuilder markers validation in CRD
var _ = Describe("CardAlert Controller embedded Validation in CRD", Label("ca", "crd"), func() {
	Context("When creating CardAlert with invalid parameters", func() {
		// error messages
		const (
			pathSpec              = "spec"
			pathStatus            = "status"
			pathSpecNodeName      = "nodeName"
			pathSpecCards         = "cards"
			pathSpecCardsModel    = "model"
			pathSpecCardsGPUUUID  = "gpuUUID"
			pathSpecCardsMIGUUIDs = "migUUIDs"
			pathSpecSource        = "source"
			pathSpecReason        = "reason"
			pathStatusState       = "state"
		)
		const (
			errMsgNodeNameInvalid   = pathSpec + "." + pathSpecNodeName + ": Invalid value: \"%s\""
			errMsgNodeNameRequired  = pathSpec + "." + pathSpecNodeName + ": Required value"
			errMsgCardsInvalid      = pathSpec + "." + pathSpecCards + ": Invalid value"
			errMsgCardsRequired     = pathSpec + "." + pathSpecCards + ": Required value"
			errMsgGPUUUIDInvalid    = pathSpec + "." + pathSpecCards + "[0]." + pathSpecCardsGPUUUID + ": Invalid value: \"%s\""
			errMsgGPUUUIDRequired   = pathSpec + "." + pathSpecCards + "[0]." + pathSpecCardsGPUUUID + ": Required value"
			errMsgModelInvalid      = pathSpec + "." + pathSpecCards + "[0]." + pathSpecCardsModel + ": Invalid value: \"%s\""
			errMsgSourceUnsupported = pathSpec + "." + pathSpecSource + ": Unsupported value"
			errMsgSourceRequired    = pathSpec + "." + pathSpecSource + ": Required value"
			errMsgStateUnsupported  = pathStatus + "." + pathStatusState + ": Unsupported value"
			errMsgModifyExisting    = "`" + pathSpec + "` may not be modified after creation."
		)
		// common const
		const (
			caNamespace = defaultNamespace
		)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupCardAlertResourcesWithoutFinalizers(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		newCardAlertObjMap := func(name, namespace string) map[string]any {
			return map[string]any{
				"apiVersion": scalev1.GroupVersion.String(),
				"kind":       "CardAlert",
				"metadata": map[string]any{
					"name":      name,
					"namespace": namespace,
				},
				pathSpec: map[string]any{
					pathSpecNodeName: "dummy-node",
					pathSpecCards: []any{
						map[string]any{
							pathSpecCardsModel:   "dummy-model",
							pathSpecCardsGPUUUID: "dummy-uuid",
						},
					},
					pathSpecSource: string(scalev1.WebhookReceiverOnTenant),
				},
			}
		}

		It("should fail to create CardAlert with invalid NodeName", func() {
			const caName string = "test-validation-create-invalid-node-name"
			var (
				nodeName = "" // empty NodeName is invalid
				errMsgs  = []string{
					fmt.Sprintf(errMsgNodeNameInvalid, nodeName),
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				Expect(unstructured.SetNestedField(caObj, nodeName, pathSpec, pathSpecNodeName)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert without NodeName", func() {
			const caName string = "test-validation-create-without-node-name"
			var (
				errMsgs = []string{
					errMsgNodeNameRequired,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecNodeName)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert with invalid zero-length Cards", func() {
			const caName string = "test-validation-create-invalid-zero-length-cards"
			var (
				cards   = []any{} // zero-length Cards is invalid
				errMsgs = []string{
					errMsgCardsInvalid,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				Expect(unstructured.SetNestedSlice(caObj, cards, pathSpec, pathSpecCards)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert with invalid nil Cards", func() {
			const caName string = "test-validation-create-invalid-nil-cards"
			var (
				cards   = []any(nil) // nil Cards is invalid
				errMsgs = []string{
					errMsgCardsRequired,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				Expect(unstructured.SetNestedSlice(caObj, cards, pathSpec, pathSpecCards)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert without Cards", func() {
			const caName string = "test-validation-create-without-cards"
			var (
				errMsgs = []string{
					errMsgCardsRequired,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecCards)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert with invalid Cards[].Model", func() {
			const caName string = "test-validation-create-invalid-model"
			var (
				cardsModel = "" // empty Model is invalid
				errMsgs    = []string{
					fmt.Sprintf(errMsgModelInvalid, cardsModel),
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				// Get Slice
				cards, _, err := unstructured.NestedSlice(caObj, pathSpec, pathSpecCards)
				Expect(err).NotTo(HaveOccurred())
				Expect(cards).To(HaveLen(1))
				// Get Map of the first element in Slice
				card, ok := cards[0].(map[string]any)
				Expect(ok).To(BeTrue())
				// Set invalid value
				Expect(unstructured.SetNestedField(card, cardsModel, pathSpecCardsModel)).To(Succeed())
				Expect(unstructured.SetNestedSlice(caObj, cards, pathSpec, pathSpecCards)).To(Succeed())
				// Create
				err = k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should succeed to create CardAlert without Cards[].Model", func() {
			const caName string = "test-validation-create-without-cards-model"
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				// Get Slice
				cards, _, err := unstructured.NestedSlice(caObj, pathSpec, pathSpecCards)
				Expect(err).NotTo(HaveOccurred())
				Expect(cards).To(HaveLen(1))
				// Get Map of the first element in Slice
				card, ok := cards[0].(map[string]any)
				Expect(ok).To(BeTrue())
				// Set invalid value
				unstructured.RemoveNestedField(card, pathSpecCardsModel)
				Expect(unstructured.SetNestedSlice(caObj, cards, pathSpec, pathSpecCards)).To(Succeed())
				// Create
				Expect(k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})).To(Succeed())
			})
		})

		It("should fail to create CardAlert without Cards[].GPUUUID", func() {
			const caName string = "test-validation-create-without-gpuuuid"
			var (
				errMsgs = []string{
					errMsgGPUUUIDRequired,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				// Get Slice
				cards, _, err := unstructured.NestedSlice(caObj, pathSpec, pathSpecCards)
				Expect(err).NotTo(HaveOccurred())
				Expect(cards).To(HaveLen(1))
				// Get Map of the first element in Slice
				card, ok := cards[0].(map[string]any)
				Expect(ok).To(BeTrue())
				// Set invalid value
				unstructured.RemoveNestedField(card, pathSpecCardsGPUUUID)
				Expect(unstructured.SetNestedSlice(caObj, cards, pathSpec, pathSpecCards)).To(Succeed())
				// Create
				err = k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert with invalid Source", func() {
			const caName string = "test-validation-create-invalid-source"
			var (
				source  = "InvalidSource" // invalid Source
				errMsgs = []string{
					errMsgSourceUnsupported,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				Expect(unstructured.SetNestedField(caObj, source, pathSpec, pathSpecSource)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardAlert without Source", func() {
			const caName string = "test-validation-create-without-source"
			var (
				errMsgs = []string{
					errMsgSourceRequired,
				}
			)
			By("creating CardAlert", func() {
				caObj := newCardAlertObjMap(caName, caNamespace)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecSource)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update CardAlert with invalid State", func() {
			const caName string = "test-validation-update-invalid-state"
			var (
				errMsgs = []string{
					errMsgStateUnsupported,
				}
			)
			By("creating CardAlert", func() {
				ca := &unstructured.Unstructured{Object: newCardAlertObjMap(caName, caNamespace)}
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			By("updating CardAlert", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				ca.Status.State = "InvalidState"
				err := k8sClientWorkload.Status().Update(ctx, ca)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update CardAlert with modifying Spec", func() {
			const caName string = "test-validation-update-modifying-spec"
			var (
				errMsgs = []string{
					errMsgModifyExisting,
				}
			)
			By("creating CardAlert", func() {
				ca := &unstructured.Unstructured{Object: newCardAlertObjMap(caName, caNamespace)}
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			By("updating CardAlert", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				ca.Spec.NodeName += "-modified"
				err := k8sClientWorkload.Update(ctx, ca)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})
})

// Test for Reconcile() method
var _ = Describe("CardAlert Controller", Label("ca", "controller"), func() {
	Context("When reconciling a resource", func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location = ptr.To("test-location")
			model    = "test-model"
			uuids    = []string{
				"gpu-uuid-1",
				"gpu-uuid-2",
				"gpu-uuid-3",
			}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
				{UUID: uuids[1]},
				{UUID: uuids[2]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(3),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is composed", func() {
			const caName = "test-ca-enable-auto-replace-composed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should not be changed before patching CardScalingRequest to Completed", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingCardScalingRequestState)).To(Succeed())
				})
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is composed, but CardScalingRequest failed", func() {
			const caName = "test-ca-enable-auto-replace-composed-csr-failed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed CardAlert because auto replacement by CardScalingRequest failed with phase",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseNodeScaleDownFailed)).To(Succeed())
				By("asserting the phase should become Failed when CardScalingRequest failed", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaFailed(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaFailed(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is cots", func() {
			const caName = "test-ca-enable-auto-replace-cots"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should not be changed before patching MachineAlert to Completed", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)).To(Succeed())
				})
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is cots, and MachineAlert is deleted by anyone before waiting", func() {
			const caName = "test-ca-enable-auto-replace-cots-ma-deleted-before-waiting"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("deleting MachineAlert by hand", func() {
					ca := &scalev1.CardAlert{}
					Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
					maName := ca.Status.ResourceRef.Name
					maNamespace := ca.Status.ResourceRef.Namespace
					Expect(testutils.CleanupMachineAlertResource(ctx, k8sClientManagement, maName, maNamespace)).To(Succeed())
				})
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is vm", func() {
			const caName = "test-ca-enable-auto-replace-vm"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"node created by VM is not supported for auto replacement",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Failed when the type of server is vm", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaFailed(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaFailed(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is unknown", Label("mock"), func() {
			const caName = "test-ca-enable-auto-replace-unknown-server-type"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"unexpected server type in LogicalServer",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get LogicalServer", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					alterFunc := func(ctx context.Context, namespacedName types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
						ls := &rpmv1.LogicalServer{}
						Expect(k8sClientManagement.Get(ctx, namespacedName, ls)).To(Succeed()) // get LogicalServer successfully in mock
						ls.Status.AppliedSpec.Type = rpmv1.AppliedServerType("")               // set unknown server type with overwriting
						ls.DeepCopyInto(obj.(*rpmv1.LogicalServer))
						return nil
					}
					Expect(mClient.SetGetWithFunc(k8sClientManagement.Get, alterFunc, []int{1})).To(Succeed()) // get invalid LogicalServer
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when the server type is unknown", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled", func() {
			const caName = "test-ca-disable-auto-replace"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled and the number of target GPU cards is over 1", func() {
			const caName = "test-ca-disable-auto-replace-multiple-gpu-uuids"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
					{
						Model:   model,
						GPUUUID: uuids[1],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			By("asserting the CEL expression in DeviceTaintRule", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				dtrName := ca.Status.ResourceRef.Name
				dtrNamespace := ca.Status.ResourceRef.Namespace
				celExpr := fmt.Sprintf("cel.bind(%s, device.attributes['%s'], %s.uuid=='%s'||%s.uuid=='%s')",
					deviceTaintRuleDevice, deviceTaintRuleCELLabelGPUNvidia,
					deviceTaintRuleDevice, cards[0].GPUUUID, deviceTaintRuleDevice, cards[1].GPUUUID)
				dtr := &resourcev1.DeviceTaintRule{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(dtrName, dtrNamespace), dtr)).To(Succeed())
				Expect(dtr.Spec.DeviceSelector.Selectors[0].CEL.Expression).To(Equal(celExpr))
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled and the number of target MIG cards is over 1", func() {
			const caName = "test-ca-disable-auto-replace-multiple-mig-uuids"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
						MIGUUIDs: []string{
							"mig-uuid-1",
							"mig-uuid-2",
						},
					},
					{
						Model:   model,
						GPUUUID: uuids[1],
						MIGUUIDs: []string{
							"mig-uuid-3",
						},
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			By("asserting the CEL expression in DeviceTaintRule", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				dtrName := ca.Status.ResourceRef.Name
				dtrNamespace := ca.Status.ResourceRef.Namespace
				celExpr := fmt.Sprintf("cel.bind(%s, device.attributes['%s'], %s.parentUUID=='%s'||%s.parentUUID=='%s')",
					deviceTaintRuleDevice, deviceTaintRuleCELLabelGPUNvidia,
					deviceTaintRuleDevice, cards[0].GPUUUID, deviceTaintRuleDevice, cards[1].GPUUUID)
				dtr := &resourcev1.DeviceTaintRule{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(dtrName, dtrNamespace), dtr)).To(Succeed())
				Expect(dtr.Spec.DeviceSelector.Selectors[0].CEL.Expression).To(Equal(celExpr))
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when another resource is processing", func() {
			const caName = "test-ca-another-resource-processing"
			const caAnotherName = "test-ca-another-resource-processing-another"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			By("creating another CardAlert", func() {
				ca := testutils.NewCardAlert(caAnotherName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("asserting the 1st CardAlert until Initializing passed", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
			})
			// Reconcile
			By("asserting the 2nd CardAlert's phase should be Pending when another resource is still processing", func() {
				Expect(executeCaReconciler(r, caAnotherName, caNamespace, scalev1.CardAlertPendingState)).To(Succeed())
			})
			By("asserting the 1st CardAlert", func() {
				// Reconcile
				By("executing reconcile", func() {
					Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
					Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
					Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
					Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
					Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
				})
				// Delete CardAlert
				Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
				// Finalize
				By("executing finalize", func() {
					Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
					Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
				})
			})
			By("asserting the 2nd CardAlert after the 1st CardAlert is completed", func() {
				// Reconcile
				By("executing reconcile", func() {
					Expect(executeCaPending(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaInitializing(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaCreatingReplaceResource(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(patchCardScalingRequestPhase(caAnotherName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
					Expect(executeCaWaitingCardScalingRequest(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaCompleted(r, caAnotherName, caNamespace)).To(Succeed())
				})
				// Delete CardAlert
				Expect(deleteCardAlertResource(ctx, caAnotherName, caNamespace)).To(Succeed())
				// Finalize
				By("executing finalize", func() {
					Expect(finalizeCaCompleted(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(finalizeCaFinalized(r, caAnotherName, caNamespace)).To(Succeed())
				})
			})
		})

		It("should successfully reconcile when Node is already deleted", func() {
			const caName = "test-ca-node-deleted"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
				Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the Node is already deleted", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when Node has no annotations for ClusterName,ClusterNamespace", func() {
			const caName = "test-ca-node-without-annotations"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get Cluster Namespace from Node",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nodeName, ""), node)).To(Succeed())
				node.Annotations = map[string]string{} // remove annotations for ClusterName and ClusterNamespace
				Expect(k8sClientWorkload.Update(ctx, node)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting error happens when Node has no annotations for ClusterName and ClusterNamespace", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when DCIMachine has no reference to LogicalServer", func() {
			const caName = "test-ca-dcimachine-without-logicalserver"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get LogicalServer name and namespace from DCIMachine",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.RemoveDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting error happens when DCIMachine has no reference to LogicalServer", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when the target UUID is already excluded with enabling auto replacement", func() {
			const caName = "test-ca-gpu-uuid-excluded"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: "excluded-gpu-uuid",
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the target UUID is already excluded", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when the target UUID is already excluded with disabling auto replacement", func() {
			const caName = "test-ca-gpu-uuid-excluded-auto-replace-disabled"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: "excluded-gpu-uuid",
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the target UUID is already excluded", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource for finalizer from each state", func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully finalize from Initial state", Label("mock"), func() {
			const caName = "test-ca-finalize-initial"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitialState)).To(HaveOccurred())
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from Initializing state", func() {
			const caName = "test-ca-finalize-initializing"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from Pending state", func() {
			const caName = "test-ca-finalize-pending"
			const caAnotherName = "test-ca-finalize-pending-another"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			By("creating another CardAlert", func() {
				ca := testutils.NewCardAlert(caAnotherName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile to Initializing state", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
			})
			By("executing reconcile to Pending state for another CardAlert", func() {
				By("asserting the 2nd CardAlert's phase should be Pending when another resource is still processing", func() {
					Expect(executeCaReconciler(r, caAnotherName, caNamespace, scalev1.CardAlertPendingState)).To(Succeed())
				})
			})
			By("executing reconcile to Completed state", func() {
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			Expect(deleteCardAlertResource(ctx, caAnotherName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
			By("executing finalize for another CardAlert", func() {
				Expect(finalizeCaPending(r, caAnotherName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caAnotherName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from CreatingReplaceResource state", func() {
			const caName = "test-ca-finalize-creating-replace-resource"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from CreatingReplaceResource state with MachineAlert created", func() {
			const caName = "test-ca-finalize-creating-replace-resource-ma-created"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)).To(HaveOccurred())
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from WaitingCardScalingRequest state", func() {
			const caName = "test-ca-finalize-waiting-card-scaling-request"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from WaitingMachineAlert state", func() {
			const caName = "test-ca-finalize-waiting-machine-alert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from DeletingMachineAlert state", func() {
			const caName = "test-ca-finalize-waiting-machine-alert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from CreatingTaintResource state", func() {
			const caName = "test-ca-finalize-creating-taint-resource"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from Completed state", func() {
			const caName = "test-ca-finalize-completed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully finalize from Failed state", func() {
			const caName = "test-ca-finalize-failed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Failed when the type of server is vm", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)).To(HaveOccurred())
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaFailed(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling called with no valid resource", func() {
		const (
			caNamespace = defaultNamespace
		)

		AfterEach(func() {
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should fail to get a resource when no resource exist", func() {
			const caName string = "test-not-exist"
			r := newCaReconciler(nil, nil)
			By("executing reconcile", func() {
				_, err := r.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(caName, caNamespace),
				})
				Expect(err).NotTo(HaveOccurred()) // when not found has occurred, should return nil as err
			})
		})
	})

	Context("When reconciling called with no reconciled deleting resource", func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location = ptr.To("test-location")
			model    = "test-model"
			uuids    = []string{
				"gpu-uuid-1",
			}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is composed", func() {
			const caName = "test-ca-enable-auto-replace-composed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				ca.Finalizers = []string{
					"unknown.finalizer/test",
				}
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitialState)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource with failure to verify idempotency", Label("mock"), func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1", "gpu-uuid-2", "gpu-uuid-3"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
				{UUID: uuids[1]},
				{UUID: uuids[2]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(3),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for CardAlert", func() {
			const caName = "test-ca-mock-fail-get-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get CardAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				By("setting mock to fail to patch CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetGet([]int{1})).To(Succeed()) // failed to get CardAlert
					wClient.SetDefault()
					r.Client = wClient.GetMockClient()
				})
				By("asserting error happens when failed to get CardAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitialState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Patch() for CardAlert", func() {
			const caName = "test-ca-mock-fail-patch-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to add finalizer to CardAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				By("setting mock to fail to patch CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetPatch([]int{1})).To(Succeed()) // failed to patch CardAlert
					wClient.SetDefault()
					r.Client = wClient.GetMockClient()
				})
				By("asserting error happens when failed to patch CardAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitialState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for CardAlert to check cached resource version", func() {
			const caName = "test-ca-mock-fail-get-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to synchronize cached resource version of CardAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				By("setting mock to fail to patch CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get CardAlert for checking synced resource version
					wClient.SetDefault()
					r.Client = wClient.GetMockClient()
				})
				By("asserting error happens when failed to get CardAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of List() for CardAlert", func() {
			const caName = "test-ca-mock-fail-list-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get other running CardAlerts",
					"failed to list CardAlerts",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				By("setting mock to fail to list CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetList([]int{1})).To(Succeed()) // failed to list CardAlert
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to list CardAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitialState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for Node", func() {
			const caName = "test-ca-mock-fail-get-node"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get Node",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get Node", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get Node
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to get Node", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for Machine", func() {
			const caName = "test-ca-mock-fail-get-machine"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get Machine from Node",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get Machine", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get Machine
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get Machine", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for DCIMachine", func() {
			const caName = "test-ca-mock-fail-get-dcimachine"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get DCIMachine from Machine",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get DCIMachine", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{2})).To(Succeed()) // failed to get DCIMachine
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get DCIMachine", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for LogicalServer", func() {
			const caName = "test-ca-mock-fail-get-logicalserver"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get LogicalServer",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get LogicalServer", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get LogicalServer
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get LogicalServer", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of List() for CardScalingRequest", func() {
			const caName = "test-ca-mock-fail-list-cardscalingrequest"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to list CardScalingRequests",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to list CardScalingRequests", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetList([]int{1})).To(Succeed()) // failed to list CardScalingRequests
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to list CardScalingRequests", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Scheme() for CardScalingRequest", func() {
			const caName = "test-ca-mock-fail-scheme-cardscalingrequest"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to set controller reference to CardScalingRequest",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get scheme for CardScalingRequests", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetScheme([]int{1})).To(Succeed()) // failed to get scheme for CardScalingRequests
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to get scheme for CardScalingRequests", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Create() for CardScalingRequest", func() {
			const caName = "test-ca-mock-fail-create-cardscalingrequest"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to create CardScalingRequest",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to create CardScalingRequests", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetCreate([]int{1})).To(Succeed()) // failed to create CardScalingRequests
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to create CardScalingRequests", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Status().Patch() for CardAlert with creating CardScalingRequest", func() {
			const caName = "test-ca-mock-fail-status-patch-cardalert-csr"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to patch CardAlert status",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of List() for MachineAlert", func() {
			const caName = "test-ca-mock-fail-list-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to list MachineAlerts",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to list MachineAlerts", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetList([]int{1})).To(Succeed()) // failed to list MachineAlerts
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to list MachineAlerts", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Create() for MachineAlert", func() {
			const caName = "test-ca-mock-fail-create-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to create MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to create MachineAlerts", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetCreate([]int{1})).To(Succeed()) // failed to create MachineAlerts
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to create MachineAlerts", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Status().Patch() for CardAlert with creating MachineAlert", func() {
			const caName = "test-ca-mock-fail-status-patch-cardalert-ma"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to patch CardAlert status",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for CardScalingRequest", func() {
			const caName = "test-ca-mock-fail-get-cardscalingrequest"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get CardScalingRequest",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get CardScalingRequest", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get CardScalingRequest
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to get CardScalingRequest", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingCardScalingRequestState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for MachineAlert with waiting", func() {
			const caName = "test-ca-mock-fail-get-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for MachineAlert with deleting", func() {
			const caName = "test-ca-mock-fail-get-machinealert-deleting"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get MachineAlert for deletion",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertDeletingMachineAlertState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Delete() for MachineAlert", func() {
			const caName = "test-ca-mock-fail-delete-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to delete MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to delete MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetDelete([]int{1})).To(Succeed()) // failed to delete MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to delete MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertDeletingMachineAlertState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for LogicalServer", func() {
			const caName = "test-ca-mock-fail-get-logicalserver"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get LogicalServer",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get LogicalServer", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get LogicalServer
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get LogicalServer", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingTaintResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of List() for DeviceTaintRule", func() {
			const caName = "test-ca-mock-fail-list-devicetaintrule"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to list DeviceTaintRules",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to list DeviceTaintRules", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetList([]int{1})).To(Succeed()) // failed to list DeviceTaintRules
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to list DeviceTaintRules", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingTaintResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Create() for DeviceTaintRule", func() {
			const caName = "test-ca-mock-fail-create-devicetaintrule"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to create DeviceTaintRule",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to create DeviceTaintRule", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetCreate([]int{1})).To(Succeed()) // failed to create DeviceTaintRule
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to create DeviceTaintRule", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingTaintResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Status().Patch() for CardAlert with creating DeviceTaintRule", func() {
			const caName = "test-ca-mock-fail-status-patch-cardalert-dtr"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to patch CardAlert status",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingTaintResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Patch() for CardAlert with removing finalizer", func() {
			const caName = "test-ca-mock-fail-patch-removing-finalizer"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to remove finalizer from CardAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetPatch([]int{1})).To(Succeed()) // failed to patch CardAlert
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource for finalizer with failure to verify idempotency", Label("mock"), func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for LogicalServer", func() {
			const caName = "test-ca-finalizer-mock-fail-get-logicalserver"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get LogicalServer",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				By("setting mock to fail to get LogicalServer", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get LogicalServer
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get LogicalServer", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of List() for MachineAlert after failure of Status().Patch() for CardAlert", func() {
			const caName = "test-ca-finalizer-mock-fail-list-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to list MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to patch status of CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch status of CardAlert
					wClient.SetDefault()
					r = newCaReconciler(nil, wClient.GetMockClient())
				})
				By("asserting error happens when failed to patch status of CardAlert", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)).To(HaveOccurred())
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				By("setting mock to fail to list MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetList([]int{1})).To(Succeed()) // failed to list MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to list MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Get() for MachineAlert", func() {
			const caName = "test-ca-finalizer-mock-fail-get-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				By("setting mock to fail to get MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to get MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Delete() for MachineAlert", func() {
			const caName = "test-ca-finalizer-mock-fail-delete-machinealert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to delete MachineAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				By("setting mock to fail to delete MachineAlert", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					Expect(mClient.SetDelete([]int{1})).To(Succeed()) // failed to delete MachineAlert
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when failed to delete MachineAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Status().Patch() for CardAlert", func() {
			const caName = "test-ca-mock-fail-status-patch-cardalert"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to patch CardAlert status",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				By("setting mock to fail to patch CardAlert status", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to patch CardAlert status
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("asserting error happens when failed to patch CardAlert status", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource", Label("ttl"), func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile to delete CardAlert when TTL is enabled and CardAlert completed successfully", func() {
			const caName = "test-ca-ttl-expired"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as 1 second", func() {
				r.enableTTLSecondsAfterFinished = true
				r.ttlSecondsAfterFinished = 1
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				By("sleeping for a few seconds to let TTL expire", func() {
					time.Sleep(1 * time.Second)
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile to pend deleting CardAlert when TTL is enabled and before TTL is expired", func() {
			const caName = "test-ca-ttl-pending-deletion"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as 2 second", func() {
				r.enableTTLSecondsAfterFinished = true
				r.ttlSecondsAfterFinished = 2
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
				By("sleeping for a few seconds to let TTL expire", func() {
					time.Sleep(2 * time.Second)
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile to skip deleting CardAlert when TTL is disabled", func() {
			const caName = "test-ca-ttl-disabled"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as disabled", func() {
				r.enableTTLSecondsAfterFinished = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile to skip deleting CardAlert when TTL is enabled and OwnerReference is set", func() {
			const caName = "test-ca-ttl-skip-deletion-with-ownerref"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as enabled", func() {
				r.enableTTLSecondsAfterFinished = true
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				ca.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion: "test.version",
						Kind:       "TestOwner",
						Name:       "test-owner",
						UID:        "test-uid",
					},
				}
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile to skip deleting CardAlert when TTL is enabled and CardAlert failed", func() {
			const caName = "test-ca-ttl-skip-deletion-with-failed-state"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as enabled", func() {
				r.enableTTLSecondsAfterFinished = true
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting error happens when reconciling CardAlert for VM type LogicalServer", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)).To(HaveOccurred())
				})
				Expect(executeCaFailed(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile to delete CardAlert when .status.completedTime is not set", func() {
			const caName = "test-ca-ttl-skip-deletion-without-completion-time"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"expected .status.completedTime is set when state is Completed, but not set",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting TTL as 1 second", func() {
				r.enableTTLSecondsAfterFinished = true
				r.ttlSecondsAfterFinished = 1
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				By("patching .status.completedTime as empty", func() {
					ca := &scalev1.CardAlert{}
					Expect(k8sClientWorkload.Get(ctx, types.NamespacedName{Name: caName, Namespace: caNamespace}, ca)).To(Succeed())
					ca.Status.CompletedTime = metav1.Time{}
					Expect(k8sClientWorkload.Status().Update(ctx, ca)).To(Succeed())
				})
				By("asserting error happens when reconciling CardAlert with Completed state but .status.completedTime is not set", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource", Label("ttl", "mock"), func() {
		const (
			caNamespace                 = defaultNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = defaultNamespace
			machineName                 = "test-machine"
			machineNamespace            = defaultNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = defaultNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = defaultNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = defaultNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Delete() for CardAlert", func() {
			const caName = "test-ca-ttl-delete-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to delete expired CardAlert",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to delete CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					Expect(wClient.SetDelete([]int{1})).To(Succeed()) // failed to delete CardAlert
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("setting TTL as 0 second", func() {
					r.enableTTLSecondsAfterFinished = true
					r.ttlSecondsAfterFinished = 0
				})
				By("asserting error happens when failed to delete CardAlert", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile after failure of Delete() for CardAlert with NotFound error", func() {
			const caName = "test-ca-ttl-delete-ca"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to delete CardAlert", func() {
					wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
					defaultFunc := k8sClientWorkload.Delete
					alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
						return apierrors.NewNotFound(schema.GroupResource{Resource: "cardalerts"}, caName)
					}
					Expect(wClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{1})).To(Succeed()) // failed to delete CardAlert
					wClient.SetDefault()
					r = newCaReconciler(wClient.GetMockClient(), nil)
				})
				By("setting TTL as 0 second", func() {
					r.enableTTLSecondsAfterFinished = true
					r.ttlSecondsAfterFinished = 0
				})
				By("asserting error happens when failed to delete CardAlert", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})

	Context("When reconciling a resource", Label("namespaced"), func() {
		const (
			caNamespace                 = testNamespace
			nodeName                    = "test-node"
			clusterName                 = "test-cluster"
			clusterNamespace            = testNamespace
			machineName                 = "test-machine"
			machineNamespace            = testNamespace
			mdName                      = "test-machinedeployment"
			mdNamespace                 = testNamespace
			dciMachineName              = "test-dci-machine"
			dciMachineNamespace         = testNamespace
			dciMachineTemplateName      = "test-dci-machine-template"
			dciMachineTemplateNamespace = testNamespace
			logicalServerName           = "test-logical-server"
			logicalServerNamespace      = rpmNamespace
			maNamespace                 = rpmNamespace
		)
		var (
			location = ptr.To("test-location")
			model    = "test-model"
			uuids    = []string{
				"gpu-uuid-1",
				"gpu-uuid-2",
				"gpu-uuid-3",
			}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
				{UUID: uuids[1]},
				{UUID: uuids[2]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(3),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupDeviceTaintRuleResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineAlertResources(ctx, k8sClientManagement, maNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClientWorkload, caNamespace)).To(Succeed())
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is composed", func() {
			const caName = "test-ca-enable-auto-replace-composed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should not be changed before patching CardScalingRequest to Completed", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingCardScalingRequestState)).To(Succeed())
				})
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
				Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is composed, but CardScalingRequest failed", func() {
			const caName = "test-ca-enable-auto-replace-composed-csr-failed"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed CardAlert because auto replacement by CardScalingRequest failed with phase",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseNodeScaleDownFailed)).To(Succeed())
				By("asserting the phase should become Failed when CardScalingRequest failed", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaFailed(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaFailed(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is cots", func() {
			const caName = "test-ca-enable-auto-replace-cots"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should not be changed before patching MachineAlert to Completed", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)).To(Succeed())
				})
				Expect(patchMachineAlertState(caName, caNamespace, rpmv1.MachineAlertStateCompleted)).To(Succeed())
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is enabled and the type of server is cots, and MachineAlert is deleted by anyone before waiting", func() {
			const caName = "test-ca-enable-auto-replace-cots-ma-deleted-before-waiting"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				By("deleting MachineAlert by hand", func() {
					ca := &scalev1.CardAlert{}
					Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
					maName := ca.Status.ResourceRef.Name
					maNamespace := ca.Status.ResourceRef.Namespace
					Expect(testutils.CleanupMachineAlertResource(ctx, k8sClientManagement, maName, maNamespace)).To(Succeed())
				})
				Expect(executeCaWaitingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaDeletingMachineAlert(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is vm", func() {
			const caName = "test-ca-enable-auto-replace-vm"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"node created by VM is not supported for auto replacement",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Failed when the type of server is vm", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
				Expect(executeCaFailed(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaFailed(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when auto replacement is enabled and the type of server is unknown", Label("mock"), func() {
			const caName = "test-ca-enable-auto-replace-unknown-server-type"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"unexpected server type in LogicalServer",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("setting mock to fail to get LogicalServer", func() {
					mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
					alterFunc := func(ctx context.Context, namespacedName types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
						ls := &rpmv1.LogicalServer{}
						Expect(k8sClientManagement.Get(ctx, namespacedName, ls)).To(Succeed()) // get LogicalServer successfully in mock
						ls.Status.AppliedSpec.Type = rpmv1.AppliedServerType("")               // set unknown server type with overwriting
						ls.DeepCopyInto(obj.(*rpmv1.LogicalServer))
						return nil
					}
					Expect(mClient.SetGetWithFunc(k8sClientManagement.Get, alterFunc, []int{1})).To(Succeed()) // get invalid LogicalServer
					mClient.SetDefault()
					r = newCaReconciler(nil, mClient.GetMockClient())
				})
				By("asserting error happens when the server type is unknown", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled", func() {
			const caName = "test-ca-disable-auto-replace"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeVM)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled and the number of target GPU cards is over 1", func() {
			const caName = "test-ca-disable-auto-replace-multiple-gpu-uuids"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
					{
						Model:   model,
						GPUUUID: uuids[1],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			By("asserting the CEL expression in DeviceTaintRule", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				dtrName := ca.Status.ResourceRef.Name
				dtrNamespace := ca.Status.ResourceRef.Namespace
				celExpr := fmt.Sprintf("cel.bind(%s, device.attributes['%s'], %s.uuid=='%s'||%s.uuid=='%s')",
					deviceTaintRuleDevice, deviceTaintRuleCELLabelGPUNvidia,
					deviceTaintRuleDevice, cards[0].GPUUUID, deviceTaintRuleDevice, cards[1].GPUUUID)
				dtr := &resourcev1.DeviceTaintRule{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(dtrName, dtrNamespace), dtr)).To(Succeed())
				Expect(dtr.Spec.DeviceSelector.Selectors[0].CEL.Expression).To(Equal(celExpr))
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when auto replacement is disabled and the number of target MIG cards is over 1", func() {
			const caName = "test-ca-disable-auto-replace-multiple-mig-uuids"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
						MIGUUIDs: []string{
							"mig-uuid-1",
							"mig-uuid-2",
						},
					},
					{
						Model:   model,
						GPUUUID: uuids[1],
						MIGUUIDs: []string{
							"mig-uuid-3",
						},
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCreatingTaintResource(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			By("asserting the CEL expression in DeviceTaintRule", func() {
				ca := &scalev1.CardAlert{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
				dtrName := ca.Status.ResourceRef.Name
				dtrNamespace := ca.Status.ResourceRef.Namespace
				celExpr := fmt.Sprintf("cel.bind(%s, device.attributes['%s'], %s.parentUUID=='%s'||%s.parentUUID=='%s')",
					deviceTaintRuleDevice, deviceTaintRuleCELLabelGPUNvidia,
					deviceTaintRuleDevice, cards[0].GPUUUID, deviceTaintRuleDevice, cards[1].GPUUUID)
				dtr := &resourcev1.DeviceTaintRule{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(dtrName, dtrNamespace), dtr)).To(Succeed())
				Expect(dtr.Spec.DeviceSelector.Selectors[0].CEL.Expression).To(Equal(celExpr))
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when another resource is processing", func() {
			const caName = "test-ca-another-resource-processing"
			const caAnotherName = "test-ca-another-resource-processing-another"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			By("creating another CardAlert", func() {
				ca := testutils.NewCardAlert(caAnotherName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("asserting the 1st CardAlert until Initializing passed", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
			})
			// Reconcile
			By("asserting the 2nd CardAlert's phase should be Pending when another resource is still processing", func() {
				Expect(executeCaReconciler(r, caAnotherName, caNamespace, scalev1.CardAlertPendingState)).To(Succeed())
			})
			By("asserting the 1st CardAlert", func() {
				// Reconcile
				By("executing reconcile", func() {
					Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
					Expect(executeCaCreatingReplaceResource(r, caName, caNamespace)).To(Succeed())
					Expect(patchCardScalingRequestPhase(caName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
					Expect(executeCaWaitingCardScalingRequest(r, caName, caNamespace)).To(Succeed())
					Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
				})
				// Delete CardAlert
				Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
				// Finalize
				By("executing finalize", func() {
					Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
					Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
				})
			})
			By("asserting the 2nd CardAlert after the 1st CardAlert is completed", func() {
				// Reconcile
				By("executing reconcile", func() {
					Expect(executeCaPending(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaInitializing(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaCreatingReplaceResource(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(patchCardScalingRequestPhase(caAnotherName, caNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
					Expect(executeCaWaitingCardScalingRequest(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(executeCaCompleted(r, caAnotherName, caNamespace)).To(Succeed())
				})
				// Delete CardAlert
				Expect(deleteCardAlertResource(ctx, caAnotherName, caNamespace)).To(Succeed())
				// Finalize
				By("executing finalize", func() {
					Expect(finalizeCaCompleted(r, caAnotherName, caNamespace)).To(Succeed())
					Expect(finalizeCaFinalized(r, caAnotherName, caNamespace)).To(Succeed())
				})
			})
		})

		It("should successfully reconcile when Node is already deleted", func() {
			const caName = "test-ca-node-deleted"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
				Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the Node is already deleted", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when Node has no annotations for ClusterName,ClusterNamespace", func() {
			const caName = "test-ca-node-without-annotations"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get Cluster Namespace from Node",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nodeName, ""), node)).To(Succeed())
				node.Annotations = map[string]string{} // remove annotations for ClusterName and ClusterNamespace
				Expect(k8sClientWorkload.Update(ctx, node)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting error happens when Node has no annotations for ClusterName and ClusterNamespace", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should unsuccessfully reconcile when DCIMachine has no reference to LogicalServer", func() {
			const caName = "test-ca-dcimachine-without-logicalserver"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: uuids[0],
					},
				}
				source  = scalev1.WebhookReceiverOnTenant
				errMsgs = []string{
					"failed to get LogicalServer name and namespace from DCIMachine",
				}
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.RemoveDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				By("asserting error happens when DCIMachine has no reference to LogicalServer", func() {
					err := executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
					Expect(err).To(HaveOccurred())
					for _, errMsg := range errMsgs {
						Expect(err.Error()).To(ContainSubstring(errMsg))
					}
				})
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaInitializing(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when the target UUID is already excluded with enabling auto replacement", func() {
			const caName = "test-ca-gpu-uuid-excluded"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: "excluded-gpu-uuid",
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the target UUID is already excluded", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})

		It("should successfully reconcile when the target UUID is already excluded with disabling auto replacement", func() {
			const caName = "test-ca-gpu-uuid-excluded-auto-replace-disabled"
			var (
				cards = []scalev1.CardInfo{
					{
						Model:   model,
						GPUUUID: "excluded-gpu-uuid",
					},
				}
				source = scalev1.WebhookReceiverOnTenant
			)
			// Setup for this test case
			By("setting up resources for test case", func() {
				Expect(testutils.PatchLogicalServerType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			// Setup before reconcile
			r := newCaReconciler(nil, nil)
			By("setting up auto replacement disabled", func() {
				r.enableAutoReplace = false
			})
			By("creating CardAlert", func() {
				ca := testutils.NewCardAlert(caName, caNamespace, nodeName, cards, source)
				Expect(k8sClientWorkload.Create(ctx, ca)).To(Succeed())
			})
			// Reconcile
			By("executing reconcile", func() {
				Expect(executeCaInitial(r, caName, caNamespace)).To(Succeed())
				Expect(executeCaInitializing(r, caName, caNamespace)).To(Succeed())
				By("asserting the phase should become Completed if the target UUID is already excluded", func() {
					Expect(executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)).To(Succeed())
				})
				Expect(executeCaCompleted(r, caName, caNamespace)).To(Succeed())
			})
			// Delete CardAlert
			Expect(deleteCardAlertResource(ctx, caName, caNamespace)).To(Succeed())
			// Finalize
			By("executing finalize", func() {
				Expect(finalizeCaCompleted(r, caName, caNamespace)).To(Succeed())
				Expect(finalizeCaFinalized(r, caName, caNamespace)).To(Succeed())
			})
		})
	})
})

// newCaReconciler returns CardAlertReconciler with mock by one line.
func newCaReconciler(
	wClient, mClient client.Client,
) *CardAlertReconciler {
	By("creating CardAlertReconciler")
	controllerReconciler := (&CardAlertReconciler{
		Recorder:                      record.NewFakeRecorder(1024),
		enableTTLSecondsAfterFinished: false,
		ttlSecondsAfterFinished:       3600,
		MachineAlertNamespace:         rpmNamespace,
		enableAutoReplace:             true,
		cardTaintPolicy:               resourcev1.DeviceTaintEffectNoExecute,
	})

	// client and scheme for workload cluster
	if ent, ok := wClient.(*mock.MockClient); ok && ent != nil {
		By("setting mock client for workload cluster")
		controllerReconciler.Client = wClient
	} else {
		By("setting real client for workload cluster")
		controllerReconciler.Client = k8sClientWorkload
		controllerReconciler.Scheme = k8sClientWorkload.Scheme()
	}

	// client for management cluster
	var err error
	var opt *helperclient.ClientHelperOption
	if ent, ok := mClient.(*mock.MockClient); ok && ent != nil {
		By("setting mock client for management cluster")
		opt = &helperclient.ClientHelperOption{
			ManagementClusterClient: mClient,
		}
	} else {
		By("setting real client for management cluster")
		opt = &helperclient.ClientHelperOption{
			ManagementClusterClient: k8sClientManagement,
		}
	}
	controllerReconciler.ClientHelper, err = helperclient.NewClientHelper(
		controllerReconciler.Client,
		opt,
	)
	Expect(err).NotTo(HaveOccurred())

	return controllerReconciler
}

// executeCaReconciler executes Reconcile() and asserts CardAlert status phase.
func executeCaReconciler(
	r *CardAlertReconciler, caName, caNamespace string, nextPhase scalev1.CardAlertState,
) error {
	By("executing Reconcile, next phase to be " + string(nextPhase))
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: testutils.NewNamespacedName(caName, caNamespace),
	})

	By("asserting CardAlert status to be " + string(nextPhase))
	ca := &scalev1.CardAlert{}
	Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
	Expect(ca.Status.State).To(Equal(nextPhase))

	return err
}

func executeCaInitial(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertInitializingState)
}

func executeCaPending(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaInitial(r, caName, caNamespace)
}

func executeCaInitializing(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	switch r.enableAutoReplace {
	case true:
		return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingReplaceResourceState)
	case false:
		return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCreatingTaintResourceState)
	default:
		return fmt.Errorf("unexpected enableAutoReplace value: %v", r.enableAutoReplace)
	}
}

func executeCaCreatingReplaceResource(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	switch t := getLogicalServerType(caName, caNamespace); t {
	case rpmv1.AppliedServerTypeComposed:
		return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingCardScalingRequestState)
	case rpmv1.AppliedServerTypeCOTS:
		return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertWaitingMachineAlertState)
	default:
		return fmt.Errorf("unexpected logical server type: %v", t)
	}
}

func executeCaWaitingCardScalingRequest(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
}

func executeCaWaitingMachineAlert(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertDeletingMachineAlertState)
}

func executeCaDeletingMachineAlert(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
}

func executeCaCreatingTaintResource(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
}

func executeCaCompleted(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertCompletedState)
}

func executeCaFailed(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFailedState)
}

func finalizeCaInitial(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaInitializing(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaPending(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return finalizeCaInitializing(r, caName, caNamespace)
}

func finalizeCaCreatingReplaceResource(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaWaitingCardScalingRequest(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaWaitingMachineAlert(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaDeletingMachineAlert(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaCreatingTaintResource(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaCompleted(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaFailed(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	return executeCaReconciler(r, caName, caNamespace, scalev1.CardAlertFinalizedState)
}

func finalizeCaFinalized(
	r *CardAlertReconciler, caName, caNamespace string,
) error {
	By("executing Reconcile, next phase is nothing as CardScalingRequest will be deleted")
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: testutils.NewNamespacedName(caName, caNamespace),
	})
	if err != nil {
		return err
	}

	By("asserting CardScalingRequest is deleted")
	csr := &scalev1.CardScalingRequest{}
	err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), csr)
	Expect(apierrors.IsNotFound(err)).To(BeTrue())

	return nil
}

func deleteCardAlertResource(
	ctx context.Context, name, namespace string,
) (err error) {
	By("deleting CardAlert resource", func() {
		ca := &scalev1.CardAlert{}
		err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), ca)
		if apierrors.IsNotFound(err) {
			err = nil // already deleted
			return
		} else if err != nil {
			return
		}
		// if exists, delete it
		err = k8sClientWorkload.Delete(ctx, ca)
		if apierrors.IsNotFound(err) {
			err = nil // already deleted
		}
	})
	return
}

func getLogicalServerType(
	caName, caNamespace string,
) rpmv1.AppliedServerType {
	ca := &scalev1.CardAlert{}
	Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)).To(Succeed())
	lsName := ca.Status.LogicalServerRef.Name
	lsNamespace := ca.Status.LogicalServerRef.Namespace
	ls := &rpmv1.LogicalServer{}
	Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(lsName, lsNamespace), ls)).To(Succeed())
	Expect(ls.Status.AppliedSpec).NotTo(BeNil())
	return ls.Status.AppliedSpec.Type
}

func patchCardScalingRequestPhase(
	caName, caNamespace string, phase scalev1.CardScalingRequestPhase,
) (err error) {
	By("patching CardScalingRequest phase to "+string(phase), func() {
		// Get CardAlert to obtain ResourceRef for CardScalingRequest
		ca := &scalev1.CardAlert{}
		err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)
		if err != nil {
			return
		}
		// Get CardScalingRequest
		csrName := ca.Status.ResourceRef.Name
		csrNamespace := ca.Status.ResourceRef.Namespace
		csr := &scalev1.CardScalingRequest{}
		err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)
		if err != nil {
			return
		}
		// Patch phase
		csr.Status.Phase = phase
		err = k8sClientWorkload.Status().Update(ctx, csr)
	})
	return
}

//nolint:unparam // state always receives rpmv1.MachineAlertStateCompleted, but it may receive other states in the future, so keep the parameter for future use.
func patchMachineAlertState(
	caName, caNamespace string, state rpmv1.MachineAlertState,
) (err error) {
	By("patching MachineAlert state to "+string(state), func() {
		// Get CardAlert to obtain ResourceRef for MachineAlert
		ca := &scalev1.CardAlert{}
		err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(caName, caNamespace), ca)
		if err != nil {
			return
		}
		// Get MachineAlert
		maName := ca.Status.ResourceRef.Name
		maNamespace := ca.Status.ResourceRef.Namespace
		ma := &rpmv1.MachineAlert{}
		err = k8sClientManagement.Get(ctx, testutils.NewNamespacedName(maName, maNamespace), ma)
		if err != nil {
			return
		}
		// Patch state
		ma.Status.State = state
		err = k8sClientManagement.Status().Update(ctx, ma)
	})
	return
}
