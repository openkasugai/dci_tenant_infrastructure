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

package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/test/testutils"
)

var _ = Describe("CardScalingRequest Webhook Methods directly calling", Label("csr", "webhook", "methods"), func() {
	Context("When calling methods directly", func() {
		It("Should deny creation if object is not CardScalingRequest", func() {
			var (
				validator *CardScalingRequestCustomValidator
			)
			By("setting up validator", func() {
				validator = &CardScalingRequestCustomValidator{}
			})
			By("executing ValidateCreate", func() {
				_, err := validator.ValidateCreate(ctx, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny creation if object is CardScalingRequest but nil", func() {
			var (
				validator *CardScalingRequestCustomValidator
				csr       *scalev1.CardScalingRequest
			)
			By("setting up validator", func() {
				validator = &CardScalingRequestCustomValidator{}
			})
			By("executing ValidateCreate", func() {
				_, err := validator.ValidateCreate(ctx, csr)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny updates if object is not CardScalingRequest", func() {
			var validator *CardScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &CardScalingRequestCustomValidator{}
			})
			By("executing ValidateUpdate", func() {
				_, err := validator.ValidateUpdate(ctx, nil, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should admit deletion", func() {
			var validator *CardScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &CardScalingRequestCustomValidator{}
			})
			By("executing ValidateDelete", func() {
				_, err := validator.ValidateDelete(ctx, (*scalev1.CardScalingRequest)(nil))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("Should deny deletion if object is not CardScalingRequest", func() {
			var validator *CardScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &CardScalingRequestCustomValidator{}
			})
			By("executing ValidateDelete", func() {
				_, err := validator.ValidateDelete(ctx, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("CardScalingRequest Webhook", Label("csr", "webhook"), func() {
	Context("When creating a CardScalingRequest", func() {
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
			location     = ptr.To("test-location")
			model        = "test-model"
			modelInvalid = "invalid-model"
			uuids        = []string{"gpu-uuid-1", "gpu-uuid-2"}
			assignments  = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
				{UUID: uuids[1]},
			}
			serverSpec = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(2),
					Assignments: assignments,
				},
				&rpmv1.NICSpec{
					Model: "dummy-nic-model",
					Num:   int32(1),
				},
			)
			serverSpecWithoutGPU = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil, nil, nil)
			serverSpecWithoutGPUAssignments = testutils.NewServerSpec(
				rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil,
				&rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(2),
					Assignments: nil,
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
			Expect(testutils.CreateDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace, location, serverSpec)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
			Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineTemplateResource(ctx, k8sClientManagement, dciMachineTemplateName, dciMachineTemplateNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("Should admit creation for add Card", func() {
			const csrName = "test-csr-create-add"
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
		})

		It("Should admit creation for add Card from serverSpec without GPUs", func() {
			const csrName = "test-csr-create-add-from-zero-gpu"
			By("regenerating LogicalServer resource to have nil GPU spec", func() {
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
				Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace,
					location, serverSpecWithoutGPU, rpmv1.RoleTypeWorker)).To(Succeed())
				Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
				Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
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
		})

		It("Should admit creation for delete Card", func() {
			const csrName = "test-csr-create-delete"
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
		})

		It("Should admit creation for delete Card to remove all GPUs", func() {
			const csrName = "test-csr-create-delete-all"
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
		})

		It("Should admit creation for replace Card", func() {
			const csrName = "test-csr-create-replace"
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
		})

		It("Should deny creation if the configured Node could not be found.", func() {
			const csrName = "test-csr-create-node-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get Node(%s)", nodeName),
			}
			By("deleting Node resource", func() {
				Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured Node Annotations could not be found.", func() {
			const csrName = "test-csr-create-node-annotation-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get cluster info from node(%s)", nodeName),
			}
			By("updating Node resource to remove annotations", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, client.ObjectKey{Name: nodeName}, node)).To(Succeed())
				node.Annotations = map[string]string{}
				Expect(k8sClientWorkload.Update(ctx, node)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured Machine could not be found.", func() {
			const csrName = "test-csr-create-machine-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get Machine from node(%s)", nodeName),
			}
			By("deleting Machine resource", func() {
				Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured MachineDeployment could not be found.", func() {
			const csrName = "test-csr-create-machine-deployment-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get MachineDeployment from Machine(%s/%s)", machineNamespace, machineName),
			}
			By("deleting MachineDeployment resource", func() {
				Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured MachineDeployment has invalid Annotations.", func() {
			const csrName = "test-csr-create-machine-deployment-invalid-annotations"
			var errMsgs = []string{
				fmt.Sprintf("expected MachineDeployment(%s/%s) has no annotations for ClusterAutoscaler, but it has them",
					mdNamespace, mdName),
			}
			By("updating MachineDeployment resource to have invalid Annotations", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, client.ObjectKey{Name: mdName, Namespace: mdNamespace}, md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMaxSizeAnnotation: "1",
					capiv1.AutoscalerMinSizeAnnotation: "1",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured DCIMachine could not be found.", func() {
			const csrName = "test-csr-create-dcimachine-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get DCIMachine from Machine(%s/%s)", machineNamespace, machineName),
			}
			By("deleting DCIMachine resource", func() {
				Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured LogicalServer could not be found.", func() {
			const csrName = "test-csr-create-logicalserver-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get LogicalServer from DCIMachine(%s/%s)", dciMachineNamespace, dciMachineName),
			}
			By("deleting LogicalServer resource", func() {
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured LogicalServer is unexpected type.", func() {
			const csrName = "test-csr-create-logicalserver-unexpected-type"
			var errMsgs = []string{
				fmt.Sprintf("expected applied spec type of LogicalServer(%s/%s) to be '%s', but got spec ",
					logicalServerNamespace, logicalServerName, rpmv1.AppliedServerTypeComposed),
			}
			By("Patching LogicalServer resource to have unexpected type", func() {
				Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeCOTS)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured LogicalServer has invalid state.", func() {
			const csrName = "test-csr-create-logicalserver-invalid-state"
			var errMsgs = []string{
				fmt.Sprintf("expected LogicalServer(%s/%s) status.state to be '%s', but got '%s'",
					logicalServerNamespace, logicalServerName, rpmv1.LogicalServerStateProvisioned, rpmv1.LogicalServerStateInitial),
			}
			By("updating LogicalServer resource to have invalid state", func() {
				Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateInitial)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					model,
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should admit creation if the configured LogicalServer has nil GPU.", func() {
			const csrName = "test-csr-create-logicalserver-without-gpu"
			By("regenerating LogicalServer resource to have nil GPU spec", func() {
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
				Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace,
					location, serverSpecWithoutGPU, rpmv1.RoleTypeWorker)).To(Succeed())
				Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
				Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
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
		})

		It("Should deny creation if the request is invalid(not found model).", func() {
			const csrName = "test-csr-create-invalid-request-not-found-model"
			var errMsgs = []string{
				"failed to generate expected serverSpec after scaling",
			}
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpAdd,
					modelInvalid, // invalid model
					1,
					nil,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured LogicalServer has nil GPU Assignments.", func() {
			const csrName = "test-csr-create-invalid-request-num-negative"
			var errMsgs = []string{
				fmt.Sprintf("expected all UUIDs in CardScalingRequest to exist in LogicalServer(%s/%s), but some do not. GPUSpec: %v",
					logicalServerNamespace, logicalServerName, serverSpecWithoutGPUAssignments.GPU),
			}
			By("regenerating LogicalServer resource to have nil GPU assignments", func() {
				Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())
				Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace,
					location, serverSpecWithoutGPUAssignments, rpmv1.RoleTypeWorker)).To(Succeed())
				Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
				Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
			})
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					uuids,
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the request is invalid(not found uuid in delete).", func() {
			const csrName = "test-csr-create-invalid-request-not-found-uuid"
			var errMsgs = []string{
				"failed to validate GPU UUIDs in scaling parameters",
				fmt.Sprintf("expected all UUIDs in CardScalingRequest to exist in LogicalServer(%s/%s), but some do not",
					logicalServerNamespace, logicalServerName),
			}
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpDelete,
					model,
					0,
					[]string{"not-found-uuid"}, // invalid uuid
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the request is invalid(not found uuid in replace).", func() {
			const csrName = "test-csr-create-invalid-request-not-found-uuid"
			var errMsgs = []string{
				"failed to validate GPU UUIDs in scaling parameters",
				fmt.Sprintf("expected all UUIDs in CardScalingRequest to exist in LogicalServer(%s/%s), but some do not",
					logicalServerNamespace, logicalServerName),
			}
			By("creating CardScalingRequest", func() {
				csr := testutils.NewCardScalingRequest(csrName, csrNamespace, nodeName)
				csr.Spec.ScalingParameter = *testutils.NewScalingParameter(
					capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					model,
					0,
					[]string{"not-found-uuid"}, // invalid uuid
				)
				err := k8sClientWorkload.Create(ctx, csr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})
	})

	Context("When updating a CardScalingRequest", func() {
		const (
			csrNamespace           = defaultNamespace
			nodeName               = "test-node"
			clusterName            = "test-cluster"
			clusterNamespace       = defaultNamespace
			machineName            = "test-machine"
			machineNamespace       = defaultNamespace
			mdName                 = "test-machinedeployment"
			mdNamespace            = defaultNamespace
			dciMachineName         = "test-dci-machine"
			dciMachineNamespace    = defaultNamespace
			logicalServerName      = "test-logical-server"
			logicalServerNamespace = defaultNamespace
		)
		var (
			location    = ptr.To("test-location")
			model       = "test-model"
			uuids       = []string{"gpu-uuid-1", "gpu-uuid-2"}
			assignments = []rpmv1.GPUAssignment{
				{UUID: uuids[0]},
				{UUID: uuids[1]},
			}
			serverSpec = testutils.NewServerSpec(rpmv1.ReqServerTypeComposed,
				nil, nil, nil, nil, &rpmv1.GPUSpec{
					Model:       model,
					Num:         int32(1),
					Assignments: assignments,
				}, nil,
			)
		)

		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineName, nil)).To(Succeed())
			Expect(testutils.CreateDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, location, "", serverSpec)).To(Succeed())
			Expect(testutils.CreateLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, location, serverSpec, rpmv1.RoleTypeWorker)).To(Succeed())

			Expect(testutils.PatchDCIMachineLogicalServerRef(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace, logicalServerName, logicalServerNamespace)).To(Succeed())
			Expect(testutils.PatchMachineInfrastructureRef(ctx, k8sClientManagement, machineName, machineNamespace, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.PatchLogicalServerState(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.LogicalServerStateProvisioned)).To(Succeed())
			Expect(testutils.PatchLogicalServerAppliedType(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace, rpmv1.AppliedServerTypeComposed)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			Expect(testutils.CleanupDCIMachineResource(ctx, k8sClientManagement, dciMachineName, dciMachineNamespace)).To(Succeed())
			Expect(testutils.CleanupLogicalServerResource(ctx, k8sClientManagement, logicalServerName, logicalServerNamespace)).To(Succeed())

			Expect(testutils.CleanupCardScalingRequestResources(ctx, k8sClientWorkload, csrNamespace)).To(Succeed())
		})

		It("Should admit update if CardScalingRequest is updated", func() {
			const csrName = "test-csr-update"
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
			By("updating CardScalingRequest", func() {
				csr := &scalev1.CardScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(csrName, csrNamespace), csr)).To(Succeed())
				csr.Annotations = map[string]string{"updated": "true"}
				Expect(k8sClientWorkload.Update(ctx, csr)).To(Succeed())
			})
		})
	})
})
