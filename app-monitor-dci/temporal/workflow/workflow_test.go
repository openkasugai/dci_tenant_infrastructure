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
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"temporal/shared"
)

func registerWorkflowTestActivities(
	env *testsuite.TestWorkflowEnvironment,
	vector func(context.Context, shared.PodUpdateStatus) (*shared.PodUpdateStatus, error),
	alert func(context.Context, shared.PodUpdateStatus, string, string) (*shared.PodUpdateStatus, error),
) {
	env.RegisterActivityWithOptions(vector, activity.RegisterOptions{Name: "VectorSignalActivity"})
	env.RegisterActivityWithOptions(alert, activity.RegisterOptions{Name: "AlertManagerSignalActivity"})
}

func cancelWorkflowAfter(env *testsuite.TestWorkflowEnvironment, delay time.Duration) {
	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, delay)
}

func requireWorkflowCanceled(t *testing.T, env *testsuite.TestWorkflowEnvironment) {
	t.Helper()

	if !env.IsWorkflowCompleted() {
		t.Fatalf("workflow did not complete")
	}

	err := env.GetWorkflowError()
	if err == nil {
		t.Fatalf("expected workflow cancellation error")
	}
	if !temporal.IsCanceledError(err) {
		t.Fatalf("expected canceled workflow error, got %v", err)
	}
}

func TestLoadHistoryEventThreshold_InvalidValueFallsBackToDefault(t *testing.T) {
	orig := os.Getenv("HISTORY_EVENT_THRESHOLD")
	defer func() {
		if orig == "" {
			_ = os.Unsetenv("HISTORY_EVENT_THRESHOLD")
			return
		}
		_ = os.Setenv("HISTORY_EVENT_THRESHOLD", orig)
	}()

	if err := os.Setenv("HISTORY_EVENT_THRESHOLD", "not-a-number"); err != nil {
		t.Fatalf("failed to set HISTORY_EVENT_THRESHOLD: %v", err)
	}

	if threshold := loadHistoryEventThreshold(); threshold != DefaultHistoryEventThreshold {
		t.Fatalf("expected default threshold %d, got %d", DefaultHistoryEventThreshold, threshold)
	}
}

func TestLoadHistoryEventThreshold_ValidValueParses(t *testing.T) {
	orig := os.Getenv("HISTORY_EVENT_THRESHOLD")
	defer func() {
		if orig == "" {
			_ = os.Unsetenv("HISTORY_EVENT_THRESHOLD")
			return
		}
		_ = os.Setenv("HISTORY_EVENT_THRESHOLD", orig)
	}()

	if err := os.Setenv("HISTORY_EVENT_THRESHOLD", "1234"); err != nil {
		t.Fatalf("failed to set HISTORY_EVENT_THRESHOLD: %v", err)
	}

	if threshold := loadHistoryEventThreshold(); threshold != 1234 {
		t.Fatalf("expected parsed threshold 1234, got %d", threshold)
	}
}

func TestMonitorWorkflow_ContinueAsNewWhenHistoryThresholdReached(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	origKeepAliveInterval := keepAliveInterval
	origCurrentHistoryLengthFn := currentHistoryLengthFn
	defer func() {
		HistoryEventThreshold = origThreshold
		keepAliveInterval = origKeepAliveInterval
		currentHistoryLengthFn = origCurrentHistoryLengthFn
	}()

	HistoryEventThreshold = 6
	keepAliveInterval = 0
	historyChecks := 0
	currentHistoryLengthFn = func(ctx workflow.Context) int {
		historyChecks++
		if historyChecks >= 2 {
			return HistoryEventThreshold
		}
		return 0
	}

	vectorCalls := 0
	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			vectorCalls++
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	update := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Namespace: "ns1",
			Name:      "pod1",
		},
	}
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(VectorSignalName, update)
	}, time.Second)
	env.ExecuteWorkflow(MonitorWorkflow, (*MonitorWorkflowState)(nil))

	if !env.IsWorkflowCompleted() {
		t.Fatalf("workflow did not complete")
	}

	err := env.GetWorkflowError()
	if err == nil {
		t.Fatalf("expected continue-as-new error")
	}
	if !strings.Contains(err.Error(), "continue as new") {
		t.Fatalf("expected continue-as-new error, got %v", err)
	}
	if !strings.Contains(err.Error(), "MonitorWorkflow") {
		t.Fatalf("expected continue-as-new target workflow in error, got %v", err)
	}
	if vectorCalls == 0 {
		t.Fatalf("expected VectorSignalActivity to be called before continue-as-new")
	}
}

// Test MonitorWorkflow handling of vector signal before workflow cancellation.
func TestMonitorWorkflow_VectorSignalAndCancel(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	defer func() {
		HistoryEventThreshold = origThreshold
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold

	// Stub activities so no real external calls are made.
	vectorCalls := 0
	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			vectorCalls++
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	// Send a single vector signal shortly after the workflow starts.
	update := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Namespace: "ns1",
			Name:      "pod1",
		},
	}
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(VectorSignalName, update)
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, (*MonitorWorkflowState)(nil))

	requireWorkflowCanceled(t, env)

	if vectorCalls == 0 {
		t.Fatalf("expected VectorSignalActivity to be called at least once")
	}
}

// Test MonitorWorkflow handling of alertmanager signal with existing state before workflow cancellation.
func TestMonitorWorkflow_AlertManagerSignalAndCancel(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	defer func() {
		HistoryEventThreshold = origThreshold
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold

	// Stub activities.
	alertCalls := 0
	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			alertCalls++
			if ns != podState.Metadata.Namespace || pod != podState.Metadata.Name {
				t.Fatalf("unexpected namespace/pod: got %s/%s, want %s/%s", ns, pod, podState.Metadata.Namespace, podState.Metadata.Name)
			}
			return &podState, nil
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	// Pre-populate workflow state so alert branch finds a matching pod.
	initialState := &MonitorWorkflowState{
		State: map[string]shared.PodUpdateStatus{
			"ns1/pod1": {
				Metadata: shared.PodMetaData{
					Namespace: "ns1",
					Name:      "pod1",
				},
			},
		},
	}

	// Build alert that matches the existing pod in state.
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

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(AlertmanagerSignalName, alerts)
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, initialState)

	requireWorkflowCanceled(t, env)

	if alertCalls == 0 {
		t.Fatalf("expected AlertManagerSignalActivity to be called at least once")
	}
}

// Test VectorSignalActivity error path within MonitorWorkflow.
func TestMonitorWorkflow_VectorActivityError(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	defer func() {
		HistoryEventThreshold = origThreshold
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold

	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return nil, fmt.Errorf("vector activity error")
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	update := shared.PodUpdateStatus{
		Metadata: shared.PodMetaData{
			Namespace: "ns1",
			Name:      "pod1",
		},
	}

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(VectorSignalName, update)
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, (*MonitorWorkflowState)(nil))

	requireWorkflowCanceled(t, env)
}

// Test AlertManagerSignalActivity error path within MonitorWorkflow.
func TestMonitorWorkflow_AlertActivityError(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	defer func() {
		HistoryEventThreshold = origThreshold
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold

	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return nil, fmt.Errorf("alert activity error")
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	initialState := &MonitorWorkflowState{
		State: map[string]shared.PodUpdateStatus{
			"ns1/pod1": {
				Metadata: shared.PodMetaData{
					Namespace: "ns1",
					Name:      "pod1",
				},
			},
		},
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

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(AlertmanagerSignalName, alerts)
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, initialState)

	requireWorkflowCanceled(t, env)
}

// Test alert path when pod is not present in state.
func TestMonitorWorkflow_AlertPodNotInState(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	defer func() {
		HistoryEventThreshold = origThreshold
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold

	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			t.Fatalf("AlertManagerSignalActivity should not be called when pod is not in state")
			return &podState, nil
		},
	)

	env.RegisterWorkflow(MonitorWorkflow)

	initialState := &MonitorWorkflowState{
		State: map[string]shared.PodUpdateStatus{},
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

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(AlertmanagerSignalName, alerts)
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, initialState)

	requireWorkflowCanceled(t, env)
}

// Test SetQueryHandler error path: workflow returns the error immediately.
func TestMonitorWorkflow_SetQueryHandlerError(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origSetQueryHandlerFn := setQueryHandlerFn
	defer func() { setQueryHandlerFn = origSetQueryHandlerFn }()

	setQueryHandlerFn = func(ctx workflow.Context, queryType string, handler interface{}) error {
		return fmt.Errorf("setQueryHandler failed")
	}

	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
	)
	env.RegisterWorkflow(MonitorWorkflow)
	env.ExecuteWorkflow(MonitorWorkflow, (*MonitorWorkflowState)(nil))

	if !env.IsWorkflowCompleted() {
		t.Fatalf("workflow did not complete")
	}
	err := env.GetWorkflowError()
	if err == nil {
		t.Fatalf("expected error from SetQueryHandler")
	}
	if !strings.Contains(err.Error(), "setQueryHandler failed") {
		t.Fatalf("expected setQueryHandler error, got %v", err)
	}
}

// Test is-draining query handler returns false before threshold is reached.
func TestMonitorWorkflow_IsDrainingQuery(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	origThreshold := HistoryEventThreshold
	origKeepAliveInterval := keepAliveInterval
	defer func() {
		HistoryEventThreshold = origThreshold
		keepAliveInterval = origKeepAliveInterval
	}()

	HistoryEventThreshold = DefaultHistoryEventThreshold
	keepAliveInterval = 0

	registerWorkflowTestActivities(env,
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
		func(ctx context.Context, podState shared.PodUpdateStatus, ns, pod string) (*shared.PodUpdateStatus, error) {
			return &podState, nil
		},
	)
	env.RegisterWorkflow(MonitorWorkflow)

	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("is-draining")
		if err != nil {
			t.Errorf("is-draining query failed: %v", err)
			return
		}
		var isDraining bool
		if err := val.Get(&isDraining); err != nil {
			t.Errorf("failed to decode is-draining response: %v", err)
			return
		}
		if isDraining {
			t.Errorf("expected isDraining=false before threshold, got true")
		}
	}, time.Second)
	cancelWorkflowAfter(env, 2*time.Second)

	env.ExecuteWorkflow(MonitorWorkflow, (*MonitorWorkflowState)(nil))
	requireWorkflowCanceled(t, env)
}
