package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
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
	executeCalls     int
	lastStartOptions client.StartWorkflowOptions
	lastWorkflow     interface{}
	// errFn returns the error for a given 1-based call number; nil means success.
	// When errFn itself is nil, every call succeeds.
	errFn func(call int) error
}

func (f *fakeClient) ExecuteWorkflow(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	f.executeCalls++
	f.lastStartOptions = opts
	f.lastWorkflow = workflow
	if f.errFn != nil {
		if err := f.errFn(f.executeCalls); err != nil {
			return nil, err
		}
	}
	return &fakeWorkflowRun{id: opts.ID}, nil
}

func (f *fakeClient) Close() {}

// alwaysErr returns an errFn that yields err on every call.
func alwaysErr(err error) func(int) error {
	return func(int) error { return err }
}

// failFirst returns an errFn that yields err for the first n calls, then succeeds.
func failFirst(n int, err error) func(int) error {
	return func(call int) error {
		if call <= n {
			return err
		}
		return nil
	}
}

// withEnv sets env vars for the duration of the test and restores them after.
func withEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		orig, had := os.LookupEnv(k)
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
		k := k
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(k, orig)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// run() integration-style tests
// -----------------------------------------------------------------------------

func TestRun_SuccessWithDefaultEnv(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	withEnv(t, map[string]string{
		"TEMPORAL_ADDRESS":   "",
		"TEMPORAL_NAMESPACE": "",
		"SHARD_COUNT":        "",
		"POD_NAME":           "", // keep Events disabled (logs only)
		"POD_NAMESPACE":      "",
	})

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		fc.options = opts
		return fc, nil
	}

	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if fc.executeCalls != 3 {
		t.Fatalf("expected 3 workflow starts (default shard count), got %d", fc.executeCalls)
	}

	// Fallback address and namespace.
	if fc.options.HostPort != "temporal-stack-frontend.temporal.svc.cluster.local:7233" {
		t.Fatalf("unexpected HostPort: got %q", fc.options.HostPort)
	}
	if fc.options.Namespace != "default" {
		t.Fatalf("unexpected Namespace: got %q", fc.options.Namespace)
	}

	// Workflow options / name / idempotency policy.
	if fc.lastStartOptions.ID != "monitor-workflow-2" {
		t.Fatalf("unexpected workflow ID: got %q", fc.lastStartOptions.ID)
	}
	if fc.lastStartOptions.TaskQueue != "monitor-task-queue" {
		t.Fatalf("unexpected TaskQueue: got %q", fc.lastStartOptions.TaskQueue)
	}
	if fc.lastStartOptions.WorkflowIDReusePolicy != enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE {
		t.Fatalf("expected ALLOW_DUPLICATE reuse policy, got %v", fc.lastStartOptions.WorkflowIDReusePolicy)
	}
	if name, ok := fc.lastWorkflow.(string); !ok || name != "MonitorWorkflow" {
		t.Fatalf("unexpected workflow name: %#v", fc.lastWorkflow)
	}
}

func TestRun_DialError(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	called := 0
	temporalClientDial = func(opts client.Options) (temporalClient, error) {
		called++
		return nil, errors.New("dial error")
	}

	if code := run(); code != 1 {
		t.Fatalf("expected exit code 1 on dial error, got %d", code)
	}
	if called != 1 {
		t.Fatalf("expected temporalClientDial to be called once, got %d", called)
	}
}

func TestRun_PermanentErrorReturnsFailure(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	withEnv(t, map[string]string{
		"TEMPORAL_ADDRESS":   "addr",
		"TEMPORAL_NAMESPACE": "ns",
		"SHARD_COUNT":        "2",
		"POD_NAME":           "",
		"POD_NAMESPACE":      "",
	})

	// A plain (non-serviceerror) error is treated as permanent: one attempt per shard.
	fc := &fakeClient{errFn: alwaysErr(errors.New("boom"))}
	temporalClientDial = func(opts client.Options) (temporalClient, error) { return fc, nil }

	if code := run(); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if fc.executeCalls != 2 {
		t.Fatalf("expected 2 attempts (one per shard, no retry), got %d", fc.executeCalls)
	}
	if fc.lastStartOptions.ID != "monitor-workflow-1" {
		t.Fatalf("unexpected last workflow ID: got %q", fc.lastStartOptions.ID)
	}
}

func TestRun_InvalidShardCountFallsBackToDefault(t *testing.T) {
	origDial := temporalClientDial
	defer func() { temporalClientDial = origDial }()

	withEnv(t, map[string]string{
		"TEMPORAL_ADDRESS":   "addr",
		"TEMPORAL_NAMESPACE": "ns",
		"SHARD_COUNT":        "invalid",
		"POD_NAME":           "",
		"POD_NAMESPACE":      "",
	})

	fc := &fakeClient{}
	temporalClientDial = func(opts client.Options) (temporalClient, error) { return fc, nil }

	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if fc.executeCalls != 3 {
		t.Fatalf("expected default shard count 3, got %d workflow starts", fc.executeCalls)
	}
}

// -----------------------------------------------------------------------------
// ensureWorkflowRunning unit tests
// -----------------------------------------------------------------------------

func TestEnsureWorkflowRunning_Success(t *testing.T) {
	fc := &fakeClient{}
	if err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fc.executeCalls != 1 {
		t.Fatalf("expected 1 call, got %d", fc.executeCalls)
	}
	if fc.lastStartOptions.WorkflowIDReusePolicy != enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE {
		t.Fatalf("expected ALLOW_DUPLICATE reuse policy")
	}
}

func TestEnsureWorkflowRunning_AlreadyStartedIsSuccess(t *testing.T) {
	alreadyStarted := serviceerror.NewWorkflowExecutionAlreadyStarted("already", "", "")
	fc := &fakeClient{errFn: alwaysErr(alreadyStarted)}

	if err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 5); err != nil {
		t.Fatalf("AlreadyStarted should be treated as success, got: %v", err)
	}
	if fc.executeCalls != 1 {
		t.Fatalf("expected 1 call (no retry on AlreadyStarted), got %d", fc.executeCalls)
	}
}

func TestEnsureWorkflowRunning_PermanentErrorNoRetry(t *testing.T) {
	fc := &fakeClient{errFn: alwaysErr(serviceerror.NewInvalidArgument("bad"))}

	err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 5)
	if err == nil {
		t.Fatal("expected error for permanent failure")
	}
	if fc.executeCalls != 1 {
		t.Fatalf("expected 1 call (no retry for permanent error), got %d", fc.executeCalls)
	}
}

func TestEnsureWorkflowRunning_RetryThenSuccess(t *testing.T) {
	orig := initialBackoff
	initialBackoff = time.Millisecond
	defer func() { initialBackoff = orig }()

	// Fail twice with a transient error, then succeed on the third call.
	fc := &fakeClient{errFn: failFirst(2, serviceerror.NewUnavailable("temporarily down"))}

	if err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 5); err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if fc.executeCalls != 3 {
		t.Fatalf("expected 3 calls (2 failures + 1 success), got %d", fc.executeCalls)
	}
}

func TestEnsureWorkflowRunning_ExhaustsRetries(t *testing.T) {
	orig := initialBackoff
	initialBackoff = time.Millisecond
	defer func() { initialBackoff = orig }()

	fc := &fakeClient{errFn: alwaysErr(serviceerror.NewUnavailable("still down"))}

	err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 3)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if fc.executeCalls != 3 {
		t.Fatalf("expected 3 attempts, got %d", fc.executeCalls)
	}
}

func TestEnsureWorkflowRunning_ContextCanceledDuringBackoff(t *testing.T) {
	orig := initialBackoff
	initialBackoff = time.Hour // ensure we block in backoff until ctx cancels
	defer func() { initialBackoff = orig }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled

	fc := &fakeClient{errFn: alwaysErr(serviceerror.NewUnavailable("down"))}

	err := ensureWorkflowRunning(ctx, fc, "monitor-workflow-0", 5)
	if err == nil {
		t.Fatal("expected error when context is canceled")
	}
	// First attempt happens, then backoff select hits ctx.Done().
	if fc.executeCalls != 1 {
		t.Fatalf("expected 1 call before context cancellation, got %d", fc.executeCalls)
	}
}

// -----------------------------------------------------------------------------
// isRetryable
// -----------------------------------------------------------------------------

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"unavailable", serviceerror.NewUnavailable("x"), true},
		{"deadline", serviceerror.NewDeadlineExceeded("x"), true},
		{"resource-exhausted", serviceerror.NewResourceExhausted(enumspb.RESOURCE_EXHAUSTED_CAUSE_SYSTEM_OVERLOADED, "x"), true},
		{"invalid-argument", serviceerror.NewInvalidArgument("x"), false},
		{"already-started", serviceerror.NewWorkflowExecutionAlreadyStarted("x", "", ""), false},
		{"plain", errors.New("x"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryable(tc.err); got != tc.want {
				t.Fatalf("isRetryable(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// config loaders
// -----------------------------------------------------------------------------

func TestLoadPositiveInt(t *testing.T) {
	const key = "TEST_LOAD_POSITIVE_INT"
	cases := []struct {
		name string
		val  string
		set  bool
		want int
	}{
		{"unset", "", false, 7},
		{"empty", "", true, 7},
		{"valid", "12", true, 12},
		{"zero", "0", true, 7},
		{"negative", "-3", true, 7},
		{"non-numeric", "abc", true, 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				_ = os.Setenv(key, tc.val)
			} else {
				_ = os.Unsetenv(key)
			}
			t.Cleanup(func() { _ = os.Unsetenv(key) })
			if got := loadPositiveInt(key, 7); got != tc.want {
				t.Fatalf("loadPositiveInt(%q) = %d, want %d", tc.val, got, tc.want)
			}
		})
	}
}

func TestLoadStartTimeout(t *testing.T) {
	cases := []struct {
		name string
		val  string
		set  bool
		want time.Duration
	}{
		{"unset", "", false, defaultStartTimeout},
		{"valid", "30s", true, 30 * time.Second},
		{"invalid", "notaduration", true, defaultStartTimeout},
		{"zero", "0s", true, defaultStartTimeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				_ = os.Setenv("START_TIMEOUT", tc.val)
			} else {
				_ = os.Unsetenv("START_TIMEOUT")
			}
			t.Cleanup(func() { _ = os.Unsetenv("START_TIMEOUT") })
			if got := loadStartTimeout(); got != tc.want {
				t.Fatalf("loadStartTimeout() = %s, want %s", got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// eventEmitter
// -----------------------------------------------------------------------------

func TestNewEventEmitter_DisabledWithoutDownwardAPI(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "", "POD_NAMESPACE": ""})
	if em := newEventEmitter(context.Background()); em != nil {
		t.Fatalf("expected nil emitter without POD_NAME/POD_NAMESPACE, got %#v", em)
	}
}

func TestNewEventEmitter_DisabledWhenInClusterConfigFails(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "pod-0", "POD_NAMESPACE": "temporal"})

	origCfg := inClusterConfigFunc
	defer func() { inClusterConfigFunc = origCfg }()
	inClusterConfigFunc = func() (*rest.Config, error) { return nil, errors.New("not in cluster") }

	if em := newEventEmitter(context.Background()); em != nil {
		t.Fatalf("expected nil emitter when in-cluster config fails, got %#v", em)
	}
}

func TestNewEventEmitter_ResolvesJobAndEmits(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "temporal-client-abc", "POD_NAMESPACE": "temporal"})

	origCfg := inClusterConfigFunc
	origNew := newForConfigFunc
	defer func() { inClusterConfigFunc = origCfg; newForConfigFunc = origNew }()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "temporal-client-abc",
			Namespace: "temporal",
			UID:       types.UID("pod-uid"),
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Job", APIVersion: "batch/v1", Name: "temporal-client", UID: types.UID("job-uid")},
			},
		},
	}
	cs := fake.NewSimpleClientset(pod)

	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) { return cs, nil }

	em := newEventEmitter(context.Background())
	if em == nil {
		t.Fatal("expected non-nil emitter")
	}

	if em.ref == nil || em.ref.Kind != "Job" || em.ref.Name != "temporal-client" || em.ref.UID != types.UID("job-uid") {
		t.Fatalf("expected Job reference, got %#v", em.ref)
	}

	// Should synchronously create an Event attributed to the Job.
	em.event(corev1.EventTypeWarning, "WorkflowStartFailed", "shard %s failed", "monitor-workflow-0")

	evs, err := cs.CoreV1().Events("temporal").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing events: %v", err)
	}
	if len(evs.Items) != 1 {
		t.Fatalf("expected 1 event created, got %d", len(evs.Items))
	}
	ev := evs.Items[0]
	if ev.InvolvedObject.Kind != "Job" || ev.InvolvedObject.Name != "temporal-client" {
		t.Fatalf("event not attributed to Job: %#v", ev.InvolvedObject)
	}
	if ev.Reason != "WorkflowStartFailed" || ev.Type != corev1.EventTypeWarning {
		t.Fatalf("unexpected event reason/type: %q/%q", ev.Reason, ev.Type)
	}
	if ev.Source.Component != "temporal-client" {
		t.Fatalf("unexpected event source component: %q", ev.Source.Component)
	}
}

func TestResolveJobReference_FallbackToPod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "standalone-pod",
			Namespace: "temporal",
			UID:       types.UID("pod-uid"),
		},
	}
	cs := fake.NewSimpleClientset(pod)

	ref, err := resolveJobReference(context.Background(), cs, "temporal", "standalone-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Kind != "Pod" || ref.Name != "standalone-pod" || ref.UID != types.UID("pod-uid") {
		t.Fatalf("expected Pod fallback reference, got %#v", ref)
	}
}

func TestResolveJobReference_PodNotFound(t *testing.T) {
	cs := fake.NewSimpleClientset()
	if _, err := resolveJobReference(context.Background(), cs, "temporal", "missing"); err == nil {
		t.Fatal("expected error when pod is not found")
	}
}

func TestEventEmitter_NilSafe(t *testing.T) {
	var em *eventEmitter // nil
	// Must not panic on a nil receiver.
	em.event(corev1.EventTypeWarning, "Reason", "message %d", 1)
}

func TestEnsureWorkflowRunning_BackoffCappedAtMax(t *testing.T) {
	origInit, origMax := initialBackoff, maxBackoff
	initialBackoff, maxBackoff = time.Nanosecond, time.Nanosecond
	defer func() { initialBackoff, maxBackoff = origInit, origMax }()

	fc := &fakeClient{errFn: alwaysErr(serviceerror.NewUnavailable("down"))}
	if err := ensureWorkflowRunning(context.Background(), fc, "monitor-workflow-0", 3); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if fc.executeCalls != 3 {
		t.Fatalf("expected 3 attempts, got %d", fc.executeCalls)
	}
}

func TestNewEventEmitter_DisabledWhenClientInitFails(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "pod-0", "POD_NAMESPACE": "temporal"})

	origCfg, origNew := inClusterConfigFunc, newForConfigFunc
	defer func() { inClusterConfigFunc, newForConfigFunc = origCfg, origNew }()
	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	newForConfigFunc = func(*rest.Config) (kubernetes.Interface, error) {
		return nil, errors.New("client init failed")
	}

	if em := newEventEmitter(context.Background()); em != nil {
		t.Fatalf("expected nil emitter when client init fails, got %#v", em)
	}
}

func TestNewEventEmitter_DisabledWhenJobResolutionFails(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "missing-pod", "POD_NAMESPACE": "temporal"})

	origCfg, origNew := inClusterConfigFunc, newForConfigFunc
	defer func() { inClusterConfigFunc, newForConfigFunc = origCfg, origNew }()
	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	newForConfigFunc = func(*rest.Config) (kubernetes.Interface, error) {
		return fake.NewSimpleClientset(), nil // no pod -> resolveJobReference errors
	}

	if em := newEventEmitter(context.Background()); em != nil {
		t.Fatalf("expected nil emitter when Job resolution fails, got %#v", em)
	}
}

func TestEventEmitter_EventCreateErrorFallsBackToLog(t *testing.T) {
	withEnv(t, map[string]string{"POD_NAME": "temporal-client-abc", "POD_NAMESPACE": "temporal"})

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "temporal-client-abc",
			Namespace: "temporal",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Job", APIVersion: "batch/v1", Name: "temporal-client", UID: types.UID("job-uid")},
			},
		},
	}
	cs := fake.NewSimpleClientset(pod)
	cs.PrependReactor("create", "events", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("api unavailable")
	})

	origCfg, origNew := inClusterConfigFunc, newForConfigFunc
	defer func() { inClusterConfigFunc, newForConfigFunc = origCfg, origNew }()
	inClusterConfigFunc = func() (*rest.Config, error) { return &rest.Config{}, nil }
	newForConfigFunc = func(*rest.Config) (kubernetes.Interface, error) { return cs, nil }

	em := newEventEmitter(context.Background())
	if em == nil {
		t.Fatal("expected non-nil emitter")
	}
	// Must not panic even though Event creation fails.
	em.event(corev1.EventTypeWarning, "WorkflowStartFailed", "shard failed")
}
