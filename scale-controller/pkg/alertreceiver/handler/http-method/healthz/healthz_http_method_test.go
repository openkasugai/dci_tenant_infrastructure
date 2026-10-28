package healthz

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"scale-controller/test/testutils"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	logger logr.Logger
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Alerts AlertReceiver handler Suite")
}

var _ = BeforeSuite(func() {
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
})

// Test for Reconcile() method
var _ = Describe("Dump AlertReceiver HTTP method", Label("dump", "receiver", "http"), func() {
	Context("When creating Dump AlertReceiver HTTP method", func() {
		It("should successfully create Dump AlertReceiver HTTP method", func() {
			By("calling NewHealthzHTTPMethod", func() {
				httpMethod, err := NewHealthzHTTPMethod(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(httpMethod).NotTo(BeNil())
				healthzHTTPMethod, ok := httpMethod.(*healthzHTTPMethod)
				Expect(ok).To(BeTrue())
				Expect(healthzHTTPMethod.logger).To(Equal(logger.WithName("healthz-http-method")))
			})
		})

		It("should unsuccessfully create Healthz AlertReceiver HTTP method with nil logger", func() {
			var (
				errMsgs = []string{
					"expected Logger to be not nil, but got nil",
				}
			)
			By("calling NewHealthzHTTPMethod", func() {
				httpMethod, err := NewHealthzHTTPMethod(logr.Logger{})
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
				httpMethod = &healthzHTTPMethod{
					logger: logger,
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
				httpMethod = &healthzHTTPMethod{
					logger: logger,
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
		It("should successfully handle POST requests with Status OK", func() {
			var (
				httpMethod = &healthzHTTPMethod{
					logger: logger,
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
				resBody = "\"OK\"\n"
			)
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			By("asserting the response status code", func() {
				Expect(writer.StatusCode).To(Equal(http.StatusOK))
			})
			By("asserting the response body", func() {
				Expect(writer.Body.String()).To(Equal(resBody))
			})
		})

		It("should unsuccessfully handle POST requests if Encode failed", func() {
			var (
				httpMethod = &healthzHTTPMethod{
					logger: logger,
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("mocking functions", func() {
				writer.WriteError = errors.New("mocked error for writer.Write()")
			})
			By("executing POST method", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				httpMethod.Post(writer, req)
			})
			// The result of Encode failure is logged as an error,
			// but the response status code is still 200 OK
			// because the error happens after writing the status code.
		})
	})
})
