package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"
	"scale-controller/test/testutils"
	"scale-controller/test/testutils/mock"
)

// Test for SetupWithManager() method
var _ = Describe("CardScalingRequest Controller Setting Up", Label("csr", "settings", "ttl"), func() {
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
			logger = ctrl.Log.WithName("cardscalingrequest-controller-test")
		)

		BeforeEach(func() {
			ctrl.NewControllerManagedBy = newControllerManagedByMock
		})
		AfterEach(func() {
			Expect(os.Unsetenv(cardScalingRequestTTLEnvVarName)).To(Succeed())
			ctrl.NewControllerManagedBy = newControllerManagedBy
		})

		It("should successfully setup with enabling TTL", func() {
			// setting variables
			var (
				settingTTL = "1"
			)

			// setting environment variables
			Expect(os.Setenv(cardScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should successfully setup with disabling TTL", func() {
			// setting environment variables
			Expect(os.Unsetenv(cardScalingRequestTTLEnvVarName)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: mockClient})
			Expect(r.SetupWithManager(mgr)).To(Succeed())
		})

		It("should unsuccessfully setup with setting invalid empty CSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingTTL = ""
				errMsgs    = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds is empty",
				}
			)

			// setting environment variables
			Expect(os.Setenv(cardScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid character CSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingTTL = "notNumber"
				errMsgs    = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds cannot convert into int32",
				}
			)

			// setting environment variables
			Expect(os.Setenv(cardScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with setting invalid minus CSR_TTL_SECONDS_AFTER_FINISHED", func() {
			// setting variables
			var (
				settingTTL = "-1"
				errMsgs    = []string{
					"failed to get TTL configuration from environment variable",
					"argument for setting TTL seconds is invalid(" + settingTTL + ")",
				}
			)

			// setting environment variables
			Expect(os.Setenv(cardScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
			mockClient := mock.NewMockClientHelper(gomock.NewController(GinkgoT()))

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: mockClient})
			err := r.SetupWithManager(mgr)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully setup with nil HelperClient", func() {
			// setting variables
			var (
				settingTTL = "3600"
				errMsgs    = []string{
					"expected ClientHelper is not nil, but got nil",
				}
			)

			// setting environment variables
			Expect(os.Setenv(cardScalingRequestTTLEnvVarName, settingTTL)).To(Succeed())

			// creating mock manager
			mgr := mock.NewMockManager(gomock.NewController(GinkgoT()))
			mgr.EXPECT().GetLogger().AnyTimes().Return(logger)

			// executing SetupWithManager
			r := (&CardScalingRequestReconciler{ClientHelper: nil})
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
			r := (&CardScalingRequestReconciler{ClientHelper: nil})
			err := r.SetupWithManager(nil)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})
	})
})

// Test for internal methods with direct calls
var _ = Describe("CardScalingRequest Controller Methods directly calling", Label("csr", "methods"), func() {
	Context("When calling methods directly", func() {
		It("should unsuccessfully call doReconcile with invalid csr.Status", func() {
			By("calling doReconcile", func() {
				const dummyPhase = "dummy"
				var errMsgs = []string{
					fmt.Sprintf("unexpected phase(%s) is detected in doReconcile", dummyPhase),
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{
					Status: scalev1.CardScalingRequestStatus{
						Phase: dummyPhase,
					},
				}
				_, err := r.doReconcile(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call doFinalize with invalid csr.Status", func() {
			By("calling doFinalize", func() {
				const dummyPhase = "dummy"
				var errMsgs = []string{
					fmt.Sprintf("unexpected phase(%s) is detected in doFinalize", dummyPhase),
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{
					Status: scalev1.CardScalingRequestStatus{
						Phase: dummyPhase,
					},
				}
				_, err := r.doFinalize(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileDCIMachineTemplateCreating with invalid csr.Status", func() {
			By("calling reconcileDCIMachineTemplateCreating", func() {
				var errMsgs = []string{
					"expected OldDCIMachineInfo and NewDCIMachineInfo are not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileDCIMachineTemplateCreating(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileMachineDeploymentCreating with invalid csr.Status", func() {
			By("calling reconcileMachineDeploymentCreating", func() {
				var errMsgs = []string{
					"expected OldDCIMachineInfo and NewDCIMachineInfo are not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileMachineDeploymentCreating(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call generateAnnotationCardScalingInfo with invalid csr", func() {
			By("calling generateAnnotationCardScalingInfo", func() {
				var errMsgs = []string{
					"expected CardScalingRequest is not nil and ScalingParameter is not empty, but got invalid value",
				}
				csr := &scalev1.CardScalingRequest{}
				_, err := generateAnnotationCardScalingInfo(csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileDCIMachineAnnotating with invalid csr.Status", func() {
			By("calling reconcileDCIMachineAnnotating", func() {
				var errMsgs = []string{
					"expected OldDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileDCIMachineAnnotating(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileNodeScaleDownCompleted with invalid csr.Status", func() {
			By("calling reconcileNodeScaleDownCompleted", func() {
				var errMsgs = []string{
					"expected NewDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileNodeScaleDownCompleted(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileNodeScalingUp with invalid csr.Status", func() {
			By("calling reconcileNodeScalingUp", func() {
				var errMsgs = []string{
					"expected NewDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileNodeScalingUp(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileTerminalPhase with invalid csr.Status", func() {
			By("calling reconcileTerminalPhase", func() {
				var errMsgs = []string{
					"expected LastTransitionTime is not nil in terminal phase, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				r.enableTTLSecondsAfterFinished = true
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileTerminalPhase(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileDCIMachineCleanup with invalid csr.Status", func() {
			By("calling reconcileDCIMachineCleanup", func() {
				var errMsgs = []string{
					"expected OldDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileDCIMachineCleanup(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileMachineDeploymentCleanup with invalid csr.Status", func() {
			By("calling reconcileMachineDeploymentCleanup", func() {
				var errMsgs = []string{
					"expected NewDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileMachineDeploymentCleanup(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileDCIMachineTemplateCleanup with invalid csr.Status", func() {
			By("calling reconcileDCIMachineTemplateCleanup", func() {
				var errMsgs = []string{
					"expected NewDCIMachineInfo is not nil, but got nil",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{}
				_, err := r.reconcileDCIMachineTemplateCleanup(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})

	Context("When calling methods directly with client", func() {
		const (
			csrNamespace           = defaultNamespace
			clusterName            = "test-cluster"
			clusterNamespace       = defaultNamespace
			mdName                 = "test-md"
			mdNamespace            = defaultNamespace
			dciMachineName         = "test-dci-machine"
			dciMachineNamespace    = defaultNamespace
			dciMachineTemplateName = "test-dci-machine-template"
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

		AfterEach(func() {
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())

			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully call reconcileMachineDeploymentCreating with invalid MachineDeployment", func() {
			By("creating Dummy MachineDeployment resource", func() {
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			})
			By("patching MachineDeployment resource to remove Bootstrap configRef", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				md.Spec.Template.Spec.Bootstrap.ConfigRef = capiv1.ContractVersionedObjectReference{}
				md.Spec.Template.Spec.Bootstrap.DataSecretName = ptr.To("dummy")
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("calling reconcileMachineDeploymentCreating", func() {
				var errMsgs = []string{
					"failed to create new MachineDeployment instance",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{
					Status: scalev1.CardScalingRequestStatus{
						ClusterNamespace: clusterNamespace,
						OldDCIMachineInfo: &scalev1.DCIMachineInfo{
							MachineDeploymentName: mdName,
						},
						NewDCIMachineInfo: &scalev1.DCIMachineInfo{
							DCIMachineTemplateName: "new-" + mdName,
						},
					},
				}
				_, err := r.reconcileMachineDeploymentCreating(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should unsuccessfully call reconcileDCIMachineCleanup with invalid csr.Status", func() {
			By("creating Dummy DCIMachine resource", func() {
				Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			})
			By("calling reconcileDCIMachineCleanup", func() {
				var errMsgs = []string{
					"failed to generate annotation info for DCIMachine",
				}
				r := newCsrReconciler(nil, nil)
				csr := &scalev1.CardScalingRequest{
					Status: scalev1.CardScalingRequestStatus{
						ClusterNamespace: clusterNamespace,
						OldDCIMachineInfo: &scalev1.DCIMachineInfo{
							DCIMachineName: dciMachineName,
						},
					},
				}
				_, err := r.reconcileDCIMachineCleanup(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})
})

// Test for kubebuilder markers validation in CRD
var _ = Describe("CardScalingRequest Controller embedded Validation in CRD", Label("csr", "crd"), func() {
	Context("When creating CardScalingRequest with invalid parameters", func() {
		// error messages
		const (
			pathSpec                      = "spec"
			pathStatus                    = "status"
			pathSpecNodeName              = "nodeName"
			pathSpecScalingParameter      = "scalingParameter"
			pathSpecScalingParameterOp    = "op"
			pathSpecScalingParameterModel = "model"
			pathSpecScalingParameterNum   = "num"
			pathSpecScalingParameterUUIDs = "uuids"
			pathStatusPhase               = "phase"
		)
		const (
			errMsgAddNeedNum           = "When op is 'add', only 'num' must be set."
			errMsgDeleteNeedNumOrUUIDs = "When op is 'delete', either 'num' or 'uuids' must be set, but not both."
			errMsgReplaceNeedUUIDs     = "When op is 'replace', only 'uuids' must be set."
			errMsgNodeNameRequired     = pathSpec + "." + pathSpecNodeName + ": Required value"
			errMsgOpUnsupported        = pathSpec + "." + pathSpecScalingParameter + "." + pathSpecScalingParameterOp + ": Unsupported value"
			errMsgOpRequired           = pathSpec + "." + pathSpecScalingParameter + "." + pathSpecScalingParameterOp + ": Required value"
			errMsgModelRequired        = pathSpec + "." + pathSpecScalingParameter + "." + pathSpecScalingParameterModel + ": Required value"
			errMsgNumInvalid           = pathSpec + "." + pathSpecScalingParameter + "." + pathSpecScalingParameterNum + ": Invalid value"
			errMsgUUIDsInvalid         = pathSpec + "." + pathSpecScalingParameter + "." + pathSpecScalingParameterUUIDs + ": Invalid value"
			errMsgPhaseUnsupported     = pathStatus + "." + pathStatusPhase + ": Unsupported value"
			errMsgModifyExisting       = "`" + pathSpec + "` may not be modified after creation."
		)
		// common const
		const (
			csrNamespace = defaultNamespace
		)

		AfterEach(func() {
			// workload
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		newCardScalingRequestObjMap := func(name, namespace string) map[string]any {
			return map[string]any{
				"apiVersion": scalev1.GroupVersion.String(),
				"kind":       "CardScalingRequest",
				"metadata": map[string]any{
					"name":      name,
					"namespace": namespace,
				},
				pathSpec: map[string]any{
					pathSpecNodeName: "dummy-node",
					pathSpecScalingParameter: map[string]any{
						pathSpecScalingParameterOp:    string(capdv1.DCIMachineAnnotationCardScalingInfoOpAdd),
						pathSpecScalingParameterModel: "dummy-model",
						pathSpecScalingParameterNum:   1,
						pathSpecScalingParameterUUIDs: []string{
							"gpu-uuid-1",
						},
					},
				},
			}
		}

		It("should fail to create CardScalingRequest without NodeName", func() {
			const csrName string = "test-validation-create-without-nodeName"
			var (
				errMsgs = []string{
					errMsgNodeNameRequired,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecNodeName)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest without Operation", func() {
			const csrName string = "test-validation-create-without-operation"
			var (
				errMsgs = []string{
					errMsgOpRequired,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with empty Operation", func() {
			const csrName string = "test-validation-create-empty-operation"
			var (
				op      = "" // empty op is invalid
				errMsgs = []string{
					errMsgOpUnsupported,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest without Model", func() {
			const csrName string = "test-validation-create-without-model"
			var (
				errMsgs = []string{
					errMsgModelRequired,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterModel)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Add Operation and no Num", func() {
			const csrName string = "test-validation-create-add-no-num"
			var (
				errMsgs = []string{
					errMsgAddNeedNum,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Add Operation and zero Num", func() {
			const csrName string = "test-validation-create-add-zero-num"
			var (
				errMsgs = []string{
					errMsgNumInvalid,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				Expect(unstructured.SetNestedField(csrObj, int64(0), pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Add Operation and uuid without Num", func() {
			const csrName string = "test-validation-create-add-uuid-without-num"
			var (
				errMsgs = []string{
					errMsgAddNeedNum,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Add Operation and UUID and Num", func() {
			const csrName string = "test-validation-create-add-uuid-and-num"
			var (
				errMsgs = []string{
					errMsgAddNeedNum,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Delete Operation and zero Num", func() {
			const csrName string = "test-validation-create-delete-zero-num"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpDelete)
				errMsgs = []string{
					errMsgNumInvalid,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				Expect(unstructured.SetNestedField(csrObj, int64(0), pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Delete Operation without uuid and Num", func() {
			const csrName string = "test-validation-create-delete-without-uuid-and-num"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpDelete)
				errMsgs = []string{
					errMsgDeleteNeedNumOrUUIDs,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Delete Operation and UUID and Num", func() {
			const csrName string = "test-validation-create-delete-uuid-and-num"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpDelete)
				errMsgs = []string{
					errMsgDeleteNeedNumOrUUIDs,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Delete Operation and zero length UUID", func() {
			const csrName string = "test-validation-create-delete-zero-length-uuid"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpDelete)
				uuids   = []any{} // zero length UUIDs is invalid
				errMsgs = []string{
					errMsgUUIDsInvalid,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				Expect(unstructured.SetNestedField(csrObj, uuids, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Replace Operation and Num without UUIDs", func() {
			const csrName string = "test-validation-create-replace-without-uuids"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpReplace)
				errMsgs = []string{
					errMsgReplaceNeedUUIDs,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Replace Operation and UUID and Num", func() {
			const csrName string = "test-validation-create-replace-uuid-and-num"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpReplace)
				errMsgs = []string{
					errMsgReplaceNeedUUIDs,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Replace Operation and zero length UUID", func() {
			const csrName string = "test-validation-create-replace-zero-length-uuid"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpReplace)
				uuids   = []any{} // zero length UUIDs is invalid
				errMsgs = []string{
					errMsgUUIDsInvalid,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				Expect(unstructured.SetNestedField(csrObj, uuids, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)).To(Succeed())
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to create CardScalingRequest with Replace Operation without UUID", func() {
			const csrName string = "test-validation-create-replace-without-uuid"
			var (
				op      = string(capdv1.DCIMachineAnnotationCardScalingInfoOpReplace)
				errMsgs = []string{
					errMsgReplaceNeedUUIDs,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				Expect(unstructured.SetNestedField(csrObj, op, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterOp)).To(Succeed())
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterNum)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				err := k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update CardScalingRequest with invalid Phase", func() {
			const csrName string = "test-validation-update-invalid-phase"
			var (
				errMsgs = []string{
					errMsgPhaseUnsupported,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				Expect(k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})).To(Succeed())
			})
			By("updating CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				csr.Status.Phase = "InvalidPhase"
				err := k8sClientWorkload.Status().Update(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})

		It("should fail to update CardScalingRequest with modifying Spec", func() {
			const csrName string = "test-validation-update-modifying-spec"
			var (
				errMsgs = []string{
					errMsgModifyExisting,
				}
			)
			By("creating CardScalingRequest", func() {
				csrObj := newCardScalingRequestObjMap(csrName, csrNamespace)
				unstructured.RemoveNestedField(csrObj, pathSpec, pathSpecScalingParameter, pathSpecScalingParameterUUIDs)
				Expect(k8sClientWorkload.Create(ctx, &unstructured.Unstructured{Object: csrObj})).To(Succeed())
			})
			By("updating CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				csr.Spec.NodeName += "-modified"
				err := k8sClientWorkload.Update(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})
})

// Test for Reconcile() method
var _ = Describe("CardScalingRequest Controller", Label("csr", "controller"), func() {
	Context("When reconciling a resource", func() {
		const (
			csrNamespace                 = defaultNamespace
			nodeName                     = "test-node"
			clusterName                  = "test-cluster"
			clusterNamespace             = defaultNamespace
			machineName                  = "test-machine"
			machineNamespace             = defaultNamespace
			mdName                       = "test-machinedeployment"
			mdNamespace                  = defaultNamespace
			dciMachineName               = "test-dci-machine"
			dciMachineNamespace          = defaultNamespace
			dciMachineTemplateName       = "test-dci-machine-template"
			dciMachineTemplateAdd2Name   = "test-dci-machine-template-add2"
			dciMachineTemplateMinus2Name = "test-dci-machine-template-minus2"
			dciMachineTemplateNamespace  = defaultNamespace
			logicalServerName            = "test-logical-server"
			logicalServerNamespace       = defaultNamespace
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
			serverSpecAdd2 = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model: model,
					Num:   int32(3) + int32(2),
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
			serverSpecMinus2 = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model: model,
					Num:   int32(3) - int32(2),
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
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateAdd2Name, dciMachineTemplateNamespace, location, serverSpecAdd2)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateMinus2Name, dciMachineTemplateNamespace, location, serverSpecMinus2)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card", func() {
			const csrName = "test-csr-adding-card"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding 2 cards (use existing DCIMachineTemplate)", func() {
			const csrName = "test-csr-adding-card-2"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					2,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card (use existing MachineDeployment with annotations, but without autoscaling)", func() {
			const csrName = "test-csr-adding-card"
			By("patching MachineDeployment to add autoscaler min size annotation", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					"dummy-annotation-key": "dummy-annotation-value",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding 2 cards (not use existing MachineDeployment with autoscaling)", func() {
			const csrName = "test-csr-adding-card-2"
			const mdNameWithAutoScaling = "test-machinedeployment-with-autoscaling"
			By("creating MachineDeployment after adding 2 cards", func() {
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdNameWithAutoScaling, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateAdd2Name, nil)).To(Succeed())
			})
			By("patching MachineDeployment to add autoscaler min size annotation", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdNameWithAutoScaling, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMinSizeAnnotation: "1",
					capiv1.AutoscalerMaxSizeAnnotation: "2",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					2,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("asserting MachineDeployment name", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(csr.Status.NewDCIMachineInfo).ToNot(BeNil())
				Expect(csr.Status.NewDCIMachineInfo.MachineDeploymentName).NotTo(Equal(mdNameWithAutoScaling))
			})
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when deleting a card with num", func() {
			const csrName = "test-csr-deleting-card-with-num"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when deleting 2 cards with num (use existing DCIMachineTemplate)", func() {
			const csrName = "test-csr-deleting-card-with-num-2"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					2,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when deleting a card with uuids", func() {
			const csrName = "test-csr-deleting-card-with-uuids"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					0,
					[]string{uuids[0]},
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when replacing a card with uuids", func() {
			const csrName = "test-csr-replacing-card-with-uuids"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					uuids,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when deleting a card with uuids, but already the target card is deleted", func() {
			const csrName = "test-csr-deleting-card-with-uuids-already-deleted"
			r := newCsrReconciler(nil, nil)
			By("patching LogicalServer to remove one GPU assignment", func() {
				gpuSpec := &rpmv1.GPUSpec{
					Model: model,
					Num:   int32(2), // reduce GPU num by 1 to simulate the target card is already deleted
					Assignments: []rpmv1.GPUAssignment{
						{UUID: uuids[1]},
						{UUID: uuids[2]},
					},
				}
				Expect(testutils.PatchLogicalServerGPUSpec(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, gpuSpec)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					0,
					[]string{uuids[0]},
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when replacing a card with uuids, but already the target card is replaced", func() {
			const csrName = "test-csr-replacing-card-with-uuids-already-replaced"
			r := newCsrReconciler(nil, nil)
			By("patching LogicalServer to remove one GPU assignment", func() {
				gpuSpec := &rpmv1.GPUSpec{
					Model: model,
					Num:   int32(2), // reduce GPU num by 1 to simulate the target card is already deleted
					Assignments: []rpmv1.GPUAssignment{
						{UUID: uuids[1]},
						{UUID: uuids[2]},
					},
				}
				Expect(testutils.PatchLogicalServerGPUSpec(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, gpuSpec)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					[]string{uuids[0]},
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when deleting a card with uuids, but already some of the target cards are deleted", func() {
			const csrName = "test-csr-deleting-card-with-uuids-some-already-deleted"
			r := newCsrReconciler(nil, nil)
			By("patching LogicalServer to remove one GPU assignment", func() {
				gpuSpec := &rpmv1.GPUSpec{
					Model: model,
					Num:   int32(2), // reduce GPU num by 1 to simulate the target card is already replaced or deleted
					Assignments: []rpmv1.GPUAssignment{
						{UUID: uuids[1]},
						{UUID: uuids[2]},
					},
				}
				Expect(testutils.PatchLogicalServerGPUSpec(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, gpuSpec)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					0,
					[]string{uuids[0], uuids[1]}, // 1 card is already deleted, but still try to delete both cards to verify the reconcile can be successful
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when replacing a card with uuids, but already some of the target cards are replaced", func() {
			const csrName = "test-csr-replacing-card-with-uuids-some-already-replaced"
			r := newCsrReconciler(nil, nil)
			By("patching LogicalServer to remove one GPU assignment", func() {
				gpuSpec := &rpmv1.GPUSpec{
					Model: model,
					Num:   int32(2), // reduce GPU num by 1 to simulate the target card is already replaced or deleted
					Assignments: []rpmv1.GPUAssignment{
						{UUID: uuids[1]},
						{UUID: uuids[2]},
					},
				}
				Expect(testutils.PatchLogicalServerGPUSpec(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, gpuSpec)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					[]string{uuids[0], uuids[1]}, // 1 card is already replaced, but still try to replace both cards to verify the reconcile can be successful
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when another CardScalingRequest exists after finished", func() {
			const csrName = "test-csr-adding-card"
			const csrAnotherName = "test-csr-adding-card-another"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("creating another CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrAnotherName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
			By("executing Reconcile for another CardScalingRequest", func() {
				rslt, err := r.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(csrAnotherName, csrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(rslt.RequeueAfter).To(Equal(requeueIntervalCardScalingRequestPending))
			})
			By("asserting CardScalingRequest status is "+string(scalev1.PhaseCardScalePending), func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrAnotherName, csrNamespace), csr)).To(Succeed())
				Expect(csr.Status.Phase).To(Equal(scalev1.PhaseCardScalePending))
			})
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrReconcile(r, csrAnotherName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrAnotherName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrAnotherName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card with waiting for scaling down NodeScalingRequest", func() {
			const csrName = "test-csr-adding-card-with-waiting-scaledown-nsr"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)).To(Succeed())
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card with waiting for scaling up NodeScalingRequest", func() {
			const csrName = "test-csr-adding-card-with-waiting-scaleup-nsr"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when Node has no annotations", func() {
			const csrName = "test-csr-without-node-annotations"
			var (
				errMsgs = []string{
					"failed to get Cluster info from Node resource",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("updating Node to remove annotations", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, client.ObjectKey{Name: nodeName}, node)).To(Succeed())
				node.Annotations = map[string]string{}
				Expect(k8sClientWorkload.Update(ctx, node)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when DCIMachine has annotations already", func() {
			const csrName = "test-csr-with-dm-annotations-already"
			var (
				errMsgs = []string{
					"expected DCIMachine is not annotated for card scaling, but already annotated",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("updating MachineDeployment to add annotations", func() {
				dm := &capdv1.DCIMachine{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(dciMachineName, dciMachineNamespace), dm)).To(Succeed())
				dm.Annotations = utils.AddDCIMachineCardScalingInfoAnnotation(dm.Annotations, "dummy-value")
				Expect(k8sClientManagement.Update(ctx, dm)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when CardScalingRequest is invalid: model name", func() {
			const csrName = "test-csr-with-invalid-model-name"
			var (
				invalidModel = "invalid-model-name"
				errMsgs      = []string{
					"failed to generate expected DCIMachineSpec after scaling",
					"failed to generate ServerSpec after scaling",
					fmt.Sprintf("expected scaling parameter model(%s) is the same as ServerSpec GPU", invalidModel),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					invalidModel,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when CardScalingRequest is invalid: too many gpu uuids", func() {
			const csrName = "test-csr-with-invalid-gpu-uuids"
			var (
				invalidUUIDs = []string{"gpu-uuid-1", "gpu-uuid-2", "gpu-uuid-3", "gpu-uuid-4"}
				errMsgs      = []string{
					"failed to generate expected DCIMachineSpec after scaling",
					"failed to generate ServerSpec after scaling",
					fmt.Sprintf("expected the number of GPUs to replace(%d) is less than or equal to the current number of GPUs(%d), but greater",
						len(invalidUUIDs), len(uuids)),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					invalidUUIDs,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when CardScalingRequest is invalid: too many gpu num to delete", func() {
			const csrName = "test-csr-with-invalid-gpu-uuids"
			var (
				dummyGPUNum = int32(10)
				errMsgs     = []string{
					"failed to generate expected DCIMachineSpec after scaling",
					"failed to generate ServerSpec after scaling",
					fmt.Sprintf("expected the number of GPUs after scaling is greater than or equal to 0, but got %d",
						len(uuids)-int(dummyGPUNum)),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					dummyGPUNum, // too many num to delete
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to generate DCIMachineSpecAfterScaling at PhaseDCIMachineTemplateCreating", func() {
			const csrName = "test-csr-with-dcimachine-spec-generation-failure"
			var (
				deleteGPUNum = int32(1)
				errMsgs      = []string{
					"failed to generate expected DCIMachineSpec after scaling",
					"failed to generate ServerSpec after scaling",
					fmt.Sprintf("expected the number of GPUs after scaling is greater than or equal to 0, but got %d", -deleteGPUNum),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					deleteGPUNum,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed))).To(Succeed())
			err := executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to create new DCIMachineTemplate", func() {
			const csrName = "test-csr-with-create-new-dcimachinetemplate-failure"
			var (
				errMsgs = []string{
					"retry to create a new DCIMachineTemplate with new name",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("creating invalid DCIMachineTemplate to cause creation failure", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				invalidTemplateName := csr.Status.NewDCIMachineInfo.DCIMachineTemplateName
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, invalidTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to create new MachineDeployment", func() {
			const csrName = "test-csr-with-create-new-machinedeployment-failure"
			var (
				errMsgs = []string{
					"retry to create a new MachineDeployment with new name",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("creating invalid MachineDeployment to cause creation failure", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				invalidTemplateName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, invalidTemplateName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to generate annotation string", func() {
			const csrName = "test-csr-with-create-annotation-string"
			var (
				errMsgs = []string{
					"failed to generate annotation info for DCIMachine",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("patching CardScalingRequest status to remove new DCIMachineInfo to cause annotation generation failure", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				csr.Status.NewDCIMachineInfo = nil
				Expect(k8sClientWorkload.Status().Update(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		})

		It("should unsuccessfully reconcile when DCIMachine annotation is already set with unknown string", func() {
			const csrName = "test-csr-with-unknown-dcimachine-annotation-string"
			var (
				errMsgs = []string{
					"expected the annotation of DCIMachine is not annotated, but annotated with unexpected value",
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("updating DCIMachine to add unknown annotation string", func() {
				dm := &capdv1.DCIMachine{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(dciMachineName, dciMachineNamespace), dm)).To(Succeed())
				dm.Annotations = utils.AddDCIMachineCardScalingInfoAnnotation(dm.Annotations, "unknown-annotation-string")
				Expect(k8sClientManagement.Update(ctx, dm)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("cleaning up DCIMachine annotation", func() {
				dm := &capdv1.DCIMachine{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(dciMachineName, dciMachineNamespace), dm)).To(Succeed())
				utils.RemoveDCIMachineCardScalingInfoAnnotation(dm.Annotations)
				Expect(k8sClientManagement.Update(ctx, dm)).To(Succeed())
			})
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to create scaling down NodeScalingRequest with created by other owner", func() {
			const csrName = "test-csr-with-create-scaledown-nsr-by-other-owner"
			var (
				errMsgs = []string{
					fmt.Sprintf("expected existing NodeScalingRequest(%s/%s) is controlled by CardScalingRequest, but not controlled", csrNamespace, csrName),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)).To(Succeed())
			By("creating invalid NodeScalingRequest to cause creation failure", func() {
				nsr := testutils.NewNodeScalingRequest(csrName, csrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(nodeName)
				nsr.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion:         "dummy",
						Kind:               "dummy",
						Name:               "dummy",
						UID:                "dummy",
						Controller:         ptr.To(true),
						BlockOwnerDeletion: ptr.To(true),
					},
				}
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting invalid NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should unsuccessfully reconcile when it failed to create scaling up NodeScalingRequest with created by other owner", func() {
			const csrName = "test-csr-with-create-scaleup-nsr-by-other-owner"
			var (
				errMsgs = []string{
					fmt.Sprintf("expected existing NodeScalingRequest(%s/%s) is controlled by CardScalingRequest, but not controlled", csrNamespace, csrName),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting)).To(Succeed())
			By("creating invalid NodeScalingRequest to cause creation failure", func() {
				nsr := testutils.NewNodeScalingRequest(csrName, csrNamespace, scalev1.TypeCreate)
				nsr.Spec.ServerSpec = serverSpec
				nsr.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion:         "dummy",
						Kind:               "dummy",
						Name:               "dummy",
						UID:                "dummy",
						Controller:         ptr.To(true),
						BlockOwnerDeletion: ptr.To(true),
					},
				}
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting invalid NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
		})
	})

	Context("When reconciling a resource with checking various serverSpec,location", func() {
		const (
			csrNamespace                 = defaultNamespace
			nodeName                     = "test-node"
			clusterName                  = "test-cluster"
			clusterNamespace             = defaultNamespace
			machineName                  = "test-machine"
			machineNamespace             = defaultNamespace
			mdName                       = "test-machinedeployment"
			mdNamespace                  = defaultNamespace
			dciMachineName               = "test-dci-machine"
			dciMachineNamespace          = defaultNamespace
			dciMachineTemplateName       = "test-dci-machine-template"
			dciMachineTemplateNamePrefix = "test-dci-machine-template-"
			dciMachineTemplateNamespace  = defaultNamespace
			logicalServerName            = "test-logical-server"
			logicalServerNamespace       = defaultNamespace
		)
		var (
			locations = []*string{
				ptr.To("validLocation"),
				ptr.To("dummyLocation"),
				nil,
			}
			types = []rpmv1.ReqServerType{
				rpmv1.ReqServerTypeComposed,
				rpmv1.ReqServerTypeCOTS,
			}
			oss = []*rpmv1.OSType{
				{Distribution: "Ubuntu", Version: "22.04"},
				{Distribution: "Dummy", Version: "Dummy"},
				nil,
			}
			mins = []*int32{
				ptr.To(int32(10)),
				ptr.To(int32(5)),
				nil,
			}
			gpus = []*rpmv1.GPUSpec{
				{Model: "A100", Num: 4},
				{Model: "A100", Num: 1},
				{Model: "Dummy", Num: 1},
				nil,
			}
			nics = []*rpmv1.NICSpec{
				{Model: "ConnectX-5", Num: 1},
				{Model: "Dummy", Num: 1},
				nil,
			}
			serverSpec = &rpmv1.ServerSpec{
				Type:       types[0],
				OS:         oss[0],
				MinCPU:     mins[0],
				MinMemory:  mins[0],
				MinStorage: mins[0],
				GPU:        gpus[0],
				NIC:        nics[0],
			}
			serverSpecBase = &rpmv1.ServerSpec{
				Type:       types[0],
				OS:         oss[0],
				MinCPU:     mins[0],
				MinMemory:  mins[0],
				MinStorage: mins[0],
				GPU:        &rpmv1.GPUSpec{Model: "A100", Num: 2},
				NIC:        nics[0],
			}
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, locations[0], serverSpecBase)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, locations[0], "", serverSpecBase)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, locations[0], serverSpecBase, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
			cnt := 0
			for _, l := range locations[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, l, ss)).To(Succeed())
			}
			for _, t := range types[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.Type = t
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, os := range oss[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.OS = os
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, cpu := range mins[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.MinCPU = cpu
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, mem := range mins[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.MinMemory = mem
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, sto := range mins[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.MinStorage = sto
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, gpu := range gpus[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.GPU = gpu
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
			for _, nic := range nics[1:] {
				cnt++
				dmtNameTarget := dciMachineTemplateNamePrefix + fmt.Sprintf("%02d", cnt)
				ss := serverSpec.DeepCopy()
				ss.NIC = nic
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameTarget, dciMachineTemplateNamespace, locations[0], ss)).To(Succeed())
			}
		})
		AfterEach(func() {
			Expect(testutils.CleanupNodeResources(ctx, k8sClientWorkload)).To(Succeed())
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card", func() {
			const csrName = "test-csr-adding-card"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					gpus[0].Model,
					2,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})
	})

	Context("When reconciling a resource in different namespace", func() {
		const (
			csrNamespace                = testNamespace
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
			logicalServerNamespace      = testNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile when adding a card", func() {
			const csrName = "test-csr-adding-card"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})
	})

	Context("When reconciling called with no valid resource", func() {
		const (
			csrNamespace = defaultNamespace
		)

		AfterEach(func() {
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should fail to get a resource when no resource exist", func() {
			const csrName string = "test-not-exist"
			r := newCsrReconciler(nil, nil)
			By("executing reconcile", func() {
				_, err := r.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(csrName, csrNamespace),
				})
				Expect(err).NotTo(HaveOccurred()) // when not found has occurred, should return nil as err
			})
		})
	})

	Context("When reconciling a resource for testing Finalize from each phase", func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		exec := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase,
		) {
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(name, namespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, name, namespace, reconcileEnd)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)).To(Succeed())
		}

		execScale := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase,
			cleanupFunc func(*CardScalingRequestReconciler, string, string, scalev1.CardScalingRequestPhase) error,
		) {
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(name, namespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, name, namespace, reconcileEnd)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(cleanupFunc(r, name, namespace, finalizeStart)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)).To(Succeed())
		}

		execScaleDown := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase,
		) {
			execScale(name, namespace, reconcileEnd, finalizeStart, executeCsrFinalizeWaitingScaleDown)
		}

		execScaleUp := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase,
		) {
			execScale(name, namespace, reconcileEnd, finalizeStart, executeCsrFinalizeWaitingScaleUp)
		}

		It("should successfully finalize from PhaseCardScaleInitial", func() {
			const csrName = "test-csr-finalize-card-scale-initial"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				csr.Finalizers = []string{scalev1.CardScalingRequestFinalizerName} // Set Finalizer
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
		})

		It("should successfully finalize from PhaseCardScalePending", func() {
			const csrName = "test-csr-finalize"
			const csrAnotherName = "test-csr-finalize-card-scale-pending"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("creating another CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrAnotherName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
			By("executing Reconcile for another CardScalingRequest", func() {
				rslt, err := r.Reconcile(ctx, reconcile.Request{
					NamespacedName: testutils.NewNamespacedName(csrAnotherName, csrNamespace),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(rslt.RequeueAfter).To(Equal(requeueIntervalCardScalingRequestPending))
			})
			By("asserting CardScalingRequest status is "+string(scalev1.PhaseCardScalePending), func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrAnotherName, csrNamespace), csr)).To(Succeed())
				Expect(csr.Status.Phase).To(Equal(scalev1.PhaseCardScalePending))
			})
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrAnotherName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrFinalize(r, csrAnotherName, csrNamespace)).To(Succeed())
		})

		It("should successfully finalize from PhaseCardScaleInitializing", func() {
			const csrName = "test-csr-finalize-card-scale-initializing"
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitial, scalev1.PhaseCardScaleInitializing)
		})

		It("should successfully finalize from PhaseDCIMachineTemplateCreating", func() {
			const csrName = "test-csr-finalize-dcimachine-template-creating"
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitializing, scalev1.PhaseDCIMachineTemplateCreating)
		})

		It("should successfully finalize from PhaseMachineDeploymentCreating", func() {
			const csrName = "test-csr-finalize-machine-deployment-creating"
			exec(csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseMachineDeploymentCreating)
		})

		It("should successfully finalize from PhaseMachineDeploymentCreating if MachineDeployment with autoscaling exists", func() {
			const csrName = "test-csr-finalize-machine-deployment-creating-with-autoscaling"
			const mdNameWithAutoScaling = "test-machinedeployment-with-autoscaling"
			const dmtNameWithAutoScaling = "test-dci-machine-template-with-autoscaling"
			By("creating DCIMachineTemplate", func() {
				ss := serverSpec.DeepCopy()
				ss.GPU.Num += 1 // Increase GPU num for autoscaling
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtNameWithAutoScaling, dciMachineTemplateNamespace, location, ss)).To(Succeed())
			})
			By("creating MachineDeployment", func() {
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdNameWithAutoScaling, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dmtNameWithAutoScaling, nil)).To(Succeed())
			})
			By("patching MachineDeployment to add autoscaler min size annotation", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdNameWithAutoScaling, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMinSizeAnnotation: "1",
					capiv1.AutoscalerMaxSizeAnnotation: "2",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			exec(csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseMachineDeploymentCreating)
		})

		It("should successfully finalize from PhaseDCIMachineAnnotating", func() {
			const csrName = "test-csr-finalize-dcimachine-annotating"
			exec(csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseDCIMachineAnnotating)
		})

		It("should successfully finalize from PhaseCardScaleInitialized", func() {
			const csrName = "test-csr-finalize-card-scale-initialized"
			execScaleDown(csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating, scalev1.PhaseDCIMachineCleanup)
		})

		It("should successfully finalize from PhaseNodeScalingDown", func() {
			const csrName = "test-csr-finalize-node-scaling-down"
			execScaleDown(csrName, csrNamespace, scalev1.PhaseCardScaleInitialized, scalev1.PhaseMachineDeploymentCleanup)
		})

		It("should successfully finalize from PhaseNodeScaleDownDeleting", func() {
			const csrName = "test-csr-finalize-node-scale-down-deleting"
			execScaleDown(csrName, csrNamespace, scalev1.PhaseNodeScalingDown, scalev1.PhaseMachineDeploymentCleanup)
		})

		It("should successfully finalize from PhaseNodeScaleDownCompleted", func() {
			const csrName = "test-csr-finalize-node-scale-down-completed"
			execScaleUp(csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting, scalev1.PhaseMachineDeploymentCleanup)
		})

		It("should successfully finalize from PhaseNodeScalingUp", func() {
			const csrName = "test-csr-finalize-node-scaling-up"
			execScaleUp(csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted, scalev1.PhaseCardScaleFinalized)
		})
	})

	Context("When reconciling a resource for testing Finalize for various conditions", func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully skip to finalize if only other finalizer is set", func() {
			const csrName = "test-csr-finalize-only-other-finalizer-set"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				csr.Finalizers = []string{"dummy/cardscalingrequest"} // Set only other Finalizer
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)).To(Succeed())
			By("patching CardScalingRequest to remove other finalizer", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				csr.Finalizers = []string{} // Remove other Finalizer
				Expect(k8sClientWorkload.Update(ctx, csr)).To(Succeed())
			})
		})

		It("should successfully finalize if valid DCIMachineTemplate not found", func() {
			const csrName = "test-csr-finalize-no-valid-dcimachine-template"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("regenerating DCIMachineTemplate", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				dmtName := csr.Status.NewDCIMachineInfo.DCIMachineTemplateName
				dmt := &capdv1.DCIMachineTemplate{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(dmtName, dciMachineTemplateNamespace), dmt)).To(Succeed())
				newDmt := dmt.DeepCopy()
				newDmt.ResourceVersion = "" // Clear ResourceVersion to create a new resource
				Expect(k8sClientManagement.Delete(ctx, dmt)).To(Succeed())
				Expect(k8sClientManagement.Create(ctx, newDmt)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
		})

		It("should successfully finalize if valid MachineDeployments not found", func() {
			const csrName = "test-csr-finalize-no-valid-machinedeployments"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting MachineDeployments", func() {
				Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
		})

		It("should successfully finalize if valid MachineDeployment not found", func() {
			const csrName = "test-csr-finalize-no-valid-machinedeployment"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("regenerating MachineDeployment", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				mdName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, dciMachineTemplateNamespace), md)).To(Succeed())
				newMd := md.DeepCopy()
				newMd.ResourceVersion = "" // Clear ResourceVersion to create a new resource
				Expect(k8sClientManagement.Delete(ctx, md)).To(Succeed())
				Expect(k8sClientManagement.Create(ctx, newMd)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should successfully finalize if valid MachineDeployment has replicas already", func() {
			const csrName = "test-csr-finalize-machinedeployment-has-replicas"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("patching MachineDeployment to add replicas", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				mdName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, dciMachineTemplateNamespace), md)).To(Succeed())
				md.Spec.Replicas = ptr.To(int32(1))
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should successfully finalize if valid MachineDeployment is the last MachineDeployment", func() {
			const csrName = "test-csr-finalize-machinedeployment-last"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting existing MachineDeployment", func() {
				Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should successfully finalize if valid DCIMachine not found", func() {
			const csrName = "test-csr-finalize-no-valid-dcimachine"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("deleting existing DCIMachine", func() {
				Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should unsuccessfully finalize if NodeScalingRequest for scaling down is DeleteFailed", func() {
			const csrName = "test-csr-finalize-node-scaling-request-delete-failed"
			var (
				errMsgs = []string{
					fmt.Sprintf("unexpected phase(%s) of NodeScalingRequest", scalev1.PhaseNodeCreating),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("patching NodeScalingRequest to Deleting", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeDeleting
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)).To(Succeed())
			By("patching NodeScalingRequest to invalid(Creating)", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeCreating
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)
			Expect(err).To(HaveOccurred())
			for _, msg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(msg))
			}
			By("patching NodeScalingRequest to DeleteFailed", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeDeleteFailed
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineCleanup)).To(Succeed())
		})

		It("should unsuccessfully finalize if NodeScalingRequest for scaling up is CreateFailed", func() {
			const csrName = "test-csr-finalize-node-scaling-request-create-failed"
			var (
				errMsgs = []string{
					fmt.Sprintf("unexpected phase(%s) of NodeScalingRequest", scalev1.PhaseNodeDeleting),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			By("patching NodeScalingRequest to Creating", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeCreating
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("patching NodeScalingRequest to invalid(Deleting)", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeDeleting
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)
			Expect(err).To(HaveOccurred())
			for _, msg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(msg))
			}
			By("patching NodeScalingRequest to CreateFailed", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
		})
	})

	Context("When reconciling a resource", Label("mock"), func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-get-csr"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to get CardScalingRequest: %s/%s", csrNamespace, csrName),
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{1})).To(Succeed()) // failed to get CardScalingRequest
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Patch() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-patch-csr"
			var (
				errMsgs = []string{
					"failed to add finalizer to CardScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{1})).To(Succeed()) // failed to patch CardScalingRequest for adding finalizer
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Patch() and Status().Patch() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-patch-status-patch-csr"
			var (
				errMsgs = []string{
					"failed to add finalizer to CardScalingRequest",
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{1})).To(Succeed())       // failed to patch CardScalingRequest for adding finalizer
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to status.patch CardScalingRequest for PhaseCardScaleInitial
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of List() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-list-csr"
			var (
				errMsgs = []string{
					"failed to get other running CardScalingRequests",
					"failed to list CardScalingRequests",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetList([]int{1})).To(Succeed()) // failed to list CardScalingRequests
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for CardScalingRequest to check sync", func() {
			const csrName = "test-csr-mock-fail-get-csr-check-sync"
			var (
				errMsgs = []string{
					"failed to synchronize cached resource version of CardScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get CardScalingRequest to check sync
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for Node", func() {
			const csrName = "test-csr-mock-fail-get-node"
			var (
				errMsgs = []string{
					"failed to get Node resource",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{4})).To(Succeed()) // failed to get Node
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for Machine", func() {
			const csrName = "test-csr-mock-fail-get-machine"
			var (
				errMsgs = []string{
					"failed to get Machine resource from Node resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get Machine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get MachineDeployment resource from Machine resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{2})).To(Succeed()) // failed to get MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-get-dcimachinetemplate"
			var (
				errMsgs = []string{
					"failed to get DCIMachineTemplate resource from MachineDeployment resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{3})).To(Succeed()) // failed to get DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for DCIMachine", func() {
			const csrName = "test-csr-mock-fail-get-dcimachine"
			var (
				errMsgs = []string{
					"failed to get DCIMachine resource from Machine resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{4})).To(Succeed()) // failed to get DCIMachine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for LogicalServer", func() {
			const csrName = "test-csr-mock-fail-get-logicalserver"
			var (
				errMsgs = []string{
					"failed to get LogicalServer resource from DCIMachine resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{5})).To(Succeed()) // failed to get LogicalServer
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					[]string{"dummy-gpu-uuid"},
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of List() for DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-list-dcimachinetemplate"
			var (
				errMsgs = []string{
					"failed to get valid DCIMachineTemplate after scaling",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed()) // failed to list DCIMachineTemplates
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of List() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-list-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get valid MachineDeployment after scaling",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{2})).To(Succeed()) // failed to list MachineDeployments
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for old DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-get-old-dcimachinetemplate"
			var (
				errMsgs = []string{
					"failed to get old DCIMachineTemplate resource from management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{5})).To(Succeed()) // failed to get old DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Create() for new DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-create-new-dcimachinetemplate"
			var (
				errMsgs = []string{
					"failed to create DCIMachineTemplate resource in management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{1})).To(Succeed()) // failed to create new DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Status().Patch() for PhaseDCIMachineTemplateCreating", func() {
			const csrName = "test-csr-mock-fail-status-dcimachinetemplatecreating"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{3})).To(Succeed()) // failed to status.patch for PhaseDCIMachineTemplateCreating
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of List() for DCIMachineTemplate after failure of Status().Patch()", func() {
			const csrName = "test-csr-mock-fail-list-dcimachinetemplate-after-status-patch"
			var (
				errMsgs = []string{
					"failed to create unique name for DCIMachineTemplate",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{3})).To(Succeed()) // failed to status.patch for PhaseDCIMachineTemplateCreating
			wClient.SetDefault()
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{3})).To(Succeed()) // failed to List for DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred()) // This error is tested in other test case, so just check occurrence here
			err = executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for new MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-new-machinedeployment"
			var (
				mdAfterScaleName     = "md-after-scale"
				dmtAfterScaleName    = "dci-machine-template-after-scale"
				serverSpecAfterScale = serverSpec.DeepCopy()
				errMsgs              = []string{
					"failed to get new MachineDeployment after scaled resource from management cluster",
				}
			)
			serverSpecAfterScale.GPU.Num += 1
			By("creating MachineDeployment satisfying the scaled up requirement", func() {
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdAfterScaleName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dmtAfterScaleName, nil)).To(Succeed())
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtAfterScaleName, dciMachineTemplateNamespace, location, serverSpecAfterScale)).To(Succeed())
			})
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{5})).To(Succeed()) // failed to get new MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Patch() for new MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-patch-new-machinedeployment"
			var (
				mdAfterScaleName     = "md-after-scale"
				dmtAfterScaleName    = "dci-machine-template-after-scale"
				serverSpecAfterScale = serverSpec.DeepCopy()
				errMsgs              = []string{
					"failed to add annotation to new MachineDeployment after scaled resource",
				}
			)
			serverSpecAfterScale.GPU.Num += 1
			By("creating MachineDeployment satisfying the scaled up requirement", func() {
				Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdAfterScaleName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dmtAfterScaleName, nil)).To(Succeed())
				Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dmtAfterScaleName, dciMachineTemplateNamespace, location, serverSpecAfterScale)).To(Succeed())
			})
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed()) // failed to patch new MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for old MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-old-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get old MachineDeployment resource from management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{6})).To(Succeed()) // failed to get old MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Create() for new MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-create-new-machinedeployment"
			var (
				errMsgs = []string{
					"failed to create MachineDeployment resource in management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetCreate([]int{2})).To(Succeed()) // failed to create new MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Status().Patch() for PhaseMachineDeploymentCreating", func() {
			const csrName = "test-csr-mock-fail-status-machinedeploymentcreating"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{4})).To(Succeed()) // failed to status.patch for PhaseMachineDeploymentCreating
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for DCIMachineTemplate after failure of Status().Patch()", func() {
			const csrName = "test-csr-mock-fail-get-dcimachinetemplate-after-status-patch"
			var (
				errMsgs = []string{
					"failed to get DCIMachineTemplate resource from management cluster",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{4})).To(Succeed()) // failed to status.patch for PhaseMachineDeploymentCreating
			wClient.SetDefault()
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{8})).To(Succeed()) // failed to Get for DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred()) // This error is tested in other test case, so just check occurrence here
			err = executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of List() for MachineDeployment after failure of Status().Patch()", func() {
			const csrName = "test-csr-mock-fail-list-machinedeployment-after-status-patch"
			var (
				errMsgs = []string{
					"failed to create unique name for MachineDeployment",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{4})).To(Succeed()) // failed to status.patch for PhaseMachineDeploymentCreating
			wClient.SetDefault()
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{3})).To(Succeed()) // failed to List for MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred()) // This error is tested in other test case, so just check occurrence here
			err = executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for new DCIMachine", func() {
			const csrName = "test-csr-mock-fail-get-new-dcimachine"
			var (
				errMsgs = []string{
					"failed to get DCIMachine resource from management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{7})).To(Succeed()) // failed to get new DCIMachine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Patch() for new DCIMachine", func() {
			const csrName = "test-csr-mock-fail-patch-new-dcimachine"
			var (
				errMsgs = []string{
					"failed to patch DCIMachine resource in management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed()) // failed to patch new DCIMachine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Scheme() for setting owner reference to scaledown NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-scheme-workload-cluster-scaledown"
			var (
				errMsgs = []string{
					"failed to set owner reference to NodeScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetScheme([]int{1})).To(Succeed()) // failed to Scheme for workload cluster
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Create() for scaledown NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-create-scaledown-nsr"
			var (
				errMsgs = []string{
					"failed to create NodeScalingRequest for scaling down",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetCreate([]int{1})).To(Succeed()) // failed to Create NodeScalingRequest for deleting
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Status().Patch() for PhaseCardScaleInitialized", func() {
			const csrName = "test-csr-mock-fail-status-cardscaleinitialized"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{6})).To(Succeed()) // failed to status.patch for PhaseCardScaleInitialized
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for NodeScalingRequest after failure of Status().Patch()", func() {
			const csrName = "test-csr-mock-fail-get-nsr-after-status-patch"
			var (
				errMsgs = []string{
					"failed to get existing NodeScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{6})).To(Succeed()) // failed to status.patch for PhaseCardScaleInitialized
			Expect(wClient.SetGet([]int{14})).To(Succeed())        // failed to Get NodeScalingRequest after status.patch
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred()) // This error is tested in other test case, so just check occurrence here
			err = executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for NodeScalingRequest for scale down", func() {
			const csrName = "test-csr-mock-fail-get-nsr-scale-down"
			var (
				errMsgs = []string{
					"failed to get NodeScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{15})).To(Succeed()) // failed to Get NodeScalingRequest for deleting
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for NodeScalingRequest for scale down to delete", func() {
			const csrName = "test-csr-mock-fail-get-nsr-scale-down-to-delete"
			var (
				errMsgs = []string{
					"failed to get existing NodeScalingRequest for scaling down",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{18})).To(Succeed()) // failed to Get NodeScalingRequest for scale down
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Delete() for NodeScalingRequest for scale down", func() {
			const csrName = "test-csr-mock-fail-delete-nsr-scale-down"
			var (
				errMsgs = []string{
					"failed to delete existing NodeScalingRequest for scaling down",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetDelete([]int{1})).To(Succeed()) // failed to Delete NodeScalingRequest for scale down
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Status().Patch() for PhaseNodeScaleDownDeleting", func() {
			const csrName = "test-csr-mock-fail-status-for-nsr-scale-down-deleting"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{8})).To(Succeed()) // failed to status.patch for PhaseNodeScaleDownDeleting
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for new DCIMachineTemplate to create NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-get-dmt-to-create-nsr"
			var (
				errMsgs = []string{
					"failed to get DCIMachineTemplate resource from management cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{8})).To(Succeed()) // failed to get new DCIMachineTemplate
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Scheme() for setting owner reference to scaleup NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-scheme-workload-cluster-scaleup"
			var (
				errMsgs = []string{
					"failed to set owner reference to NodeScalingRequest for scaling up",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetScheme([]int{2})).To(Succeed()) // failed to Scheme for workload cluster client
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Create() for scaleup NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-create-scaleup-nsr"
			var (
				errMsgs = []string{
					"failed to create NodeScalingRequest for scaling up",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetCreate([]int{2})).To(Succeed()) // failed to Create NodeScalingRequest for scaling up
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Status().Patch() for PhaseNodeScaleDownCompleted", func() {
			const csrName = "test-csr-mock-fail-status-nodescaledowncompleted"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{9})).To(Succeed()) // failed to status.patch for PhaseNodeScaleDownCompleted
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for scaleup NodeScalingRequest after failure of Status().Patch()", func() {
			const csrName = "test-csr-mock-fail-get-scaleup-nsr-after-status-patch"
			var (
				errMsgs = []string{
					"failed to get existing NodeScalingRequest for scaling up",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{9})).To(Succeed()) // failed to status.patch for PhaseNodeScaleDownCompleted
			Expect(wClient.SetGet([]int{22})).To(Succeed())        // failed to Get NodeScalingRequest after status.patch
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred()) // This error is tested in other test case, so just check occurrence here
			err = executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Create() for scaleup NodeScalingRequest with DeletionTimestamp", func() {
			const csrName = "test-csr-mock-fail-create-nsr-with-deletion-timestamp"
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)).To(Succeed())
			By("patching dummy finalizer to NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Finalizers = append(nsr.Finalizers, "dummy.finalizer/test")
				Expect(k8sClientWorkload.Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("removing dummy finalizer from NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Finalizers = []string{}
				Expect(k8sClientWorkload.Update(ctx, nsr)).To(Succeed())
			})
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for scaleup NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-get-scaleup-nsr"
			var (
				errMsgs = []string{
					"failed to get NodeScalingRequest for scaling up",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{23})).To(Succeed()) // failed to Get NodeScalingRequest
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for new MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-new-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get MachineDeployment annotation for card scaling",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{9})).To(Succeed()) // failed to get new MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Patch() for new MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-patch-new-machinedeployment"
			var (
				errMsgs = []string{
					"failed to remove annotation from MachineDeployment after scaling",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{2})).To(Succeed()) // failed to patch new MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for Node", func() {
			const csrName = "test-csr-mock-fail-get-node"
			var (
				errMsgs = []string{
					"failed to get Node resource",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{24})).To(Succeed()) // failed to Get Node
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for created Machine", func() {
			const csrName = "test-csr-mock-fail-get-created-machine"
			var (
				errMsgs = []string{
					"failed to get Machine resource from Node resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{10})).To(Succeed()) // failed to get new Machine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile after failure of Get() for created DCIMachine", func() {
			const csrName = "test-csr-mock-fail-get-created-dcimachine"
			var (
				errMsgs = []string{
					"failed to get DCIMachine resource from Machine resource",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{11})).To(Succeed()) // failed to get new DCIMachine
			mClient.SetDefault()
			r := newCsrReconciler(nil, mClient.GetMockClient())
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			err := executeCsrReconcile(r, csrName, csrNamespace)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})
	})

	Context("When reconciling a resource for testing Finalize", Label("mock"), func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		exec := func(
			name, namespace string, reconcileEnd, finalizeStart, finalizeRestart scalev1.CardScalingRequestPhase, wClient, mClient client.Client, errMsgs []string,
		) {
			r := newCsrReconciler(nil, nil) // generate reconciler with real clients
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(name, namespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, name, namespace, reconcileEnd)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			r = newCsrReconciler(wClient, mClient) // regenerate reconciler with mock clients
			err := executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrFinalizeFromPhase(r, name, namespace, finalizeRestart)).To(Succeed())
		}

		// execCardScaleInitialized is used to test after DCIMachineAnnotating phase is completed.
		// To finalize from the CardScaleInitialized phase, there are many branches depending on the status of NodeScalingRequest, so this case is separated.
		// This function test failure of executeCsrFinalizeFromPhase.
		execCardScaleInitialized := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase, wClient, mClient client.Client, errMsgs []string,
			cleanupFunc func(r *CardScalingRequestReconciler, name, namespace string, phase scalev1.CardScalingRequestPhase) error,
		) {
			// Reconcile
			r := newCsrReconciler(nil, nil) // generate reconciler with real clients
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(name, namespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, name, namespace, reconcileEnd)).To(Succeed())
			// Finalize
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			r = newCsrReconciler(wClient, mClient) // regenerate reconciler with mock clients
			Expect(cleanupFunc(r, name, namespace, finalizeStart)).To(Succeed())
			err := executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)).To(Succeed())

		}

		// execScale is used to test after DCIMachineAnnotating phase is completed.
		// To finalize from the NodeScalingRequest scaling phase, there are many branches depending on the status of NodeScalingRequest, so this case is separated.
		// This function test failure of cleanupFunc.
		execScale := func(
			name, namespace string, reconcileEnd, finalizeStart scalev1.CardScalingRequestPhase, wClient, mClient client.Client, errMsgs []string,
			cleanupFunc func(r *CardScalingRequestReconciler, name, namespace string, phase scalev1.CardScalingRequestPhase) error,
		) {
			// Reconcile
			r := newCsrReconciler(nil, nil) // generate reconciler with real clients
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(name, namespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, name, namespace, reconcileEnd)).To(Succeed())
			// Finalize
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(name, namespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			r = newCsrReconciler(wClient, mClient) // regenerate reconciler with mock clients
			err := cleanupFunc(r, name, namespace, finalizeStart)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(cleanupFunc(r, name, namespace, finalizeStart)).To(Succeed())
			Expect(executeCsrFinalizeFromPhase(r, name, namespace, finalizeStart)).To(Succeed())
		}

		It("should successfully finalize after failure of Patch() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-patch-csr"
			var (
				errMsgs = []string{
					"failed to remove finalizer from CardScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{1})).To(Succeed()) // failed to patch CardScalingRequest
			wClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitial, scalev1.PhaseCardScaleInitializing, scalev1.PhaseCardScaleInitializing, wClient.GetMockClient(), nil, errMsgs)
		})

		It("should successfully finalize after failure of Get() for DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-get-dci-machine-template"
			var (
				errMsgs = []string{
					"failed to get existing DCIMachineTemplate",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get DCIMachineTemplate
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitializing, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseDCIMachineTemplateCreating, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of List() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-list-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get MachineDeployments using the DCIMachineTemplate",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed()) // failed to list MachineDeployment
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseDCIMachineTemplateCleanup, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of Delete() for DCIMachineTemplate", func() {
			const csrName = "test-csr-mock-fail-delete-dci-machine-template"
			var (
				errMsgs = []string{
					"failed to delete existing DCIMachineTemplate",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{1})).To(Succeed()) // failed to delete DCIMachineTemplate
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseDCIMachineTemplateCleanup, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of Get() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-machinedeployment"
			var (
				errMsgs = []string{
					"failed to get existing MachineDeployment",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get MachineDeployment
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseMachineDeploymentCreating, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of List() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-get-machinedeployment"
			var (
				errMsgs = []string{
					"failed to check if MachineDeployment is the last one in the cluster",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetList([]int{1})).To(Succeed()) // failed to list MachineDeployment
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseDCIMachineAnnotating, scalev1.PhaseMachineDeploymentCleanup, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of Patch() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-patch-machinedeployment"
			var (
				errMsgs = []string{
					"failed to patch MachineDeployment to remove annotation",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed()) // failed to patch MachineDeployment
			mClient.SetDefault()
			r := newCsrReconciler(nil, nil) // generate reconciler with real clients
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)).To(Succeed())
			By("patching MachineDeployment to add replicas", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				newMDName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(newMDName, mdNamespace), md)).To(Succeed())
				md.Spec.Replicas = ptr.To(int32(1))
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			r = newCsrReconciler(nil, mClient.GetMockClient()) // regenerate reconciler with mock clients
			err := executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)).To(Succeed())
		})

		It("should successfully finalize after failure of Delete() for MachineDeployment", func() {
			const csrName = "test-csr-mock-fail-delete-machinedeployment"
			var (
				errMsgs = []string{
					"failed to delete existing MachineDeployment",
				}
			)
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetDelete([]int{1})).To(Succeed()) // failed to delete MachineDeployment
			mClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseDCIMachineAnnotating, scalev1.PhaseMachineDeploymentCleanup, nil, mClient.GetMockClient(), errMsgs)
		})

		It("should successfully finalize after failure of Get() for DCIMachine", func() {
			const csrName = "test-csr-mock-fail-get-dci-machine"
			var (
				errMsgs = []string{
					"failed to get DCIMachine resource from management cluster",
				}
				reconcileEnd  = scalev1.PhaseDCIMachineAnnotating
				finalizeStart = scalev1.PhaseDCIMachineCleanup
				cleanupFunc   = executeCsrFinalizeWaitingScaleDown
			)
			// Setup
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetGet([]int{1})).To(Succeed()) // failed to get DCIMachine
			mClient.SetDefault()
			// Execute
			execCardScaleInitialized(csrName, csrNamespace, reconcileEnd, finalizeStart, nil, mClient.GetMockClient(), errMsgs, cleanupFunc)
		})

		It("should successfully finalize after failure of Patch() for DCIMachine", func() {
			const csrName = "test-csr-mock-fail-patch-dci-machine"
			var (
				errMsgs = []string{
					"failed to patch DCIMachine resource in management cluster",
				}
				reconcileEnd  = scalev1.PhaseDCIMachineAnnotating
				finalizeStart = scalev1.PhaseDCIMachineCleanup
				cleanupFunc   = executeCsrFinalizeWaitingScaleDown
			)
			// Setup
			mClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientManagement)
			Expect(mClient.SetPatch([]int{1})).To(Succeed()) // failed to patch DCIMachine
			mClient.SetDefault()
			// Execute
			execCardScaleInitialized(csrName, csrNamespace, reconcileEnd, finalizeStart, nil, mClient.GetMockClient(), errMsgs, cleanupFunc)
		})

		It("should successfully finalize after failure of Get() for scaling down NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-get-scaledown-nsr"
			var (
				errMsgs = []string{
					"failed to get NodeScalingRequest for scaling down",
				}
				reconcileEnd  = scalev1.PhaseDCIMachineAnnotating
				finalizeStart = scalev1.PhaseDCIMachineCleanup
				cleanupFunc   = executeCsrFinalizeWaitingScaleDown
			)
			// Setup
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get NodeScalingRequest
			wClient.SetDefault()
			// Execute
			execScale(csrName, csrNamespace, reconcileEnd, finalizeStart, wClient.GetMockClient(), nil, errMsgs, cleanupFunc)
		})

		It("should successfully finalize after failure of Get() for scaling up NodeScalingRequest", func() {
			const csrName = "test-csr-mock-fail-get-scaleup-nsr"
			var (
				errMsgs = []string{
					"failed to get NodeScalingRequest for scaling up",
				}
				reconcileEnd  = scalev1.PhaseNodeScaleDownDeleting
				finalizeStart = scalev1.PhaseMachineDeploymentCleanup
				cleanupFunc   = executeCsrFinalizeWaitingScaleUp
			)
			// Setup
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetGet([]int{2})).To(Succeed()) // failed to get NodeScalingRequest
			wClient.SetDefault()
			// Execute
			execScale(csrName, csrNamespace, reconcileEnd, finalizeStart, wClient.GetMockClient(), nil, errMsgs, cleanupFunc)
		})

		It("should successfully finalize after failure of Status().Patch() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-status-patch-csr"
			var (
				errMsgs = []string{
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to status.patch CardScalingRequest
			wClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitializing, scalev1.PhaseDCIMachineTemplateCreating, scalev1.PhaseDCIMachineTemplateCreating, wClient.GetMockClient(), nil, errMsgs)
		})

		It("should successfully finalize after failure of Patch() and Status().Patch() for CardScalingRequest", func() {
			const csrName = "test-csr-mock-fail-patch-status-patch-csr"
			var (
				errMsgs = []string{
					"failed to remove finalizer from CardScalingRequest",
					"failed to patch CardScalingRequest status",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetPatch([]int{1})).To(Succeed())       // failed to patch CardScalingRequest
			Expect(wClient.SetStatusPatch([]int{1})).To(Succeed()) // failed to status.patch CardScalingRequest
			wClient.SetDefault()
			exec(csrName, csrNamespace, scalev1.PhaseCardScaleInitial, scalev1.PhaseCardScaleInitializing, scalev1.PhaseCardScaleInitializing, wClient.GetMockClient(), nil, errMsgs)
		})
	})

	Context("When reconciling a resource", Label("ttl"), func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to delete CardScalingRequest when TTL is enabled and CardScalingRequest completed successfully", func() {
			const csrName = "test-csr-deleting-csr-after-ttl"
			r := newCsrReconciler(nil, nil)
			r.enableTTLSecondsAfterFinished = true
			r.ttlSecondsAfterFinished = 1 // 1 second
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			By("waiting for CardScalingRequest to be deleted after TTL", func() {
				time.Sleep(1 * time.Second)
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to pend deleting CardScalingRequest when TTL is enabled and before TTL is expired", func() {
			const csrName = "test-csr-deleting-csr-before-ttl-expiration"
			r := newCsrReconciler(nil, nil)
			r.enableTTLSecondsAfterFinished = true
			r.ttlSecondsAfterFinished = 2 // 2 seconds
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("waiting for CardScalingRequest to be deleted after TTL", func() {
				time.Sleep(2 * time.Second)
			})
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to skip deleting CardScalingRequest when TTL is disabled", func() {
			const csrName = "test-csr-to-skip-deleting-disabled-ttl"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to skip deleting CardScalingRequest when TTL is enabled and OwnerReference is set", func() {
			const csrName = "test-csr-to-skip-deleting-enabled-ttl-ownerref"
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				csr.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "OwnerKind",
						Name:       "owner-name",
						UID:        "owner-uid",
					},
				}

				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to skip deleting CardScalingRequest when TTL is enabled and CardScalingRequest for scaledown failed", func() {
			const csrName = "test-csr-to-skip-deleting-enabled-ttl-scaling-down-failed"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale down, transition to %s", scalev1.PhaseNodeScaleDownFailed),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)

				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)).To(Succeed())
			By("patching NodeScalingRequest status to PhaseNodeDeleteFailed", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeDeleteFailed
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownFailed)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownFailed)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownFailed)).To(Succeed())
		})

		It("should successfully reconcile to skip deleting CardScalingRequest when TTL is enabled and CardScalingRequest for scaleup failed", func() {
			const csrName = "test-csr-to-skip-deleting-enabled-ttl-scaling-up-failed"
			var (
				errMsgs = []string{
					fmt.Sprintf("failed to scale up, transition to %s", scalev1.PhaseNodeScaleUpFailed),
				}
			)
			r := newCsrReconciler(nil, nil)
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)

				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcileToPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)).To(Succeed())
			By("patching NodeScalingRequest status to PhaseNodeCreateFailed", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
				nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
				Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
			})
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleUpFailed)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleUpFailed)).To(Succeed())
			By("deleting CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseNodeScaleUpFailed)).To(Succeed())
		})
	})

	Context("When reconciling a resource", Label("ttl", "mock"), func() {
		const (
			csrNamespace                = defaultNamespace
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
			logicalServerNamespace      = defaultNamespace
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
			Expect(testutils.CleanupMachineResources(ctx, k8sClientManagement, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResources(ctx, k8sClientManagement, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResources(ctx, k8sClientManagement, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResources(ctx, k8sClientManagement, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResources(ctx, k8sClientManagement, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to delete CardScalingRequest when TTL is enabled and CardScalingRequest completed unsuccessfully once", func() {
			const csrName = "test-csr-deleting-csr-failure-once-after-ttl"
			var (
				errMsgs = []string{
					"failed to delete expired CardScalingRequest",
				}
			)
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			Expect(wClient.SetDelete([]int{2})).To(Succeed()) // failed to delete CardScalingRequest
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			r.enableTTLSecondsAfterFinished = true
			r.ttlSecondsAfterFinished = 0
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			err := executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)
			Expect(err).To(HaveOccurred())
			for _, errMsg := range errMsgs {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})

		It("should successfully reconcile to delete CardScalingRequest when TTL is enabled and CardScalingRequest already deleted", func() {
			const csrName = "test-csr-deleting-csr-already-deleted-after-ttl"
			wClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClientWorkload)
			defaultFunc := k8sClientWorkload.Delete
			alterFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
				return apierrors.NewNotFound(schema.GroupResource{Resource: "nodescalingrequests"}, csrName)
			}
			Expect(wClient.SetDeleteWithFunc(defaultFunc, alterFunc, []int{2})).To(Succeed()) // failed to delete CardScalingRequest
			wClient.SetDefault()
			r := newCsrReconciler(wClient.GetMockClient(), nil)
			r.enableTTLSecondsAfterFinished = true
			r.ttlSecondsAfterFinished = 0
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				Expect(k8sClientWorkload.Create(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrReconcile(r, csrName, csrNamespace)).To(Succeed())
			Expect(executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)).To(Succeed())
			By("deleting CardScalingRequest to pass Finalizer", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				Expect(k8sClientWorkload.Delete(ctx, csr)).To(Succeed())
			})
			Expect(executeCsrFinalize(r, csrName, csrNamespace)).To(Succeed())
		})
	})
})

// newCsrReconciler returns CardScalingRequestReconciler with mock by one line.
func newCsrReconciler(
	wClient, mClient client.Client,
) *CardScalingRequestReconciler {
	By("creating CardScalingRequestReconciler")
	controllerReconciler := (&CardScalingRequestReconciler{
		Recorder:                      events.NewFakeRecorder(1024),
		enableTTLSecondsAfterFinished: false,
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

// executeCsrReconciler executes Reconcile() and asserts CardScalingRequest status phase.
func executeCsrReconciler(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, nextPhase scalev1.CardScalingRequestPhase,
) error {
	By("executing Reconcile, next phase is " + string(nextPhase))
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: testutils.NewNamespacedName(csrName, csrNamespace),
	})
	if err != nil {
		return err
	}

	By("asserting CardScalingRequest status is " + string(nextPhase))
	csr := &scalev1.CardScalingRequest{}
	Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
	Expect(csr.Status.Phase).To(Equal(nextPhase))

	return nil
}

// executeCsrReconcileCardScaleInitial executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileCardScaleInitial(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitializing)
}

// executeCsrReconcileCardScaleInitializing executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileCardScaleInitializing(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCreating)
}

// executeCsrReconcileDCIMachineTemplateCreating executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileDCIMachineTemplateCreating(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCreating)
}

// executeCsrReconcileMachineDeploymentCreating executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileMachineDeploymentCreating(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineAnnotating)
}

// executeCsrReconcileDCIMachineAnnotating executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileDCIMachineAnnotating(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitialized)
}

// executeCsrReconcileCardScaleInitialized executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileCardScaleInitialized(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingDown)
}

// executeCsrReconcileNodeScalingDown executes Reconcile() for CardScalingRequest in each phase.
// It also patches NodeScalingRequest status to Deleted.
func executeCsrReconcileNodeScalingDown(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	By("patching NodeScalingRequest status to Deleted")
	nsr := &scalev1.NodeScalingRequest{}
	Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
	nsr.Status.Phase = scalev1.PhaseNodeDeleted
	Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())

	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownDeleting)
}

// executeCsrReconcileNodeScaleDownDeleting executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileNodeScaleDownDeleting(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScaleDownCompleted)
}

// executeCsrReconcileNodeScaleDownCompleted executes Reconcile() for CardScalingRequest in each phase.
func executeCsrReconcileNodeScaleDownCompleted(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseNodeScalingUp)
}

// executeCsrReconcileNodeScalingUp executes Reconcile() for CardScalingRequest in each phase.
// It also patches NodeScalingRequest status to Created.
func executeCsrReconcileNodeScalingUp(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	By("patching NodeScalingRequest status to Created")
	nsr := &scalev1.NodeScalingRequest{}
	Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)).To(Succeed())
	nsr.Status.Phase = scalev1.PhaseNodeCreated
	Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())

	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleCompleted)
}

// executeCsrFinalizeWaitingScaleDown executes Reconcile() for CardScalingRequest in each phase.
// It also patches NodeScalingRequest status to Deleted.
func executeCsrFinalizeWaitingScaleDown(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, expectedPhase scalev1.CardScalingRequestPhase,
) error {
	By("patching NodeScalingRequest status to Deleted")
	nsr := &scalev1.NodeScalingRequest{}
	err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)
	if err == nil {
		nsr.Status.Phase = scalev1.PhaseNodeDeleted
		Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
	} else if !apierrors.IsNotFound(err) {
		Expect(err).To(BeNil())
	}

	return executeCsrReconciler(r, csrName, csrNamespace, expectedPhase)
}

// executeCsrFinalizeWaitingScaleUp executes Reconcile() for CardScalingRequest in each phase.
// It also patches NodeScalingRequest status to Created.
func executeCsrFinalizeWaitingScaleUp(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, expectedPhase scalev1.CardScalingRequestPhase,
) error {
	By("patching NodeScalingRequest status to Created")
	nsr := &scalev1.NodeScalingRequest{}
	err := k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), nsr)
	if err == nil {
		nsr.Status.Phase = scalev1.PhaseNodeCreated
		Expect(k8sClientWorkload.Status().Update(ctx, nsr)).To(Succeed())
	} else if !apierrors.IsNotFound(err) {
		Expect(err).To(BeNil())
	}

	return executeCsrReconciler(r, csrName, csrNamespace, expectedPhase)
}

// executeCsrFinalizeDCIMachineCleanup executes Reconcile() for CardScalingRequest in each phase.
func executeCsrFinalizeDCIMachineCleanup(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseMachineDeploymentCleanup)
}

// executeCsrFinalizeMachineDeploymentCleanup executes Reconcile() for CardScalingRequest in each phase.
func executeCsrFinalizeMachineDeploymentCleanup(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseDCIMachineTemplateCleanup)
}

// executeCsrFinalizeDCIMachineTemplateCleanup executes Reconcile() for CardScalingRequest in each phase.
func executeCsrFinalizeDCIMachineTemplateCleanup(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconciler(r, csrName, csrNamespace, scalev1.PhaseCardScaleFinalized)
}

// executeCsrFinalizeCardScaleFinalized executes Reconcile() for CardScalingRequest in each phase.
func executeCsrFinalizeCardScaleFinalized(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	By("executing Reconcile, next phase is nothing as CardScalingRequest will be deleted")
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: testutils.NewNamespacedName(csrName, csrNamespace),
	})
	if err != nil {
		return err
	}

	By("asserting CardScalingRequest is deleted")
	csr := &scalev1.CardScalingRequest{}
	err = k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)
	Expect(apierrors.IsNotFound(err)).To(BeTrue())

	return nil
}

// executePhaseFunc defines function type for executing each phase.
type executePhaseFunc func(r *CardScalingRequestReconciler, csrName, csrNamespace string) error

// funcIndex is used to map phase to index for executing phases in order.
type funcIndex int

// Below const must be in order of functions to call.
const (
	phaseIndexCardScalePending funcIndex = iota
	phaseIndexCardScaleInitial
	phaseIndexCardScaleInitializing
	phaseIndexDCIMachineTemplateCreating
	phaseIndexMachineDeploymentCreating
	phaseIndexDCIMachineAnnotating
	phaseIndexCardScaleInitialized
	phaseIndexNodeScalingDown
	phaseIndexNodeScaleDownDeleting
	phaseIndexNodeScaleDownCompleted
	phaseIndexNodeScalingUp
	phaseIndexReconcileSentinel // sentinel

	phaseIndexFinalizeWaitingScaleDown
	phaseIndexFinalizeWaitingScaleUp
	phaseIndexFinalizeDCIMachineCleanup
	phaseIndexFinalizeMachineDeploymentCleanup
	phaseIndexFinalizeDCIMachineTemplateCleanup
	phaseIndexFinalizeCardScaleFinalized
	phaseIndexFinalizeSentinel // sentinel
)

// mapFromPhaseToIndexForReconcile maps CardScalingRequestPhase to funcIndex for Reconcile().
var mapFromPhaseToIndexForReconcile = map[scalev1.CardScalingRequestPhase]funcIndex{
	scalev1.PhaseCardScalePending:           phaseIndexCardScalePending,
	scalev1.PhaseCardScaleInitial:           phaseIndexCardScaleInitial,
	scalev1.PhaseCardScaleInitializing:      phaseIndexCardScaleInitializing,
	scalev1.PhaseDCIMachineTemplateCreating: phaseIndexDCIMachineTemplateCreating,
	scalev1.PhaseMachineDeploymentCreating:  phaseIndexMachineDeploymentCreating,
	scalev1.PhaseDCIMachineAnnotating:       phaseIndexDCIMachineAnnotating,
	scalev1.PhaseCardScaleInitialized:       phaseIndexCardScaleInitialized,
	scalev1.PhaseNodeScalingDown:            phaseIndexNodeScalingDown,
	scalev1.PhaseNodeScaleDownDeleting:      phaseIndexNodeScaleDownDeleting,
	scalev1.PhaseNodeScaleDownCompleted:     phaseIndexNodeScaleDownCompleted,
	scalev1.PhaseNodeScalingUp:              phaseIndexNodeScalingUp,
	scalev1.PhaseCardScaleCompleted:         phaseIndexReconcileSentinel, // sentinel
}

// mapFromPhaseToIndexForFinalize maps CardScalingRequestPhase to funcIndex for Finalize().
var mapFromPhaseToIndexForFinalize = map[scalev1.CardScalingRequestPhase]funcIndex{
	scalev1.PhaseNodeScaleDownDeleting: phaseIndexFinalizeWaitingScaleDown,
	scalev1.PhaseNodeScalingDown:       phaseIndexFinalizeWaitingScaleDown,
	scalev1.PhaseCardScaleInitialized:  phaseIndexFinalizeWaitingScaleDown,

	scalev1.PhaseNodeScalingUp:          phaseIndexFinalizeWaitingScaleUp,
	scalev1.PhaseNodeScaleDownCompleted: phaseIndexFinalizeWaitingScaleUp,

	scalev1.PhaseDCIMachineCleanup:    phaseIndexFinalizeDCIMachineCleanup,
	scalev1.PhaseDCIMachineAnnotating: phaseIndexFinalizeDCIMachineCleanup,
	scalev1.PhaseNodeScaleDownFailed:  phaseIndexFinalizeDCIMachineCleanup,

	scalev1.PhaseMachineDeploymentCleanup:  phaseIndexFinalizeMachineDeploymentCleanup,
	scalev1.PhaseMachineDeploymentCreating: phaseIndexFinalizeMachineDeploymentCleanup,
	scalev1.PhaseNodeScaleUpFailed:         phaseIndexFinalizeMachineDeploymentCleanup,

	scalev1.PhaseDCIMachineTemplateCleanup:  phaseIndexFinalizeDCIMachineTemplateCleanup,
	scalev1.PhaseDCIMachineTemplateCreating: phaseIndexFinalizeDCIMachineTemplateCleanup,

	scalev1.PhaseCardScaleFinalized:    phaseIndexFinalizeCardScaleFinalized,
	scalev1.PhaseCardScaleInitial:      phaseIndexFinalizeCardScaleFinalized,
	scalev1.PhaseCardScalePending:      phaseIndexFinalizeCardScaleFinalized,
	scalev1.PhaseCardScaleInitializing: phaseIndexFinalizeCardScaleFinalized,
	scalev1.PhaseCardScaleCompleted:    phaseIndexFinalizeCardScaleFinalized,
}

// mapFromIndexToFunc maps funcIndex to executePhaseFunc.
var mapFromIndexToFunc = map[funcIndex]executePhaseFunc{
	phaseIndexCardScaleInitial:           executeCsrReconcileCardScaleInitial,
	phaseIndexCardScaleInitializing:      executeCsrReconcileCardScaleInitializing,
	phaseIndexDCIMachineTemplateCreating: executeCsrReconcileDCIMachineTemplateCreating,
	phaseIndexMachineDeploymentCreating:  executeCsrReconcileMachineDeploymentCreating,
	phaseIndexDCIMachineAnnotating:       executeCsrReconcileDCIMachineAnnotating,
	phaseIndexCardScaleInitialized:       executeCsrReconcileCardScaleInitialized,
	phaseIndexNodeScalingDown:            executeCsrReconcileNodeScalingDown,
	phaseIndexNodeScaleDownDeleting:      executeCsrReconcileNodeScaleDownDeleting,
	phaseIndexNodeScaleDownCompleted:     executeCsrReconcileNodeScaleDownCompleted,
	phaseIndexNodeScalingUp:              executeCsrReconcileNodeScalingUp,
	phaseIndexReconcileSentinel:          nil, // sentinel
	// Note: CardScalePending are handled specially.

	phaseIndexFinalizeDCIMachineCleanup:         executeCsrFinalizeDCIMachineCleanup,
	phaseIndexFinalizeMachineDeploymentCleanup:  executeCsrFinalizeMachineDeploymentCleanup,
	phaseIndexFinalizeDCIMachineTemplateCleanup: executeCsrFinalizeDCIMachineTemplateCleanup,
	phaseIndexFinalizeCardScaleFinalized:        executeCsrFinalizeCardScaleFinalized,
	phaseIndexFinalizeSentinel:                  nil, // sentinel
	// Note: WaitingScaleDown and WaitingScaleUp are handled specially.
}

// executeCsrReconcilerWithPhases executes Reconcile() for CardScalingRequest from start phase to end phase.
func executeCsrReconcilerWithPhases(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, start, end scalev1.CardScalingRequestPhase,
	mapFromPhaseToIndex map[scalev1.CardScalingRequestPhase]funcIndex,
) error {
	startIdx, exists := mapFromPhaseToIndex[start]
	if !exists {
		return fmt.Errorf("no execute function for phase: %s", start)
	}
	endIdx, exists := mapFromPhaseToIndex[end]
	if !exists {
		return fmt.Errorf("no execute function for phase: %s", end)
	}
	if startIdx > endIdx {
		return fmt.Errorf("given phases are invalid: %s -> %s", start, end)
	}
	for i := startIdx; i <= endIdx; i++ {
		execFunc, exists := mapFromIndexToFunc[i]
		if !exists || execFunc == nil {
			continue
		}
		if err := execFunc(r, csrName, csrNamespace); err != nil {
			return err
		}
	}
	return nil
}

// executeCsrReconcileFromPhase executes Reconcile() for CardScalingRequest from start phase to CardScaleCompleted phase.
func executeCsrReconcileFromPhase(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, start scalev1.CardScalingRequestPhase,
) error {
	return executeCsrReconcilerWithPhases(r, csrName, csrNamespace, start, scalev1.PhaseCardScaleCompleted,
		mapFromPhaseToIndexForReconcile)
}

// executeCsrReconcileToPhase executes Reconcile() for CardScalingRequest from CardScaleInitial phase to end phase.
func executeCsrReconcileToPhase(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, end scalev1.CardScalingRequestPhase,
) error {
	return executeCsrReconcilerWithPhases(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial, end,
		mapFromPhaseToIndexForReconcile)
}

// executeCsrReconcile executes Reconcile() for CardScalingRequest from CardScaleInitial phase to CardScaleCompleted phase.
func executeCsrReconcile(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrReconcileFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleInitial)
}

// executeCsrFinalizeFromPhase executes Reconcile() for CardScalingRequest from start phase to CardScaleFinalized phase.
func executeCsrFinalizeFromPhase(
	r *CardScalingRequestReconciler, csrName, csrNamespace string, start scalev1.CardScalingRequestPhase,
) error {
	return executeCsrReconcilerWithPhases(r, csrName, csrNamespace, start, scalev1.PhaseCardScaleFinalized,
		mapFromPhaseToIndexForFinalize)
}

// executeCsrFinalize executes Reconcile() for CardScalingRequest CardScaleFinalized phase.
func executeCsrFinalize(
	r *CardScalingRequestReconciler, csrName, csrNamespace string,
) error {
	return executeCsrFinalizeFromPhase(r, csrName, csrNamespace, scalev1.PhaseCardScaleFinalized)
}
