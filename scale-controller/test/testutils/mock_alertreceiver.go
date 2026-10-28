package testutils

import (
	"context"
	"net/http"
	"strings"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"scale-controller/pkg/alertreceiver/common"
)

// mockWriter is a mock implementation of http.ResponseWriter for testing purposes.
type mockWriter struct {
	StatusCode int
	HttpHeader http.Header
	Body       strings.Builder
	WriteError error
}

func (w *mockWriter) Header() http.Header {
	return w.HttpHeader
}
func (w *mockWriter) Write(
	b []byte,
) (int, error) {
	if w.WriteError != nil {
		return 0, w.WriteError
	}
	return w.Body.Write(b)
}
func (w *mockWriter) WriteHeader(
	statusCode int,
) {
	w.StatusCode = statusCode
}

var _ http.ResponseWriter = &mockWriter{}

func NewMockWriter() *mockWriter {
	return &mockWriter{
		HttpHeader: make(http.Header),
	}
}

// mockHandlerHTTPMethod is a mock implementation of common.AlertsHTTPMethod for testing purposes.
type mockHandlerHTTPMethod struct{}

func (*mockHandlerHTTPMethod) Get(_ http.ResponseWriter, _ *http.Request)  {}
func (*mockHandlerHTTPMethod) Post(_ http.ResponseWriter, _ *http.Request) {}

var _ common.HandlerHTTPMethod = &mockHandlerHTTPMethod{}

func NewMockHandlerHTTPMethod() *mockHandlerHTTPMethod {
	return &mockHandlerHTTPMethod{}
}

// mockHandler is a mock implementation of common.AlertHandler for testing purposes.
type mockHandler struct {
	readiness bool
}

func (*mockHandler) Handler(_ http.ResponseWriter, _ *http.Request) {}
func (*mockHandler) Get(_ http.ResponseWriter, _ *http.Request)     {}
func (*mockHandler) Post(_ http.ResponseWriter, _ *http.Request)    {}

var _ common.Handler = &mockHandler{}

func NewMockHandler(
	readiness bool,
) *mockHandler {
	return &mockHandler{
		readiness: readiness,
	}
}

// mockAlertReceiver is a mock implementation of common.AlertReceiver for testing purposes.
type mockAlertReceiver struct {
	serveError error
}

func (m *mockAlertReceiver) Serve(
	_ context.Context,
) error {
	return m.serveError
}

var _ common.AlertReceiver = &mockAlertReceiver{}

func NewMockAlertReceiver(
	serveError error,
) *mockAlertReceiver {
	return &mockAlertReceiver{
		serveError: serveError,
	}
}

// NewMockNewAlertReceiver returns a mock function for common.AlertReceiverNewFunc
// that returns the provided AlertReceiver and error.
func NewMockNewAlertReceiver(
	r common.AlertReceiver, err error,
) common.AlertReceiverNewFunc {
	return func(*common.AlertReceiverNewOptions) (common.AlertReceiver, error) {
		return r, err
	}
}

// NewMockInClusterConfig returns a mock function for rest.InClusterConfig
// that returns the provided config and error.
func NewMockInClusterConfig(
	config *rest.Config, err error,
) func() (*rest.Config, error) {
	return func() (*rest.Config, error) {
		return config, err
	}
}

// NewMockNewClient returns a mock function for client.New
// that returns the provided client and error.
func NewMockNewClient(
	c client.Client, err error,
) func(*rest.Config, client.Options) (client.Client, error) {
	return func(*rest.Config, client.Options) (client.Client, error) {
		return c, err
	}
}
