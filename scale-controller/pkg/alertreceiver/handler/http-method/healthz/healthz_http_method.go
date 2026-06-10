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

// This package provides an example implementation for HTTP methods used in /alerts endpoint
// and has not been verified to work in practice.
//
// Use this package as a reference to implement your own HTTP methods
// and set them in the AlertsHTTPMethod field of the options for NewAlertReceiver().
package healthz

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-logr/logr"

	"scale-controller/pkg/alertreceiver/common"
)

// healthzHTTPMethod defines a handler for health check endpoint.
type healthzHTTPMethod struct {
	logger logr.Logger
}

// Ensure healthzHTTPMethod implements the HandlerHTTPMethod interface.
var _ common.HandlerHTTPMethod = &healthzHTTPMethod{}

// NewHealthzHTTPMethod is a function to create a new instance of healthzHTTPMethod as HandlerHTTPMethod.
func NewHealthzHTTPMethod(
	logger logr.Logger,
) (common.HandlerHTTPMethod, error) {
	if logger.GetSink() == nil {
		return nil, errors.New("expected Logger to be not nil, but got nil")
	}

	logger = logger.WithName("healthz-http-method")

	return &healthzHTTPMethod{
		logger: logger,
	}, nil
}

// Get is a function to handle GET requests to the /healthz endpoint.
func (h *healthzHTTPMethod) Get(
	w http.ResponseWriter, _ *http.Request,
) {
	h.responseOK(w)
}

// Post is a function to handle POST requests to the /healthz endpoint.
func (h *healthzHTTPMethod) Post(
	w http.ResponseWriter, _ *http.Request,
) {
	h.responseOK(w)
}

// responseOK is a helper function to send a 200 OK response with "OK" message in JSON format.
// Alert Manager expects an 200 OK response, otherwise send_resolved will never work
func (h *healthzHTTPMethod) responseOK(w http.ResponseWriter) {
	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := enc.Encode("OK"); err != nil {
		h.logger.Error(err, "error encoding messages for healthzHTTPMethod")
	}
}
