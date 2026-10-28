package healthz

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/test/testutils"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	logger logr.Logger
)

func TestHealthzHandler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Healthz AlertReceiver handler Suite")
}

var _ = BeforeSuite(func() {
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
})

// Test for Reconcile() method
var _ = Describe("Healthz AlertReceiver handler", Label("healthz", "receiver"), func() {
	Context("When creating a Healthz AlertReceiver handler", func() {
		It("should successfully create Healthz AlertReceiver handler", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger: logger,
				}
			)
			By("calling NewHealthzHandler", func() {
				receiver, err := NewHealthzHandler(options)
				Expect(err).NotTo(HaveOccurred())
				Expect(receiver).NotTo(BeNil())
			})
		})

		It("should successfully create Healthz AlertReceiver handler with provided HTTP method", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger:            logger,
					HealthzHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
			)
			By("calling NewHealthzHandler", func() {
				receiver, err := NewHealthzHandler(options)
				Expect(err).NotTo(HaveOccurred())
				Expect(receiver).NotTo(BeNil())
			})
		})

		It("should unsuccessfully create Healthz AlertReceiver handler with nil options", func() {
			var (
				options *common.AlertReceiverNewOptions = nil
				errMsgs                                 = []string{
					"expected Logger in options to be not nil, but got nil",
				}
			)
			By("calling NewHealthzHandler", func() {
				receiver, err := NewHealthzHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})

		It("should unsuccessfully create Healthz AlertReceiver handler when creation of Healthz HTTP method failed", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger: logger,
				}
				errMsgs = []string{
					"failed to create Healthz HTTP method",
				}
			)
			By("mocking function to return error", func() {
				newHealthzHTTPMethod = func(_ logr.Logger) (common.HandlerHTTPMethod, error) {
					return nil, errors.New("mocked error for NewHealthzHTTPMethod")
				}
			})
			By("calling NewHealthzHandler", func() {
				receiver, err := NewHealthzHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})
	})
})
