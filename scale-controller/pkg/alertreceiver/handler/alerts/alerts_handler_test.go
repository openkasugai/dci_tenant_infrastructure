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

package alerts

import (
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

func TestAlertsHandler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Alerts AlertReceiver handler Suite")
}

var _ = BeforeSuite(func() {
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
})

// Test for Reconcile() method
var _ = Describe("Alerts AlertReceiver handler", Label("alerts", "receiver"), func() {
	Context("When creating a Alerts AlertReceiver handler", func() {
		It("should successfully create Alerts AlertReceiver handler", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger:           logger,
					AlertsHTTPMethod: testutils.NewMockHandlerHTTPMethod(),
				}
			)
			By("calling NewAlertsHandler", func() {
				receiver, err := NewAlertsHandler(options)
				Expect(err).NotTo(HaveOccurred())
				Expect(receiver).NotTo(BeNil())
			})
		})

		It("should unsuccessfully create Alerts AlertReceiver handler with nil options", func() {
			var (
				options *common.AlertReceiverNewOptions = nil
				errMsgs                                 = []string{
					"expected Logger in options to be not nil, but got nil",
				}
			)
			By("calling NewAlertsHandler", func() {
				receiver, err := NewAlertsHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})

		It("should unsuccessfully create Alerts AlertReceiver handler without AlertsHTTPMethod", func() {
			var (
				options = &common.AlertReceiverNewOptions{
					Logger: logger,
				}
				errMsgs = []string{
					"expected Alerts HTTP method handler to be not nil, but got nil",
				}
			)
			By("calling NewAlertsHandler", func() {
				receiver, err := NewAlertsHandler(options)
				Expect(err).To(HaveOccurred())
				for _, errMsg := range errMsgs {
					Expect(err.Error()).To(ContainSubstring(errMsg))
				}
				Expect(receiver).To(BeNil())
			})
		})
	})
})
