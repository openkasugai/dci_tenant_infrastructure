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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	examplecomv1 "temporal/api/v1"
	"temporal/shared"
)

type activityUsages struct {
	CpuUtilization        string `json:"cpuUtilization"`
	GpuTotalUtilization   string `json:"gpuTotalUtilization"`
	GpuSMUtilization      string `json:"gpuSMUtilization"`
	GpuGraphicUtilization string `json:"gpuGraphicUtilization"`
	GpuMemoryUtilization  string `json:"gpuMemoryUtilization"`
}

type interceptingClient struct {
	ctrlclient.Client
	getFunc    func(context.Context, ctrlclient.ObjectKey, ctrlclient.Object, ...ctrlclient.GetOption) error
	createFunc func(context.Context, ctrlclient.Object, ...ctrlclient.CreateOption) error
	patchFunc  func(context.Context, ctrlclient.Object, ctrlclient.Patch, ...ctrlclient.PatchOption) error
}

func (c *interceptingClient) Get(ctx context.Context, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
	if c.getFunc != nil {
		return c.getFunc(ctx, key, obj, opts...)
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *interceptingClient) Create(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
	if c.createFunc != nil {
		return c.createFunc(ctx, obj, opts...)
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *interceptingClient) Patch(ctx context.Context, obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.PatchOption) error {
	if c.patchFunc != nil {
		return c.patchFunc(ctx, obj, patch, opts...)
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func newActivityTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()

	sch := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sch); err != nil {
		t.Fatalf("failed to add client-go scheme: %v", err)
	}
	if err := examplecomv1.AddToScheme(sch); err != nil {
		t.Fatalf("failed to add DCIAppState scheme: %v", err)
	}

	return sch
}

func newActivityTestClientWithObjects(t *testing.T, objects ...ctrlclient.Object) (ctrlclient.Client, *runtime.Scheme) {
	t.Helper()

	sch := newActivityTestScheme(t)
	return crfake.NewClientBuilder().WithScheme(sch).WithObjects(objects...).Build(), sch
}

func newActivityTestClient(t *testing.T, podState shared.PodUpdateStatus) (ctrlclient.Client, *runtime.Scheme) {
	t.Helper()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podState.Metadata.Name,
			Namespace: podState.Metadata.Namespace,
			UID:       "test-pod-uid",
		},
	}

	return newActivityTestClientWithObjects(t, pod)
}

func invokeVectorSignalActivity(t *testing.T, ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
	t.Helper()

	k8sClient, sch := newActivityTestClient(t, podState)
	return VectorSignalActivity(ctx, k8sClient, podState, sch)
}

func invokeAlertManagerSignalActivity(t *testing.T, ctx context.Context, podState shared.PodUpdateStatus, alertNamespace string, alertPod string) (*shared.PodUpdateStatus, error) {
	t.Helper()

	k8sClient, sch := newActivityTestClient(t, podState)
	return AlertManagerSignalActivity(ctx, k8sClient, podState, sch, alertNamespace, alertPod)
}

func mustDecodeUsages(t *testing.T, updated *shared.PodUpdateStatus) activityUsages {
	t.Helper()

	var usages activityUsages
	if err := json.Unmarshal(updated.Status.AppStatus.Usages, &usages); err != nil {
		t.Fatalf("failed to unmarshal usages: %v", err)
	}

	return usages
}

// helper to build a simple Prometheus success response
func buildPrometheusSuccessResponse(value string) []byte {
	resp := PrometheusResponse{
		Status: "success",
	}
	resp.Data.ResultType = "vector"
	resp.Data.Result = []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}{
		{
			Metric: map[string]string{"__name__": "dummy_metric"},
			Value:  []interface{}{"1234567890", value},
		},
	}

	b, _ := json.Marshal(resp)
	return b
}

// errorReadCloser is a helper that always returns a configured error from Read.
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (int, error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

// helper to create *string from string literal
func strPtr(s string) *string {
	return &s
}

// Test ExactPrometheusValue for normal successful case
func TestExactPrometheusValue_Success(t *testing.T) {
	body := buildPrometheusSuccessResponse("0.5")
	got, err := ExactPrometheusValue(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "0.5" {
		t.Fatalf("expected value 0.5, got %s", got)
	}
}

// Test ExactPrometheusValue when JSON is invalid
func TestExactPrometheusValue_UnmarshalError(t *testing.T) {
	body := []byte("not-json")
	if _, err := ExactPrometheusValue(body); err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
}

// Test ExactPrometheusValue when status is not success
func TestExactPrometheusValue_StatusError(t *testing.T) {
	resp := PrometheusResponse{Status: "error"}
	body, _ := json.Marshal(resp)
	if _, err := ExactPrometheusValue(body); err == nil {
		t.Fatalf("expected error for non-success status, got nil")
	}
}

// Test ExactPrometheusValue when result is empty
func TestExactPrometheusValue_EmptyResult(t *testing.T) {
	resp := PrometheusResponse{Status: "success"}
	body, _ := json.Marshal(resp)
	if _, err := ExactPrometheusValue(body); err == nil {
		t.Fatalf("expected error for empty result, got nil")
	}
}

// Test ExactPrometheusValue when value is not string
func TestExactPrometheusValue_NonStringValue(t *testing.T) {
	resp := PrometheusResponse{Status: "success"}
	resp.Data.Result = []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}{
		{
			Metric: map[string]string{"__name__": "dummy_metric"},
			Value:  []interface{}{"1234567890", 1.23},
		},
	}
	body, _ := json.Marshal(resp)
	if _, err := ExactPrometheusValue(body); err == nil {
		t.Fatalf("expected error for non-string value, got nil")
	}
}

// Test checkPodGPUFromPrometheus for GPU resource detection
func TestCheckPodGPUFromPrometheus_GPUDetected(t *testing.T) {
	// Create a test server that returns a response including gpu resource
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PrometheusResponse{Status: "success"}
		resp.Data.Result = []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}{
			{
				Metric: map[string]string{"resource": "nvidia_gpu"},
				Value:  []interface{}{"1234567890", "1"},
			},
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		return http.Get(srv.URL)
	}

	got, err := checkPodGPUFromPrometheus("ignored", "default", "test-pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !got {
		t.Fatalf("expected GPU to be detected, got false")
	}
}

// Test checkPodGPUFromPrometheus when Prometheus is unreachable
func TestCheckPodGPUFromPrometheus_ConnectionError(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, fmt.Errorf("connection error")
	}

	if _, err := checkPodGPUFromPrometheus("ignored", "default", "test-pod"); err == nil {
		t.Fatalf("expected connection error, got nil")
	}
}

// Test CheckPodGPURequest when ResourceClaim check returns true
func TestCheckPodGPURequest_ResourceClaimTrue(t *testing.T) {
	// Stub ResourceClaim check to return true without error
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return true, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("should not be called")
	}

	got, err := CheckPodGPURequest("http://example", "default", "pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !got {
		t.Fatalf("expected true when ResourceClaim reports GPU, got false")
	}
}

// Test CheckPodGPURequest when ResourceClaim check fails with an error
func TestCheckPodGPURequest_ResourceClaimReturnsError(t *testing.T) {
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	fallbackCalled := false
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, fmt.Errorf("some error")
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		fallbackCalled = true
		return true, nil
	}

	got, err := CheckPodGPURequest("http://example", "default", "pod")
	if err == nil {
		t.Fatalf("expected error from ResourceClaim, got nil")
	}
	if got {
		t.Fatalf("expected false when ResourceClaim check fails, got true")
	}
	if fallbackCalled {
		t.Fatalf("expected Prometheus fallback not to be called when ResourceClaim check returns error")
	}
	if !strings.Contains(err.Error(), "failed to check ResourceClaim") {
		t.Fatalf("expected wrapped ResourceClaim error, got %v", err)
	}
}

// Test CheckPodGPURequest when no GPU is allocated
func TestCheckPodGPURequest_NoGPU(t *testing.T) {
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	got, err := CheckPodGPURequest("http://example", "default", "pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got {
		t.Fatalf("expected false when no GPU is allocated, got true")
	}
}

// Test VectorSignalActivity for a successful path
func TestVectorSignalActivity_Success(t *testing.T) {
	// Stub all outbound Prometheus GETs and HTTP functions to return valid responses
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// Stub GPU detection to indicate GPU is present, so GPU metric branches run
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	// Configure environment variables used inside the activity.
	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	if err := os.Setenv("CPU_UTIL_RANGE", "5m"); err != nil {
		t.Fatalf("failed to set CPU_UTIL_RANGE: %v", err)
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	updated, err := invokeVectorSignalActivity(t, ctx, podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}

	if httpDoCalled {
		t.Fatalf("expected VectorSignalActivity not to send an ArgoEvents request")
	}
}

// Test AlertManagerSignalActivity for a successful path
func TestAlertManagerSignalActivity_Success(t *testing.T) {
	// Stub all outbound Prometheus GETs and HTTP functions to return valid responses
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// Configure environment variables used inside the activity
	if err := os.Setenv("PROMETHEUS_URL", "http://prometheus"); err != nil {
		t.Fatalf("failed to set PROMETHEUS_URL: %v", err)
	}
	if err := os.Setenv("CPU_UTIL_RANGE", "5m"); err != nil {
		t.Fatalf("failed to set CPU_UTIL_RANGE: %v", err)
	}

	// Simulate GPU detection succeeding via Prometheus fallback
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "test-pod-alert",
			Namespace: "alert-ns",
		},
	}

	ctx := context.Background()
	updated, err := invokeAlertManagerSignalActivity(t, ctx, podState, "alert-ns", "test-pod-alert")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}

	if httpDoCalled {
		t.Fatalf("expected AlertManagerSignalActivity not to send an ArgoEvents request")
	}
}

// Test AlertManagerSignalActivity when CPU Prometheus query fails and GPU is not allocated
func TestAlertManagerSignalActivity_CPUError_NoGPU(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	// Force CPU HTTP error
	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, fmt.Errorf("http error")
	}

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// No GPU allocated and no Prometheus GPU error
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-error-alert-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "cpu-error-alert-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
	if httpDoCalled {
		t.Fatalf("expected AlertManagerSignalActivity not to send an ArgoEvents request")
	}
}

// Test AlertManagerSignalActivity when CPU response body read fails
func TestAlertManagerSignalActivity_CPUReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	// CPU HTTP succeeds but body read fails
	httpGetFunc = func(url string) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errorReadCloser{err: fmt.Errorf("cpu read error")},
			Header:     make(http.Header),
		}
		return resp, nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// No GPU allocated
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-read-error-alert-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "cpu-read-error-alert-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test AlertManagerSignalActivity when CPU response body is too large
func TestAlertManagerSignalActivity_CPUResponseTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	// CPU response body >= maxResponseSize
	httpGetFunc = func(url string) (*http.Response, error) {
		b := make([]byte, int(maxResponseSize))
		for i := range b {
			b[i] = 'a'
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// No GPU allocated
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-large-alert-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "cpu-large-alert-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test AlertManagerSignalActivity when ExactPrometheusValue fails for CPU
func TestAlertManagerSignalActivity_CPUExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		resp := PrometheusResponse{Status: "error"}
		body, _ := json.Marshal(resp)
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// No GPU allocated
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-value-error-alert-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "cpu-value-error-alert-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU allocation check fails
func TestAlertManagerSignalActivity_GPUCheckError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	// CPU metrics succeed
	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.9")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// ResourceClaim and Prometheus GPU checks both fail
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, fmt.Errorf("rc error")
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("gpu check error")
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-error-alert-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "gpu-error-alert-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}
	usages := mustDecodeUsages(t, updated)
	if usages.GpuTotalUtilization != "N/A" || usages.GpuSMUtilization != "N/A" || usages.GpuGraphicUtilization != "N/A" || usages.GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected all GPU usages to be N/A, got %+v", usages)
	}
}

// Test AlertManagerSignalActivity when JSON marshaling of podState fails
func TestAlertManagerSignalActivity_MarshalError(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.5")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	jsonMarshalFunc = func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "marshal-error-alert-pod",
			Namespace: "alert-ns",
		},
	}

	if _, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "marshal-error-alert-pod"); err == nil {
		t.Fatalf("expected error when JSON marshaling fails, got nil")
	}
}

// Test AlertManagerSignalActivity when PROMETHEUS_URL is not set so that the default cluster URL is used,
// and CPU_UTIL_RANGE is also defaulted.
func TestAlertManagerSignalActivity_DefaultURLsAndCPUUtilRange(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	_ = os.Unsetenv("PROMETHEUS_URL")
	_ = os.Unsetenv("CPU_UTIL_RANGE")

	var usedCPUURL string
	httpGetFunc = func(rawURL string) (*http.Response, error) {
		usedCPUURL = rawURL
		body := buildPrometheusSuccessResponse("0.5")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// No GPU allocated
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-default-url-pod",
			Namespace: "alert-ns",
		},
	}

	if _, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-default-url-pod"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasPrefix(usedCPUURL, "http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090") {
		t.Fatalf("expected default Prometheus URL to be used, got: %s", usedCPUURL)
	}
	if httpDoCalled {
		t.Fatalf("expected AlertManagerSignalActivity not to send an ArgoEvents request")
	}
}

// Test AlertManagerSignalActivity when GPU Total query returns HTTP error
func TestAlertManagerSignalActivity_GPUTotalHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			return nil, fmt.Errorf("gpu total http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-total-http-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-total-http-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Total response body is too large
func TestAlertManagerSignalActivity_GPUTotalBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-total-body-large-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-total-body-large-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test AlertManagerSignalActivity when ExactPrometheusValue fails for GPU Total
func TestAlertManagerSignalActivity_GPUTotalExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-total-value-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-total-value-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Total response body read fails
func TestAlertManagerSignalActivity_GPUTotalReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu total read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-total-read-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-total-read-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU SM query returns HTTP error
func TestAlertManagerSignalActivity_GPUSMHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			return nil, fmt.Errorf("gpu sm http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-sm-http-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-sm-http-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU SM response body is too large
func TestAlertManagerSignalActivity_GPUSMBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-sm-body-large-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-sm-body-large-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test AlertManagerSignalActivity when ExactPrometheusValue fails for GPU SM
func TestAlertManagerSignalActivity_GPUSMExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-sm-value-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-sm-value-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU SM response body read fails
func TestAlertManagerSignalActivity_GPUSMReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu sm read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-sm-read-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-sm-read-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Graphic query returns HTTP error
func TestAlertManagerSignalActivity_GPUGraphicHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			return nil, fmt.Errorf("gpu graphic http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-graphic-http-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-graphic-http-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Graphic response body is too large
func TestAlertManagerSignalActivity_GPUGraphicBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-graphic-body-large-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-graphic-body-large-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test AlertManagerSignalActivity when ExactPrometheusValue fails for GPU Graphic
func TestAlertManagerSignalActivity_GPUGraphicExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-graphic-value-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-graphic-value-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Graphic response body read fails
func TestAlertManagerSignalActivity_GPUGraphicReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu graphic read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-graphic-read-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-graphic-read-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Memory query returns HTTP error
func TestAlertManagerSignalActivity_GPUMemoryHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			return nil, fmt.Errorf("gpu memory http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-memory-http-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-memory-http-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Memory response body is too large
func TestAlertManagerSignalActivity_GPUMemoryBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-memory-body-large-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-memory-body-large-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test AlertManagerSignalActivity when ExactPrometheusValue fails for GPU Memory
func TestAlertManagerSignalActivity_GPUMemoryExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-memory-value-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-memory-value-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test AlertManagerSignalActivity when GPU Memory response body read fails
func TestAlertManagerSignalActivity_GPUMemoryReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalGPU := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalGPU
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_DEV_FB_USED"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu memory read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "alert-gpu-memory-read-error-pod",
			Namespace: "alert-ns",
		},
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), podState, "alert-ns", "alert-gpu-memory-read-error-pod")
	if err != nil {
		t.Fatalf("expected no error from AlertManagerSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Additional sanity test for checkPodGPUFromPrometheus when status is not success in response
func TestCheckPodGPUFromPrometheus_StatusNotSuccess(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		resp := PrometheusResponse{Status: "error"}
		b, _ := json.Marshal(resp)
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	if _, err := checkPodGPUFromPrometheus("ignored", "ns", "pod"); err == nil {
		t.Fatalf("expected error when Prometheus status is not success, got nil")
	}
}

// Additional test for checkPodGPUFromPrometheus when response JSON is malformed
func TestCheckPodGPUFromPrometheus_InvalidJSON(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("not-json"))
		return recorder.Result(), nil
	}

	if _, err := checkPodGPUFromPrometheus("ignored", "ns", "pod"); err == nil {
		t.Fatalf("expected error for invalid Prometheus JSON, got nil")
	}
}

// Additional test for checkPodGPUFromPrometheus when no GPU is present
func TestCheckPodGPUFromPrometheus_NoGPU(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		resp := PrometheusResponse{Status: "success"}
		// Result does not contain any gpu-related resource
		resp.Data.Result = []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}{
			{
				Metric: map[string]string{"resource": "cpu"},
				Value:  []interface{}{"1234567890", "1"},
			},
		}
		b, _ := json.Marshal(resp)
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	got, err := checkPodGPUFromPrometheus("ignored", "ns", "pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got {
		t.Fatalf("expected false when no GPU resource is present, got true")
	}
}

// Test checkPodGPUFromPrometheus when the response body is too large
func TestCheckPodGPUFromPrometheus_BodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		b := make([]byte, int(maxResponseSize))
		for i := range b {
			b[i] = 'a'
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	if _, err := checkPodGPUFromPrometheus("ignored", "ns", "pod"); err == nil {
		t.Fatalf("expected error for too large Prometheus response body, got nil")
	}
}

// Test checkPodGPUFromPrometheus when reading the response body fails
func TestCheckPodGPUFromPrometheus_ReadError(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errorReadCloser{err: fmt.Errorf("read error")},
			Header:     make(http.Header),
		}
		return resp, nil
	}

	if _, err := checkPodGPUFromPrometheus("ignored", "ns", "pod"); err == nil {
		t.Fatalf("expected error when reading Prometheus response body fails, got nil")
	}
}

// Additional test for checkPodGPUFromPrometheus when MIG resources are present
func TestCheckPodGPUFromPrometheus_MIGDetected(t *testing.T) {
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		resp := PrometheusResponse{Status: "success"}
		resp.Data.Result = []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}{
			{
				Metric: map[string]string{"resource": "nvidia_mig"},
				Value:  []interface{}{"1234567890", "1"},
			},
		}
		b, _ := json.Marshal(resp)
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	got, err := checkPodGPUFromPrometheus("ignored", "ns", "pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !got {
		t.Fatalf("expected MIG to be detected as GPU, got false")
	}
}

// Test that Prometheus query URL encoding in VectorSignalActivity is well-formed
func TestVectorSignalActivity_PrometheusURL(t *testing.T) {
	var capturedQuery string

	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(rawURL string) (*http.Response, error) {
		u, _ := url.Parse(rawURL)
		capturedQuery = u.RawQuery
		body := buildPrometheusSuccessResponse("0.9")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) { return false, nil }

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "test-pod-url",
			Namespace: "ns-url",
		},
	}

	if _, err := invokeVectorSignalActivity(t, context.Background(), podState); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(capturedQuery, "query=") {
		t.Fatalf("expected Prometheus query parameter to be present, got: %s", capturedQuery)
	}
}

// Test VectorSignalActivity when PROMETHEUS_URL is not set so that the default cluster URL is used.
func TestVectorSignalActivity_DefaultURLs(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	_ = os.Unsetenv("PROMETHEUS_URL")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	var usedCPUURL string
	httpGetFunc = func(rawURL string) (*http.Response, error) {
		usedCPUURL = rawURL
		body := buildPrometheusSuccessResponse("0.9")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// Ensure GPU path is not taken
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "default-url-pod",
			Namespace: "default",
		},
	}

	if _, err := invokeVectorSignalActivity(t, context.Background(), podState); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasPrefix(usedCPUURL, "http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090") {
		t.Fatalf("expected default Prometheus URL to be used, got: %s", usedCPUURL)
	}
	if httpDoCalled {
		t.Fatalf("expected VectorSignalActivity not to send an ArgoEvents request")
	}
}

func TestVectorSignalActivity_DefaultCPUUtilRange(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Unsetenv("CPU_UTIL_RANGE")

	var usedCPUURL string
	httpGetFunc = func(rawURL string) (*http.Response, error) {
		usedCPUURL = rawURL
		body := buildPrometheusSuccessResponse("0.9")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "default-cpu-range-pod",
			Namespace: "default",
		},
	}

	if _, err := invokeVectorSignalActivity(t, context.Background(), podState); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parsedURL, err := url.Parse(usedCPUURL)
	if err != nil {
		t.Fatalf("failed to parse captured URL: %v", err)
	}

	decodedQuery := parsedURL.Query().Get("query")
	if !strings.Contains(decodedQuery, "[5m]") {
		t.Fatalf("expected default CPU_UTIL_RANGE [5m] in query, got: %s", decodedQuery)
	}
}

// Test VectorSignalActivity when Prometheus CPU query fails and GPU is not allocated
func TestVectorSignalActivity_CPUError_NoGPU(t *testing.T) {
	// Force Prometheus HTTP error via stub
	originalGet := httpGetFunc
	defer func() { httpGetFunc = originalGet }()

	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, fmt.Errorf("http error")
	}

	originalDo := httpDoFunc
	defer func() { httpDoFunc = originalDo }()

	httpDoCalled := false
	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		httpDoCalled = true
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	// No GPU allocated, and no Prometheus GPU query error
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}
	// CPU utilization should be N/A due to HTTP error
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
	if httpDoCalled {
		t.Fatalf("expected VectorSignalActivity not to send an ArgoEvents request")
	}
}

// Test VectorSignalActivity when ExactPrometheusValue fails for CPU metrics
func TestVectorSignalActivity_CPUExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	// Return a body with status=error so ExactPrometheusValue fails
	httpGetFunc = func(url string) (*http.Response, error) {
		resp := PrometheusResponse{Status: "error"}
		b, _ := json.Marshal(resp)
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-value-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test VectorSignalActivity when CPU response body is too large
func TestVectorSignalActivity_CPUResponseTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	// CPU response body >= maxResponseSize
	httpGetFunc = func(url string) (*http.Response, error) {
		b := make([]byte, int(maxResponseSize))
		for i := range b {
			b[i] = 'a'
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// Ensure GPU path is not taken
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-large-body-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test VectorSignalActivity when CPU response body read fails
func TestVectorSignalActivity_CPUReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	// CPU HTTP succeeds but body read fails
	httpGetFunc = func(url string) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errorReadCloser{err: fmt.Errorf("cpu read error")},
			Header:     make(http.Header),
		}
		return resp, nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	// Ensure GPU path is not taken
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "cpu-read-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).CpuUtilization != "N/A" {
		t.Fatalf("expected CpuUtilization to be N/A, got %s", mustDecodeUsages(t, updated).CpuUtilization)
	}
}

// Test VectorSignalActivity when JSON marshaling of podState fails
func TestVectorSignalActivity_MarshalError(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.5")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	jsonMarshalFunc = func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "marshal-error-pod",
			Namespace: "default",
		},
	}

	if _, err := invokeVectorSignalActivity(t, context.Background(), podState); err == nil {
		t.Fatalf("expected error when JSON marshaling fails, got nil")
	}
}

// Test VectorSignalActivity when GPU allocation check returns an error from Prometheus path
func TestVectorSignalActivity_GPUCheckErrorFromPrometheus(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.6")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, fmt.Errorf("rc error")
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("prometheus returned error status 'error' for GPU check")
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}
	usages := mustDecodeUsages(t, updated)
	if usages.GpuTotalUtilization != "N/A" || usages.GpuSMUtilization != "N/A" || usages.GpuGraphicUtilization != "N/A" || usages.GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected all GPU usages to be N/A, got %+v", usages)
	}
}

// Test VectorSignalActivity when GPU allocation check fails with an invalid-response error
// so that the "invalid response" logging branch is covered.
func TestVectorSignalActivity_GPUCheckErrorInvalidResponse(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		body := buildPrometheusSuccessResponse("0.6")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	// Force CheckPodGPURequest to return an error that does NOT contain
	// "prometheus returned error status" so that the error logging branch is used.
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) {
		return false, nil
	}
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("some other gpu check error")
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-invalid-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated PodUpdateStatus")
	}
	usages := mustDecodeUsages(t, updated)
	if usages.GpuTotalUtilization != "N/A" || usages.GpuSMUtilization != "N/A" || usages.GpuGraphicUtilization != "N/A" || usages.GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected all GPU usages to be N/A, got %+v", usages)
	}
}

// Test VectorSignalActivity when GPU Total query returns HTTP error
func TestVectorSignalActivity_GPUTotalHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			return nil, fmt.Errorf("gpu total http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-total-http-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test VectorSignalActivity when GPU Total response body is too large
func TestVectorSignalActivity_GPUTotalBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-total-body-large-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test VectorSignalActivity when GPU Total response body read fails
func TestVectorSignalActivity_GPUTotalReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu total read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-total-read-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test VectorSignalActivity when ExactPrometheusValue fails for GPU Total
func TestVectorSignalActivity_GPUTotalExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_GPU_UTIL") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-total-value-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuTotalUtilization)
	}
}

// Test VectorSignalActivity when GPU SM query returns HTTP error
func TestVectorSignalActivity_GPUSMHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			return nil, fmt.Errorf("gpu sm http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-sm-http-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test VectorSignalActivity when GPU SM response body is too large
func TestVectorSignalActivity_GPUSMBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-sm-body-large-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test VectorSignalActivity when ExactPrometheusValue fails for GPU SM
func TestVectorSignalActivity_GPUSMExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-sm-value-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test VectorSignalActivity when GPU SM response body read fails
func TestVectorSignalActivity_GPUSMReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_PROF_SM_ACTIVE"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu sm read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-sm-read-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuSMUtilization != "N/A" {
		t.Fatalf("expected GpuSMUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuSMUtilization)
	}
}

// Test VectorSignalActivity when GPU Graphic query returns HTTP error
func TestVectorSignalActivity_GPUGraphicHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			return nil, fmt.Errorf("gpu graphic http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-graphic-http-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test VectorSignalActivity when GPU Graphic response body is too large
func TestVectorSignalActivity_GPUGraphicBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-graphic-body-large-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test VectorSignalActivity when ExactPrometheusValue fails for GPU Graphic
func TestVectorSignalActivity_GPUGraphicExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-graphic-value-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test VectorSignalActivity when GPU Graphic response body read fails
func TestVectorSignalActivity_GPUGraphicReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_PROF_GR_ENGINE_ACTIVE"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu graphic read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-graphic-read-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuGraphicUtilization != "N/A" {
		t.Fatalf("expected GpuGraphicUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuGraphicUtilization)
	}
}

// Test VectorSignalActivity when GPU Memory query returns HTTP error
func TestVectorSignalActivity_GPUMemoryHTTPError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			return nil, fmt.Errorf("gpu memory http error")
		}
		body := buildPrometheusSuccessResponse("0.8")
		recorder := httptest.NewRecorder()
		recorder.Write(body)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-memory-http-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test VectorSignalActivity when GPU Memory response body is too large
func TestVectorSignalActivity_GPUMemoryBodyTooLarge(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			b = make([]byte, int(maxResponseSize))
			for i := range b {
				b[i] = 'a'
			}
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-memory-body-large-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test VectorSignalActivity when ExactPrometheusValue fails for GPU Memory
func TestVectorSignalActivity_GPUMemoryExactValueError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		var b []byte
		if strings.Contains(url, "DCGM_FI_DEV_FB_USED") {
			resp := PrometheusResponse{Status: "error"}
			b, _ = json.Marshal(resp)
		} else {
			b = buildPrometheusSuccessResponse("0.8")
		}
		recorder := httptest.NewRecorder()
		recorder.Write(b)
		return recorder.Result(), nil
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-memory-value-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test VectorSignalActivity when GPU Memory response body read fails
func TestVectorSignalActivity_GPUMemoryReadError(t *testing.T) {
	originalGet := httpGetFunc
	originalDo := httpDoFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		httpDoFunc = originalDo
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "container_cpu_usage_seconds_total"):
			body := buildPrometheusSuccessResponse("0.6")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		case strings.Contains(url, "DCGM_FI_DEV_FB_USED"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{err: fmt.Errorf("gpu memory read error")},
				Header:     make(http.Header),
			}, nil
		default:
			body := buildPrometheusSuccessResponse("0.3")
			recorder := httptest.NewRecorder()
			recorder.Write(body)
			return recorder.Result(), nil
		}
	}

	httpDoFunc = func(req *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write([]byte("ok"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal

	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return true, nil
	}

	_ = os.Setenv("PROMETHEUS_URL", "http://prometheus")
	_ = os.Setenv("CPU_UTIL_RANGE", "5m")

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Name:      "gpu-memory-read-error-pod",
			Namespace: "default",
		},
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error from VectorSignalActivity, got %v", err)
	}
	if mustDecodeUsages(t, updated).GpuMemoryUtilization != "N/A" {
		t.Fatalf("expected GpuMemoryUtilization to be N/A, got %s", mustDecodeUsages(t, updated).GpuMemoryUtilization)
	}
}

// Test allocation helpers for v1beta1, v1beta2, and v1 ResourceClaims
func TestCheckResourceClaimAllocationV1beta1(t *testing.T) {
	claimWithGPU := &resourcev1beta1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "gpu-claim"},
		Status: resourcev1beta1.ResourceClaimStatus{
			Allocation: &resourcev1beta1.AllocationResult{
				Devices: resourcev1beta1.DeviceAllocationResult{
					Results: []resourcev1beta1.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	if !checkResourceClaimAllocationV1beta1(claimWithGPU) {
		t.Fatalf("expected true for claimWithGPU")
	}

	claimNoGPU := &resourcev1beta1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "no-gpu-claim"},
		Status: resourcev1beta1.ResourceClaimStatus{
			Allocation: &resourcev1beta1.AllocationResult{
				Devices: resourcev1beta1.DeviceAllocationResult{
					Results: []resourcev1beta1.DeviceRequestAllocationResult{{Driver: "other.com"}},
				},
			},
		},
	}
	if checkResourceClaimAllocationV1beta1(claimNoGPU) {
		t.Fatalf("expected false for claimNoGPU")
	}
}

func TestCheckResourceClaimAllocationV1beta2(t *testing.T) {
	claimWithGPU := &resourcev1beta2.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "gpu-claim"},
		Status: resourcev1beta2.ResourceClaimStatus{
			Allocation: &resourcev1beta2.AllocationResult{
				Devices: resourcev1beta2.DeviceAllocationResult{
					Results: []resourcev1beta2.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	if !checkResourceClaimAllocationV1beta2(claimWithGPU) {
		t.Fatalf("expected true for claimWithGPU")
	}

	claimNoGPU := &resourcev1beta2.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "no-gpu-claim"},
		Status: resourcev1beta2.ResourceClaimStatus{
			Allocation: &resourcev1beta2.AllocationResult{
				Devices: resourcev1beta2.DeviceAllocationResult{
					Results: []resourcev1beta2.DeviceRequestAllocationResult{{Driver: "other.com"}},
				},
			},
		},
	}
	if checkResourceClaimAllocationV1beta2(claimNoGPU) {
		t.Fatalf("expected false for claimNoGPU")
	}
}

func TestCheckResourceClaimAllocationV1(t *testing.T) {
	claimWithGPU := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "gpu-claim"},
		Status: resourcev1.ResourceClaimStatus{
			Allocation: &resourcev1.AllocationResult{
				Devices: resourcev1.DeviceAllocationResult{
					Results: []resourcev1.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	if !checkResourceClaimAllocationV1(claimWithGPU) {
		t.Fatalf("expected true for claimWithGPU")
	}

	claimNoGPU := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "no-gpu-claim"},
		Status: resourcev1.ResourceClaimStatus{
			Allocation: &resourcev1.AllocationResult{
				Devices: resourcev1.DeviceAllocationResult{
					Results: []resourcev1.DeviceRequestAllocationResult{{Driver: "other.com"}},
				},
			},
		},
	}
	if checkResourceClaimAllocationV1(claimNoGPU) {
		t.Fatalf("expected false for claimNoGPU")
	}
}

// Tests for checkPodResourceClaim using injected fake Kubernetes clients

// Test when in-cluster config retrieval fails
func TestCheckPodResourceClaim_InClusterConfigError(t *testing.T) {
	origCfg := inClusterConfigFunc
	defer func() { inClusterConfigFunc = origCfg }()

	inClusterConfigFunc = func() (*rest.Config, error) {
		return nil, fmt.Errorf("config error")
	}

	hasGPU, err := checkPodResourceClaim("ns", "pod")
	if err == nil {
		t.Fatalf("expected error from checkPodResourceClaim, got nil")
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false on config error, got true")
	}
}

// Test when creating Kubernetes clientset fails
func TestCheckPodResourceClaim_NewForConfigError(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return nil, fmt.Errorf("newForConfig error")
	}

	hasGPU, err := checkPodResourceClaim("ns", "pod")
	if err == nil {
		t.Fatalf("expected error from checkPodResourceClaim, got nil")
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false on clientset error, got true")
	}
}

// Test when Pod retrieval fails from Kubernetes API
func TestCheckPodResourceClaim_PodGetError(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}
	// Fake clientset with no Pods; Get will return an error
	client := fake.NewSimpleClientset()
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return client, nil
	}

	hasGPU, err := checkPodResourceClaim("ns", "missing-pod")
	if err == nil {
		t.Fatalf("expected error when Pod is missing, got nil")
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when Pod is missing, got true")
	}
}

// Test when Pod has no ResourceClaims
func TestCheckPodResourceClaim_NoResourceClaims(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-no-rc",
			Namespace: "ns",
		},
		Spec: corev1.PodSpec{},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return client, nil
	}

	hasGPU, err := checkPodResourceClaim("ns", "pod-no-rc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when no ResourceClaims, got true")
	}
}

// Test when Pod has ResourceClaims but no resolved ResourceClaimName in status
func TestCheckPodResourceClaim_UnresolvedResourceClaim(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-rc-unresolved",
			Namespace: "ns",
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			// No matching ResourceClaimStatuses with ResourceClaimName set
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{},
		},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return client, nil
	}

	hasGPU, err := checkPodResourceClaim("ns", "pod-rc-unresolved")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when ResourceClaim is unresolved, got true")
	}
}

// Test when Pod has a v1beta1 ResourceClaim with GPU (nvidia.com driver)
func TestCheckPodResourceClaim_V1beta1_GPUFound(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Setenv("RESOURCE_API_VERSION", "v1beta1")

	ns := "ns-v1beta1"
	podName := "pod-v1beta1"
	claimName := "gpu-claim-v1beta1"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr(claimName),
			}},
		},
	}
	claim := &resourcev1beta1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: ns,
		},
		Status: resourcev1beta1.ResourceClaimStatus{
			Allocation: &resourcev1beta1.AllocationResult{
				Devices: resourcev1beta1.DeviceAllocationResult{
					Results: []resourcev1beta1.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	client := fake.NewSimpleClientset(pod, claim)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hasGPU {
		t.Fatalf("expected hasGPU to be true for v1beta1 GPU claim")
	}
}

// Test when Pod has a v1beta2 ResourceClaim with GPU
func TestCheckPodResourceClaim_V1beta2_GPUFound(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Setenv("RESOURCE_API_VERSION", "v1beta2")

	ns := "ns-v1beta2"
	podName := "pod-v1beta2"
	claimName := "gpu-claim-v1beta2"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr(claimName),
			}},
		},
	}
	claim := &resourcev1beta2.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: ns,
		},
		Status: resourcev1beta2.ResourceClaimStatus{
			Allocation: &resourcev1beta2.AllocationResult{
				Devices: resourcev1beta2.DeviceAllocationResult{
					Results: []resourcev1beta2.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	client := fake.NewSimpleClientset(pod, claim)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hasGPU {
		t.Fatalf("expected hasGPU to be true for v1beta2 GPU claim")
	}
}

// Test when Pod has a v1 ResourceClaim with GPU
func TestCheckPodResourceClaim_V1_GPUFound(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Unsetenv("RESOURCE_API_VERSION") // default to v1

	ns := "ns-v1"
	podName := "pod-v1"
	claimName := "gpu-claim-v1"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr(claimName),
			}},
		},
	}
	claim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: ns,
		},
		Status: resourcev1.ResourceClaimStatus{
			Allocation: &resourcev1.AllocationResult{
				Devices: resourcev1.DeviceAllocationResult{
					Results: []resourcev1.DeviceRequestAllocationResult{{Driver: "gpu.nvidia.com"}},
				},
			},
		},
	}
	client := fake.NewSimpleClientset(pod, claim)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hasGPU {
		t.Fatalf("expected hasGPU to be true for v1 GPU claim")
	}
}

// Test error path of ResourceClaim Get for each API version
func TestCheckPodResourceClaim_V1beta1_ClaimGetError(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Setenv("RESOURCE_API_VERSION", "v1beta1")

	ns := "ns-v1beta1-err"
	podName := "pod-v1beta1-err"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr("missing-claim"),
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error even if ResourceClaim Get fails, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when ResourceClaim Get fails, got true")
	}
}

func TestCheckPodResourceClaim_V1beta2_ClaimGetError(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Setenv("RESOURCE_API_VERSION", "v1beta2")

	ns := "ns-v1beta2-err"
	podName := "pod-v1beta2-err"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr("missing-claim"),
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error even if ResourceClaim Get fails, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when ResourceClaim Get fails, got true")
	}
}

func TestCheckPodResourceClaim_V1_ClaimGetError(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Unsetenv("RESOURCE_API_VERSION") // default to v1

	ns := "ns-v1-err"
	podName := "pod-v1-err"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr("missing-claim"),
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim(ns, podName)
	if err != nil {
		t.Fatalf("expected no error even if ResourceClaim Get fails, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false when ResourceClaim Get fails, got true")
	}
}

func TestPodStateCrateOrPatch_MarshalAppStatusError(t *testing.T) {
	originalMarshal := jsonMarshalFunc
	defer func() { jsonMarshalFunc = originalMarshal }()

	jsonMarshalFunc = func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal appstatus error")
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "pod-marshal-error", Namespace: "default"},
	}

	client, sch := newActivityTestClient(t, podState)
	err := PodStateCrateOrPatch(context.Background(), client, podState, sch)
	if err == nil || !strings.Contains(err.Error(), "marshal appstatus error") {
		t.Fatalf("expected marshal appstatus error, got %v", err)
	}
}

func TestPodStateCrateOrPatch_DCIAppStateGetUnexpectedError(t *testing.T) {
	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "pod-get-error", Namespace: "default"},
	}

	baseClient, sch := newActivityTestClient(t, podState)
	client := &interceptingClient{
		Client: baseClient,
		getFunc: func(ctx context.Context, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
			if _, ok := obj.(*examplecomv1.DCIAppState); ok {
				return fmt.Errorf("dci get error")
			}
			return baseClient.Get(ctx, key, obj, opts...)
		},
	}

	err := PodStateCrateOrPatch(context.Background(), client, podState, sch)
	if err == nil || !strings.Contains(err.Error(), "dci get error") {
		t.Fatalf("expected DCIAppState get error, got %v", err)
	}
}

func TestPodStateCrateOrPatch_PodGetErrorOnCreatePath(t *testing.T) {
	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "missing-pod", Namespace: "default"},
	}

	client, sch := newActivityTestClientWithObjects(t)
	err := PodStateCrateOrPatch(context.Background(), client, podState, sch)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected pod get error on create path, got %v", err)
	}
}

func TestPodStateCrateOrPatch_CreateError(t *testing.T) {
	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "create-error-pod", Namespace: "default"},
	}

	baseClient, sch := newActivityTestClient(t, podState)
	client := &interceptingClient{
		Client: baseClient,
		createFunc: func(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
			return fmt.Errorf("create error")
		},
	}

	err := PodStateCrateOrPatch(context.Background(), client, podState, sch)
	if err == nil || !strings.Contains(err.Error(), "create error") {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestPodStateCrateOrPatch_SetControllerReferenceError(t *testing.T) {
	originalSetControllerReference := ctrlSetControllerReferenceFunc
	defer func() { ctrlSetControllerReferenceFunc = originalSetControllerReference }()

	ctrlSetControllerReferenceFunc = func(owner, object metav1.Object, scheme *runtime.Scheme, opts ...controllerutil.OwnerReferenceOption) error {
		return fmt.Errorf("set controller reference error")
	}

	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "set-controller-ref-pod", Namespace: "default"},
	}

	client, sch := newActivityTestClient(t, podState)
	if err := PodStateCrateOrPatch(context.Background(), client, podState, sch); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	created := &examplecomv1.DCIAppState{}
	if err := client.Get(context.Background(), ctrlclient.ObjectKey{Name: podState.Metadata.Name, Namespace: podState.Metadata.Namespace}, created); err != nil {
		t.Fatalf("expected created DCIAppState, got %v", err)
	}
}

func TestPodStateCrateOrPatch_PatchSuccess(t *testing.T) {
	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "patch-success-pod", Namespace: "default"},
		Spec:     shared.PodSpec{Type: "computation"},
		Status: shared.PodStatus{
			AppStatus: shared.AppStatus{Usages: json.RawMessage(`{"cpuUtilization":"0.7"}`)},
		},
	}

	existing := &examplecomv1.DCIAppState{
		ObjectMeta: metav1.ObjectMeta{Name: podState.Metadata.Name, Namespace: podState.Metadata.Namespace},
		Spec:       examplecomv1.DCIAppStateSpec{Type: "connection"},
	}

	client, sch := newActivityTestClientWithObjects(t, existing)
	if err := PodStateCrateOrPatch(context.Background(), client, podState, sch); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated := &examplecomv1.DCIAppState{}
	if err := client.Get(context.Background(), ctrlclient.ObjectKey{Name: podState.Metadata.Name, Namespace: podState.Metadata.Namespace}, updated); err != nil {
		t.Fatalf("failed to fetch patched DCIAppState: %v", err)
	}
	if !strings.Contains(string(updated.Status.AppStatus.Raw), "cpuUtilization") {
		t.Fatalf("expected patched app status to include cpuUtilization, got %s", string(updated.Status.AppStatus.Raw))
	}
}

func TestPodStateCrateOrPatch_PatchError(t *testing.T) {
	podState := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{Name: "patch-error-pod", Namespace: "default"},
	}

	existing := &examplecomv1.DCIAppState{
		ObjectMeta: metav1.ObjectMeta{Name: podState.Metadata.Name, Namespace: podState.Metadata.Namespace},
	}

	baseClient, sch := newActivityTestClientWithObjects(t, existing)
	client := &interceptingClient{
		Client: baseClient,
		patchFunc: func(ctx context.Context, obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.PatchOption) error {
			return fmt.Errorf("patch error")
		},
	}

	err := PodStateCrateOrPatch(context.Background(), client, podState, sch)
	if err == nil || !strings.Contains(err.Error(), "patch error") {
		t.Fatalf("expected patch error, got %v", err)
	}
}

func TestVectorSignalActivity_PodStateCreateOrPatchError(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write(buildPrometheusSuccessResponse("0.5"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) { return false, nil }

	podState := shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "vector-create-or-patch-error", Namespace: "default"}}
	baseClient, sch := newActivityTestClient(t, podState)
	client := &interceptingClient{
		Client: baseClient,
		createFunc: func(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
			return fmt.Errorf("create error")
		},
	}

	if _, err := VectorSignalActivity(context.Background(), client, podState, sch); err == nil || !strings.Contains(err.Error(), "failed to create or patch podStatus") {
		t.Fatalf("expected create or patch error, got %v", err)
	}
}

func TestAlertManagerSignalActivity_PodStateCreateOrPatchError(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write(buildPrometheusSuccessResponse("0.5"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) { return false, nil }

	podState := shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "alert-create-or-patch-error", Namespace: "default"}}
	baseClient, sch := newActivityTestClient(t, podState)
	client := &interceptingClient{
		Client: baseClient,
		createFunc: func(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
			return fmt.Errorf("create error")
		},
	}

	if _, err := AlertManagerSignalActivity(context.Background(), client, podState, sch, podState.Metadata.Namespace, podState.Metadata.Name); err == nil || !strings.Contains(err.Error(), "failed to create or patch podStatus") {
		t.Fatalf("expected create or patch error, got %v", err)
	}
}

func TestVectorSignalActivity_GPUCheckTemporaryIssue(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write(buildPrometheusSuccessResponse("0.6"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("prometheus returned error status 'error' for GPU check")
	}

	updated, err := invokeVectorSignalActivity(t, context.Background(), shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "vector-temp-issue", Namespace: "default"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	usages := mustDecodeUsages(t, updated)
	if usages.GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", usages.GpuTotalUtilization)
	}
}

func TestAlertManagerSignalActivity_GPUCheckTemporaryIssue(t *testing.T) {
	originalGet := httpGetFunc
	originalMarshal := jsonMarshalFunc
	originalRC := checkPodResourceClaimFunc
	originalProm := checkPodGPUFromPrometheusFunc
	defer func() {
		httpGetFunc = originalGet
		jsonMarshalFunc = originalMarshal
		checkPodResourceClaimFunc = originalRC
		checkPodGPUFromPrometheusFunc = originalProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write(buildPrometheusSuccessResponse("0.6"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) {
		return false, fmt.Errorf("prometheus returned error status 'error' for GPU check")
	}

	updated, err := invokeAlertManagerSignalActivity(t, context.Background(), shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "alert-temp-issue", Namespace: "default"}}, "default", "alert-temp-issue")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	usages := mustDecodeUsages(t, updated)
	if usages.GpuTotalUtilization != "N/A" {
		t.Fatalf("expected GpuTotalUtilization to be N/A, got %s", usages.GpuTotalUtilization)
	}
}

func TestCheckPodResourceClaim_UnsupportedAPIVersion(t *testing.T) {
	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	origVersion := os.Getenv("RESOURCE_API_VERSION")
	defer func() {
		inClusterConfigFunc = origCfg
		newForConfigFunc = origNew
		_ = os.Setenv("RESOURCE_API_VERSION", origVersion)
	}()

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	_ = os.Setenv("RESOURCE_API_VERSION", "invalid-version")

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-invalid-api", Namespace: "ns"},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "rc1"}},
		},
		Status: corev1.PodStatus{
			ResourceClaimStatuses: []corev1.PodResourceClaimStatus{{
				Name:              "rc1",
				ResourceClaimName: strPtr("claim1"),
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return client, nil }

	hasGPU, err := checkPodResourceClaim("ns", "pod-invalid-api")
	if err == nil || !strings.Contains(err.Error(), "unsupported RESOURCE_API_VERSION") {
		t.Fatalf("expected unsupported RESOURCE_API_VERSION error, got %v", err)
	}
	if hasGPU {
		t.Fatalf("expected hasGPU to be false, got true")
	}
}
