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
	"errors"
	"net/http"

	"github.com/go-logr/logr"

	"scale-controller/pkg/alertreceiver/common"
)

// alertHandler defines a handler for handling webhook messages from AlertManager.
type alertHandler struct {
	common.HandlerHTTPMethod
	logger logr.Logger
}

// Ensure alertHandler implements the Handler interface.
var _ common.Handler = &alertHandler{}

// NewAlertHandler is a function to create a new instance of alertHandler as Handler.
func NewAlertHandler(
	logger logr.Logger, httpMethod common.HandlerHTTPMethod,
) (common.Handler, error) {
	if logger.GetSink() == nil {
		return nil, errors.New("expected Logger has a non-nil sink, but got nil")
	}
	logger = logger.WithName("handler")

	if httpMethod == nil {
		return nil, errors.New("expected HTTP method handler to be not nil, but got nil")
	}

	logger.Info("Successfully created Handler instance")

	return &alertHandler{
		logger:            logger,
		HandlerHTTPMethod: httpMethod,
	}, nil
}

// Handler is a function to handle requests to the /alerts endpoint.
func (h *alertHandler) Handler(
	w http.ResponseWriter, req *http.Request,
) {
	switch req.Method {
	case http.MethodGet:
		h.logger.Info("Received GET request")
		h.Get(w, req)
	case http.MethodPost:
		h.logger.Info("Received POST request")
		h.Post(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
