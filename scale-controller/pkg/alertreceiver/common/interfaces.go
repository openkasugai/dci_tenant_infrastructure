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

package common

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
)

// AlertReceiver defines the interface for receiving alerts from Prometheus AlertManager webhook.
type AlertReceiver interface {
	// Serve serves the alert receiver.
	Serve(context.Context) error
}

// Handler defines the interface for handling requests to the target endpoint.
type Handler interface {
	// Handler handles requests to the target endpoint.
	Handler(w http.ResponseWriter, req *http.Request)

	// HandlerHTTPMethod implements the interface for handling HTTP methods for the target endpoint.
	HandlerHTTPMethod
}

// HandlerHTTPMethod defines the interface for handling HTTP methods for the target endpoint.
type HandlerHTTPMethod interface {
	// Get handles GET requests to the target endpoint.
	Get(w http.ResponseWriter, req *http.Request)

	// Post handles POST requests to the target endpoint.
	Post(w http.ResponseWriter, req *http.Request)
}

// AlertReceiverNewFunc is a function type for creating a AlertReceiver interface.
type AlertReceiverNewFunc func(*AlertReceiverNewOptions) (AlertReceiver, error)

// AlertHandlerNewFunc is a function type for creating a Handler interface.
type AlertHandlerNewFunc func(logr.Logger, HandlerHTTPMethod) (Handler, error)

// AlertReceiverNewOptions defines the options for creating a AlertReceiverHandler interface.
type AlertReceiverNewOptions struct {
	// ListenAddress is the address for the alert receiver to listen on.
	ListenAddress string

	// ListenPort is the port for the alert receiver to listen on.
	ListenPort string

	// Logger for logging in the handler
	Logger logr.Logger

	// AlertsHTTPMethod is the handler for HTTP methods for the /alerts endpoint. If nil, error will be returned.
	AlertsHTTPMethod HandlerHTTPMethod

	// HealthzHTTPMethod is the handler for HTTP methods for the /healthz endpoint. If nil, error will be returned.
	HealthzHTTPMethod HandlerHTTPMethod

	// Add options here if needed in the future
}
