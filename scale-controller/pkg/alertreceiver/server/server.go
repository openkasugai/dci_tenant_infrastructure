package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"

	"scale-controller/pkg/alertreceiver/common"
	"scale-controller/pkg/alertreceiver/handler/alerts"
	"scale-controller/pkg/alertreceiver/handler/healthz"
)

// listenEndpoint defines the address and port to listen.
type listenEndpoint struct {
	address string // IP address to listen
	port    string // Port number to listen
}

type endpointHandler struct {
	// endpoint is the endpoint path for the alert receiver handler.
	endpoint string

	// handler is the handler for the endpoint.
	handler common.Handler
}

// alertReceiver defines a struct that implements the AlertReceiver interface.
type alertReceiver struct {
	listen   listenEndpoint    // listen endpoint
	handlers []endpointHandler // handler for http requests
	logger   logr.Logger       // logger for logging
}

// Ensure alertReceiver implements the AlertReceiver interface.
var _ common.AlertReceiver = &alertReceiver{}

var (
	newAlertsHandler  = alerts.NewAlertsHandler
	newHealthzHandler = healthz.NewHealthzHandler
)

const (
	// Default listen address for the alert receiver server.
	listenAddressDefault = "0.0.0.0"

	// Default listen port for the alert receiver server.
	listenPortDefault = "8080"
)

// gracefulShutdownTimeout defines the timeout duration for graceful shutdown of the server when termination signal is received.
const gracefulShutdownTimeout = 10 * time.Second

// NewAlertReceiver is a function to create a new alert receiver instance.
//   - Gets listen endpoint from options.
//   - Creates handler for endpoints of alert receiver.
func NewAlertReceiver(
	options *common.AlertReceiverNewOptions,
) (common.AlertReceiver, error) {
	if options == nil || options.Logger.GetSink() == nil {
		return nil, errors.New("expected options and Logger in options to be not nil, but got nil")
	}
	logger := options.Logger
	logger.Info("Alert receiver setup")

	// Get listenConfig
	listenConfig := getListenEndpointConfig(options)

	// Create alertReceiverHandler for endpoints of alert receiver
	handlers, err := newEndpointHandlers(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler for alert receiver: %w", err)
	}

	return &alertReceiver{
		listen:   listenConfig,
		handlers: handlers,
		logger:   logger,
	}, nil
}

// getListenEndpointConfig is a function to get the listen endpoint config from environment variables or uses default values.
func getListenEndpointConfig(
	options *common.AlertReceiverNewOptions,
) listenEndpoint {
	// Set default values for listen address and port at first
	addr := listenAddressDefault
	port := listenPortDefault

	if options != nil {
		if options.ListenAddress != "" {
			// Override value if ListenAddress is set in options
			addr = options.ListenAddress
		}
		if options.ListenPort != "" {
			// Override value if ListenPort is set in options
			port = options.ListenPort
		}
	}

	return listenEndpoint{
		address: addr,
		port:    port,
	}
}

// newEndpointHandlers is a function to create a new instance of alertHandler as AlertReceiverHandler.
func newEndpointHandlers(
	options *common.AlertReceiverNewOptions,
) ([]endpointHandler, error) {
	logger := options.Logger

	// Create handlers for /alerts and /healthz endpoints if not provided in options
	alertsHandler, err := newAlertsHandler(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create alerts handler: %w", err)
	}
	healthzHandler, err := newHealthzHandler(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create healthz handler: %w", err)
	}

	logger.Info("Successfully created Handler instances for endpoints")

	return []endpointHandler{
		{
			endpoint: alerts.Endpoint,
			handler:  alertsHandler,
		},
		{
			endpoint: healthz.Endpoint,
			handler:  healthzHandler,
		},
	}, nil
}

// Serve is a function to serve the alert receiver.
// endpoints are as follows:
//   - /healthz: health check endpoint
//   - /alerts: endpoint to receive alerts from Prometheus AlertManager
func (r *alertReceiver) Serve(ctx context.Context) error {
	r.logger.Info("Alert receiver started", "address", r.listen.address, "port", r.listen.port)
	addr := r.listen.address + ":" + r.listen.port

	// Set handlers for endpoints from handler interfaces
	for _, handler := range r.handlers {
		r.logger.Info("Setting up handler for endpoint", "endpoint", handler.endpoint)
		http.HandleFunc(handler.endpoint, handler.handler.Handler)
	}

	// Set timeouts for server for security enhancement
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create a context that is canceled when termination signal is received to trigger graceful shutdown of the server
	ctxSignal, cancelSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancelSignal()

	// Start server without TLS because communication in the same cluster is considered secure.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			r.logger.Error(err, "Server stopped")
		}
		cancelSignal()
	}()

	// Wait for termination signal
	<-ctxSignal.Done()

	// If the original context is already canceled(e.g. due to Interrupt signal,...),
	// use a new background context for shutdown to avoid using a canceled context.
	if ctx.Err() != nil {
		ctx = context.Background()
	}

	// Detect termination signal and shutdown server
	r.logger.Info("Shutting down server...")
	ctxShutdown, cancelShutdown := context.WithTimeout(ctx, gracefulShutdownTimeout)
	defer cancelShutdown()

	return server.Shutdown(ctxShutdown)
}
