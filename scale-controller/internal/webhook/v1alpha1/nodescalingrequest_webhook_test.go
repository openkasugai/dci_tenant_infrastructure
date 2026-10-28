package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/test/testutils"
)

var _ = Describe("NodeScalingRequest Webhook Methods directly calling", Label("nsr", "webhook", "methods"), func() {
	const (
		nsrNamespace = defaultNamespace
	)

	Context("When calling methods directly", func() {
		It("Should deny creation if object is not NodeScalingRequest", func() {
			var (
				validator *NodeScalingRequestCustomValidator
			)
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("executing ValidateCreate", func() {
				_, err := validator.ValidateCreate(ctx, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny creation if object is NodeScalingRequest but nil", func() {
			var (
				validator *NodeScalingRequestCustomValidator
				nsr       *scalev1.NodeScalingRequest
			)
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("executing ValidateCreate", func() {
				_, err := validator.ValidateCreate(ctx, nsr)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny creation if Spec.Node is nil", func() {
			const nsrName = "test-nsr-delete-node-nil"
			var (
				validator *NodeScalingRequestCustomValidator
				nsr       *scalev1.NodeScalingRequest
			)
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("preparing NodeScalingRequest", func() {
				nsr = testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
			})
			By("executing ValidateCreate", func() {
				_, err := validator.ValidateCreate(ctx, nsr)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny updates if object is not NodeScalingRequest", func() {
			var validator *NodeScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("executing ValidateUpdate", func() {
				_, err := validator.ValidateUpdate(ctx, nil, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should deny deletion if object is not NodeScalingRequest", func() {
			var validator *NodeScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("executing ValidateDelete", func() {
				_, err := validator.ValidateDelete(ctx, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		It("Should admit deletion if object is NodeScalingRequest", func() {
			var validator *NodeScalingRequestCustomValidator
			By("setting up validator", func() {
				validator = &NodeScalingRequestCustomValidator{}
			})
			By("executing ValidateDelete", func() {
				_, err := validator.ValidateDelete(ctx, &scalev1.NodeScalingRequest{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

var _ = Describe("NodeScalingRequest Webhook", Label("nsr", "webhook"), func() {
	const (
		nsrNamespace                = defaultNamespace
		nodeName                    = "test-node"
		nodeNotExistingName         = "test-node-no-existing"
		clusterName                 = "test-cluster"
		clusterNamespace            = defaultNamespace
		machineName                 = "test-machine"
		machineNamespace            = defaultNamespace
		mdName                      = "test-machinedeployment"
		mdNamespace                 = defaultNamespace
		dciMachineTemplateName      = "test-dci-machine-template"
		dciMachineTemplateNamespace = defaultNamespace
	)

	Context("When creating a NodeScalingRequest", func() {
		BeforeEach(func() {
			Expect(testutils.CreateNodeResource(ctx, k8sClientWorkload, nodeName, clusterName, clusterNamespace, machineName)).To(Succeed())
			Expect(testutils.CreateMachineResource(ctx, k8sClientManagement, machineName, machineNamespace, mdName, clusterName)).To(Succeed())
			Expect(testutils.CreateMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace, clusterName, capiv1.GroupVersion.Group, templateKind, dciMachineTemplateName, nil)).To(Succeed())
		})

		AfterEach(func() {
			Expect(testutils.CleanupNodeResource(ctx, k8sClientWorkload, nodeName)).To(Succeed())
			Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())

			Expect(testutils.CleanupNodeScalingRequestResources(ctx, k8sClientWorkload, nsrNamespace)).To(Succeed())
		})

		It("Should admit creation if NodeRequestType is TypeCreate", func() {
			const nsrName = "test-nsr-create-create"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeCreate)
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
		})

		It("Should admit creation if NodeRequestType is TypeDelete", func() {
			const nsrName = "test-nsr-create-delete"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(nodeName)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
		})

		It("Should deny creation if the configured Node could not be found.", func() {
			const nsrName = "test-nsr-create-delete-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get Node(%s)", nodeNotExistingName),
			}
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(string(nodeNotExistingName))
				err := k8sClientWorkload.Create(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the configured Node has no annotations for cluster info.", func() {
			const nsrName = "test-nsr-create-delete-no-cluster-info"
			var errMsgs = []string{
				fmt.Sprintf("failed to get Cluster namespace from Node(%s)", nodeName),
			}
			By("removing cluster info annotations from Node", func() {
				node := &corev1.Node{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nodeName, ""), node)).To(Succeed())
				node.Annotations = map[string]string{}
				Expect(k8sClientWorkload.Update(ctx, node)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(string(nodeName))
				err := k8sClientWorkload.Create(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the Machine associated with configured Node could not be found.", func() {
			const nsrName = "test-nsr-create-delete-machine-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get Machine from Node(%s)", nodeName),
			}
			By("deleting Machine associated with Node", func() {
				Expect(testutils.CleanupMachineResource(ctx, k8sClientManagement, machineName, machineNamespace)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(string(nodeName))
				err := k8sClientWorkload.Create(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the MachineDeployment associated with configured Node could not be found.", func() {
			const nsrName = "test-nsr-create-delete-machinedeployment-not-found"
			var errMsgs = []string{
				fmt.Sprintf("failed to get MachineDeployment from Machine(%s/%s)", machineNamespace, machineName),
			}
			By("deleting MachineDeployment associated with Node", func() {
				Expect(testutils.CleanupMachineDeploymentResource(ctx, k8sClientManagement, mdName, mdNamespace)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(string(nodeName))
				err := k8sClientWorkload.Create(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})

		It("Should deny creation if the MachineDeployment associated with configured Node has auto-scaling enabled.", func() {
			const nsrName = "test-nsr-create-delete-machinedeployment-auto-scaling-enabled"
			var errMsgs = []string{
				fmt.Sprintf("expected the MachineDeployment(%s/%s) associated with the Node(%s) has auto-scaling disabled, but has auto-scaling enabled", mdNamespace, mdName, nodeName),
			}
			By("patching MachineDeployment associated with Node to enable auto-scaling", func() {
				md := &capiv1.MachineDeployment{}
				Expect(k8sClientManagement.Get(ctx, testutils.NewNamespacedName(mdName, mdNamespace), md)).To(Succeed())
				md.Annotations = map[string]string{
					capiv1.AutoscalerMaxSizeAnnotation: "1",
					capiv1.AutoscalerMinSizeAnnotation: "1",
				}
				Expect(k8sClientManagement.Update(ctx, md)).To(Succeed())
			})
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeDelete)
				nsr.Spec.Node = ptr.To(string(nodeName))
				err := k8sClientWorkload.Create(ctx, nsr)
				Expect(err).To(HaveOccurred())
				for _, msg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(msg))
				}
			})
		})
	})

	Context("When updating a NodeScalingRequest", func() {
		It("Should admit update if NodeScalingRequest is updated", func() {
			const nsrName = "test-nsr-update"
			By("creating NodeScalingRequest", func() {
				nsr := testutils.NewNodeScalingRequest(nsrName, nsrNamespace, scalev1.TypeCreate)
				nsr.Spec.ServerSpec = testutils.NewServerSpecTemplate(rpmv1.ReqServerTypeComposed)
				Expect(k8sClientWorkload.Create(ctx, nsr)).To(Succeed())
			})
			By("updating NodeScalingRequest", func() {
				nsr := &scalev1.NodeScalingRequest{}
				Expect(k8sClientWorkload.Get(ctx, testutils.NewNamespacedName(nsrName, nsrNamespace), nsr)).To(Succeed())
				nsr.Annotations = map[string]string{"updated": "true"}
				Expect(k8sClientWorkload.Update(ctx, nsr)).To(Succeed())
			})
		})
	})
})
