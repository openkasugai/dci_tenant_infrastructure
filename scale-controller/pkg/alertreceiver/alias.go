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
