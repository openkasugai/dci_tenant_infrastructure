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
	"slices"
	"time"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"

	"scale-controller/test/testutils"
	"scale-controller/test/testutils/mock"
)

// Test for SetupWithManager() method
var _ = Describe("NodeScalingRequest Controller Setting Up", Label("nsr", "settings", "ttl"), func() {
	Context("When setting up with manager", func() {
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
			logger = ctrl.Log.WithName("nodescalingrequest-controller-test")
		)

		BeforeEach(func() {
			Expect(os.Unsetenv(helperclient.ManagementKubeconfigEnvVarName)).To(Succeed())
			ctrl.NewControllerManagedBy = newControllerManagedByMock
		})
		AfterEach(func() {
			Expect(os.Unsetenv(utils.ProviderTypeEnvVarName)).To(Succeed())
			Expect(os.Unsetenv(nodeScalingRequestTTLEnvVarName)).To(Succeed())
			ctrl.NewControllerManagedBy = newControllerManagedBy
		})

		It("should successfully setup with enabling TTL", func() {
			// setting variables
			var (
				settingProvider = string(utils.ProviderTypeDCI)
				settingTTL      = "1"
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())
			Expect(os.Setenv(nodeScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			Expect(reconciler.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with disabling TTL", func() {
			// setting variables
			var (
				settingProvider = string(utils.ProviderTypeDCI)
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())
			Expect(os.Unsetenv(nodeScalingRequestTTLEnvVarName)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			Expect(reconciler.SetupWithManager(mgr)).To(Succeed())
		})

		It("should unsuccessfully setup with setting invalid empty provider name", func() {
			// setting variables
			var (
				settingProvider = ""
				errMsgs         = []string{
					"provider type is empty: ",
				}
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid provider name", func() {
			// setting variables
			var (
				settingProvider = "dummyProvider"
				errMsgs         = []string{
					fmt.Sprintf("provider type(%s) is invalid", settingProvider),
				}
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid empty NSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingTTL = ""
				errMsgs    = []string{
					"argument for setting TTL seconds is empty",
				}
			)

			// setting environment variables
			Expect(os.Setenv(nodeScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid character NSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingProvider = string(utils.ProviderTypeDocker)
				settingTTL      = "notNumber"
				errMsgs         = []string{
					"argument for setting TTL seconds cannot convert into int32",
				}
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())
			Expect(os.Setenv(nodeScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid minus NSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingProvider = string(utils.ProviderTypeDCI)
				settingTTL      = "-1"
				errMsgs         = []string{
					"argument for setting TTL seconds is invalid(" + settingTTL + ")",
				}
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())
			Expect(os.Setenv(nodeScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: mockClient})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting with nil ManagementClusterClient", func() {
			// setting variables
			var (
				settingProvider = string(utils.ProviderTypeDCI)
				settingTTL      = "3600"
				errMsgs         = []string{
					"expected ManagementClusterClient is not nil, but got nil",
				}
			)

			// setting environment variables
			Expect(os.Setenv(utils.ProviderTypeEnvVarName, settingProvider)).To(Succeed())
			Expect(os.Setenv(nodeScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: nil})
			err := reconciler.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting with nil manager", func() {
			// setting variables
			var (
				errMsgs = []string{
					"expected manager is not nil, but got nil",
				}
			)

			// executing SetupWithManager
			reconciler := (&NodeScalingRequestReconciler{ClientHelper: nil})
			err := reconciler.SetupWithManager(nil)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})
	})
})

// Test for internal methods with direct calls
var _ = Describe("NodeScalingRequest Controller Methods directly calling", Label("nsr", "methods"), func() {
	Context("When calling methods directly", func() {
		It("should unsuccessfully call startScaleUp with invalid NodeScalingRequest", func() {
			By("calling startScaleUp", func() {
				var errMsgs = []string{
					fmt.Sprintf("parameter is nil: .spec.location(%v) .spec.serverSpec(%v)", nil, nil),
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "create")
				_, err := reconciler.startScaleUp(ctx, nsr, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(nsr.Status.Phase).To(Equal(scalev1.PhaseNodeCreateFailed))
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
		})

		It("should unsuccessfully call startScaleUp with invalid provider type", func() {
			By("calling startScaleUp", func() {
				var (
					dummyProviderType = utils.ProviderType("dummyProvider")
					errMsgs           = []string{
						fmt.Sprintf("provider type(%s) is invalid", dummyProviderType),
					}
				)
				reconciler := newNodeScalingRequestReconciler(dummyProviderType, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				_, err := reconciler.startScaleUp(ctx, nsr, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "InvalidProviderType", errMsgs)
			})
		})

		It("should unsuccessfully call startScaleUpWithCreatingResource with invalid MachineDeployment", func() {
			By("calling startScaleUpWithCreatingResource", func() {
				var errMsgs = []string{
					fmt.Sprintf("parameter is nil: mdBase(%v)", nil),
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-md", defaultNamespace, "create")
				_, err := reconciler.startScaleUpWithCreatingResource(ctx, nsr, nsr, nil, nil, "", "")
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(nsr.Status.Phase).To(Equal(scalev1.PhaseNodeCreateFailed))
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
		})

		It("should unsuccessfully call startScaleDown with invalid NodeScalingRequest", func() {
			By("calling startScaleDown", func() {
				var errMsgs = []string{
					fmt.Sprintf("parameter is nil: .spec.node(%v)", nil),
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "delete")
				_, err := reconciler.startScaleDown(ctx, nsr, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(nsr.Status.Phase).To(Equal(scalev1.PhaseNodeDeleteFailed))
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
		})

		It("should unsuccessfully call proceedScaleUp with invalid NodeScalingRequest", func() {
			By("calling proceedScaleUp", func() {
				var errMsgs = []string{
					fmt.Sprintf("parameter is nil: .status.desiredReplicas(%v)", nil),
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "create")
				_, err := reconciler.proceedScaleUp(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(nsr.Status.Phase).To(Equal(scalev1.PhaseNodeCreateFailed))
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
		})

		It("should unsuccessfully call proceedScaleDown with invalid NodeScalingRequest", func() {
			By("calling proceedScaleDown", func() {
				var errMsgs = []string{
					"spec.node is nil",
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "delete")
				_, err := reconciler.proceedScaleDown(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(nsr.Status.Phase).To(Equal(scalev1.PhaseNodeDeleteFailed))
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
		})

		It("should unsuccessfully call cleanUpAfterFinished with invalid NodeScalingRequest", func() {
			By("calling cleanUpAfterFinished", func() {
				var errMsgs = []string{
					"invalid NodeScalingRequest: should contain Condition with 'Completed'",
				}
				reconciler := newNodeScalingRequestReconciler(utils.ProviderTypeDCI, nil, nil)
				nsr := testutils.NewNodeScalingRequest("test-invalid-nsr", defaultNamespace, "delete")
				_, err := reconciler.cleanUpAfterFinished(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call deleteMachineTemplateUnused with invalid provider type", func() {
			By("calling deleteMachineTemplateUnused", func() {
				var (
					dummyProviderType = utils.ProviderType("dummyProvider")
					errMsgs           = []string{
						fmt.Sprintf("provider type(%s) is invalid", dummyProviderType),
					}
				)
				reconciler := newNodeScalingRequestReconciler(dummyProviderType, nil, nil)
				err := reconciler.deleteMachineTemplateUnused(ctx, nil)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})

	Context("When calling methods directly with client", func() {
		const (
			nsrNamespace = defaultNamespace
		)

		AfterEach(func() {
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
		})

		It("should unsuccessfully call startScaleUpWithCreatingResource with invalid provider type", func() {
			By("calling startScaleUpWithCreatingResource", func() {
				const nsrName = "test-invalid-md-with-client"
				var (
					dummyProviderType = utils.ProviderType("dummyProvider")
					errMsgs           = []string{
						fmt.Sprintf("provider type(%s) is invalid", dummyProviderType),
					}
				)
				By("creating NodeScalingRequest", func() {
					nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
					nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
					Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
				})
				reconciler := newNodeScalingRequestReconciler(dummyProviderType, nil, nil)
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				_, err := reconciler.startScaleUpWithCreatingResource(ctx, nsr, nsr, &capiv1.MachineDeployment{}, nil, "", "")
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "InvalidProviderType", errMsgs)
			})
		})
	})
})

// Test for kubebuilder markers validation in CRD
var _ = Describe("NodeScalingRequest Controller embedded Validation in CRD", Label("nsr", "crd"), func() {
	Context("When creating NodeScalingRequest with invalid parameters", func() {
		// error messages
		const (
			pathSpec                         = "spec"
			pathStatus                       = "status"
			pathSpecNodeRequestType          = "nodeRequestType"
			pathSpecNode                     = "node"
			pathSpecLocation                 = "location"
			pathSpecServerSpec               = "serverSpec"
			pathServerSpecType               = "type"
			pathServerSpecOS                 = "os"
			pathServerSpecOSDistribution     = "distribution"
			pathServerSpecOSVersion          = "version"
			pathServerSpecMinCPU             = "minCPU"
			pathServerSpecMinMemory          = "minMemory"
			pathServerSpecMinStorage         = "minStorage"
			pathServerSpecGPU                = "gpu"
			pathServerSpecGPUModel           = "model"
			pathServerSpecGPUNum             = "num"
			pathServerSpecGPUAssignments     = "assignments"
			pathServerSpecGPUAssignmentsUUID = "uuid"
			pathServerSpecNIC                = "nic"
			pathServerSpecNICModel           = "model"
			pathServerSpecNICNum             = "num"
			pathStatusPhase                  = "phase"
		)
		const (
			errMsgCreateInvalidSet   = pathSpecNodeRequestType + " is create, but " + pathSpec + "." + pathSpecNode + " is set"
			errMsgCreateInvalidUnset = pathSpecNodeRequestType + " is create, but " + pathSpec + "." + pathSpecServerSpec + " is not set"
			errMsgDeleteInvalidSetS  = pathSpecNodeRequestType + " is delete, but " + pathSpec + "." + pathSpecServerSpec + " is set"
			errMsgDeleteInvalidSetL  = pathSpecNodeRequestType + " is delete, but " + pathSpec + "." + pathSpecLocation + " is set"
			errMsgDeleteInvalidUnset = pathSpecNodeRequestType + " is delete, but " + pathSpec + "." + pathSpecNode + " is not set"
			errMsgTypeRequired       = pathSpec + "." + pathSpecNodeRequestType + ": Required value"
			errMsgTypeUnsupported    = pathSpec + "." + pathSpecNodeRequestType + ": Unsupported value"
			errMsgPhaseUnsupported   = pathStatus + "." + pathStatusPhase + ": Unsupported value"
			errMsgModifyExisting     = "`" + pathSpec + "` may not be modified after creation."
		)
		// common const
		const (
			nsrNamespace = defaultNamespace
		)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
		})

		newNodeScalingRequestObjMap := func(name, namespace string, reqType scalev1.NodeScalingRequestType) map[string]any {
			return map[string]any{
				"apiVersion": scalev1.GroupVersion.String(),
				"kind":       "NodeScalingRequest",
				"metadata": map[string]any{
					"name":      name,
					"namespace": namespace,
				},
				pathSpec: map[string]any{
					pathSpecNodeRequestType: reqType,
					pathSpecNode:            "dummy-node",
					pathSpecLocation:        "dummy-location",
					pathSpecServerSpec: map[string]any{
						pathServerSpecType: rpmv1.ReqServerTypeComposed,
						pathServerSpecOS: map[string]any{
							pathServerSpecOSDistribution: "dummy-distribution",
							pathServerSpecOSVersion:      "dummy-version",
						},
						pathServerSpecMinCPU:     10,
						pathServerSpecMinMemory:  192,
						pathServerSpecMinStorage: 960,
						pathServerSpecGPU: map[string]any{
							pathServerSpecGPUModel: "connect-x5",
							pathServerSpecGPUNum:   1,
						},
						pathServerSpecNIC: map[string]any{
							pathServerSpecNICModel: "dummy-nic-model",
							pathServerSpecNICNum:   1,
						},
					},
				},
			}
		}

		It("should fail to create NodeScalingRequest with Type(create) and node", func() {
			const nsrName = "test-validation-create-with-node"
			var (
				errMsgs = []string{
					errMsgCreateInvalidSet,
				}
			)
			By("creating NodeScalingRequest", func() {
				caObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecServerSpec)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(create) without serverSpec", func() {
			const nsrName = "test-validation-create-without-serverspec"
			var (
				errMsgs = []string{
					errMsgCreateInvalidUnset,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecNode)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecServerSpec)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(create) and Node without serverSpec", func() {
			const nsrName = "test-validation-create-without-serverspec-with-node"
			var (
				errMsgs = []string{
					errMsgCreateInvalidSet,
					errMsgCreateInvalidUnset,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecServerSpec)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete) and serverSpec", func() {
			const nsrName = "test-validation-delete-with-serverspec"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidSetS,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecLocation)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete) and location", func() {
			const nsrName = "test-validation-delete-with-location"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidSetL,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecServerSpec)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete) without node", func() {
			const nsrName = "test-validation-delete-without-node"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidUnset,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecNode)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecServerSpec)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecLocation)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete) and serverSpec without node", func() {
			const nsrName = "test-validation-delete-with-serverspec-without-node"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidUnset,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecNode)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecLocation)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete) and location without node", func() {
			const nsrName = "test-validation-delete-with-location-without-node"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidUnset,
					errMsgDeleteInvalidSetL,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecNode)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecServerSpec)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with Type(delete), location and serverSpec without node", func() {
			const nsrName = "test-validation-delete-with-location-serverspec-without-node"
			var (
				errMsgs = []string{
					errMsgDeleteInvalidUnset,
					errMsgDeleteInvalidSetS,
					errMsgDeleteInvalidSetL,
				}
			)
			By("creating NodeScalingRequest", func() {
				nsrObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeDelete)
				unstructured.RemoveNestedField(nsrObj, pathSpec, pathSpecNode)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: nsrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest without Type", func() {
			const nsrName = "test-validation-type-nothing"
			var (
				errMsgs = []string{
					errMsgTypeRequired,
				}
			)
			By("creating NodeScalingRequest", func() {
				caObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecNodeRequestType)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create NodeScalingRequest with invalid Type", func() {
			const nsrName = "test-validation-type-invalid"
			var (
				reqType = scalev1.NodeScalingRequestType("invalid")
				errMsgs = []string{
					errMsgTypeUnsupported,
				}
			)
			By("creating NodeScalingRequest", func() {
				caObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, reqType)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update NodeScalingRequest spec", func() {
			const nsrName = "test-validation-update-spec"
			var (
				errMsgs = []string{
					errMsgModifyExisting,
				}
			)
			By("creating NodeScalingRequest", func() {
				caObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecNode)
				Expect(k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})).To(Succeed())
			})
			By("updating NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				nsr.Spec.NodeRequestType = scalev1.TypeDelete
				nsr.Spec.Node = ptr.To("dummy-node")
				err := k8sClientWorkload.Update(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update NodeScalingRequest with setting status.phase not expected string", func() {
			const nsrName = "test-validation-set-status-spec-invalid"
			var (
				errMsgs = []string{
					errMsgPhaseUnsupported,
				}
			)
			By("creating NodeScalingRequest", func() {
				caObj := newNodeScalingRequestObjMap(nsrName, nsrNamespace, scalev1.TypeCreate)
				unstructured.RemoveNestedField(caObj, pathSpec, pathSpecNode)
				Expect(k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: caObj})).To(Succeed())
			})
			By("updating NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = "invalidPhase"
				err := k8sClientWorkload.Status().Update(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})
})

// Test for mapperNodeToNodeScalingRequest() method
var _ = Describe("NodeScalingRequest Controller mapper method for Node Event", Label("nsr", "mapper", "node"), func() {
	const providerType = utils.ProviderTypeDCI

	Context("When calling mapper method", func() {
		// common const
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			machineName      = "fake-machine-00"
			nodeName         = "fake-node-00"
		)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
		})

		It("should succeed to create NodeScalingRequest requests with zero length", func() {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing mapper for Node", func() {
				reqs := controllerReconciler.mapperNodeToNodeScalingRequest(ctx, nil)
				Expect(reqs).To(BeEmpty())
			})
		})

		It("should succeed to create NodeScalingRequest requests with not finished NodeScalingRequest", func() {
			const nsrName = "test-handler-node-event-1-not-finished"
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing mapper for Node", func() {
				reqs := controllerReconciler.mapperNodeToNodeScalingRequest(ctx, nil)
				Expect(reqs).To(HaveLen(1))
			})
		})

		It("should succeed to create NodeScalingRequest requests with finished NodeScalingRequest", func() {
			const (
				nsrNotFinishedName = "test-handler-node-event-2-not-finished"
				nsrFinishedName    = "test-handler-node-event-2-finished"
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrNotFinishedName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrFinishedName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("patching NodeScalingRequest to be finished", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrFinishedName, nsrNamespace), nsr)).To(Succeed())
				nsrLast := nsr.DeepCopy()
				nsr.Status.Phase = scalev1.PhaseNodeCreated
				Expect(k8sClientWorkload.Status().Patch(ctx, nsr, client.MergeFrom(nsrLast))).To(Succeed())
			})
			By("executing mapper for Node", func() {
				reqs := controllerReconciler.mapperNodeToNodeScalingRequest(ctx, nil)
				Expect(reqs).To(HaveLen(1))
				Expect(reqs[0].Name).To(Equal(nsrNotFinishedName))
			})
		})
	})

	Context("When calling mapper method", Label("mock"), func() {
		It("should succeed to create NodeScalingRequest requests with zero length", func() {
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetList([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("executing mapper for Node", func() {
				reqs := controllerReconciler.mapperNodeToNodeScalingRequest(ctx, nil)
				Expect(reqs).To(BeNil())
			})
		})
	})
})

// Test for Reconcile() method
var _ = Describe("NodeScalingRequest Controller", Label("nsr", "controller"), func() {
	const providerType = utils.ProviderTypeDCI
	mdTemplateAPIGroup := capiv1.GroupVersion.Group
	mdTemplateKind := templateKind

	Context("When reconciling a resource scaling up", func() {
		const (
			nsrNamespace       = defaultNamespace
			clusterName        = "fake-cluster-00"
			clusterNamespace   = defaultNamespace
			mdName             = "fake-md-00"
			mdNamespace        = defaultNamespace
			dmtNameNoResource  = "fake-dci-machine-template-00" // no resource
			dmtNameAllResource = "fake-dci-machine-template-01" // all resource
			dmtNamespace       = defaultNamespace
			machineName        = "fake-machine-00"
			machineNamespace   = defaultNamespace
			nodeName           = "fake-node-00"
		)
		var (
			replicas              = int32(1)
			location              = "DC1"
			os                    = rpmv1.OSType{Distribution: "ubuntu", Version: "24.04"}
			minCPU                = int32(10)
			minMemory             = int32(10)
			minStorage            = int32(10)
			gpu                   = rpmv1.GPUSpec{Model: "H100", Num: 1}
			nic                   = rpmv1.NICSpec{Model: "ConnectX-5", Num: 1}
			serverSpecNoResource  = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
			serverSpecAllResource = testutils.NewServerSpec(rpmv1.ReqServerTypeComposed, &os, &minCPU, &minMemory, &minStorage, &gpu, &nic)
			mdNames               = []string{mdName}
			dmtNames              = []string{dmtNameNoResource, dmtNameAllResource}
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtNameNoResource, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameNoResource, dmtNamespace, nil, serverSpecNoResource)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameAllResource, dmtNamespace, &location, serverSpecAllResource)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up existing MachineDeployment", func() {
			const nsrName = "test-scale-up-existing-md-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupExistingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, mdNamespace, replicas+1)
		})

		It("should successfully reconcile the resource creating new MachineDeployment with existing DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-with-existing-template"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = serverSpecAllResource
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNameAllResource, replicas)
		})

		It("should successfully reconcile the resource creating new MachineDeployment and DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-and-template"
			ss := serverSpecAllResource.DeepCopy()
			ss.Type = rpmv1.ReqServerTypeCOTS
			loc := location + "-new"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &loc
				nsr.Spec.ServerSpec = ss
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingDCIMachineTemplate(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNames, dmtNamespace, loc, ss, replicas)
		})

		It("should successfully reconcile the resource scaling up existing MachineDeployment with nil replicas", func() {
			const nsrName = "test-scale-up-existing-md-with-nil-replicas-success"
			Expect(testutils.PatchMachineDeploymentReplicas(ctx, k8sClientManagement, mdName, mdNamespace, nil, nil)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupExistingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, mdNamespace, 1)
		})
	})

	Context("When reconciling a resource scaling down", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var (
			replicas   = int32(1)
			serverSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, serverSpec)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down", func() {
			const nsrName = "test-scale-down-success"
			Expect(testutils.PatchMachineDeploymentReplicas(ctx, k8sClientManagement, mdName, mdNamespace, nil, ptr.To(int32(1)))).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, replicas)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-success"
				mdDummyName   = "fake-md-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment and DCIMachineTemplate", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-dmt-success"
				mdDummyName   = "fake-md-dummy"
				dmtDummyName  = "fake-dci-machine-template-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtDummyName, dmtNamespace, nil, serverSpec)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})

		It("should successfully reconcile the resource scaling down without deleting MachineDeployment because of last MachineDeployment", func() {
			const nsrName = "test-scale-down-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, replicas-1)
		})
	})

	Context("When reconciling a resource scaling up with checking various serverSpec,location", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtNamePrefix    = "fake-dmt-"
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas         = int32(1)
			locationValid    = "validLocation"
			locationDummy    = "dummyLocation"
			minParamsValid   = int32(10)
			minParamsInvalid = int32(5)
			locations        = []*string{
				&locationValid,
				&locationDummy,
				nil,
			}
			types = []rpmv1.ReqServerType{
				rpmv1.ReqServerTypeCOTS,
				rpmv1.ReqServerTypeComposed,
			}
			oss = []*rpmv1.OSType{
				{Distribution: "Ubuntu", Version: "22.04"},
				{Distribution: "Dummy", Version: "Dummy"},
				nil,
			}
			mins = []*int32{
				&minParamsValid,
				&minParamsInvalid,
				nil,
			}
			gpus = []*rpmv1.GPUSpec{
				{Model: "A100", Num: 2},
				{Model: "A100", Num: 1},
				{Model: "Dummy", Num: 1},
				nil,
			}
			nics = []*rpmv1.NICSpec{
				{Model: "ConnectX-5", Num: 1},
				{Model: "Dummy", Num: 1},
				nil,
			}
			templateNum = len(locations[1:]) + len(types[1:]) + len(oss[1:]) + len(mins[1:]) + len(mins[1:]) + len(mins[1:]) + len(gpus[1:]) + len(nics[1:])
			mdNames     = []string{
				mdName,
			}
			dmtNames []string
		)
		for cnt := range templateNum {
			dmtNames = append(dmtNames, dmtNamePrefix+fmt.Sprintf("%02d", cnt+1))
		}

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			var (
				l      = locations[0]
				t      = types[0]
				osType = oss[0]
				cpu    = mins[0]
				mem    = mins[0]
				sto    = mins[0]
				gpu    = gpus[0]
				nic    = nics[0]
			)
			cnt := 0
			for _, l := range locations[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, t := range types[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, os := range oss[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, os, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, cpu := range mins[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, mem := range mins[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, sto := range mins[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, gpu := range gpus[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
			for _, nic := range nics[1:] {
				cnt++
				dmtNameTarget := dmtNamePrefix + fmt.Sprintf("%02d", cnt)
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dmtNamespace, l, testutils.NewServerSpec(t, osType, cpu, mem, sto, gpu, nic))).To(Succeed())
			}
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up by creating new MachineDeployment and DCIMachineTemplate", func() {
			const nsrName = "test-scale-up-without-valid-dmt-success"
			var (
				serverSpec      = testutils.NewServerSpec(types[0], oss[0], mins[0], mins[0], mins[0], gpus[0], nics[0])
				desiredReplicas = int32(1)
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = locations[0]
				nsr.Spec.ServerSpec = serverSpec
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, dmtNamespace, locations[0], serverSpec, nsrNamespace, nsrName)
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameCreated, dmtNamespace)).To(Succeed())
		})
	})

	Context("When reconciling a resource scaling up while MachineDeployment used by other NodeScalingRequest", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas   = int32(1)
			serverSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, serverSpec)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up", func() {
			const nsrName = "test-scale-up-failed-while-md-used-other-nsr"
			var (
				desiredReplicas = replicas + 1
				dummyUID        = "dummy-uid"
				errMsgs         = []string{
					fmt.Sprintf("MachineDeployment (%s) is in progress by other NodeScalingRequest (UID=%s)", mdName, dummyUID),
				}
			)
			By("patching MachineDeployment with other UID", func() {
				Expect(testutils.PatchMachineDeploymentAddInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace, dummyUID)).To(Succeed())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(dummyUID))
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(dummyUID))
			})
			Expect(testutils.PatchMachineDeploymentRemoveInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			// patch replicas
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("When reconciling a resource scaling down while MachineDeployment used by other NodeScalingRequest", func() {
		const (
			clusterName      = "fake-cluster-01"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-01"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-01"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-01"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-01"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-01"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down after other NodeScalingRequest scaling finished", func() {
			const nsrName = "test-scale-down-success"
			var (
				desiredReplicas = replicas - 1
				dummyUID        = "dummy-uid"
				errMsgs         = []string{
					fmt.Sprintf("MachineDeployment (%s) is in progress by other NodeScalingRequest (UID=%s)", mdName, dummyUID),
				}
			)
			By("patching MachineDeployment with other UID", func() {
				Expect(testutils.PatchMachineDeploymentAddInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace, dummyUID)).To(Succeed())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(dummyUID))
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(dummyUID))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			Expect(testutils.PatchMachineDeploymentRemoveInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			// patch replicas
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			// cleanup node
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("When reconciling a resource scaling down without valid resources", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		executeNsrScaledownWithoutValidResource := func(
			name, namespace, mdName, reason string, desiredReplicas *int32,
		) {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("patching MachineDeployment ready replicas to simulate scale down", func() {
				if mdName != "" {
					Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
				}
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		}

		executeNsrScaledownWithoutValidResourceFinalize := func(
			name, namespace, mdName, reason string, desiredReplicas *int32,
		) {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		}

		It("should successfully reconcile the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, "", "MachineAlreadyDeleted", nil)
		})

		It("should successfully reconcile the resource scaling down if MachineDeployment not exists", func() {
			const nsrName = "test-scale-down-without-valid-md-fail"
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, "", "MachineDeploymentAlreadyDeleted", nil)
		})

		It("should successfully reconcile the resource scaling down if MachineSet not exists", func() {
			const nsrName = "test-scale-down-without-ms"
			Expect(testutils.CleanupMachineSetResource(ctx, k8sClientManagement, msName, msNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, mdName, "ScalingRequestSucceeded", ptr.To(replicas-1))
		})

		It("should successfully finalize the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, "", "MachineAlreadyDeleted", nil)
		})

		It("should successfully finalize the resource scaling down if MachineDeployment not exists", func() {
			const nsrName = "test-scale-down-without-valid-md-fail"
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, "", "MachineDeploymentAlreadyDeleted", nil)
		})

		It("should successfully finalize the resource scaling down if MachineSet not exists", func() {
			const nsrName = "test-scale-down-without-ms"
			Expect(testutils.CleanupMachineSetResource(ctx, k8sClientManagement, msName, msNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, mdName, "ScalingRequestSucceeded", ptr.To(replicas-1))
		})

		It("should successfully reconcile the resource scaling down if DCIMachineTemplate not exists", func() {
			const nsrName = "test-scale-down-without-dci-machine-template"
			const mdDummyName = "fake-md-dummy"
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdDummyName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), ptr.To(int32(0)))
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("skipping to simulate scale down since MachineDeployment not exists")
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), ptr.To(int32(0)))
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling down without simulating cluster-api partially", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			var (
				desiredReplicas = replicas - 1
			)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "Requeueing", nil)
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource while another resource is running", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		executeNsrParallel := func(
			nsrRunningName, nsrPendingName string, phase scalev1.NodeScalingRequestPhase, desiredReplicas int32,
		) {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop for "+nsrRunningName, func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrRunningName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest: "+nsrRunningName, func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrRunningName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, phase, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile loop for "+nsrPendingName, func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrPendingName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest: "+nsrPendingName, func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrPendingName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScalePending, "", "", ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionFalse, "Pending", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Requested)).To(BeNil())
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest: "+nsrRunningName, func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrRunningName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer for "+nsrRunningName, func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrRunningName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("cleaning up NodeScalingRequest: "+nsrPendingName, func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrPendingName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer for "+nsrPendingName, func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrPendingName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if running NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrRunningName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("checking if failure NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrPendingName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		}

		It("should successfully reconcile the resource scaling up while scaling up", func() {
			const (
				nsrRunningName = "test-scale-up-running-success"
				nsrPendingName = "test-scale-up-failure-fail"
			)
			By("creating NodeScalingRequest: "+nsrRunningName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrRunningName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("creating NodeScalingRequest: "+nsrPendingName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrPendingName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrParallel(nsrRunningName, nsrPendingName, scalev1.PhaseNodeCreating, replicas+1)

		})

		It("should successfully reconcile the resource scaling down while scaling up", func() {
			const (
				nsrRunningName = "test-scale-up-running-success"
				nsrPendingName = "test-scale-down-failure-fail"
			)
			By("creating NodeScalingRequest: "+nsrRunningName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrRunningName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("creating NodeScalingRequest: "+nsrPendingName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrPendingName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrParallel(nsrRunningName, nsrPendingName, scalev1.PhaseNodeCreating, replicas+1)
		})

		It("should successfully reconcile the resource scaling up while scaling down", func() {
			const (
				nsrRunningName = "test-scale-down-running-success"
				nsrPendingName = "test-scale-up-failure-fail"
			)
			By("creating NodeScalingRequest: "+nsrRunningName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrRunningName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("creating NodeScalingRequest: "+nsrPendingName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrPendingName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrParallel(nsrRunningName, nsrPendingName, scalev1.PhaseNodeDeleting, replicas-1)
		})

		It("should successfully reconcile the resource scaling down while scaling down", func() {
			const (
				nsrRunningName = "test-scale-down-running-success"
				nsrPendingName = "test-scale-down-failure-fail"
			)
			By("creating NodeScalingRequest: "+nsrRunningName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrRunningName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("creating NodeScalingRequest: "+nsrPendingName, func() {
				nsr := testutils.NewNodeScalingRequest(nsrPendingName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrParallel(nsrRunningName, nsrPendingName, scalev1.PhaseNodeDeleting, replicas-1)
		})
	})

	Context("When reconciling a resource with nodes without cluster annotation", func() {
		const (
			clusterName      = ""
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			By("deleting annotation from Node", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, types.NamespacedName{Name: nodeName}, node)).To(Succeed())
				nodePatched := node.DeepCopy()
				delete(nodePatched.Annotations, capiv1.ClusterNameAnnotation)
				Expect(k8sClientWorkload.Patch(ctx, nodePatched, client.MergeFrom(node))).To(Succeed())
			})
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up", func() {
			const nsrName = "test-scale-up-with-nodes-without-annotation-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to get cluster namespace: no node found with annotations %s and %s", capiv1.ClusterNameAnnotation, capiv1.ClusterNamespaceAnnotation),
				}
				mdNameEmpty = ""
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred()) // error is nil
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, "", "", &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionFalse, "GettingClusterInfoFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Requested)).To(BeNil())
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource with unknown finalizer", func() {
		const (
			nsrNamespace       = defaultNamespace
			clusterName        = "fake-cluster-00"
			clusterNamespace   = defaultNamespace
			mdName             = "fake-md-00"
			mdNamespace        = defaultNamespace
			dmtNameNoResource  = "fake-dci-machine-template-00" // no resource
			dmtNameAllResource = "fake-dci-machine-template-01" // all resource
			dmtNamespace       = defaultNamespace
			machineName        = "fake-machine-00"
			machineNamespace   = defaultNamespace
			nodeName           = "fake-node-00"
		)
		var (
			replicas              = int32(1)
			location1             = "DC1"
			location2             = "DC2"
			os                    = rpmv1.OSType{Distribution: "ubuntu", Version: "24.04"}
			minCPU                = int32(10)
			minMemory             = int32(10)
			minStorage            = int32(10)
			gpu                   = rpmv1.GPUSpec{Model: "H100", Num: 1}
			nic                   = rpmv1.NICSpec{Model: "ConnectX-5", Num: 1}
			serverSpecNoResource  = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
			serverSpecAllResource = testutils.NewServerSpec(rpmv1.ReqServerTypeComposed, &os, &minCPU, &minMemory, &minStorage, &gpu, &nic)
			mdNames               = []string{mdName}
			dmtNames              = []string{dmtNameNoResource, dmtNameAllResource}
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtNameNoResource, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameNoResource, dmtNamespace, nil, serverSpecNoResource)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameAllResource, dmtNamespace, &location1, serverSpecAllResource)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource with unknown finalizer", func() {
			const nsrName = "test-unknown-finalizer"
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			var (
				desiredReplicas = int32(1)
				serverSpec      *rpmv1.ServerSpec
			)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &location2
				serverSpec = serverSpecAllResource.DeepCopy()
				serverSpec.Type = rpmv1.ReqServerTypeCOTS
				nsr.Spec.ServerSpec = serverSpec
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("adding unknown finalizer to NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				nsrSaved := nsr.DeepCopy()
				controllerutil.AddFinalizer(nsr, "unknown/finalizer")
				Expect(k8sClientWorkload.Patch(ctx, nsr, client.MergeFrom(nsrSaved))).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			// get md name and template name created in reconcile loop
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, mdNamespace, &location2, serverSpec, nsrNamespace, nsrName)
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			// patch replicas
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, clusterNamespace, 0)).To(Succeed())
			By("executing reconcile loop with patching MachineDeployment", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("executing reconcile finalizer to check unknown finalizer exists", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(Succeed()) // succeed to get because unknown finalizer exists
				Expect(nsr.Finalizers).To(ContainElement("unknown/finalizer"))
			})
			By("removing unknown finalizer", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				nsrSaved := nsr.DeepCopy()
				nsr.Finalizers = nil
				Expect(k8sClientWorkload.Patch(ctx, nsr, client.MergeFrom(nsrSaved))).To(Succeed())
			})
		})
	})

	Context("When reconciling called with no valid resource", func() {
		const (
			nsrNamespace = defaultNamespace
		)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
		})

		It("should fail to get a resource when no resource exist", func() {
			const nsrName = "test-not-exist"
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred()) // when not found has occurred, should return nil as error
			})
		})

		It("should fail to get a resource when only other name resource exist", func() {
			const (
				nsrNameDummy    = "test-dummy"
				nsrNameNotExist = "test-not-exist"
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrNameDummy, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrNameNotExist, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred()) // when not found has occurred, should return nil as error
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrNameDummy, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrNameDummy, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("When reconciling a resource with no Nodes", func() {
		It("should unsuccessfully reconcile the resource scaling up", func() {
			Skip("This case cannot be reached because webhook block the data in the real cluster")
		})

		It("should unsuccessfully reconcile the resource scaling down", func() {
			Skip("This case cannot be reached because webhook block the data in the real cluster")
		})
	})

	Context("When reconciling a resource scaling up with failure, but recover process", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up if valid md is created after failure", func() {
			const nsrName = "test-scale-up-with-adding-md-success"
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment list to scale up",
					fmt.Sprintf("expected at least one MachineDeployment in the target cluster(%s/%s), but got none", clusterNamespace, clusterName),
				}
				desiredReplicas = replicas + 1
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "TargetResourceNotFound", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("When reconciling a resource scaling up with failure and finalize resources", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if valid md not exist, but finalize resources", func() {
			const (
				nsrName                  = "test-scale-up-with-no-valid-md-fail"
				clusterNameInvalid       = "fake-cluster-01-invalid"
				mdNameInvalidKind        = "fake-md-01-invalid"
				mdNameInvalidClusterName = "fake-md-02-invalid"
			)
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment list to scale up",
					fmt.Sprintf("expected at least one MachineDeployment in the target cluster(%s/%s), but got none", clusterNamespace, clusterName),
				}
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdNameInvalidKind, mdNamespace, clusterName, capdockerv1.GroupVersion.Group, templateKindDocker, dmtName, &replicas)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdNameInvalidClusterName, mdNamespace, clusterNameInvalid, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupFailureStartScaleup(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "TargetResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling up if md has autoscaler min size annotation", func() {
			const nsrName = "test-scale-up-with-md-autoscaler-min-size-fail"
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment list to scale up",
					fmt.Sprintf("expected at least one MachineDeployment in the target cluster(%s/%s), but got none", clusterNamespace, clusterName),
				}
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			By("patching MachineDeployment to add autoscaler min size annotation", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMinSizeAnnotation: "1",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupFailureStartScaleup(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "TargetResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling up if md has autoscaler max size annotation", func() {
			const nsrName = "test-scale-up-with-md-autoscaler-max-size-fail"
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment list to scale up",
					fmt.Sprintf("expected at least one MachineDeployment in the target cluster(%s/%s), but got none", clusterNamespace, clusterName),
				}
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			By("patching MachineDeployment to add autoscaler max size annotation", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMaxSizeAnnotation: "1",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupFailureStartScaleup(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "TargetResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling up if valid md deleted during proceeding scale up, but finalize resources", func() {
			const nsrName = "test-scale-up-proceeding-without-valid-md-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("not found target MachineDeployment(%s) in clusterNamespace(%s)", mdName, clusterNamespace),
				}
				insertFunc = func() {
					Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
				}
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupFailureProceedScaleup(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(mdName), ptr.To(replicas+1), errMsgs, "ScalingRequestSucceeded", "RequestingClientFailed", insertFunc)
		})
	})

	Context("When reconciling a resource scaling down with failure and finalize resources", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var replicas = int32(1)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down if md has invalid clusterName, but finalize resources", func() {
			const (
				nsrName          = "test-scale-down-with-invalid-cluster-name-md-fail"
				clusterDummyName = "fake-cluster-dummy"
			)
			var (
				errMsgs = []string{
					fmt.Sprintf("not found valid clusterName(%s) for MachineDeployment(%s)", clusterName, mdName),
				}
			)
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterDummyName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownFailureStartScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "ValidResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling down if md has invalid Replicas(nil), but finalize resources", func() {
			const nsrName = "test-scale-down-with-invalid-replicas-nil-md-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("not found valid Replicas for MachineDeployment(%s): Replicas is nil", mdName),
				}
			)
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, nil)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownFailureStartScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "ValidResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling down if md has invalid Replicas(too small), but finalize resources", func() {
			const nsrName = "test-scale-down-with-invalid-replicas-too-small-md-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("not found valid Replicas for MachineDeployment(%s), Replicas=%d", mdName, 0),
				}
				replicasTooSmall = int32(0)
			)
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicasTooSmall)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownFailureStartScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "ValidResourceNotFound")
		})

		It("should unsuccessfully reconcile the resource scaling down if node has no machine name", func() {
			const nsrName = "test-scale-down-without-valid-node-annotation-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request Machine resource in clusterNamespace(%s) through client", clusterNamespace),
					"failed to get Machine resource",
					fmt.Sprintf("failed to get resource(%s/%s) of kind Machine", clusterNamespace, ""),
				}
			)
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, "")).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownFailureStartScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "RequestingClientFailed")
		})

		It("should unsuccessfully reconcile the resource scaling down if machine has no machinedeployment name", func() {
			const nsrName = "test-scale-down-without-valid-machine-label-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request MachineDeployment in clusterNamespace(%s) through client", clusterNamespace),
					"failed to get MachineDeployment resource from management cluster",
					fmt.Sprintf("failed to get resource(%s/%s) of kind MachineDeployment", mdNamespace, ""),
				}
			)
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, "", clusterName)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownFailureStartScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, ptr.To(""), nil, errMsgs, "RequestingClientFailed")
		})
	})

	Context("When reconciling a resource proceeding scale down with failure and finalize resources", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNameDummy = "fake-cluster-01-dummy"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down with invalid cluster name", func() {
			const nsrName = "test-scale-down-with-invalid-cluster-name-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("not match clusterName(%s) with clusterName(%s) in MachineDeployment(%s)", clusterName, clusterNameDummy, mdName),
				}
				desiredReplicas = replicas - 1
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("modifying clusterName in MachineDeployment", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				mdPatched := md.DeepCopy()
				mdPatched.Spec.ClusterName = clusterNameDummy
				Expect(k8sClientManagement.Patch(ctx, mdPatched, client.MergeFrom(md))).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidResourceNotFound", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			By("asserting MachineDeployment labels are finalized", func() {
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})
	})

	Context("When reconciling a resource proceeding scaling up", Label("ttl"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up to delete it after it finished", func() {
			const nsrName = "test-scale-up-delete-resource-success"
			var (
				desiredReplicas = replicas + 1
				ttlSeconds      = 1
				mdName          = mdName
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("setting up TTL for NodeScalingRequest", func() {
				controllerReconciler.enableTTLSecondsAfterFinished = true
				controllerReconciler.ttlSecondsAfterFinished = ttlSeconds
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("waiting for ttl for NodeScalingRequest", func() {
				time.Sleep(time.Second * time.Duration(ttlSeconds))
			})
			By("executing reconcile loop for ttl", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should successfully reconcile the resource scaling up with disabled TTL", func() {
			const nsrName = "test-scale-up-with-ttl-disabled-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			r := newNodeScalingRequestReconciler(providerType, nil, nil)
			r.enableTTLSecondsAfterFinished = false
			executeNsrScaleupExistingMachineDeploymentWithReconciler(r, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, mdNamespace, replicas+1)
		})
	})

	Context("When reconciling a resource proceeding scaling up", Label("ttl", "mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Delete() for NodeScalingRequest failed", func() {
			const nsrName = "test-scale-up-delete-resource-fail"
			var (
				desiredReplicas = replicas + 1
				ttlSeconds      = 1
				mdName          = mdName
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetDelete([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("setting up TTL for NodeScalingRequest", func() {
				controllerReconciler.enableTTLSecondsAfterFinished = true
				controllerReconciler.ttlSecondsAfterFinished = ttlSeconds
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("waiting for ttl for NodeScalingRequest", func() {
				time.Sleep(time.Second * time.Duration(ttlSeconds))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred()) // return mock error
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Delete() for NodeScalingRequest failed in not found", func() {
			const nsrName = "test-scale-up-delete-resource-fail"
			var (
				desiredReplicas = replicas + 1
				ttlSeconds      = 1
				mdName          = mdName
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			defaultFunc := k8sClientWorkload.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "nodescalingrequests"}, nsrName)
			}
			Expect(wClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("setting up TTL for NodeScalingRequest", func() {
				controllerReconciler.enableTTLSecondsAfterFinished = true
				controllerReconciler.ttlSecondsAfterFinished = ttlSeconds
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("waiting for ttl for NodeScalingRequest", func() {
				time.Sleep(time.Second * time.Duration(ttlSeconds))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("cleaning up NodeScalingRequest instead of ttl", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling with mock client for NodeScalingRequest reconcile", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nodeNameInvalid  = "fake-node-01-invalid"
			nsrNamespace     = defaultNamespace
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource if Get() for NodeScalingRequest failed", func() {
			const nsrName = "test-mock-w-get1-fail"
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource if List() for NodeScalingRequest failed", func() {
			const nsrName = "test-mock-w-list1-status1-fail"
			var (
				errMsgs = []string{
					"failed to get other NodeScalingRequest is running",
					"failed to list NodeScalingRequests",
				}
				mdNameEmpty = ""
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed())
			Expect(wClient.SetList([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, "", "", &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource if List() for Node failed", func() {
			const nsrName = "test-mock-w-list2-fail"
			var (
				errMsgs = []string{
					"failed to get cluster namespace",
					fmt.Sprintf("no node found with annotations %s and %s", capiv1.ClusterNameAnnotation, capiv1.ClusterNamespaceAnnotation),
				}
				mdNameEmpty = ""
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetList([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, "", "", &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionFalse, "GettingClusterInfoFailed", errMsgs)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource if Patch() for NodeScalingRequest.Status failed", func() {
			const nsrName = "test-mock-w-status12-fail"
			var (
				nodeName = nodeName
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1, 2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred()) // not target of this test
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource if Get() for checking sync of NodeScalingRequest.Status failed", func() {
			const nsrName = "test-mock-w-get3-fail"
			var (
				nodeName    = nodeName
				mdNameEmpty = ""
				errMsgs     = []string{
					"failed to synchronize cached resource version of NodeScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{3})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "MachineAlreadyDeleted", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource if Get() for checking sync of NodeScalingRequest.Status failed", func() {
			const nsrName = "test-mock-w-get2-with-sync-cancel"
			var (
				nodeName    = nodeName
				mdNameEmpty = ""
				errMsgs     = []string{
					"failed to synchronize cached resource version of NodeScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			defaultFunc := func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				Expect(k8sClientWorkload.Get(ctx, key, obj, opts...)).To(Succeed())
				obj.SetResourceVersion("0") // override ResourceVersion to make status not sync
				return nil
			}
			alterFunc := k8sClientWorkload.Get
			Expect(wClient.SetGetWithFunc(defaultFunc, alterFunc, []int{})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			ctxLocal, cancelLocal := context.WithCancel(ctx)
			defer cancelLocal()
			By("cancel context after 2 second", func() {
				go func() {
					time.Sleep(2 * time.Second)
					cancelLocal()
				}()
			})
			By("executing reconcile loop with cancelable context", func() {
				_, err := controllerReconciler.Reconcile(ctxLocal, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "MachineAlreadyDeleted", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("When reconciling a resource scaling up with mock client", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNameWithoutMD = "fake-dci-machine-template-01"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas        = int32(1)
			desiredReplicas = replicas + 1
			location        = "data-center"
			locationDummy   = "data-not-match"
			dmtNames        = []string{
				dmtName,
				dmtNameWithoutMD,
			}
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameWithoutMD, dmtNamespace, &locationDummy, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if List() for DCIMachineTemplate failed", func() {
			const nsrName = "test-mock-m-list1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request DCIMachineTemplate list in clusterNamespace(%s) through client", clusterNamespace),
					"failed to list DCIMachineTemplate resources to get allocatable DCIMachineTemplates",
					fmt.Sprintf("failed to list resources of kind DCIMachineTemplate in namespace %s", dmtNamespace),
				}
				mdName      = mdName
				mdNameEmpty = ""
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if List() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-list2-fail"
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment list to scale up",
					"failed to list MachineDeployment resources to get allocatable MachineDeployments",
					fmt.Sprintf("failed to list resources of kind MachineDeployment in namespace %s", mdNamespace),
				}
				mdName      = mdName
				mdNameEmpty = ""
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "TargetResourceNotFound", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest.Status failed in startScaleUp", func() {
			const nsrName = "test-mock-w-status1-fail"
			var (
				errMsgs = []string{
					"failed to patch pre-request status",
				}
				mdName = mdName
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "PatchingStatusFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-patch1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("valid MachineDeployment(%s) found, but failed to scale up: clusterName(%s), clusterNamespace(%s)", mdName, clusterName, defaultNamespace),
				}
				mdName = mdName
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if another resource processing failed after pre-request", func() {
			const nsrName = "test-pre-request-pending-with-mock-m-patch1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("valid MachineDeployment(%s) found, but failed to scale up: clusterName(%s), clusterNamespace(%s)", mdName, clusterName, defaultNamespace),
				}
				mdName = mdName
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("creating another NodeScalingRequest to check pending state", func() {
				nsr := testutils.NewNodeScalingRequest("another-"+nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop for another NodeScalingRequest", func() {
				rslt, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName("another-"+nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(rslt.RequeueAfter).To(Equal(requeueIntervalNodeScalingRequestPending))
			})
			By("asserting another NodeScalingRequest is in pending state", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName("another-"+nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScalePending, "", "", ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionFalse, "Pending", nil)
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest.Status failed in Reconcile", func() {
			const nsrName = "test-mock-w-status2-fail"
			var (
				mdName = mdName
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionUnknown, "Requesting", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest.Status failed in startScaleUpWithCreatingResource", func() {
			const nsrName = "test-mock-w-status1-fail"
			var (
				errMsgs = []string{
					"failed to patch pre-request status",
				}
				mdNames = []string{
					mdName,
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "PatchingStatusFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, mdNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed), nsrNamespace, nsrName)
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameCreated, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Create() for DCIMachineTemplate failed", func() {
			const nsrName = "test-mock-m-create1-fail"
			var (
				errMsgs = []string{
					"failed to create DCIMachineTemplate(",
					fmt.Sprintf(") in clusterNamespace(%s)", defaultNamespace),
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "CreatingResourceFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, mdNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed), nsrNamespace, nsrName)
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameCreated, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Create() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-create2-fail"
			var (
				errMsgs = []string{
					"failed to create MachineDeployment(",
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &location
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed), nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "CreatingResourceFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop 3rd", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop 4th", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameCreated, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Create() for MachineDeployment without creating DCIMachineTemplate failed", func() {
			const nsrName = "test-mock-m-create1-without-creating-dmt-fail"
			var (
				errMsgs = []string{
					"failed to create MachineDeployment(",
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "CreatingResourceFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for MachineDeployment failed in proceedScaleUp", func() {
			const nsrName = "test-mock-m-patch2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to remove in-progress labels from MachineDeployment: %s", mdName),
				}
				mdName = mdName
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if cleanupMachineDeployment() failed(get) in finalizer", func() {
			const nsrName = "test-mock-m-get1"
			var (
				errMsgs = []string{
					"failed to clean up MachineDeployment",
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			errMsgs = append(errMsgs, fmt.Sprintf("failed to get MachineDeployment (%s) in namespace (%s)", mdNameCreated, mdNamespace))
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer with failure", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if cleanupMachineDeployment() failed(patch) in finalizer", func() {
			const nsrName = "test-mock-m-patch1"
			var (
				errMsgs = []string{
					"failed to clean up MachineDeployment",
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			errMsgs = append(errMsgs, fmt.Sprintf("failed to patch resource(%s/%s) of kind MachineDeployment", mdNamespace, mdNameCreated))
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer with failure", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest failed in finalizer", func() {
			const nsrName = "test-mock-m-patch2"
			var (
				errMsgs = []string{
					"failed to remove finalizer",
				}
				mdNames = []string{
					mdName,
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer with failure", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest failed in finalizer", func() {
			const nsrName = "test-mock-m-patch2"
			var (
				errMsgs = []string{
					"failed to remove finalizer",
				}
				mdNames = []string{
					mdName,
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer with failure", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for NodeScalingRequest to add finalizer failed", func() {
			const nsrName = "test-mock-w-patch1"
			var (
				errMsgs = []string{
					"failed to add finalizer",
				}
				mdNames = []string{
					mdName,
				}
				mdNameEmpty = ""
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &locationDummy
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, "", "", &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionFalse, "AddingFinalizerFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Requested)).To(BeNil())
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameWithoutMD, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Patch() for MachineDeployment failed in proceedScaleUp", func() {
			const nsrName = "test-mock-m-patch1"
			var (
				errMsgs = []string{
					"failed to remove in-progress labels from MachineDeployment: ",
				}
				mdNames = []string{
					mdName,
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, mdNamespace, nil, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed), nsrNamespace, nsrName)
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
			})
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameCreated, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Succeed())
		})
	})

	Context("When reconciling a resource scaling down without deleting MachineDeployment with mock client", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-01"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-01"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-01"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-01"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-01"
			machineDummyName = "fake-machine-02"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-01"
			nodeDummyName    = "fake-node-02"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas = int32(2)
			location = "data-center"
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeDummyName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineDummyName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for Node failed", func() {
			const nsrName = "test-mock-w-get2-fail"
			var (
				errMsgs = []string{
					"failed to request Node list to WorkloadCluster",
				}
				nodeName        = nodeName
				mdName          = mdName
				mdNameEmpty     = ""
				desiredReplicas = replicas - 1
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for Machine failed", func() {
			const nsrName = "test-mock-w-get1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request Machine resource in clusterNamespace(%s) through client", clusterNamespace),
					"failed to get Machine resource",
					fmt.Sprintf("failed to get resource(%s/%s) of kind Machine", machineNamespace, machineName),
				}
				nodeName        = nodeName
				mdName          = mdName
				mdNameEmpty     = ""
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for MachineDeployment failed", func() {
			const nsrName = "test-mock-w-get2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request MachineDeployment in clusterNamespace(%s) through client", clusterNamespace),
					"failed to get MachineDeployment resource from management cluster",
					fmt.Sprintf("failed to get resource(%s/%s) of kind MachineDeployment", mdNamespace, mdName),
				}
				nodeName        = nodeName
				mdName          = mdName
				mdNameEmpty     = ""
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for NodeScalingRequest failed", func() {
			const nsrName = "test-mock-w-status1-patch2-fail"
			var (
				errMsgs = []string{
					"failed to patch pre-request status",
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "PatchingStatusFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for MachineSet failed", func() {
			const nsrName = "test-mock-m-list1-fail"
			var (
				errMsgs = []string{
					"failed to get valid MachineSet with MachineDeployment UID",
					fmt.Sprintf(" in namespace(%s)", mdNamespace),
					"failed to get MachineSet owned by MachineDeployment with UID ",
					fmt.Sprintf("failed to list resources of kind MachineSet in namespace %s", mdNamespace),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "TargetResourceNotFound", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for MachineSet failed", func() {
			const nsrName = "test-mock-m-patch1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to patch annotation(%s) to true through client", capiv1.PausedAnnotation),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-patch2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to patch resource(%s/%s) of kind MachineDeployment", mdNamespace, mdName),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Delete() for Machine failed", func() {
			const nsrName = "test-mock-m-delete1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to delete Machine(%s) in Namespace(%s)", machineName, clusterNamespace),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "DeletingMachineFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for NodeScalingRequest failed in Reconcile", func() {
			const nsrName = "test-mock-w-patch2-status2-fail"
			var (
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{2})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionUnknown, "Requesting", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "MachineAlreadyDeleted", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "MachineAlreadyDeleted", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for Node failed in proceedScaleDown", func() {
			const nsrName = "test-mock-w-get5-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to get Node(%s) for reasons other than 'not found'", nodeName),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{5})).To(Succeed())
			wClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient.GetMockClient(), nil)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for MachineDeployment failed in proceedScaleDown", func() {
			const nsrName = "test-mock-m-get3-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("not found target MachineDeployment(%s) in clusterNamespace(%s)", mdName, mdNamespace),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{3})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for MachineSet failed", func() {
			const nsrName = "test-mock-m-list2-fail"
			var (
				mdUID   = getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
				errMsgs = []string{
					fmt.Sprintf("failed to get valid MachineSet with MachineDeployment UID(%v) in namespace(%s)", mdUID, clusterNamespace),
				}
				nodeName        = nodeName
				desiredReplicas = replicas - 1
				mdName          = mdName
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for MachineSet failed in proceedScaleDown", func() {
			const nsrName = "test-mock-m-patch3-fail"
			var (
				mdUID   = getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
				errMsgs = []string{
					fmt.Sprintf("failed to remove paused annotation from MachineSet binding MachineDeployment UID(%v) in namespace(%s)", mdUID, clusterNamespace),
				}
				nodeName        = nodeName
				desiredReplicas = replicas - 1
				mdName          = mdName
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{3})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for MachineDeployment failed in proceedScaleDown", func() {
			const nsrName = "test-mock-m-patch4-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to remove in-progress labels from MachineDeployment: %s", mdName),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{4})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling down with deleting MachineDeployment and DCIMachineTemplate with mock client", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdDummyName      = "fake-md-dummy"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtDummyName     = "fake-dci-machine-template-dummy"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas      = int32(1)
			replicasDummy = int32(0)
			location      = "data-center"
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtDummyName, &replicasDummy)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtDummyName, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for valid MachineDeployment failed", func() {
			const nsrName = "test-mock-m-list1-fail-machinedeployment-not-exist"
			var (
				errMsgs = []string{
					"failed to check if valid MachineDeployment exist or not",
					"failed to list MachineDeployment resources to check last MachineDeployment",
					fmt.Sprintf("failed to list resources of kind MachineDeployment in namespace %s", clusterNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ValidResourceNotFound", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for DCIMachineTemplate failed", func() {
			const nsrName = "test-mock-m-get3-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DCIMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to get DCIMachineTemplate(%s) in Namespace(%s)", dmtName, dmtNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{3})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-list2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DCIMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to list resources of kind MachineDeployment in namespace %s", mdNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Delete() for DCIMachineTemplate failed", func() {
			const nsrName = "test-mock-m-delete1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DCIMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DCIMachineTemplate(%s) in Namespace(%s)", dmtName, dmtNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should successfully reconcile the resource scaling down if Delete() for DCIMachineTemplate failed as not found", func() {
			const nsrName = "test-mock-m-delete1-fail-not-found"
			var (
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			defaultFunc := k8sClientManagement.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "dcimachinetemplates"}, dmtDummyName)
			}
			Expect(mClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Delete() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-delete2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should successfully reconcile the resource scaling down if Delete() for MachineDeployment failed as not found", func() {
			const nsrName = "test-mock-m-delete2-is-not-found"
			var (
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			defaultFunc := k8sClientManagement.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "machinedeployments"}, mdName)
			}
			Expect(mClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling down finalizer executed in the process of scaling", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-01"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-01"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-01"
			msNamespace      = defaultNamespace
			dmtName          = "fake-dci-machine-template-01"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-01"
			machineDummyName = "fake-machine-02"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-01"
			nodeDummyName    = "fake-node-02"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas = int32(2)
			location = "data-center"
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeDummyName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, &location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS))).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineDummyName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for MachineSet failed in finalizer", func() {
			const nsrName = "test-mock-m-list2-fail"
			var (
				mdUID   = getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
				errMsgs = []string{
					"failed to clean up on finalize",
					"failed to clean up MachineSet",
					fmt.Sprintf("failed to get MachineSet binding MachineDeployment UID (%s) in namespace (%s)", mdUID, msNamespace),
					fmt.Sprintf("failed to get MachineSet owned by MachineDeployment with UID %s", mdUID),
					fmt.Sprintf(" failed to list resources of kind MachineSet in namespace %s", msNamespace),
				}
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Patch() for MachineSet failed in finalizer", func() {
			const nsrName = "test-mock-m-patch3-fail"
			var (
				nodeName        = nodeName
				mdName          = mdName
				desiredReplicas = replicas - 1
				errMsgs         = []string{
					"failed to clean up MachineSet",
					fmt.Sprintf("failed to patch MachineSet (%s) in namespace (%s)", msName, msNamespace),
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{3})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = &nodeName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
		})
	})

	Context("When reconciling a resource scaling up in namespace("+testNamespace+")", Label("namespaced"), func() {
		const (
			nsrNamespace       = testNamespace
			clusterName        = "fake-cluster-00"
			clusterNamespace   = testNamespace
			mdName             = "fake-md-00"
			mdNamespace        = testNamespace
			dmtNameNoResource  = "fake-dci-machine-template-00" // no resource
			dmtNameAllResource = "fake-dci-machine-template-01" // all resource
			dmtNamespace       = testNamespace
			machineName        = "fake-machine-00"
			machineNamespace   = testNamespace
			nodeName           = "fake-node-00"
		)
		var (
			replicas              = int32(1)
			location              = "DC1"
			os                    = rpmv1.OSType{Distribution: "ubuntu", Version: "24.04"}
			minCPU                = int32(10)
			minMemory             = int32(10)
			minStorage            = int32(10)
			gpu                   = rpmv1.GPUSpec{Model: "H100", Num: 1}
			nic                   = rpmv1.NICSpec{Model: "ConnectX-5", Num: 1}
			serverSpecNoResource  = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
			serverSpecAllResource = testutils.NewServerSpec(rpmv1.ReqServerTypeComposed, &os, &minCPU, &minMemory, &minStorage, &gpu, &nic)
			mdNames               = []string{mdName}
			dmtNames              = []string{dmtNameNoResource, dmtNameAllResource}
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtNameNoResource, &replicas)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameNoResource, dmtNamespace, nil, serverSpecNoResource)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameAllResource, dmtNamespace, &location, serverSpecAllResource)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up existing MachineDeployment", func() {
			const nsrName = "test-scale-up-existing-md-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupExistingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, mdNamespace, replicas+1)
		})

		It("should successfully reconcile the resource creating new MachineDeployment with existing DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-with-existing-template"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = serverSpecAllResource
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNameAllResource, replicas)
		})

		It("should successfully reconcile the resource creating new MachineDeployment and DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-and-template"
			ss := serverSpecAllResource.DeepCopy()
			ss.Type = rpmv1.ReqServerTypeCOTS
			loc := location + "-new"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &loc
				nsr.Spec.ServerSpec = ss
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingDCIMachineTemplate(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNames, dmtNamespace, loc, ss, replicas)
		})
	})

	Context("When reconciling a resource scaling down in namespace("+testNamespace+")", Label("namespaced"), func() {
		const (
			nsrNamespace     = testNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = testNamespace
			mdName           = "fake-md-00"
			mdNamespace      = testNamespace
			msName           = "fake-ms-00"
			msNamespace      = testNamespace
			dmtName          = "fake-dci-machine-template-00"
			dmtNamespace     = testNamespace
			machineName      = "fake-machine-00"
			machineNamespace = testNamespace
			nodeName         = "fake-node-00"
		)
		var (
			replicas   = int32(1)
			serverSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace, nil, serverSpec)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down", func() {
			const nsrName = "test-scale-down-success"
			Expect(testutils.PatchMachineDeploymentReplicas(ctx, k8sClientManagement, mdName, mdNamespace, nil, ptr.To(int32(1)))).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, replicas)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-success"
				mdDummyName   = "fake-md-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment and DCIMachineTemplate", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-dmt-success"
				mdDummyName   = "fake-md-dummy"
				dmtDummyName  = "fake-dci-machine-template-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtDummyName, dmtNamespace, nil, serverSpec)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})

		It("should successfully reconcile the resource scaling down without deleting MachineDeployment because of last MachineDeployment", func() {
			const nsrName = "test-scale-down-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, replicas-1)
		})
	})
})

// Test for Reconcile() method for Docker Provider mode
var _ = Describe("NodeScalingRequest Controller for Docker Provider", Label("nsr", "controller", "docker"), func() {
	const providerType = utils.ProviderTypeDocker
	mdTemplateAPIGroup := capdockerv1.GroupVersion.Group
	mdTemplateKind := templateKindDocker

	Context("When reconciling a resource scaling up", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var (
			replicas = int32(1)
			mdNames  = []string{mdName}
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling up existing MachineDeployment", func() {
			const nsrName = "test-scale-up-existing-md-success"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				nsr.Spec.Location = ptr.To(dmtName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupExistingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, mdNamespace, replicas+1)
		})

		It("should successfully reconcile the resource creating new MachineDeployment with existing DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-with-existing-template"
			const dmtNewName = "fake-docker-machine-template-new"
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtNewName, dmtNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				nsr.Spec.Location = ptr.To(dmtNewName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNewName, replicas)
		})

		It("should successfully reconcile the resource creating new MachineDeployment and DCIMachineTemplate", func() {
			const nsrName = "test-create-new-md-and-template"
			const dmtNewName = "fake-docker-machine-template-new"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				nsr.Spec.Location = ptr.To(dmtNewName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaleupWithCreatingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNewName, replicas)
		})
	})

	Context("When reconciling a resource scaling down", func() {
		const (
			nsrNamespace     = defaultNamespace
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down", func() {
			const nsrName = "test-scale-down-success"
			Expect(testutils.PatchMachineDeploymentReplicas(ctx, k8sClientManagement, mdName, mdNamespace, nil, ptr.To(int32(1)))).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledown(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, replicas)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-success"
				mdDummyName   = "fake-md-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})

		It("should successfully reconcile the resource scaling down with deleting MachineDeployment and DCIMachineTemplate", func() {
			const (
				nsrName       = "test-scale-down-with-deleting-md-dmt-success"
				mdDummyName   = "fake-md-dummy"
				dmtDummyName  = "fake-dci-machine-template-dummy"
				replicasDummy = 0
			)
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtDummyName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, ptr.To(int32(replicasDummy)))).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithDeletingMachineDeployment(providerType, nil, nil, nsrName, nsrNamespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, replicas-1)
		})
	})

	Context("When reconciling a resource scaling up with nil location", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up", func() {
			const nsrName = "test-docker-scale-up-with-nil-location"
			var (
				serverSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS) // dummy data
				errMsgs    = []string{
					fmt.Sprintf("parameter is nil: .spec.location(%v) .spec.serverSpec(%v)", nil, serverSpec),
				}
				mdNameEmpty = ""
			)
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			controllerReconciler.providerType = providerType
			Expect(controllerReconciler.providerType).To(Equal(providerType))
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = serverSpec
				nsr.Spec.Location = nil
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreateFailed, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Requested)).To(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "ValidationFailed", errMsgs)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling down without valid resources", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		executeNsrScaledownWithoutValidResource := func(
			name, namespace, mdName, reason string, desiredReplicas *int32,
		) {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("patching MachineDeployment ready replicas to simulate scale down", func() {
				if mdName != "" {
					Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
				}
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		}

		executeNsrScaledownWithoutValidResourceFinalize := func(
			name, namespace, mdName, reason string, desiredReplicas *int32,
		) {
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, reason, nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(name, namespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		}

		It("should successfully reconcile the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, "", "MachineAlreadyDeleted", nil)
		})

		It("should successfully reconcile the resource scaling down if MachineDeployment not exists", func() {
			const nsrName = "test-scale-down-without-valid-md-fail"
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, "", "MachineDeploymentAlreadyDeleted", nil)
		})

		It("should successfully reconcile the resource scaling down if MachineSet not exists", func() {
			const nsrName = "test-scale-down-without-ms"
			Expect(testutils.CleanupMachineSetResource(ctx, k8sClientManagement, msName, msNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResource(nsrName, nsrNamespace, mdName, "ScalingRequestSucceeded", ptr.To(replicas-1))
		})

		It("should successfully finalize the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, "", "MachineAlreadyDeleted", nil)
		})

		It("should successfully finalize the resource scaling down if MachineDeployment not exists", func() {
			const nsrName = "test-scale-down-without-valid-md-fail"
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, "", "MachineDeploymentAlreadyDeleted", nil)
		})

		It("should successfully finalize the resource scaling down if MachineSet not exists", func() {
			const nsrName = "test-scale-down-without-ms"
			Expect(testutils.CleanupMachineSetResource(ctx, k8sClientManagement, msName, msNamespace)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			executeNsrScaledownWithoutValidResourceFinalize(nsrName, nsrNamespace, mdName, "ScalingRequestSucceeded", ptr.To(replicas-1))
		})

		It("should successfully reconcile the resource scaling down if DockerMachineTemplate not exists", func() {
			const nsrName = "test-scale-down-without-docker-machine-template"
			const mdDummyName = "fake-md-dummy"
			Expect(testutils.CleanupDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdDummyName, mdNamespace, 0)).To(Succeed())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), ptr.To(int32(0)))
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("skipping to simulate scale down since MachineDeployment not exists")
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), ptr.To(int32(0)))
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling down without simulating cluster-api partially", func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should successfully reconcile the resource scaling down if Machine not exists", func() {
			const nsrName = "test-scale-down-without-valid-machine-fail"
			var (
				desiredReplicas = replicas - 1
			)
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, nil)
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "Requeueing", nil)
			})
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("When reconciling a resource scaling up with mock client", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtNameWithoutMD = "fake-docker-machine-template-01"
			dmtNameNotUsed   = "fake-docker-machine-template-02"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var replicas = int32(1)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtNameWithoutMD, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if List() for DockerMachineTemplate failed", func() {
			const nsrName = "test-mock-m-list1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to request DockerMachineTemplate list in clusterNamespace(%s) through client", clusterNamespace),
					"failed to list DockerMachineTemplate resources to get allocatable DockerMachineTemplates",
					fmt.Sprintf("failed to list resources of kind DockerMachineTemplate in namespace %s", mdNamespace),
				}
				dmtName     = dmtNameWithoutMD
				mdNameEmpty = ""
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				nsr.Spec.Location = &dmtName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, &mdNameEmpty, nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "RequestingClientFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling up if Create() for DockerMachineTemplate failed", func() {
			const nsrName = "test-mock-m-create1-fail"
			var (
				errMsgs = []string{
					"failed to create DockerMachineTemplate(",
					fmt.Sprintf(") in clusterNamespace(%s)", defaultNamespace),
				}
				dmtName = dmtNameNotUsed
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				nsr.Spec.Location = &dmtName
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "CreatingResourceFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, []string{mdName}, clusterNamespace, clusterName, dmtName, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, clusterNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, clusterNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling up if Create() for MachineDeployment without creating DockerMachineTemplate failed", func() {
			const nsrName = "test-mock-m-create1-without-creating-dmt-fail"
			var (
				errMsgs = []string{
					"failed to create MachineDeployment(",
				}
				dmtName = dmtNameWithoutMD
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "create")
				nsr.Spec.Location = &dmtName
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeCOTS)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "CreatingResourceFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, []string{mdName}, clusterNamespace, clusterName, dmtName, nsrNamespace, nsrName)
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, nil, &replicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdNameCreated, clusterNamespace)).To(Succeed())
		})
	})

	Context("When reconciling a resource scaling down with deleting MachineDeployment and DockerMachineTemplate with mock client", Label("mock"), func() {
		const (
			clusterName      = "fake-cluster-00"
			clusterNamespace = defaultNamespace
			mdName           = "fake-md-00"
			mdDummyName      = "fake-md-dummy"
			mdNamespace      = defaultNamespace
			msName           = "fake-ms-00"
			msNamespace      = defaultNamespace
			dmtName          = "fake-docker-machine-template-00"
			dmtDummyName     = "fake-docker-machine-template-dummy"
			dmtNamespace     = defaultNamespace
			machineName      = "fake-machine-00"
			machineNamespace = defaultNamespace
			nodeName         = "fake-node-00"
			nsrNamespace     = defaultNamespace
		)
		var (
			replicas      = int32(1)
			replicasDummy = int32(0)
		)

		BeforeEach(func() {
			// workload
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			// management
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtName, &replicas)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdDummyName, mdNamespace, clusterName, mdTemplateAPIGroup, mdTemplateKind, dmtDummyName, &replicasDummy)).To(Succeed())
			Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
			mdUID := getMachineDeploymentUID(ctx, k8sClientManagement, mdName, mdNamespace)
			Expect(testutils.CreateMachineSetResource(ctx, k8sClientManagement, msName, msNamespace, clusterName, mdName, &replicas, mdUID)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateDockerMachineTemplateResource(ctx, k8sClientManagement, dmtDummyName, dmtNamespace)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
		})
		AfterEach(func() {
			// workload
			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			// management
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineSetResources(ctx, k8sClientManagement, msNamespace)).To(Succeed())
			Expect(testutils.CleanupDockerMachineTemplateResources(ctx, k8sClientManagement, dmtNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for valid MachineDeployment failed", func() {
			const nsrName = "test-mock-m-list1-fail-machinedeployment-not-exist"
			var (
				errMsgs = []string{
					"failed to check if valid MachineDeployment exist or not",
					"failed to list MachineDeployment resources to check last MachineDeployment",
					fmt.Sprintf("failed to list resources of kind MachineDeployment in namespace %s", clusterNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ValidResourceNotFound", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Get() for DockerMachineTemplate failed", func() {
			const nsrName = "test-mock-m-get3-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DockerMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to get DockerMachineTemplate(%s) in Namespace(%s)", dmtName, dmtNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{3})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if List() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-list2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DockerMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to list resources of kind MachineDeployment in namespace %s", mdNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Delete() for DockerMachineTemplate failed", func() {
			const nsrName = "test-mock-m-delete1-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DockerMachineTemplate for MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete DockerMachineTemplate(%s) in Namespace(%s)", dmtName, dmtNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
				Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
				Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should successfully reconcile the resource scaling down if Delete() for DockerMachineTemplate failed as not found", func() {
			const nsrName = "test-mock-m-delete1-fail-not-found"
			var (
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			defaultFunc := k8sClientManagement.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "DockerMachinetemplates"}, dmtDummyName)
			}
			Expect(mClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{1})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should unsuccessfully reconcile the resource scaling down if Delete() for MachineDeployment failed", func() {
			const nsrName = "test-mock-m-delete2-fail"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
					fmt.Sprintf("failed to delete MachineDeployment(%s) in Namespace(%s)", mdName, mdNamespace),
				}
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, ptr.To(""), nil)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, "ScalingRequestFailed", errMsgs)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("should successfully reconcile the resource scaling down if Delete() for MachineDeployment failed as not found", func() {
			const nsrName = "test-mock-m-delete2-is-not-found"
			var (
				desiredReplicas = int32(0)
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			defaultFunc := k8sClientManagement.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "machinedeployments"}, mdName)
			}
			Expect(mClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{2})).To(Succeed())
			mClient.SetDefault()
			controllerReconciler := newNodeScalingRequestReconciler(providerType, nil, mClient.GetMockClient())
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, "delete")
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
			})
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			By("executing reconcile loop", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("asserting NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, ptr.To(mdName), &desiredReplicas)
				Expect(nsr.Status.Conditions).NotTo(BeNil())
				assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
				assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
			})
			By("cleaning up NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			By("executing reconcile finalizer", func() {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
			})
			By("checking if NodeScalingRequest is already deleted", func() {
				nsr := &scalev1.NodeScalingRequest{}
				err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})

// newNodeScalingRequestReconciler returns NodeScalingRequestReconciler with mock by one line.
func newNodeScalingRequestReconciler(
	providerType utils.ProviderType, wClient, mClient client.Client,
) *NodeScalingRequestReconciler {
	By("creating NodeScalingRequestReconciler with mock")
	controllerReconciler := (&NodeScalingRequestReconciler{
		Recorder:                      record.NewFakeRecorder(1024),
		providerType:                  providerType,
		enableTTLSecondsAfterFinished: true,
		ttlSecondsAfterFinished:       3600,
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

// getMachineDeploymentCreated returns the name of the MachineDeployment created by the reconcile loop.
func getMachineDeploymentCreated(
	ctx context.Context, c client.Client, mdNames []string, mdNamespace, clusterName, dmtName, nsrNamespace, nsrName string,
) string {
	var (
		mdCreated     = (*capiv1.MachineDeployment)(nil)
		mdNameCreated = ""
	)
	// get the MachineDeployment created by reconcile loop
	mdList := &capiv1.MachineDeploymentList{}
	Expect(c.List(ctx, mdList, &client.ListOptions{Namespace: mdNamespace})).To(Succeed())
	for idx := range mdList.Items {
		md := &mdList.Items[idx]
		if slices.Contains(mdNames, md.Name) {
			continue // if the name of the MachineDeployment is in the existing list, skip it
		}
		// created MachineDeployment
		mdCreated = md
		mdNameCreated = md.Name
		break
	}
	Expect(mdCreated).NotTo(BeNil())
	Expect(mdNameCreated).NotTo(Equal(""))
	// asserting MachineDeployment created by reconcile loop
	Expect(mdCreated.Namespace).To(Equal(mdNamespace))
	Expect(mdCreated.ObjectMeta.Labels).NotTo(BeNil())
	Expect(mdCreated.ObjectMeta.Labels[scalev1.CreatedByNamespaceLabel]).To(Equal(nsrNamespace))
	Expect(mdCreated.ObjectMeta.Labels[scalev1.CreatedByNameLabel]).To(Equal(nsrName))
	Expect(mdCreated.Spec.ClusterName).To(Equal(clusterName))
	Expect(mdCreated.Spec.Replicas).NotTo(BeNil())
	Expect(*mdCreated.Spec.Replicas).To(Equal(int32(1)))
	Expect(mdCreated.Spec.Selector.MatchLabels).NotTo(BeNil())
	Expect(mdCreated.Spec.Selector.MatchLabels[capiv1.ClusterNameLabel]).To(Equal(clusterName))
	Expect(mdCreated.Spec.Template.Spec.ClusterName).To(Equal(clusterName))
	Expect(mdCreated.Spec.Template.Spec.Bootstrap.ConfigRef).NotTo(BeNil())
	Expect(mdCreated.Spec.Template.Spec.InfrastructureRef.Name).To(Equal(dmtName))
	Expect(mdCreated.Status.ReadyReplicas).To(BeNil())
	return mdNameCreated
}

// getDCIMachineTemplateCreated returns the name of the DCIMachineTemplate created by the reconcile loop.
func getDCIMachineTemplateCreated(
	ctx context.Context, c client.Client, dmtNames []string, dmtNamespace string, location *string, serverSpec *rpmv1.ServerSpec, nsrNamespace, nsrName string,
) string {
	var (
		dmtCreated     = (*capdv1.DCIMachineTemplate)(nil)
		dmtNameCreated = ""
	)
	// get the DCIMachineTemplate created by reconcile loop
	dmtList := &capdv1.DCIMachineTemplateList{}
	Expect(c.List(ctx, dmtList, &client.ListOptions{Namespace: dmtNamespace})).To(Succeed())
	for idx := range dmtList.Items {
		dmt := &dmtList.Items[idx]
		if slices.Contains(dmtNames, dmt.Name) {
			continue // if the name of the DCIMachineTemplate is in the existing list, skip it
		}
		// created machineDeployment
		dmtCreated = dmt
		dmtNameCreated = dmt.Name
		break
	}
	Expect(dmtCreated).NotTo(BeNil())
	Expect(dmtNameCreated).NotTo(Equal(""))
	// asserting DCIMachineTemplate created by reconcile loop
	Expect(dmtCreated.Namespace).To(Equal(dmtNamespace))
	By(fmt.Sprintf("%v", dmtCreated))
	Expect(dmtCreated.Labels).NotTo(BeNil())
	Expect(dmtCreated.Labels[scalev1.CreatedByNamespaceLabel]).To(Equal(nsrNamespace))
	Expect(dmtCreated.Labels[scalev1.CreatedByNameLabel]).To(Equal(nsrName))
	if location == nil {
		// skip checking Location, this field will be set by DCI provider in real case,
	} else {
		Expect(dmtCreated.Spec.Template.Spec.Location).NotTo(BeNil())
		Expect(*dmtCreated.Spec.Template.Spec.Location).To(Equal(*location))
	}
	Expect(dmtCreated.Spec.Template.Spec.ServerSpec).To(Equal(*serverSpec))
	// skip checking ProviderID, this field will be set by DCI provider in real case,
	return dmtNameCreated
}

// getMachineDeploymentUID returns the UID of the specified MachineDeployment.
func getMachineDeploymentUID(
	ctx context.Context, c client.Client, mdName, mdNamespace string,
) types.UID {
	md := &capiv1.MachineDeployment{}
	Expect(c.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
	return md.UID
}

// getMachineDeploymentInProgressLabel returns the value of the InProgress label of the specified MachineDeployment.
func getMachineDeploymentInProgressLabel(
	ctx context.Context, c client.Client, mdName, mdNamespace string,
) string {
	md := &capiv1.MachineDeployment{}
	Expect(c.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
	return md.Labels[scalev1.InProgressUIDLabel]
}

// getMachineSetPausedAnnotation returns the value of the paused annotation of the specified MachineSet.
func getMachineSetPausedAnnotation(
	ctx context.Context, c client.Client, msName, msNamespace string,
) string {
	ms := &capiv1.MachineSet{}
	Expect(c.Get(ctx, testutils.NewNamespacedName(msName, msNamespace), ms)).To(Succeed())
	return ms.Annotations[capiv1.PausedAnnotation]
}

// assertNodeScalingRequestStatus asserts nsr.Status other than Conditions[*] by 1 line.
func assertNodeScalingRequestStatus(
	nsr *scalev1.NodeScalingRequest, phase scalev1.NodeScalingRequestPhase, clusterName, clusterNamespace string,
	mdName *string, replicas *int32,
) {
	Expect(nsr).NotTo(BeNil())
	Expect(nsr.Status.Phase).To(Equal(phase))
	Expect(nsr.Status.ClusterName).To(Equal(clusterName))
	Expect(nsr.Status.ClusterNamespace).To(Equal(clusterNamespace))
	if mdName == nil {
		Expect(nsr.Status.MachineDeploymentName).To(HavePrefix("worker-md-"))
	} else {
		Expect(nsr.Status.MachineDeploymentName).To(Equal(*mdName))
	}
	if replicas == nil {
		Expect(nsr.Status.DesiredReplicas).To(BeNil())
	} else {
		Expect(nsr.Status.DesiredReplicas).NotTo(BeNil())
		Expect(*nsr.Status.DesiredReplicas).To(Equal(*replicas))
	}
}

// assertNodeScalingRequestStatusCondition asserts nsr.Status.Conditions[*] by 1 line.
func assertNodeScalingRequestStatusCondition(
	nsr *scalev1.NodeScalingRequest, typeCondition string, statusCondition metav1.ConditionStatus,
	reasonCondition string, messages []string,
) {
	Expect(nsr).NotTo(BeNil())
	cond := meta.FindStatusCondition(nsr.Status.Conditions, typeCondition)
	Expect(cond).NotTo(BeNil())
	Expect(cond.Status).To(Equal(statusCondition))
	Expect(cond.Reason).To(Equal(reasonCondition))
	if len(messages) == 0 {
		Expect(cond.Message).To(Equal(""))
	} else {
		for _, message := range messages {
			Expect(cond.Message).To(ContainSubstring(message))
		}
	}
}

// executeNsrScaleupExistingMachineDeploymentWithReconciler executes scale-up scenario on existing MachineDeployment
// with using provided reconciler.
func executeNsrScaleupExistingMachineDeploymentWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	name, namespace, clusterName, clusterNamespace, mdName, mdNamespace string, desiredReplicas int32,
) {
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
	})
	By("executing reconcile loop without patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "Requeueing", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
	})
	// patch replicas
	Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
	By("executing reconcile loop with patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
	})
	By("executing reconcile loop after finished request", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
	})
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

// executeNsrScaleupExistingMachineDeployment executes scale-up scenario on existing MachineDeployment
// with creating reconciler inside.
//
//nolint:unparam // parameters are kept for redundancy
func executeNsrScaleupExistingMachineDeployment(
	providerType utils.ProviderType, wClient, mClient client.Client,
	name, namespace, clusterName, clusterNamespace, mdName, mdNamespace string, desiredReplicas int32,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaleupExistingMachineDeploymentWithReconciler(r, name, namespace, clusterName, clusterNamespace, mdName, mdNamespace, desiredReplicas)
}

// executeNsrScaleupWithCreatingMachineDeploymentWithReconciler executes scale-up scenario with creating MachineDeployment
// with using provided reconciler.
func executeNsrScaleupWithCreatingMachineDeploymentWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	name, namespace, clusterName, clusterNamespace string, mdNames []string, mdNamespace, dmtName string,
	desiredReplicas int32,
) {
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	// get md name created in reconcile loop
	mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtName, namespace, name)
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
	})
	// patch replicas
	Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, clusterNamespace, 0)).To(Succeed())
	By("executing reconcile loop with patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
	})
	By("executing reconcile loop after finished request", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
	})
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("cleaning up machineDeployment created by reconcile loop: "+clusterNamespace+"/"+mdNameCreated, func() {
		md := &capiv1.MachineDeployment{}
		Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdNameCreated, clusterNamespace), md)).To(Succeed())
		Expect(k8sClientManagement.Delete(ctx, md)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

// executeNsrScaleupWithCreatingMachineDeployment executes scale-up scenario with creating MachineDeployment
// with creating reconciler inside.
//
//nolint:unparam // parameters are kept for redundancy
func executeNsrScaleupWithCreatingMachineDeployment(
	providerType utils.ProviderType, wClient, mClient client.Client,
	name, namespace, clusterName, clusterNamespace string, mdNames []string, mdNamespace, dmtName string,
	desiredReplicas int32,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaleupWithCreatingMachineDeploymentWithReconciler(r, name, namespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtName, desiredReplicas)
}

// executeNsrScaleupWithCreatingDCIMachineTemplateWithReconciler executes scale-up scenario with creating DCIMachineTemplate
// and MachineDeployment with using provided reconciler.
func executeNsrScaleupWithCreatingDCIMachineTemplateWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	name, namespace, clusterName, clusterNamespace string, mdNames []string, mdNamespace string, dmtNames []string, dmtNamespace string,
	location string, serverSpec *rpmv1.ServerSpec, desiredReplicas int32,
) {
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	// get md name and template name created in reconcile loop
	dmtNameCreated := getDCIMachineTemplateCreated(ctx, k8sClientManagement, dmtNames, dmtNamespace, &location, serverSpec, namespace, name)
	mdNameCreated := getMachineDeploymentCreated(ctx, k8sClientManagement, mdNames, mdNamespace, clusterName, dmtNameCreated, namespace, name)
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(string(nsr.UID)))
	})
	// patch replicas
	Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdNameCreated, clusterNamespace, 0)).To(Succeed())
	By("executing reconcile loop with patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
	})
	By("executing reconcile loop after finished request", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreated, clusterName, clusterNamespace, &mdNameCreated, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdNameCreated, mdNamespace)).To(Equal(""))
	})
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("cleaning up MachineDeployment created by reconcile loop: "+clusterNamespace+"/"+mdNameCreated, func() {
		md := &capiv1.MachineDeployment{}
		Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdNameCreated, clusterNamespace), md)).To(Succeed())
		Expect(k8sClientManagement.Delete(ctx, md)).To(Succeed())
	})
	By("cleaning up DCIMachineTemplate created by reconcile loop: "+clusterNamespace+"/"+dmtNameCreated, func() {
		dmt := &capdv1.DCIMachineTemplate{}
		Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(dmtNameCreated, clusterNamespace), dmt)).To(Succeed())
		Expect(k8sClientManagement.Delete(ctx, dmt)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

// executeNsrScaleupWithCreatingDCIMachineTemplate executes scale-up scenario with creating DCIMachineTemplate
// and MachineDeployment with creating reconciler inside.
func executeNsrScaleupWithCreatingDCIMachineTemplate(
	providerType utils.ProviderType, wClient, mClient client.Client,
	name, namespace, clusterName, clusterNamespace string, mdNames []string, mdNamespace string, dmtNames []string, dmtNamespace string,
	location string, serverSpec *rpmv1.ServerSpec, desiredReplicas int32,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaleupWithCreatingDCIMachineTemplateWithReconciler(r, name, namespace, clusterName, clusterNamespace, mdNames, mdNamespace, dmtNames, dmtNamespace, location, serverSpec, desiredReplicas)
}

// executeNsrScaledownWithReconciler executes scale-down scenario with using provided reconciler.
func executeNsrScaledownWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	name, namespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace string,
	desiredReplicas int32,
) {
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
		Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
	})
	By("executing reconcile loop without patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "Requeueing", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(string(nsr.UID)))
		Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal("true"))
	})
	// cleanup node
	Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
	// patch replicas
	Expect(testutils.PatchMachineDeploymentReadyReplicas(ctx, k8sClientManagement, mdName, mdNamespace, 0)).To(Succeed())
	By("executing reconcile loop with patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
		Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
	})
	By("executing reconcile loop after finished request", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
		Expect(getMachineDeploymentInProgressLabel(ctx, k8sClientManagement, mdName, mdNamespace)).To(Equal(""))
		Expect(getMachineSetPausedAnnotation(ctx, k8sClientManagement, msName, msNamespace)).To(Equal(""))
	})
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

// executeNsrScaledown executes scale-down scenario with creating reconciler inside.
//
//nolint:unparam // parameters are kept for redundancy
func executeNsrScaledown(
	providerType utils.ProviderType, wClient, mClient client.Client,
	name, namespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace string,
	desiredReplicas int32,
) {
	controllerReconciler := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaledownWithReconciler(controllerReconciler, name, namespace, clusterName, clusterNamespace, nodeName, mdName, mdNamespace, msName, msNamespace, desiredReplicas)
}

// executeNsrScaledownWithDeletingMachineDeploymentWithReconciler executes scale-down scenario with deleting MachineDeployment
// with using provided reconciler.
func executeNsrScaledownWithDeletingMachineDeploymentWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	name, namespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName string,
	desiredReplicas int32,
) {
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
	})
	By("executing reconcile loop without patching MachineDeployment", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleting, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, "Requeueing", nil)
	})
	// cleanup node and machine
	Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
	Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
	})
	By("executing reconcile loop after finished request", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeDeleted, clusterName, clusterNamespace, &mdName, &desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionTrue, "CompletionSucceeded", nil)
	})
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(name, namespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

// executeNsrScaledownWithDeletingMachineDeployment executes scale-down scenario with deleting MachineDeployment
// with creating reconciler inside.
//
//nolint:unparam // parameters are kept for redundancy
func executeNsrScaledownWithDeletingMachineDeployment(
	providerType utils.ProviderType, wClient, mClient client.Client,
	name, namespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName string,
	desiredReplicas int32,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaledownWithDeletingMachineDeploymentWithReconciler(r, name, namespace, clusterName, clusterNamespace, nodeName, machineName, machineNamespace, mdName, desiredReplicas)
}

// executeNsrScaleupFailureStartScaleupWithReconciler for failure case of startScaleUp with using provided reconciler.
func executeNsrScaleupFailureStartScaleupWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition string,
) {
	// reconcile start scale-up
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).To(HaveOccurred())
		for _, errMsg := range errMsgs {
			Expect(err.Error()).To(ContainSubstring(errMsg))
		}
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, mdName, desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, statusRequestedCondition, errMsgs)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
	})
	// finalize
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("checking if NodeScalingRequest is already deleted", func() {
		nsr := &scalev1.NodeScalingRequest{}
		err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
}

// executeNsrScaleupFailureStartScaleup for failure case of startScaleUp with creating reconciler inside.
func executeNsrScaleupFailureStartScaleup(
	providerType utils.ProviderType, wClient, mClient client.Client,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition string,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaleupFailureStartScaleupWithReconciler(r, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, desiredReplicas, errMsgs, statusRequestedCondition)
}

// executeNsrScaleupFailureProceedScaleupWithReconciler for failure case of proceedScaleUp with using provided reconciler.
func executeNsrScaleupFailureProceedScaleupWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition, statusCompletedCondition string, insertFunc func(),
) {
	// reconcile start scale-up
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, mdName, desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, statusRequestedCondition, nil)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
	})
	// execute insertFunc to simulate scale-up failure
	insertFunc()
	// reconcile proceed scale-up failure
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).To(HaveOccurred())
		for _, errMsg := range errMsgs {
			Expect(err.Error()).To(ContainSubstring(errMsg))
		}
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeCreating, clusterName, clusterNamespace, mdName, desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionTrue, statusRequestedCondition, nil)
		assertNodeScalingRequestStatusCondition(nsr, Completed, metav1.ConditionFalse, statusCompletedCondition, errMsgs)
	})
	// finalize
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("checking if NodeScalingRequest is already deleted", func() {
		nsr := &scalev1.NodeScalingRequest{}
		err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
}

// executeNsrScaleupFailureProceedScaleup for failure case of proceedScaleUp with creating reconciler inside.
func executeNsrScaleupFailureProceedScaleup(
	providerType utils.ProviderType, wClient, mClient client.Client,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition, statusCompletedCondition string, insertFunc func(),
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaleupFailureProceedScaleupWithReconciler(r, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, desiredReplicas, errMsgs, statusRequestedCondition, statusCompletedCondition, insertFunc)
}

// executeNsrScaledownFailureStartScaledownWithReconciler for failure case of startScaleDown with using provided reconciler.
func executeNsrScaledownFailureStartScaledownWithReconciler(
	controllerReconciler *NodeScalingRequestReconciler,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition string,
) {
	// reconcile start scale-down failure
	By("executing reconcile loop", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).To(HaveOccurred())
		for _, errMsg := range errMsgs {
			Expect(err.Error()).To(ContainSubstring(errMsg))
		}
	})
	By("asserting NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		assertNodeScalingRequestStatus(nsr, scalev1.PhaseNodeScaleInitial, clusterName, clusterNamespace, mdName, desiredReplicas)
		Expect(nsr.Status.Conditions).NotTo(BeNil())
		assertNodeScalingRequestStatusCondition(nsr, Initialized, metav1.ConditionTrue, "InitializationSucceeded", nil)
		assertNodeScalingRequestStatusCondition(nsr, Requested, metav1.ConditionFalse, statusRequestedCondition, errMsgs)
		Expect(meta.FindStatusCondition(nsr.Status.Conditions, Completed)).To(BeNil())
	})
	// finalize
	By("cleaning up NodeScalingRequest", func() {
		nsr := &scalev1.NodeScalingRequest{}
		Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
		Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
	})
	By("executing reconcile finalizer", func() {
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: testutils.NewNamespacedName(nsrName, nsrNamespace),
		})
		Expect(err).NotTo(HaveOccurred())
	})
	By("checking if NodeScalingRequest is already deleted", func() {
		nsr := &scalev1.NodeScalingRequest{}
		err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
}

// executeNsrScaledownFailureStartScaledown for failure case of startScaleDown with creating reconciler inside.
//
//nolint:unparam // parameters are kept for redundancy
func executeNsrScaledownFailureStartScaledown(
	providerType utils.ProviderType, wClient, mClient client.Client,
	nsrName, nsrNamespace, clusterName, clusterNamespace string, mdName *string, desiredReplicas *int32,
	errMsgs []string, statusRequestedCondition string,
) {
	r := newNodeScalingRequestReconciler(providerType, wClient, mClient)
	executeNsrScaledownFailureStartScaledownWithReconciler(r, nsrName, nsrNamespace, clusterName, clusterNamespace, mdName, desiredReplicas, errMsgs, statusRequestedCondition)
}
