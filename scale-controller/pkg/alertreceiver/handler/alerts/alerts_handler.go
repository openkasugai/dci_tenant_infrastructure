package alerts

import (
	"errors"

	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/pkg/alertreceiver/handler"
)

const Endpoint = "/alerts"

// NewAlertsHandler is a function to create a new instance of alertAlertsHandler as AlertReceiverAlertsHandler.
func NewAlertsHandler(
	options *common.AlertReceiverNewOptions,
) (common.Handler, error) {
	if options == nil || options.Logger.GetSink() == nil {
		return nil, errors.New("expected Logger in options to be not nil, but got nil")
	}
	logger := options.Logger.WithName("alertsHandler")

	httpMethod := options.AlertsHTTPMethod
	if httpMethod == nil {
		return nil, errors.New("expected Alerts HTTP method handler to be not nil, but got nil")
	}

	return handler.NewAlertHandler(logger, httpMethod)
}
