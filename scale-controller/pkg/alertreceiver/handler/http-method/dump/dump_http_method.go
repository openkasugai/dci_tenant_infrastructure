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

// This package provides an example implementation for HTTP methods used in HandlerHTTPMethod interface
// and has not been verified to work in practice.
//
// Use this package as a reference to implement your own HTTP methods
// and set them in the AlertsHTTPMethod field of the options for NewAlertReceiver().
package dump

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-logr/logr"

	"scale-controller/pkg/alertreceiver/common"
)

// dumpAlertHTTPMethod defines a handler for dumping AlertManager webhook messages.
type dumpAlertHTTPMethod struct {
	logger logr.Logger
}

// Ensure dumpAlertHTTPMethod implements the AlertsHTTPMethod interface.
var _ common.HandlerHTTPMethod = &dumpAlertHTTPMethod{}

// NewDumpHTTPMethod is a function to create a new instance of dumpAlertHTTPMethod as AlertsHTTPMethod.
func NewDumpHTTPMethod(
	options *common.AlertReceiverNewOptions,
) (common.HandlerHTTPMethod, error) {
	if options == nil || options.Logger.GetSink() == nil {
		return nil, errors.New("expected options and Logger in options to be not nil, but got nil")
	}

	logger := options.Logger.WithName("dump-alert-receiver")
	logger.Info("Creating dumpAlertReceiver")

	return &dumpAlertHTTPMethod{
		logger: logger,
	}, nil
}

// Get is a function to handle GET requests to the /alerts endpoint.
// Alert Manager expects an 200 OK response, otherwise send_resolved will never work
func (h *dumpAlertHTTPMethod) Get(
	w http.ResponseWriter, _ *http.Request,
) {
	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := enc.Encode("OK"); err != nil {
		h.logger.Error(err, "error encoding messages for getHandler")
	}
}

// Post is a function to handle POST requests to the /alerts endpoint.
func (h *dumpAlertHTTPMethod) Post(
	w http.ResponseWriter, req *http.Request,
) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()

	message := &common.AlertManagerHookMessage{}
	if err := dec.Decode(message); err != nil {
		h.logger.Error(err, "failed to decode message")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Dump the AlertManagerHookMessage
	h.dumpAlertManagerHookMessage(message)
}

// dumpAlertManagerHookMessage dumps the contents of the AlertManagerHookMessage.
func (h *dumpAlertHTTPMethod) dumpAlertManagerHookMessage(
	message *common.AlertManagerHookMessage,
) {
	h.logger.Info("Dumping AlertManager Hook Message",
		"Version", message.Version,
		"GroupKey", message.GroupKey,
		"TruncatedAlerts", message.TruncatedAlerts,
		"Status", message.Status,
		"Receiver", message.Receiver,
		"GroupLabels", message.GroupLabels,
		"CommonLabels", message.CommonLabels,
		"CommonAnnotations", message.CommonAnnotations,
		"ExternalURL", message.ExternalURL,
		"Alerts", message.Alerts,
	)
}
