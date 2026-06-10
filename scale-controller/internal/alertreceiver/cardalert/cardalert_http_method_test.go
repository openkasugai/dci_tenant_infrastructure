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

package cardalert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/test/testutils"
)

// Test for Reconcile() method
var _ = Describe("Card AlertReceiver HTTP method", Label("ca", "receiver", "http"), func() {
	Context("When creating Card AlertReceiver HTTP method", func() {
		const caNamespace = defaultNamespace
		var (
			storedInClusterConfig    func() (*rest.Config, error)
			storedNewClient          func(*rest.Config, client.Options) (client.Client, error)
			storedScalev1AddToScheme func(scheme *runtime.Scheme) error
			storedCorev1AddToScheme  func(scheme *runtime.Scheme) error
		)

		BeforeEach(func() {
			storedInClusterConfig = InClusterConfig
			storedNewClient = NewClient
			storedScalev1AddToScheme = scalev1.AddToScheme
			storedCorev1AddToScheme = corev1.AddToScheme
		})

		AfterEach(func() {
			// Restore the original functions after each test
			InClusterConfig = storedInClusterConfig
			NewClient = storedNewClient
			scalev1.AddToScheme = storedScalev1AddToScheme
			corev1.AddToScheme = storedCorev1AddToScheme
		})

		It("should successfully create Card AlertReceiver HTTP method", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				c = k8sClient
			)
			By("mocking functions", func() {
				InClusterConfig = testutils.NewMockInClusterConfig(&rest.Config{}, nil)
				NewClient = testutils.NewMockNewClient(c, nil)
			})
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).NotTo(HaveOccurred())
				Expect(httpMethod).NotTo(BeNil())
				Expect(httpMethod.logger).To(Equal(logger))
				Expect(httpMethod.client).To(Equal(c))
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method with nil options", func() {
			var (
				options *CardAlertReceiverNewOptions = nil
				errMsgs                              = []string{
					"expected options or Logger in options is not nil, but got nil",
				}
			)
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method with empty CardAlert Namespace", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: "",
				}
				errMsgs = []string{
					"expected CardAlertNamespace in options to be not empty, but got empty",
				}
			)
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method if scalev1.AddToScheme failed", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				errMsgs = []string{
					"failed to add scale-controller scheme",
				}
			)
			By("mocking functions", func() {
				scalev1.AddToScheme = func(_ *runtime.Scheme) error {
					return errors.New("mocked error for scalev1.AddToScheme()")
				}
			})
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method if corev1.AddToScheme failed", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				errMsgs = []string{
					"failed to add corev1 scheme",
				}
			)
			By("mocking functions", func() {
				corev1.AddToScheme = func(_ *runtime.Scheme) error {
					return errors.New("mocked error for corev1.AddToScheme()")
				}
			})
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method if InClusterConfig failed", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				errMsgs = []string{
					"failed to get in-cluster config",
				}
			)
			By("mocking functions", func() {
				InClusterConfig = testutils.NewMockInClusterConfig(nil, errors.New("mocked error for InClusterConfig()"))
			})
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver HTTP method if NewClient failed", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				errMsgs = []string{
					"failed to create client for In-Cluster",
				}
			)
			By("mocking functions", func() {
				InClusterConfig = testutils.NewMockInClusterConfig(&rest.Config{}, nil)
				NewClient = testutils.NewMockNewClient(nil, errors.New("mocked error for NewClient()"))
			})
			By("calling newCardAlertHandler", func() {
				httpMethod, err := newCardAlertHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(httpMethod).To(BeNil())
			})
		})
	})

	Context("When receiving an alert of GET", Label("GET"), func() {
		It("should successfully handle GET requests with Status OK", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					logger: logger,
					client: k8sClient,
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
				resBody = "\"OK\"\n"
			)
			By("executing GET method", func() {
				req, err := http.NewRequest("GET", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Get(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("asserting the response body", func() {
				Expect(writer.Body.String()).To(Equal(resBody))
			})
		})

		It("should unsuccessfully handle GET requests if Encode failed", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					logger: logger,
					client: k8sClient,
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("mocking functions", func() {
				writer.WriteError = errors.New("mocked error for writer.Write()")
			})
			By("executing GET method", func() {
				req, err := http.NewRequest("GET", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Get(writer, req)
			})
			// The result of Encode failure is logged as an error,
			// but the response status code is still 200 OK
			// because the error happens after writing the status code.
		})
	})

	Context("When receiving an alert of POST", Label("POST"), func() {
		const (
			caNamespace = defaultNamespace
		)

		AfterEach(func() {
			Expect(testutils.CleanupCardAlertResources(ctx, k8sClient, caNamespace)).To(Succeed())
		})

		It("should successfully handle POST requests with creating CardAlert resource", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
					},
				}
			)
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("checking if CardAlert resource is created with correct spec", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(1))
				ca := caList.Items[0]
				Expect(ca.Spec.NodeName).To(Equal("test-node"))
				Expect(ca.Spec.Cards[0].Model).To(Equal("test-gpu-model"))
				Expect(ca.Spec.Cards[0].GPUUUID).To(Equal("test-gpu-uuid"))
				Expect(ca.Spec.Cards[0].MIGUUIDs).To(HaveLen(1))
				Expect(ca.Spec.Cards[0].MIGUUIDs[0]).To(Equal("test-mig-uuid"))
			})
		})

		It("should successfully handle POST requests without creating CardAlert resource if Status is resolved", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusResolved,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
					},
				}
			)
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("checking if CardAlert resource is NOT created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(0))
			})
		})

		It("should successfully handle POST requests without creating CardAlert resource if AlertName is not expected one", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid",
								common.AlertManagerAlertLabelAlertNameKey: "unexpected-alert-name",
							},
						},
					},
				}
			)
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("checking if CardAlert resource is NOT created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(0))
			})
		})

		It("should successfully handle POST requests without creating CardAlert resource if alert has no required labels", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
					},
				}
			)
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("checking if CardAlert resource is NOT created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(0))
			})
		})

		It("should successfully handle POST requests with creating only valid CardAlert resource if unexpected alerts exist", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-0",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-0",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-0",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid-0",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-1",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-1",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-1",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid-1",
								common.AlertManagerAlertLabelAlertNameKey: "unexpected-alert-name",
							},
						},
						{
							Status: common.AlertManagerAlertStatusResolved,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-2",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-2",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-2",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid-2",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-3",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-3",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-3",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid-3",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-3",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-3",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-3",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid-4",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-5",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-5",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-5",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node-5",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model-5",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid-6",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
					},
				}
			)
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("checking if CardAlert resources are created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(3))
				var ca *scalev1.CardAlert
				for idx := range caList.Items {
					ca = &caList.Items[idx]
					switch ca.Spec.NodeName {
					case "test-node-0":
						Expect(ca.Spec.Cards).To(HaveLen(1))
						Expect(ca.Spec.Cards[0].Model).To(Equal("test-gpu-model-0"))
						Expect(ca.Spec.Cards[0].GPUUUID).To(Equal("test-gpu-uuid-0"))
						Expect(ca.Spec.Cards[0].MIGUUIDs).To(HaveLen(1))
						Expect(ca.Spec.Cards[0].MIGUUIDs[0]).To(Equal("test-mig-uuid-0"))
					case "test-node-3":
						Expect(ca.Spec.Cards).To(HaveLen(1))
						Expect(ca.Spec.Cards[0].Model).To(Equal("test-gpu-model-3"))
						Expect(ca.Spec.Cards[0].GPUUUID).To(Equal("test-gpu-uuid-3"))
						Expect(ca.Spec.Cards[0].MIGUUIDs).To(HaveLen(2))
						Expect(ca.Spec.Cards[0].MIGUUIDs[0]).To(Equal("test-mig-uuid-3"))
						Expect(ca.Spec.Cards[0].MIGUUIDs[1]).To(Equal("test-mig-uuid-4"))
					case "test-node-5":
						Expect(ca.Spec.Cards).To(HaveLen(2))
						Expect(ca.Spec.Cards[0].Model).To(Equal("test-gpu-model-5"))
						Expect(ca.Spec.Cards[0].GPUUUID).To(Equal("test-gpu-uuid-5"))
						Expect(ca.Spec.Cards[0].MIGUUIDs).To(HaveLen(0))
						Expect(ca.Spec.Cards[1].Model).To(Equal("test-gpu-model-5"))
						Expect(ca.Spec.Cards[1].GPUUUID).To(Equal("test-gpu-uuid-6"))
						Expect(ca.Spec.Cards[1].MIGUUIDs).To(HaveLen(0))
					default:
						Fail(fmt.Sprintf("unexpected NodeName in created CardAlert resource: %s", caList.Items[idx].Spec.NodeName))
					}
				}
			})
		})

		It("should unsuccessfully handle POST requests with Status BadRequest if body is empty", func() {
			var (
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             k8sClient,
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusBadRequest))
			})
			By("asserting the response body", func() {
				Expect(writer.Body.String()).To(Equal("invalid request body\n"))
			})
			By("checking if CardAlert resource is NOT created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(0))
			})
		})

		It("should unsuccessfully handle POST requests when creation for CardAlert resource failed", func() {
			var (
				mockClient = testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClient)
				httpMethod = &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					logger:             logger,
					client:             mockClient.GetMockClient(),
				}
				writer       = testutils.NewMockWriter()
				reqBodyAlert = common.AlertManagerHookMessage{
					Alerts: []common.AlertManagerAlert{
						{
							Status: common.AlertManagerAlertStatusFiring,
							Labels: map[string]string{
								alertManagerAlertLabelNodeNameKey:         "test-node",
								alertManagerAlertLabelGPUModelKey:         "test-gpu-model",
								alertManagerAlertLabelGPUUUIDKey:          "test-gpu-uuid",
								alertManagerAlertLabelMIGUUIDKey:          "test-mig-uuid",
								common.AlertManagerAlertLabelAlertNameKey: alertManagerAlertCardAlertName,
							},
						},
					},
				}
			)
			By("mocking functions", func() {
				Expect(mockClient.SetCreate([]int{1})).To(Succeed()) // failed to create CardAlert
				mockClient.SetDefault()
			})
			reqBody, err := json.Marshal(reqBodyAlert)
			Expect(err).NotTo(HaveOccurred())
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(string(reqBody))))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusInternalServerError))
			})
			By("checking if CardAlert resource is NOT created", func() {
				caList := &scalev1.CardAlertList{}
				Expect(k8sClient.List(ctx, caList, client.InNamespace(caNamespace))).To(Succeed())
				Expect(caList.Items).To(HaveLen(0))
			})
		})
	})
})
