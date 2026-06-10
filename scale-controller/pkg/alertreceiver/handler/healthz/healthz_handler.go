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

package healthz

import (
	"errors"
	"fmt"

	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/pkg/alertreceiver/handler"
	method "scale-controller/pkg/alertreceiver/handler/http-method/healthz"
)

const Endpoint = "/healthz"

var newHealthzHTTPMethod = method.NewHealthzHTTPMethod

// NewHealthzHandler is a function to create a new instance of alertHealthzHandler as AlertReceiverHealthzHandler.
func NewHealthzHandler(
	options *common.AlertReceiverNewOptions,
) (common.Handler, error) {
	if options == nil || options.Logger.GetSink() == nil {
		return nil, errors.New("expected Logger in options to be not nil, but got nil")
	}
	logger := options.Logger.WithName("healthzHandler")

	httpMethod := options.HealthzHTTPMethod
	if httpMethod == nil {
		logger.Info("Healthz HTTP method handler is not provided, using default healthzHTTPMethod")
		var err error
		httpMethod, err = newHealthzHTTPMethod(logger)
		if err != nil {
			// This case is only for failsafe
			return nil, fmt.Errorf("failed to create Healthz HTTP method: %w", err)
		}
	}
	return handler.NewAlertHandler(logger, httpMethod)
}
