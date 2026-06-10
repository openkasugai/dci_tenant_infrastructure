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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"temporal/shared"
)

// fakeClient implements temporalClient for testing.
type fakeClient struct {
	options client.Options
	closed  bool
}

func (f *fakeClient) Close() { f.closed = true }

type bareTemporalClient struct{}

func (c *bareTemporalClient) Close() {}

// fakeWorker implements temporalWorker for testing.
type fakeWorker struct {
	registeredWorkflows  []interface{}
	registeredActivities []interface{}
	runErr               error
	runCalled            int
}

func (w *fakeWorker) RegisterWorkflow(fn interface{}) {
	w.registeredWorkflows = append(w.registeredWorkflows, fn)
}

func (w *fakeWorker) RegisterActivity(fn interface{}) {
	w.registeredActivities = append(w.registeredActivities, fn)
}

func (w *fakeWorker) RegisterActivityWithOptions(fn interface{}, _ activity.RegisterOptions) {
	w.RegisterActivity(fn)
}

func (w *fakeWorker) Run(ch <-chan interface{}) error {
	w.runCalled++
	return w.runErr
}

// Test run with default environment variables (fallback values) and successful worker run.
func TestRun_SuccessWithDefaultEnv(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
	}()

	// Save and clear env so defaults are used.
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

	fw := &fakeWorker{}
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		if c != fc {
			t.Fatalf("unexpected client passed to workerNew")
		}
		if taskQueue != "monitor-task-queue" {
			t.Fatalf("unexpected taskQueue: %q", taskQueue)
		}
		return fw
	}

	run()

	// Verify Dial options use defaults.
	if fc.options.HostPort != "temporal-stack-frontend.temporal.svc.cluster.local:7233" {
		t.Fatalf("unexpected HostPort: %q", fc.options.HostPort)
	}
	if fc.options.Namespace != "default" {
		t.Fatalf("unexpected Namespace: %q", fc.options.Namespace)
	}

	// Verify worker registration and run were invoked.
	if fw.runCalled != 1 {
		t.Fatalf("expected worker.Run to be called once, got %d", fw.runCalled)
	}
	if len(fw.registeredWorkflows) == 0 {
		t.Fatalf("expected at least one workflow to be registered")
	}
	if len(fw.registeredActivities) < 2 {
		t.Fatalf("expected at least two activities to be registered")
	}
}

// Test run when Temporal client dial fails.
func TestRun_DialError(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
	}()

	called := 0
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		called++
		return nil, errors.New("dial error")
	}

	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		t.Fatalf("workerNew should not be called when dial fails")
		return nil
	}

	run()

	if called != 1 {
		t.Fatalf("expected temporalClientDial to be called once, got %d", called)
	}
}

// Test run when worker.Run returns a graceful stop error (context canceled/worker stopped).
func TestRun_WorkerRunGracefulStop(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
	}()

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	fw := &fakeWorker{runErr: errors.New("worker stopped")}
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		return fw
	}

	run()

	if fw.runCalled != 1 {
		t.Fatalf("expected worker.Run to be called once, got %d", fw.runCalled)
	}
}

// Test run when worker.Run returns an unexpected error.
func TestRun_WorkerRunUnexpectedError(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	origGetConfigOrDie := getConfigOrDie
	origK8sClientNew := k8sClientNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
		getConfigOrDie = origGetConfigOrDie
		k8sClientNew = origK8sClientNew
	}()

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	getConfigOrDie = func() *rest.Config {
		return &rest.Config{}
	}
	client, _ := newActivityTestClient(t, shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "worker-run-pod", Namespace: "default"}})
	k8sClientNew = func(cfg *rest.Config, options ctrlclient.Options) (ctrlclient.Client, error) {
		return client, nil
	}

	fw := &fakeWorker{runErr: errors.New("some other error")}
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		return fw
	}

	run()

	if fw.runCalled != 1 {
		t.Fatalf("expected worker.Run to be called once, got %d", fw.runCalled)
	}
}

func TestRun_K8sClientNewError(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	origGetConfigOrDie := getConfigOrDie
	origK8sClientNew := k8sClientNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
		getConfigOrDie = origGetConfigOrDie
		k8sClientNew = origK8sClientNew
	}()

	getConfigOrDie = func() *rest.Config {
		return &rest.Config{}
	}

	dialCalled := false
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		dialCalled = true
		return &fakeClient{}, nil
	}

	workerNewCalled := false
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		workerNewCalled = true
		return &fakeWorker{}
	}

	k8sClientNew = func(cfg *rest.Config, options ctrlclient.Options) (ctrlclient.Client, error) {
		return nil, errors.New("k8s client create error")
	}

	run()

	if dialCalled {
		t.Fatalf("expected temporalClientDial not to be called when k8sclient.New fails")
	}
	if workerNewCalled {
		t.Fatalf("expected workerNew not to be called when k8sclient.New fails")
	}
}

func TestWorkerNew_PanicsWhenClientDoesNotImplementSDKClient(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatalf("expected workerNew to panic for non-SDK client")
		} else if !strings.Contains(recovered.(string), "temporal client does not implement client.Client") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	workerNew(&bareTemporalClient{}, "monitor-task-queue", worker.Options{})
}

func TestActivities_VectorSignalActivityWrapper(t *testing.T) {
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

	podState := shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "wrapper-vector-pod", Namespace: "default"}}
	client, sch := newActivityTestClient(t, podState)
	acts := &Activities{K8sClient: client, Scheme: sch}

	updated, err := acts.VectorSignalActivity(context.Background(), podState)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated pod state")
	}
}

func TestActivities_AlertManagerSignalActivityWrapper(t *testing.T) {
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

	podState := shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "wrapper-alert-pod", Namespace: "default"}}
	client, sch := newActivityTestClient(t, podState)
	acts := &Activities{K8sClient: client, Scheme: sch}

	updated, err := acts.AlertManagerSignalActivity(context.Background(), podState, "default", "wrapper-alert-pod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected non-nil updated pod state")
	}
}

func TestMain_DelegatesToRun(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	origGetConfigOrDie := getConfigOrDie
	origK8sClientNew := k8sClientNew
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
		getConfigOrDie = origGetConfigOrDie
		k8sClientNew = origK8sClientNew
	}()

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}
	getConfigOrDie = func() *rest.Config {
		return &rest.Config{}
	}
	client, _ := newActivityTestClient(t, shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "main-pod", Namespace: "default"}})
	k8sClientNew = func(cfg *rest.Config, options ctrlclient.Options) (ctrlclient.Client, error) {
		return client, nil
	}

	fw := &fakeWorker{}
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		return fw
	}

	main()

	if fw.runCalled != 1 {
		t.Fatalf("expected worker.Run to be called once from main, got %d", fw.runCalled)
	}
}

func TestRun_RegisteredActivityClosuresExecute(t *testing.T) {
	origDial := temporalClientDial
	origWorkerNew := workerNew
	origGetConfigOrDie := getConfigOrDie
	origK8sClientNew := k8sClientNew
	origGet := httpGetFunc
	origMarshal := jsonMarshalFunc
	origRC := checkPodResourceClaimFunc
	origProm := checkPodGPUFromPrometheusFunc
	defer func() {
		temporalClientDial = origDial
		workerNew = origWorkerNew
		getConfigOrDie = origGetConfigOrDie
		k8sClientNew = origK8sClientNew
		httpGetFunc = origGet
		jsonMarshalFunc = origMarshal
		checkPodResourceClaimFunc = origRC
		checkPodGPUFromPrometheusFunc = origProm
	}()

	httpGetFunc = func(url string) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		recorder.Write(buildPrometheusSuccessResponse("0.5"))
		return recorder.Result(), nil
	}
	jsonMarshalFunc = json.Marshal
	checkPodResourceClaimFunc = func(namespace, podName string) (bool, error) { return false, nil }
	checkPodGPUFromPrometheusFunc = func(promURL, namespace, podName string) (bool, error) { return false, nil }

	getConfigOrDie = func() *rest.Config {
		return &rest.Config{}
	}
	podState := shared.PodUpdateStatus{Metadata: shared.PodMetaData{Name: "registered-activity-pod", Namespace: "default"}}
	k8sClient, _ := newActivityTestClient(t, podState)
	k8sClientNew = func(cfg *rest.Config, options ctrlclient.Options) (ctrlclient.Client, error) {
		return k8sClient, nil
	}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return &fakeClient{}, nil
	}

	fw := &fakeWorker{}
	workerNew = func(c temporalClient, taskQueue string, opts worker.Options) temporalWorker {
		return fw
	}

	run()

	if len(fw.registeredActivities) < 2 {
		t.Fatalf("expected two registered activities, got %d", len(fw.registeredActivities))
	}

	vectorClosure, ok := fw.registeredActivities[0].(func(context.Context, shared.PodUpdateStatus) (*shared.PodUpdateStatus, error))
	if !ok {
		t.Fatalf("expected first registered activity to be a vector closure")
	}
	updated, err := vectorClosure(context.Background(), podState)
	if err != nil {
		t.Fatalf("expected vector closure to succeed, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected vector closure to return pod state")
	}

	alertClosure, ok := fw.registeredActivities[1].(func(context.Context, shared.PodUpdateStatus, string, string) (*shared.PodUpdateStatus, error))
	if !ok {
		t.Fatalf("expected second registered activity to be an alert closure")
	}
	updated, err = alertClosure(context.Background(), podState, "default", "registered-activity-pod")
	if err != nil {
		t.Fatalf("expected alert closure to succeed, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected alert closure to return pod state")
	}
}
