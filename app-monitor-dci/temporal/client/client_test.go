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
	"errors"
	"os"
	"testing"

	"go.temporal.io/sdk/client"
)

// fakeWorkflowRun implements client.WorkflowRun for testing.
type fakeWorkflowRun struct {
	id string
}

func (f *fakeWorkflowRun) GetID() string                                       { return f.id }
func (f *fakeWorkflowRun) GetRunID() string                                    { return "run-" + f.id }
func (f *fakeWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error { return nil }
func (f *fakeWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return nil
}

// fakeClient implements the temporalClient interface used by run.
type fakeClient struct {
	options          client.Options
	executeCalled    bool
	executeCalls     int
	executeErr       error
	lastStartOptions client.StartWorkflowOptions
	lastWorkflow     interface{}
}

func (f *fakeClient) ExecuteWorkflow(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	f.executeCalled = true
	f.executeCalls++
	f.lastStartOptions = opts
	f.lastWorkflow = workflow
	if f.executeErr != nil {
		return nil, f.executeErr
	}
	return &fakeWorkflowRun{id: "test-workflow-id"}, nil
}

func (f *fakeClient) Close() {}

// Test run with default environment variables (using fallback values) and successful workflow start.
func TestRun_SuccessWithDefaultEnv(t *testing.T) {
	// Save and restore globals and environment.
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

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

	run()

	if !fc.executeCalled {
		t.Fatalf("expected ExecuteWorkflow to be called")
	}

	// Ensure fallback address and namespace are used.
	if fc.options.HostPort != "temporal-stack-frontend.temporal.svc.cluster.local:7233" {
		t.Fatalf("unexpected HostPort: got %q", fc.options.HostPort)
	}
	if fc.options.Namespace != "default" {
		t.Fatalf("unexpected Namespace: got %q", fc.options.Namespace)
	}

	// Ensure workflow options and name are correct.
	if fc.lastStartOptions.ID != "monitor-workflow-2" {
		t.Fatalf("unexpected workflow ID: got %q", fc.lastStartOptions.ID)
	}
	if fc.lastStartOptions.TaskQueue != "monitor-task-queue" {
		t.Fatalf("unexpected TaskQueue: got %q", fc.lastStartOptions.TaskQueue)
	}
	if name, ok := fc.lastWorkflow.(string); !ok || name != "MonitorWorkflow" {
		t.Fatalf("unexpected workflow name: %#v", fc.lastWorkflow)
	}
}

// Test run when client.Dial (via temporalClientDial) fails.
func TestRun_DialError(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	called := 0
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		called++
		return nil, errors.New("dial error")
	}

	run() // should not panic even if dial fails

	if called != 1 {
		t.Fatalf("expected temporalClientDial to be called once, got %d", called)
	}
}

// Test run when ExecuteWorkflow returns an error.
func TestRun_ExecuteWorkflowError(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	fc := &fakeClient{executeErr: errors.New("execute error")}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	run() // should log error and return without panic

	if !fc.executeCalled {
		t.Fatalf("expected ExecuteWorkflow to be called")
	}
}

func TestRun_InvalidShardCountFallsBackToDefault(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	origAddr := os.Getenv("TEMPORAL_ADDRESS")
	origNS := os.Getenv("TEMPORAL_NAMESPACE")
	origShardCount := os.Getenv("SHARD_COUNT")
	defer func() {
		_ = os.Setenv("TEMPORAL_ADDRESS", origAddr)
		_ = os.Setenv("TEMPORAL_NAMESPACE", origNS)
		_ = os.Setenv("SHARD_COUNT", origShardCount)
	}()

	_ = os.Setenv("TEMPORAL_ADDRESS", "addr")
	_ = os.Setenv("TEMPORAL_NAMESPACE", "ns")
	_ = os.Setenv("SHARD_COUNT", "invalid")

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	run()

	if fc.executeCalls != 3 {
		t.Fatalf("expected default shard count 3, got %d workflow starts", fc.executeCalls)
	}
	if fc.lastStartOptions.ID != "monitor-workflow-2" {
		t.Fatalf("unexpected last workflow ID: got %q", fc.lastStartOptions.ID)
	}
}

func TestRun_AllShardStartsFail(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	origAddr := os.Getenv("TEMPORAL_ADDRESS")
	origNS := os.Getenv("TEMPORAL_NAMESPACE")
	origShardCount := os.Getenv("SHARD_COUNT")
	defer func() {
		_ = os.Setenv("TEMPORAL_ADDRESS", origAddr)
		_ = os.Setenv("TEMPORAL_NAMESPACE", origNS)
		_ = os.Setenv("SHARD_COUNT", origShardCount)
	}()

	_ = os.Setenv("TEMPORAL_ADDRESS", "addr")
	_ = os.Setenv("TEMPORAL_NAMESPACE", "ns")
	_ = os.Setenv("SHARD_COUNT", "2")

	fc := &fakeClient{executeErr: errors.New("execute error")}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		return fc, nil
	}

	run()

	if fc.executeCalls != 2 {
		t.Fatalf("expected 2 workflow start attempts, got %d", fc.executeCalls)
	}
	if fc.lastStartOptions.ID != "monitor-workflow-1" {
		t.Fatalf("unexpected last workflow ID: got %q", fc.lastStartOptions.ID)
	}
}
