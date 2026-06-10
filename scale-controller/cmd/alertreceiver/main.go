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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	cardalert "scale-controller/internal/alertreceiver/cardalert"
)

const (
	cardAlertNamespaceEnvVarName = "CARD_ALERT_NAMESPACE"
	listenAddressEnvVarName      = "ALERT_RECEIVER_LISTEN_ADDRESS"
	listenPortEnvVarName         = "ALERT_RECEIVER_LISTEN_PORT"
)

// main is the main function for the alert receiver.
func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

// run is a function to run the alert receiver.
func run() int {
	// Create logger
	klog.InitFlags(nil)
	logger := textlogger.NewLogger(textlogger.NewConfig()).WithName("card-alert-receiver")
	logger.Info("Starting alert receiver")

	// Setup alert receiver
	alertReceiver, err := cardalert.NewAlertReceiver(&cardalert.CardAlertReceiverNewOptions{
		CardAlertNamespace: os.Getenv(cardAlertNamespaceEnvVarName),
		AlertReceiverNewOptions: cardalert.AlertReceiverNewOptions{
			Logger:        logger,
			ListenAddress: os.Getenv(listenAddressEnvVarName),
			ListenPort:    os.Getenv(listenPortEnvVarName),
		},
	})
	if err != nil {
		logger.Error(err, "Failed to setup alert receiver")
		return 1
	}

	// Run alert receiver
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := alertReceiver.Serve(ctx); err != nil {
		logger.Error(err, "Stopped alert receiver")
	}

	return 0
}
