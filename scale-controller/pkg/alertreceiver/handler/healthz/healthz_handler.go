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
