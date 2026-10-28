package alertreceiver

import (
	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/pkg/alertreceiver/handler"
	"scale-controller/pkg/alertreceiver/server"
)

type (
	// AlertReceiver is an alias for common.AlertReceiver.
	AlertReceiver = common.AlertReceiver

	// AlertReceiverNewFunc is an alias for common.AlertReceiverNewFunc.
	AlertReceiverNewFunc = common.AlertReceiverNewFunc

	// AlertHandlerNewFunc is an alias for common.AlertHandlerNewFunc.
	AlertHandlerNewFunc = common.AlertHandlerNewFunc

	// AlertReceiverNewOptions is an alias for common.AlertReceiverNewOptions.
	AlertReceiverNewOptions = common.AlertReceiverNewOptions

	// AlertHTTPMethod is an alias for common.HandlerHTTPMethod.
	AlertHTTPMethod = common.HandlerHTTPMethod
)

var (
	// NewAlertReceiver is a function to create a AlertReceiver interface.
	// You can change the implementation by changing this variable.
	NewAlertReceiver AlertReceiverNewFunc = server.NewAlertReceiver

	// NewAlertHandler is a function to create a AlertHandler interface.
	// You can change the implementation by changing this variable.
	NewAlertHandler AlertHandlerNewFunc = handler.NewAlertHandler
)
