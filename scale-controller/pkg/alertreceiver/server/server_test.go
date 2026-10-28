package server

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/pkg/alertreceiver/handler/alerts"
	"scale-controller/pkg/alertreceiver/handler/healthz"
	"scale-controller/test/testutils"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	logger logr.Logger
	ctx    context.Context
)

func TestReceiver(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "AlertReceiver receiver Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
})

// Test for Reconcile() method
var _ = Describe("AlertReceiver receiver", Label("receiver"), func() {
	Context("When creating a AlertReceiver receiver", func() {
		AfterEach(func() {
			By("resetting functions for creating handlers to default", func() {
				newAlertsHandler = alerts.NewAlertsHandler
				newHealthzHandler = healthz.NewHealthzHandler
			})
		})

		It("should successfully create AlertReceiver receiver", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger:           logger,
					AlertsHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
			)
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).NotTo(HaveOccurred())
				r, ok := receiver.(*alertReceiver)
				Expect(ok).To(BeTrue())
				Expect(r.logger).To(Equal(logger))
				Expect(r.handlers).NotTo(BeNil())
				Expect(r.listen.address).To(Equal(listenAddressDefault))
				Expect(r.listen.port).To(Equal(listenPortDefault))
			})
		})

		It("should successfully create AlertReceiver receiver with providing endpoint options", func() {
			var (
				addr    = "8.8.8.8"
				port    = "0"
				options = &common.AlertReceiverNewOptions{
					Logger:           logger,
					AlertsHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
					ListenAddress:    addr,
					ListenPort:       port,
				}
			)
			By("calling NewAlertReceiver", func() {
				receiver, err := NewAlertReceiver(options)
				Expect(err).NotTo(HaveOccurred())
				r, ok := receiver.(*alertReceiver)
				Expect(ok).To(BeTrue())
				Expect(r.logger).To(Equal(logger))
				Expect(r.handlers).NotTo(BeNil())
				Expect(r.listen.address).To(Equal(addr))
				Expect(r.listen.port).To(Equal(port))
			})
		})

		It("should unsuccessfully create AlertReceiver receiver with nil options", func() {
			var (
				options *common.AlertReceiverNewOptions = nil
				errMsgs                                 = []string{
					"expected options and Logger in options to be not nil, but got nil",
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

		It("should unsuccessfully create AlertReceiver receiver when alerts handler creation failed", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger: logger,
				}
				errMsgs = []string{
					"failed to create handler for alert receiver",
					"failed to create alerts handler",
				}
			)
			By("mocking function to return error", func() {
				newAlertsHandler = func(_ *common.AlertReceiverNewOptions) (common.Handler, error) {
					return nil, errors.New("mocked error for newAlertsHandler")
				}
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

		It("should unsuccessfully create AlertReceiver receiver when healthz handler creation failed", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger: logger,
				}
				errMsgs = []string{
					"failed to create handler for alert receiver",
					"failed to create healthz handler",
				}
			)
			By("mocking function to return error", func() {
				newAlertsHandler = func(_ *common.AlertReceiverNewOptions) (common.Handler, error) {
					return nil, nil
				}
				newHealthzHandler = func(_ *common.AlertReceiverNewOptions) (common.Handler, error) {
					return nil, errors.New("mocked error for newHealthzHandler")
				}
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

	Context("When serving alert receiver", func() {
		AfterEach(func() {
			By("resetting default serve mux", func() {
				// Reset default serve mux to avoid conflicts with other tests
				http.DefaultServeMux = http.NewServeMux()
			})
		})

		It("should unsuccessfully serve with graceful shutdown success", func() {
			var (
				addr = "dummy-address"
				port = "dummy-port"
			)
			By("calling Serve", func() {
				receiver, err := NewAlertReceiver(&common.AlertReceiverNewOptions{
					Logger:           logger,
					ListenAddress:    addr,
					ListenPort:       port,
					AlertsHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(receiver).NotTo(BeNil())
				ctxTest, cancelTest := context.WithCancel(ctx)
				cancelTest()
				Expect(receiver.Serve(ctxTest)).To(Succeed())
			})
		})
	})
})
