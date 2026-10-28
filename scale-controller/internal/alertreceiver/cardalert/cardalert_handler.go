package cardalert

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"scale-controller/pkg/alertreceiver"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertReceiverNewOptions is alias of alertreceiver.AlertReceiverNewOptions.
type AlertReceiverNewOptions = alertreceiver.AlertReceiverNewOptions

// CardAlertReceiverNewOptions defines the options for creating CardAlertReceiver.
type CardAlertReceiverNewOptions struct {
	AlertReceiverNewOptions AlertReceiverNewOptions
	CardAlertNamespace      string
}

// cardAlertReceiver is a struct that implements the AlertReceiver interface
// for handling CardAlert related alerts.
type cardAlertReceiver struct {
	alertReceiver alertreceiver.AlertReceiver
	httpMethod    *cardAlertHTTPMethod
}

// Ensure cardAlertReceiver implements the AlertReceiver interface
var _ alertreceiver.AlertReceiver = &cardAlertReceiver{}

// NewAlertReceiver is a function to create a new AlertReceiver for CardAlert with the given options.
func NewAlertReceiver(
	options *CardAlertReceiverNewOptions,
) (alertreceiver.AlertReceiver, error) {
	if options == nil || options.AlertReceiverNewOptions.Logger.GetSink() == nil {
		return nil, fmt.Errorf("expected options or Logger is not nil, but got nil")
	}
	logger := options.AlertReceiverNewOptions.Logger
	logger.Info("Setting up CardAlert handler")

	// Create the HTTP method for the /alerts endpoint for CardAlert and set it into options
	httpMethod, err := newCardAlertHandler(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create AlertsHTTPMethod for CardAlert: %w", err)
	}
	options.AlertReceiverNewOptions.AlertsHTTPMethod = httpMethod

	alertReceiver, err := alertreceiver.NewAlertReceiver(&options.AlertReceiverNewOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create AlertReceiver for CardAlert: %w", err)
	}

	return &cardAlertReceiver{
		alertReceiver: alertReceiver,
		httpMethod:    httpMethod,
	}, nil
}

func (r *cardAlertReceiver) Serve(
	ctx context.Context,
) error {
	// Create namespace for CardAlert if not exists
	nsKey := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.httpMethod.cardAlertNamespace,
		},
	}
	if err := r.httpMethod.client.Create(ctx, nsKey); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace for CardAlert: %w", err)
	}

	return r.alertReceiver.Serve(ctx)
}
