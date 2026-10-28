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
