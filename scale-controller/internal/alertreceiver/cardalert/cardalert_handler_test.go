package cardalert

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"scale-controller/pkg/alertreceiver"
	"scale-controller/test/testutils"
)

// Test for Reconcile() method
var _ = Describe("Card AlertReceiver", Label("ca", "receiver"), func() {
	Context("When creating a Card AlertReceiver", func() {
		const caNamespace = "test-namespace"
		var (
			storedInClusterConfig func() (*rest.Config, error)
			storedNewClient       func(*rest.Config, client.Options) (client.Client, error)
		)

		BeforeEach(func() {
			storedInClusterConfig = InClusterConfig
			storedNewClient = NewClient
		})

		AfterEach(func() {
			// Restore the original functions after each test
			InClusterConfig = storedInClusterConfig
			NewClient = storedNewClient
		})

		It("should successfully create Card AlertReceiver", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
			)
			By("mocking functions", func() {
				InClusterConfig = testutils.NewMockInClusterConfig(&rest.Config{}, nil)
				NewClient = testutils.NewMockNewClient(nil, nil)
			})
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).NotTo(HaveOccurred())
				cardAlertReceiver, ok := receiver.(*cardAlertReceiver)
				Expect(ok).To(BeTrue())
				Expect(cardAlertReceiver.httpMethod).NotTo(BeNil())
				Expect(cardAlertReceiver.alertReceiver).NotTo(BeNil())
				Expect(options.AlertReceiverNewOptions.AlertsHTTPMethod).To(Equal(cardAlertReceiver.httpMethod))
			})
		})

		It("should unsuccessfully create Card AlertReceiver with nil options", func() {
			var (
				options *CardAlertReceiverNewOptions = nil
				errMsgs                              = []string{
					"expected options or Logger is not nil, but got nil",
				}
			)
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver without CardAlert Namespace", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
				}
				errMsgs = []string{
					"failed to create AlertsHTTPMethod for CardAlert",
				}
			)
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})

		It("should unsuccessfully create Card AlertReceiver if alertreceiver.NewAlertReceiver() failed", func() {
			var (
				options = &CardAlertReceiverNewOptions{
					AlertReceiverNewOptions: AlertReceiverNewOptions{
						Logger: logger,
					},
					CardAlertNamespace: caNamespace,
				}
				errMsgs = []string{
					"failed to create AlertReceiver for CardAlert",
				}
			)
			By("mocking functions", func() {
				InClusterConfig = testutils.NewMockInClusterConfig(&rest.Config{}, nil)
				NewClient = testutils.NewMockNewClient(nil, nil)
				alertreceiver.NewAlertReceiver = testutils.NewMockNewAlertReceiver(nil, errors.New("mocked error for alertreceiver.NewAlertReceiver()"))
			})
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})
	})

	Context("When serving a Card AlertReceiver", func() {
		const caNamespace = "test-namespace"

		It("should successfully serve Card AlertReceiver", func() {
			receiver := &cardAlertReceiver{
				httpMethod: &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					client:             k8sClient,
				},
				alertReceiver: testutils.NewMockAlertReceiver(nil),
			}
			By("calling NewAlertReceiver", func() {
				Expect(receiver.Serve(ctx)).To(Succeed())
			})
		})

		It("should unsuccessfully serve Card AlertReceiver", func() {
			receiver := &cardAlertReceiver{
				httpMethod: &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					client:             k8sClient,
				},
				alertReceiver: testutils.NewMockAlertReceiver(errors.New("mocked error for alertReceiver.Serve()")),
			}
			By("calling NewAlertReceiver", func() {
				err := receiver.Serve(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		It("should unsuccessfully create CardAlert Namespace", Label("mock"), func() {
			var (
				errMsgs = []string{
					"failed to create namespace for CardAlert",
				}
			)
			mockClient := testutils.NewMockClientHelper(gomock.NewController(GinkgoT()), k8sClient)
			Expect(mockClient.SetCreate([]int{1})).To(Succeed()) // failed to create Namespace
			mockClient.SetDefault()
			receiver := &cardAlertReceiver{
				httpMethod: &cardAlertHTTPMethod{
					cardAlertNamespace: caNamespace,
					client:             mockClient.GetMockClient(),
				},
				alertReceiver: testutils.NewMockAlertReceiver(nil),
			}
			By("calling NewAlertReceiver", func() {
				err := receiver.Serve(ctx)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
			})
		})
	})
})
