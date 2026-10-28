package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"temporal/shared"

	enumspb "go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/klog/v2"
)

// signalRecord stores one SignalWorkflow call for multi-call inspection in tests.
type signalRecord struct {
	workflowID string
	signalName string
	arg        interface{}
}

// fakeClient implements temporalClient for testing.
type fakeClient struct {
	options         client.Options
	signalCalled    int
	lastWorkflowID  string
	lastRunID       string
	lastSignalName  string
	lastSignalArg   interface{}
	signals         []signalRecord
	signalErr       error
	queryErr        error
	queryErrsByCall []error
	queryCallCount  int
	queryGetErr     error
	queryBoolVal    bool
	describeErr     error // error returned by DescribeWorkflowExecution
	workflowClosed  bool  // true = simulate a terminated/closed workflow
	closed          bool
}

func (f *fakeClient) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	f.signalCalled++
	f.lastWorkflowID = workflowID
	f.lastRunID = runID
	f.lastSignalName = signalName
	f.lastSignalArg = arg
	f.signals = append(f.signals, signalRecord{workflowID: workflowID, signalName: signalName, arg: arg})
	return f.signalErr
}

// fakeEncodedValue is a minimal converter.EncodedValue stub for testing.
type fakeEncodedValue struct {
	boolVal bool
	getErr  error
}

func (v *fakeEncodedValue) HasValue() bool { return true }
func (v *fakeEncodedValue) Get(valuePtr interface{}) error {
	if v.getErr != nil {
		return v.getErr
	}
	if p, ok := valuePtr.(*bool); ok {
		*p = v.boolVal
		return nil
	}
	return fmt.Errorf("fakeEncodedValue: unexpected type %T", valuePtr)
}

func (f *fakeClient) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	f.queryCallCount++
	// Use per-call errors when provided (for retry simulation).
	if len(f.queryErrsByCall) > 0 {
		err := f.queryErrsByCall[0]
		f.queryErrsByCall = f.queryErrsByCall[1:]
		if err != nil {
			return nil, err
		}
		return &fakeEncodedValue{boolVal: f.queryBoolVal, getErr: f.queryGetErr}, nil
	}
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return &fakeEncodedValue{boolVal: f.queryBoolVal, getErr: f.queryGetErr}, nil
}

func (f *fakeClient) Close() { f.closed = true }

// DescribeWorkflowExecution returns a running execution by default.
// Set workflowClosed=true to simulate a terminated workflow (CloseTime != nil).
// Set describeErr to simulate a connection error.
func (f *fakeClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	info := &workflowpb.WorkflowExecutionInfo{
		Status: enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
	}
	if f.workflowClosed {
		info.CloseTime = timestamppb.Now()
		info.Status = enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED
	}
	return &workflowservice.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: info,
	}, nil
}

func mustMarshalVectorSignalBody(t *testing.T, namespace, pod string) []byte {
	t.Helper()

	inner := shared.InnerMetadata{
		Message: "msg",
		Metadata: shared.PodMetaData{
			Namespace: namespace,
			Name:      pod,
		},
		Spec: shared.PodSpec{Type: "t1"},
		Status: shared.PodStatus{
			AppStatus: shared.AppStatus{},
		},
	}

	innerBytes, err := json.Marshal(inner)
	if err != nil {
		t.Fatalf("failed to marshal inner: %v", err)
	}

	outer := shared.OuterStruct{Metadata: string(innerBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	return body
}

// Test run uses default Temporal address/namespace and starts HTTP server.
func TestRun_UsesDefaultEnvAndStartsServer(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
	}()

	// Clear env so defaults are used.
	origAddr := os.Getenv("TEMPORAL_ADDRESS")
	origNS := os.Getenv("TEMPORAL_NAMESPACE")
	defer func() {
		_ = os.Setenv("TEMPORAL_ADDRESS", origAddr)
		_ = os.Setenv("TEMPORAL_NAMESPACE", origNS)
	}()
	_ = os.Unsetenv("TEMPORAL_ADDRESS")
	_ = os.Unsetenv("TEMPORAL_NAMESPACE")

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		fc.options = opts
		return fc, nil
	}

	var listenedAddr string
	var listenedHandler http.Handler
	httpListenAndServe = func(addr string, handler http.Handler) error {
		listenedAddr = addr
		listenedHandler = handler
		return nil
	}

	// Use a fresh ServeMux for this test.
	http.DefaultServeMux = http.NewServeMux()

	run()

	if fc.options.HostPort != "temporal-stack-frontend.temporal.svc.cluster.local:7233" {
		t.Fatalf("unexpected HostPort: %q", fc.options.HostPort)
	}
	if fc.options.Namespace != "default" {
		t.Fatalf("unexpected Namespace: %q", fc.options.Namespace)
	}
	if listenedAddr != ":8080" {
		t.Fatalf("expected server to listen on :8080, got %q", listenedAddr)
	}
	// Handler is nil in our wrapper when using default mux.
	if listenedHandler != nil {
		t.Fatalf("expected listenedHandler to be nil when using DefaultServeMux")
	}
}

func TestLoadShardCount_InvalidValueUsesDefault(t *testing.T) {
	origShardCount := os.Getenv("SHARD_COUNT")
	defer func() {
		_ = os.Setenv("SHARD_COUNT", origShardCount)
	}()

	tests := []struct {
		name  string
		value string
	}{
		{name: "non numeric", value: "invalid"},
		{name: "zero", value: "0"},
		{name: "negative", value: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("SHARD_COUNT", tt.value)

			if got := loadShardCount(); got != defaultShardCount {
				t.Fatalf("expected default shard count %d, got %d", defaultShardCount, got)
			}
		})
	}
}

// Test vector-signal handler end-to-end with a valid payload.
func TestRun_VectorSignalHandler_Success(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	// Mock podExistsFunc to always return true (pod exists)
	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	httpListenAndServe = func(addr string, handler http.Handler) error {
		// Do not actually start a server in tests.
		return nil
	}

	// Fresh mux for handler registration.
	http.DefaultServeMux = http.NewServeMux()

	run()

	// Build a valid InnerMetadata and wrap it into OuterStruct.
	inner := shared.InnerMetadata{
		Message: "msg",
		Metadata: shared.PodMetaData{
			Namespace: "ns1",
			Name:      "pod1",
		},
		Spec: shared.PodSpec{Type: "t1"},
		Status: shared.PodStatus{
			AppStatus: shared.AppStatus{},
		},
	}
	innerBytes, err := json.Marshal(inner)
	if err != nil {
		t.Fatalf("failed to marshal inner: %v", err)
	}
	outer := shared.OuterStruct{Metadata: string(innerBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Body.String() != "Signal sent." {
		t.Fatalf("unexpected response body: %q", rec.Body.String())
	}

	if fc.signalCalled != 1 {
		t.Fatalf("expected SignalWorkflow to be called once, got %d", fc.signalCalled)
	}
	if fc.lastWorkflowID != shardedWorkflowID("pod1") || fc.lastRunID != "" || fc.lastSignalName != "vector-signal" {
		t.Fatalf("unexpected SignalWorkflow identifiers: %q %q %q", fc.lastWorkflowID, fc.lastRunID, fc.lastSignalName)
	}

	update, ok := fc.lastSignalArg.(shared.PodUpdateStatus)
	if !ok {
		t.Fatalf("expected argument type PodUpdateStatus, got %T", fc.lastSignalArg)
	}
	if update.Metadata.Namespace != "ns1" || update.Metadata.Name != "pod1" {
		t.Fatalf("unexpected metadata in update: %+v", update.Metadata)
	}
}

func TestRun_VectorSignalHandler_SignalWorkflowErrorAfterValidPayload(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{signalErr: errors.New("signal error")}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()
	run()

	body := mustMarshalVectorSignalBody(t, "ns1", "pod1")
	req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(rec, req)

	if fc.signalCalled != 1 {
		t.Fatalf("expected SignalWorkflow to be called once, got %d", fc.signalCalled)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty response body on signal error, got %q", rec.Body.String())
	}
}

func TestRun_VectorSignalHandler_ResponseWriteErrorAfterValidPayload(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()
	run()

	body := mustMarshalVectorSignalBody(t, "ns1", "pod1")
	req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	writer := &errorWriter{}

	http.DefaultServeMux.ServeHTTP(writer, req)

	if fc.signalCalled != 1 {
		t.Fatalf("expected SignalWorkflow to be called once, got %d", fc.signalCalled)
	}
}

// Test alertmanager-signal handler end-to-end with a valid payload.
func TestRun_AlertmanagerSignalHandler_Success(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	// Mock podExistsFunc to always return true (pod exists)
	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()

	run()

	alerts := shared.PodAlerts{
		Alerts: []shared.Alert{
			{
				Labels: shared.AlertLabels{
					Namespace: "ns1",
					Pod:       "pod1",
				},
			},
		},
	}
	alertsBytes, err := json.Marshal(alerts)
	if err != nil {
		t.Fatalf("failed to marshal alerts: %v", err)
	}
	outer := shared.OuterStruct{Metadata: string(alertsBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Body.String() != "Signal sent." {
		t.Fatalf("unexpected response body: %q", rec.Body.String())
	}

	if fc.signalCalled != 1 {
		t.Fatalf("expected SignalWorkflow to be called once, got %d", fc.signalCalled)
	}
	if fc.lastWorkflowID != shardedWorkflowID("pod1") || fc.lastRunID != "" || fc.lastSignalName != "alertmanager-signal" {
		t.Fatalf("unexpected SignalWorkflow identifiers: %q %q %q", fc.lastWorkflowID, fc.lastRunID, fc.lastSignalName)
	}
}

// Test run when Temporal client dial fails.
func TestRun_DialError(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
	}()

	called := 0
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		called++
		return nil, errors.New("dial error")
	}

	httpListenAndServe = func(addr string, handler http.Handler) error {
		t.Fatalf("ListenAndServe should not be called when dial fails")
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()

	run()

	if called != 1 {
		t.Fatalf("expected temporalClientDial to be called once, got %d", called)
	}
}

// Test run when HTTP server returns an error after successful client dial.
func TestRun_HTTPServerError(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
	}()

	// Ensure env is set to avoid relying on defaults (not strictly necessary but explicit).
	origAddr := os.Getenv("TEMPORAL_ADDRESS")
	origNS := os.Getenv("TEMPORAL_NAMESPACE")
	defer func() {
		_ = os.Setenv("TEMPORAL_ADDRESS", origAddr)
		_ = os.Setenv("TEMPORAL_NAMESPACE", origNS)
	}()
	_ = os.Setenv("TEMPORAL_ADDRESS", "addr")
	_ = os.Setenv("TEMPORAL_NAMESPACE", "ns")

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	called := 0
	httpListenAndServe = func(addr string, handler http.Handler) error {
		called++
		return errors.New("listen error")
	}

	http.DefaultServeMux = http.NewServeMux()

	run()

	if called != 1 {
		t.Fatalf("expected httpListenAndServe to be called once, got %d", called)
	}
}

// errorReader is an io.Reader that always returns an error.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

// errorWriter is a ResponseWriter whose Write method always errors.
type errorWriter struct {
	headers http.Header
}

func (e *errorWriter) Header() http.Header {
	if e.headers == nil {
		e.headers = http.Header{}
	}
	return e.headers
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

func (e *errorWriter) WriteHeader(statusCode int) {}

// Test error branches in vector-signal handler: read error, too large body, bad JSON, SignalWorkflow error, and write error.
func TestRun_VectorSignalHandler_ErrorBranches(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	// Default: Pod exists (for most test cases)
	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	httpListenAndServe = func(addr string, handler http.Handler) error { return nil }

	http.DefaultServeMux = http.NewServeMux()

	run()

	// 1) Read error
	req := httptest.NewRequest(http.MethodPost, "/vector-signal", io.NopCloser(&errorReader{}))
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 2) Body too large
	largeBody := bytes.Repeat([]byte("a"), int(maxRequestSize)+1)
	req = httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(largeBody))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 3) Invalid outer JSON
	req = httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader([]byte("not-json")))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 4) Invalid inner JSON in metadata
	outer := shared.OuterStruct{Metadata: "not-json"}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 5) SignalWorkflow error
	inner := shared.InnerMetadata{}
	innerBytes, err := json.Marshal(inner)
	if err != nil {
		t.Fatalf("failed to marshal inner: %v", err)
	}
	outer = shared.OuterStruct{Metadata: string(innerBytes)}
	body, err = json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}
	fc.signalErr = errors.New("signal error")
	req = httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 6) Response write error
	fc.signalErr = nil
	req = httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
	ew := &errorWriter{}
	http.DefaultServeMux.ServeHTTP(ew, req)
}

// Test error branches in alertmanager-signal handler similar to vector-signal.
func TestRun_AlertmanagerSignalHandler_ErrorBranches(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	// Default: Pod exists (for most test cases)
	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	httpListenAndServe = func(addr string, handler http.Handler) error { return nil }

	http.DefaultServeMux = http.NewServeMux()

	run()

	// 1) Read error
	req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", io.NopCloser(&errorReader{}))
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 2) Body too large
	largeBody := bytes.Repeat([]byte("a"), int(maxRequestSize)+1)
	req = httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(largeBody))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 3) Invalid outer JSON
	req = httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader([]byte("not-json")))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 4) Invalid alerts JSON in metadata
	outer := shared.OuterStruct{Metadata: "not-json"}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 5) SignalWorkflow error — include a real alert so bufferOrSignal is actually called
	alerts5 := shared.PodAlerts{
		Alerts: []shared.Alert{
			{Labels: shared.AlertLabels{Namespace: "ns1", Pod: "pod1"}},
		},
	}
	alertsBytes5, err := json.Marshal(alerts5)
	if err != nil {
		t.Fatalf("failed to marshal alerts: %v", err)
	}
	outer = shared.OuterStruct{Metadata: string(alertsBytes5)}
	body, err = json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}
	fc.signalErr = errors.New("signal error")
	req = httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// 6) Response write error — include an alert so the signal succeeds before the write
	fc.signalErr = nil
	req = httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	ew := &errorWriter{}
	http.DefaultServeMux.ServeHTTP(ew, req)
}

// Test vector-signal handler Pod existence check error cases
func TestRun_VectorSignalHandler_PodChecks(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()
	run()

	// Test 1: Empty namespace
	t.Run("EmptyNamespace", func(t *testing.T) {
		inner := shared.InnerMetadata{
			Message: "msg",
			Metadata: shared.PodMetaData{
				Namespace: "", // Empty namespace
				Name:      "pod1",
			},
		}
		innerBytes, _ := json.Marshal(inner)
		outer := shared.OuterStruct{Metadata: string(innerBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		// Should return OK but not signal (early return)
		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 2: Empty pod name
	t.Run("EmptyPodName", func(t *testing.T) {
		fc.signalCalled = 0
		inner := shared.InnerMetadata{
			Message: "msg",
			Metadata: shared.PodMetaData{
				Namespace: "ns1",
				Name:      "", // Empty pod name
			},
		}
		innerBytes, _ := json.Marshal(inner)
		outer := shared.OuterStruct{Metadata: string(innerBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 3: podExistsFunc returns error
	t.Run("PodExistsFuncError", func(t *testing.T) {
		fc.signalCalled = 0
		podExistsFunc = func(namespace, podName string) (bool, error) {
			return false, errors.New("k8s api error")
		}

		inner := shared.InnerMetadata{
			Message: "msg",
			Metadata: shared.PodMetaData{
				Namespace: "ns1",
				Name:      "pod1",
			},
		}
		innerBytes, _ := json.Marshal(inner)
		outer := shared.OuterStruct{Metadata: string(innerBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 4: Pod does not exist
	t.Run("PodNotFound", func(t *testing.T) {
		fc.signalCalled = 0
		podExistsFunc = func(namespace, podName string) (bool, error) {
			return false, nil // Pod does not exist
		}

		inner := shared.InnerMetadata{
			Message: "msg",
			Metadata: shared.PodMetaData{
				Namespace: "ns1",
				Name:      "pod1",
			},
		}
		innerBytes, _ := json.Marshal(inner)
		outer := shared.OuterStruct{Metadata: string(innerBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/vector-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})
}

// Test alertmanager-signal handler Pod existence check error cases
func TestRun_AlertmanagerSignalHandler_PodChecks(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()
	run()

	// Test 1: Empty namespace in alert
	t.Run("EmptyNamespace", func(t *testing.T) {
		alerts := shared.PodAlerts{
			Alerts: []shared.Alert{
				{
					Labels: shared.AlertLabels{
						Namespace: "", // Empty
						Pod:       "pod1",
					},
				},
			},
		}
		alertsBytes, _ := json.Marshal(alerts)
		outer := shared.OuterStruct{Metadata: string(alertsBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 2: Empty pod in alert
	t.Run("EmptyPodName", func(t *testing.T) {
		fc.signalCalled = 0
		alerts := shared.PodAlerts{
			Alerts: []shared.Alert{
				{
					Labels: shared.AlertLabels{
						Namespace: "ns1",
						Pod:       "", // Empty
					},
				},
			},
		}
		alertsBytes, _ := json.Marshal(alerts)
		outer := shared.OuterStruct{Metadata: string(alertsBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 3: podExistsFunc returns error
	t.Run("PodExistsFuncError", func(t *testing.T) {
		fc.signalCalled = 0
		podExistsFunc = func(namespace, podName string) (bool, error) {
			return false, errors.New("k8s api error")
		}

		alerts := shared.PodAlerts{
			Alerts: []shared.Alert{
				{
					Labels: shared.AlertLabels{
						Namespace: "ns1",
						Pod:       "pod1",
					},
				},
			},
		}
		alertsBytes, _ := json.Marshal(alerts)
		outer := shared.OuterStruct{Metadata: string(alertsBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})

	// Test 4: Pod does not exist (Warning case)
	t.Run("PodNotFound", func(t *testing.T) {
		fc.signalCalled = 0
		podExistsFunc = func(namespace, podName string) (bool, error) {
			return false, nil // Pod does not exist
		}

		alerts := shared.PodAlerts{
			Alerts: []shared.Alert{
				{
					Labels: shared.AlertLabels{
						Namespace: "ns1",
						Pod:       "pod1",
					},
				},
			},
		}
		alertsBytes, _ := json.Marshal(alerts)
		outer := shared.OuterStruct{Metadata: string(alertsBytes)}
		body, _ := json.Marshal(outer)

		req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status OK, got %d", rec.Code)
		}
		if fc.signalCalled != 0 {
			t.Errorf("expected SignalWorkflow not to be called, but was called %d times", fc.signalCalled)
		}
	})
}

// Test podExists function directly
func TestPodExists(t *testing.T) {
	origInClusterConfig := inClusterConfigFunc
	origNewForConfig := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origInClusterConfig
		newForConfigFunc = origNewForConfig
	}()

	// Test 1: Pod exists
	t.Run("PodExists", func(t *testing.T) {
		inClusterConfigFunc = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-ns",
			},
		}
		clientset := fake.NewSimpleClientset(pod)

		newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
			return clientset, nil
		}

		exists, err := podExists("test-ns", "test-pod")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected pod to exist")
		}
	})

	// Test 2: Pod does not exist
	t.Run("PodNotExist", func(t *testing.T) {
		inClusterConfigFunc = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

		clientset := fake.NewSimpleClientset() // No pods

		newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
			return clientset, nil
		}

		exists, err := podExists("test-ns", "nonexistent-pod")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected pod to not exist")
		}
	})

	// Test 3: InClusterConfig error
	t.Run("InClusterConfigError", func(t *testing.T) {
		inClusterConfigFunc = func() (*rest.Config, error) {
			return nil, errors.New("config error")
		}

		exists, err := podExists("test-ns", "test-pod")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if exists {
			t.Error("expected exists to be false on error")
		}
	})

	// Test 4: NewForConfig error
	t.Run("NewForConfigError", func(t *testing.T) {
		inClusterConfigFunc = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

		newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
			return nil, errors.New("client error")
		}

		exists, err := podExists("test-ns", "test-pod")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if exists {
			t.Error("expected exists to be false on error")
		}
	})

	// Test 5: API error (other than NotFound)
	t.Run("APIError", func(t *testing.T) {
		inClusterConfigFunc = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

		// Create a clientset that will return an error
		clientset := fake.NewSimpleClientset()
		// Inject a reactor to return a specific error
		clientset.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, apierrors.NewInternalError(errors.New("internal error"))
		})

		newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
			return clientset, nil
		}

		exists, err := podExists("test-ns", "test-pod")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if exists {
			t.Error("expected exists to be false on error")
		}
	})
}

func TestLoadMaxSignalBufSize(t *testing.T) {
	orig := os.Getenv("MAX_SIGNAL_BUF_SIZE")
	defer func() { _ = os.Setenv("MAX_SIGNAL_BUF_SIZE", orig) }()

	// Valid value: returns parsed integer
	_ = os.Setenv("MAX_SIGNAL_BUF_SIZE", "300")
	if got := loadMaxSignalBufSize(); got != 300 {
		t.Errorf("expected 300, got %d", got)
	}

	// Invalid string: falls back to default
	_ = os.Setenv("MAX_SIGNAL_BUF_SIZE", "invalid")
	if got := loadMaxSignalBufSize(); got != defaultMaxSignalBufSize {
		t.Errorf("expected default %d, got %d", defaultMaxSignalBufSize, got)
	}

	// Zero: falls back to default
	_ = os.Setenv("MAX_SIGNAL_BUF_SIZE", "0")
	if got := loadMaxSignalBufSize(); got != defaultMaxSignalBufSize {
		t.Errorf("expected default %d, got %d", defaultMaxSignalBufSize, got)
	}

	// Negative: falls back to default
	_ = os.Setenv("MAX_SIGNAL_BUF_SIZE", "-1")
	if got := loadMaxSignalBufSize(); got != defaultMaxSignalBufSize {
		t.Errorf("expected default %d, got %d", defaultMaxSignalBufSize, got)
	}

	// Unset: returns default
	_ = os.Unsetenv("MAX_SIGNAL_BUF_SIZE")
	if got := loadMaxSignalBufSize(); got != defaultMaxSignalBufSize {
		t.Errorf("expected default %d, got %d", defaultMaxSignalBufSize, got)
	}
}

// Test queryIsDraining returns true when QueryWorkflow returns an error.
func TestQueryIsDraining_QueryWorkflowError(t *testing.T) {
	fc := &fakeClient{queryErr: fmt.Errorf("query failed")}
	got := queryIsDraining(context.Background(), fc, "monitor-workflow-0")
	if !got {
		t.Errorf("expected true (treat as draining) on QueryWorkflow error, got false")
	}
}

// Test queryIsDraining returns true when resp.Get() returns an error.
func TestQueryIsDraining_GetError(t *testing.T) {
	fc := &fakeClient{queryGetErr: fmt.Errorf("decode failed")}
	got := queryIsDraining(context.Background(), fc, "monitor-workflow-0")
	if !got {
		t.Errorf("expected true (treat as draining) on Get error, got false")
	}
}

// Test bufferOrSignal buffers the signal when workflow is draining (normal case, no overflow).
func TestBufferOrSignal_Draining_Buffered(t *testing.T) {
	// queryErr makes queryIsDraining return true (draining)
	fc := &fakeClient{queryErr: fmt.Errorf("draining")}

	origBuf := signalBuf
	origSize := maxSignalBufSize
	defer func() {
		signalBuf = origBuf
		maxSignalBufSize = origSize
	}()
	signalBuf = nil
	maxSignalBufSize = 10

	err := bufferOrSignal(context.Background(), fc, "monitor-workflow-0", "vector-signal", "payload")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	signalBufMu.Lock()
	defer signalBufMu.Unlock()
	if len(signalBuf) != 1 {
		t.Fatalf("expected 1 buffered signal, got %d", len(signalBuf))
	}
	if signalBuf[0].signalName != "vector-signal" {
		t.Errorf("expected signal name 'vector-signal', got %s", signalBuf[0].signalName)
	}
}

// Test bufferOrSignal drops the oldest signal when the buffer is full (overflow case).
func TestBufferOrSignal_Draining_Overflow(t *testing.T) {
	// queryErr makes queryIsDraining return true (draining)
	fc := &fakeClient{queryErr: fmt.Errorf("draining")}

	origBuf := signalBuf
	origSize := maxSignalBufSize
	defer func() {
		signalBuf = origBuf
		maxSignalBufSize = origSize
	}()
	maxSignalBufSize = 2
	signalBuf = []bufferedSignal{
		{workflowID: "monitor-workflow-0", signalName: "oldest-signal", arg: nil},
		{workflowID: "monitor-workflow-0", signalName: "second-signal", arg: nil},
	}

	err := bufferOrSignal(context.Background(), fc, "monitor-workflow-0", "new-signal", "payload")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	signalBufMu.Lock()
	defer signalBufMu.Unlock()
	if len(signalBuf) != 2 {
		t.Fatalf("expected buffer size 2 after overflow, got %d", len(signalBuf))
	}
	// oldest-signal should have been dropped; buffer should now be [second-signal, new-signal]
	if signalBuf[0].signalName != "second-signal" {
		t.Errorf("expected first signal to be 'second-signal', got %s", signalBuf[0].signalName)
	}
	if signalBuf[1].signalName != "new-signal" {
		t.Errorf("expected second signal to be 'new-signal', got %s", signalBuf[1].signalName)
	}
}

// Test flushSignalBuffer does nothing when the buffer is empty (early return path).
func TestFlushSignalBuffer_EmptyBuffer(t *testing.T) {
	fc := &fakeClient{}

	origBuf := signalBuf
	defer func() { signalBuf = origBuf }()
	signalBuf = nil

	flushSignalBuffer(context.Background(), fc)

	// SignalWorkflow should not have been called
	if fc.signalCalled != 0 {
		t.Errorf("expected no SignalWorkflow calls, got %d", fc.signalCalled)
	}
}

// Test flushSignalBuffer flushes buffered signals when workflow is not draining.
func TestFlushSignalBuffer_FlushesSignals(t *testing.T) {
	// queryErr=nil, boolVal=false → not draining → signals get sent
	fc := &fakeClient{}

	origBuf := signalBuf
	defer func() { signalBuf = origBuf }()
	signalBuf = []bufferedSignal{
		{workflowID: "monitor-workflow-0", signalName: "vector-signal", arg: "payload1"},
		{workflowID: "monitor-workflow-0", signalName: "vector-signal", arg: "payload2"},
	}

	flushSignalBuffer(context.Background(), fc)

	if fc.signalCalled != 2 {
		t.Errorf("expected 2 SignalWorkflow calls, got %d", fc.signalCalled)
	}
	signalBufMu.Lock()
	defer signalBufMu.Unlock()
	if len(signalBuf) != 0 {
		t.Errorf("expected empty buffer after flush, got %d", len(signalBuf))
	}
}

// Test flushSignalBuffer re-buffers signals when workflow is still draining during flush.
func TestFlushSignalBuffer_StillDraining(t *testing.T) {
	// queryBoolVal=true → queryIsDraining returns true → signal kept in remaining
	fc := &fakeClient{queryBoolVal: true}

	origBuf := signalBuf
	defer func() { signalBuf = origBuf }()
	signalBuf = []bufferedSignal{
		{workflowID: "monitor-workflow-0", signalName: "vector-signal", arg: nil},
	}

	flushSignalBuffer(context.Background(), fc)

	// SignalWorkflow should not have been called
	if fc.signalCalled != 0 {
		t.Errorf("expected no SignalWorkflow calls, got %d", fc.signalCalled)
	}
	signalBufMu.Lock()
	defer signalBufMu.Unlock()
	// Signal should be re-buffered as remaining
	if len(signalBuf) != 1 {
		t.Fatalf("expected 1 re-buffered signal, got %d", len(signalBuf))
	}
	if signalBuf[0].signalName != "vector-signal" {
		t.Errorf("expected re-buffered signal name 'vector-signal', got %s", signalBuf[0].signalName)
	}
}

// Test flushSignalBuffer re-buffers signals when SignalWorkflow returns an error.
func TestFlushSignalBuffer_SignalWorkflowError(t *testing.T) {
	// queryBoolVal=false → not draining → tries SignalWorkflow, which fails
	fc := &fakeClient{signalErr: fmt.Errorf("signal failed")}

	origBuf := signalBuf
	defer func() { signalBuf = origBuf }()
	signalBuf = []bufferedSignal{
		{workflowID: "monitor-workflow-0", signalName: "vector-signal", arg: nil},
	}

	flushSignalBuffer(context.Background(), fc)

	if fc.signalCalled != 1 {
		t.Errorf("expected 1 SignalWorkflow call, got %d", fc.signalCalled)
	}
	signalBufMu.Lock()
	defer signalBufMu.Unlock()
	// Signal should be re-buffered due to error
	if len(signalBuf) != 1 {
		t.Fatalf("expected 1 re-buffered signal after error, got %d", len(signalBuf))
	}
	if signalBuf[0].signalName != "vector-signal" {
		t.Errorf("expected re-buffered signal name 'vector-signal', got %s", signalBuf[0].signalName)
	}
}

// Test startDrainLoop fires flushSignalBuffer via ticker.
func TestStartDrainLoop_FiresFlush(t *testing.T) {
	fc := &fakeClient{} // queryBoolVal=false → not draining → SignalWorkflow called

	origBuf := signalBuf
	origInterval := drainLoopInterval
	defer func() {
		signalBuf = origBuf
		drainLoopInterval = origInterval
	}()
	drainLoopInterval = 10 * time.Millisecond
	signalBuf = []bufferedSignal{
		{workflowID: "monitor-workflow-0", signalName: "vector-signal", arg: nil},
	}

	ctx, cancel := context.WithCancel(context.Background())
	startDrainLoop(ctx, fc)

	// Wait for at least one tick to fire and flush
	time.Sleep(50 * time.Millisecond)
	cancel()

	if fc.signalCalled == 0 {
		t.Errorf("expected SignalWorkflow to be called by drain loop, got 0 calls")
	}
}

// TestRun_AlertmanagerSignalHandler_MultiShardRouting verifies that alerts are
// distributed to each shard based on each pod's FNV hash, i.e. every shard
// receives exactly the alerts whose pods map to it.
func TestRun_AlertmanagerSignalHandler_MultiShardRouting(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error { return nil }

	http.DefaultServeMux = http.NewServeMux()
	run()

	// Use multiple pods; shardedWorkflowID determines expected routing dynamically.
	testPods := []string{"pod-alpha", "pod-beta", "pod-gamma", "pod-delta", "pod-epsilon"}

	expectedBuckets := make(map[string][]string) // workflowID → pod names
	for _, p := range testPods {
		wfID := shardedWorkflowID(p)
		expectedBuckets[wfID] = append(expectedBuckets[wfID], p)
	}
	if len(expectedBuckets) < 2 {
		t.Skip("all test pods hash to the same shard; adjust pod names for multi-shard coverage")
	}

	alerts := shared.PodAlerts{}
	for _, p := range testPods {
		alerts.Alerts = append(alerts.Alerts, shared.Alert{
			Labels: shared.AlertLabels{Namespace: "ns1", Pod: p},
		})
	}
	alertsBytes, err := json.Marshal(alerts)
	if err != nil {
		t.Fatalf("failed to marshal alerts: %v", err)
	}
	outer := shared.OuterStruct{Metadata: string(alertsBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Body.String() != "Signal sent." {
		t.Fatalf("unexpected response body: %q", rec.Body.String())
	}

	// One SignalWorkflow call per distinct shard.
	if fc.signalCalled != len(expectedBuckets) {
		t.Fatalf("expected %d SignalWorkflow calls (one per shard), got %d", len(expectedBuckets), fc.signalCalled)
	}

	// Each recorded signal must target the correct workflow and contain only
	// alerts whose pods hash to that workflow.
	for _, sig := range fc.signals {
		if sig.signalName != "alertmanager-signal" {
			t.Errorf("unexpected signal name: %q", sig.signalName)
		}
		bucket, ok := sig.arg.(*shared.PodAlerts)
		if !ok {
			t.Fatalf("expected *shared.PodAlerts arg, got %T", sig.arg)
		}
		for _, alert := range bucket.Alerts {
			pod := alert.Labels.GetPod()
			if shardedWorkflowID(pod) != sig.workflowID {
				t.Errorf("alert for pod %q routed to workflow %q, expected %q",
					pod, sig.workflowID, shardedWorkflowID(pod))
			}
		}
	}
}

// TestRun_AlertmanagerSignalHandler_EmptyAlerts verifies that a request with
// zero alerts sends no signal. This differs from the old implementation that
// always routed to monitor-workflow-0 even for an empty alert list.
func TestRun_AlertmanagerSignalHandler_EmptyAlerts(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error { return nil }

	http.DefaultServeMux = http.NewServeMux()
	run()

	alerts := shared.PodAlerts{} // no alerts
	alertsBytes, err := json.Marshal(alerts)
	if err != nil {
		t.Fatalf("failed to marshal alerts: %v", err)
	}
	outer := shared.OuterStruct{Metadata: string(alertsBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if fc.signalCalled != 0 {
		t.Fatalf("expected no SignalWorkflow calls for empty alert list, got %d", fc.signalCalled)
	}
}

// TestRun_AlertmanagerSignalHandler_SignalError_MultiShard verifies that when
// SignalWorkflow returns an error the handler returns early without writing a
// response body.
func TestRun_AlertmanagerSignalHandler_SignalError_MultiShard(t *testing.T) {
	origDial := temporalClientDial
	origListen := httpListenAndServe
	origPodExists := podExistsFunc
	defer func() {
		temporalClientDial = origDial
		httpListenAndServe = origListen
		podExistsFunc = origPodExists
	}()

	podExistsFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}

	fc := &fakeClient{signalErr: errors.New("signal failed")}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	httpListenAndServe = func(addr string, handler http.Handler) error { return nil }

	http.DefaultServeMux = http.NewServeMux()
	run()

	testPods := []string{"pod-alpha", "pod-beta", "pod-gamma", "pod-delta", "pod-epsilon"}
	alerts := shared.PodAlerts{}
	for _, p := range testPods {
		alerts.Alerts = append(alerts.Alerts, shared.Alert{
			Labels: shared.AlertLabels{Namespace: "ns1", Pod: p},
		})
	}
	alertsBytes, err := json.Marshal(alerts)
	if err != nil {
		t.Fatalf("failed to marshal alerts: %v", err)
	}
	outer := shared.OuterStruct{Metadata: string(alertsBytes)}
	body, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("failed to marshal outer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/alertmanager-signal", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, req)

	// Handler must return early after the first error; response body must be empty.
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty response body on signal error, got %q", rec.Body.String())
	}
}

// --- checkWorkflowAlive tests ---

func captureKlogOutput(t *testing.T, fn func()) string {
	t.Helper()

	initLoggingOnce.Do(func() {
		klog.InitFlags(nil)
	})

	logToStderr := flag.Lookup("logtostderr")
	if logToStderr == nil {
		t.Fatal("logtostderr flag is not initialized")
	}
	prevLogToStderr := logToStderr.Value.String()

	alsoLogToStderr := flag.Lookup("alsologtostderr")
	prevAlsoLogToStderr := ""
	if alsoLogToStderr != nil {
		prevAlsoLogToStderr = alsoLogToStderr.Value.String()
	}

	var buf bytes.Buffer
	if err := flag.Set("logtostderr", "false"); err != nil {
		t.Fatalf("failed to disable logtostderr: %v", err)
	}
	if alsoLogToStderr != nil {
		if err := flag.Set("alsologtostderr", "false"); err != nil {
			t.Fatalf("failed to disable alsologtostderr: %v", err)
		}
	}
	klog.SetOutput(&buf)

	defer func() {
		klog.Flush()
		klog.SetOutput(os.Stderr)
		_ = flag.Set("logtostderr", prevLogToStderr)
		if alsoLogToStderr != nil {
			_ = flag.Set("alsologtostderr", prevAlsoLogToStderr)
		}
	}()

	fn()
	klog.Flush()
	return buf.String()
}

// Test that no error is logged when the workflow is healthy (running).
func TestCheckWorkflowAlive_HealthyWorkflow(t *testing.T) {
	fc := &fakeClient{} // workflowClosed=false → running
	logs := captureKlogOutput(t, func() {
		checkWorkflowAlive(context.Background(), fc, "monitor-workflow-0")
	})

	if logs != "" {
		t.Fatalf("expected no log for healthy workflow, got: %s", logs)
	}
}

// Test that an ERROR log is emitted when DescribeWorkflowExecution shows the
// workflow is closed (CloseTime != nil), which covers terminate/cancel/fail.
func TestCheckWorkflowAlive_WorkflowClosed(t *testing.T) {
	fc := &fakeClient{workflowClosed: true}
	logs := captureKlogOutput(t, func() {
		checkWorkflowAlive(context.Background(), fc, "monitor-workflow-0")
	})

	if !strings.Contains(logs, "NOT running") {
		t.Fatalf("expected 'NOT running' log, got: %s", logs)
	}
}

// Test that an ERROR log is emitted when DescribeWorkflowExecution itself fails
// (e.g. network error or workflow truly not found).
func TestCheckWorkflowAlive_DescribeError(t *testing.T) {
	fc := &fakeClient{describeErr: errors.New("connection refused")}
	logs := captureKlogOutput(t, func() {
		checkWorkflowAlive(context.Background(), fc, "monitor-workflow-0")
	})

	if !strings.Contains(logs, "health check error") {
		t.Fatalf("expected 'health check error' log, got: %s", logs)
	}
}

// Test that NO log is emitted when ctx is cancelled (gateway shutting down).
func TestCheckWorkflowAlive_CtxCancelled_NoFalseAlert(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	fc := &fakeClient{describeErr: errors.New("context cancelled")}
	logs := captureKlogOutput(t, func() {
		checkWorkflowAlive(ctx, fc, "monitor-workflow-0")
	})

	if logs != "" {
		t.Fatalf("expected no log on ctx cancellation, got: %s", logs)
	}
}
