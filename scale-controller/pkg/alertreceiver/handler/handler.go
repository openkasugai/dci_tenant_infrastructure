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
		h.Get(w, req)
	case http.MethodPost:
		h.Post(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
