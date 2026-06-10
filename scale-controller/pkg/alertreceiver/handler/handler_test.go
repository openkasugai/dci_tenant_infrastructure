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

package handler

import (
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

func TestHandler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "AlertReceiver handler Suite")
}

var _ = BeforeSuite(func() {
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
})

// Test for Reconcile() method
var _ = Describe("AlertReceiver handler", Label("receiver"), func() {
	Context("When creating a AlertReceiver handler", func() {
		It("should successfully create AlertReceiver handler", func() {
			var (
				httpMethod = testutils.NewMockHandlerHTTPMethod()
			)
			By("calling NewAlertHandler", func() {
				receiver, err := NewAlertHandler(logger, httpMethod)
				Expect(err).NotTo(HaveOccurred())
				alertsHandler, ok := receiver.(*alertHandler)
				Expect(ok).To(BeTrue())
				Expect(alertsHandler.logger).To(Equal(logger.WithName("handler")))
				Expect(alertsHandler.Handler).NotTo(BeNil())
			})
		})

		It("should unsuccessfully create AlertReceiver handler with nil options", func() {
			var (
				errMsgs = []string{
					"expected Logger has a non-nil sink, but got nil",
				}
			)
			By("calling NewHandler", func() {
				receiver, err := NewAlertHandler(logr.Logger{}, nil)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})

		It("should unsuccessfully create AlertReceiver handler with nil handler", func() {
			var (
				errMsgs = []string{
					"expected HTTP method handler to be not nil, but got nil",
				}
			)
			By("calling NewHandler", func() {
				receiver, err := NewAlertHandler(logger, nil)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})
	})

	Context("When serving alerts requests", func() {
		It("should successfully serve Alerts AlertReceiver handler for GET requests", func() {
			var (
				receiver = &alertHandler{
					logger:            logger.WithName("alerts-handler"),
					HandlerHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("calling AlertsHandler", func() {
				req, err := http.NewRequest("GET", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				receiver.Handler(writer, req)
			})
		})

		It("should successfully serve Alerts AlertReceiver handler for POST requests", func() {
			var (
				receiver = &alertHandler{
					logger:            logger.WithName("alerts-handler"),
					HandlerHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("calling AlertsHandler", func() {
				req, err := http.NewRequest("POST", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				receiver.Handler(writer, req)
			})
		})

		It("should unsuccessfully serve Alerts AlertReceiver handler for unknown requests", func() {
			var (
				receiver = &alertHandler{
					logger:            logger.WithName("alerts-handler"),
					HandlerHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
				writer  = testutils.NewMockWriter()
				reqBody = ``
			)
			By("calling AlertsHandler", func() {
				req, err := http.NewRequest("UNKNOWN", "", io.Reader(strings.NewReader(reqBody)))
				Expect(err).NotTo(HaveOccurred())
				receiver.Handler(writer, req)
				Expect(writer.StatusCode).To(Equal(http.StatusMethodNotAllowed))
				Expect(writer.Body.String()).To(ContainSubstring("method not allowed"))
			})
		})
	})
})
