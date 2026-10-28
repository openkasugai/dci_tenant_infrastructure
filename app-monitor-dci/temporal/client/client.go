package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	taskQueue    = "monitor-task-queue"
	workflowType = "MonitorWorkflow"

	defaultShardCount       = 3
	defaultMaxStartAttempts = 5
	defaultStartTimeout     = 2 * time.Minute
)

// temporalClient is a narrow interface used by run for easier testing.
type temporalClient interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow any, args ...any) (client.WorkflowRun, error)
	Close()
}

// temporalClientDial is an indirection for client.Dial so tests can replace it.
var temporalClientDial = func(options client.Options) (temporalClient, error) {
	return client.Dial(options)
}

// Indirections for the Kubernetes client used to emit Events (replaceable in tests).
var (
	inClusterConfigFunc = rest.InClusterConfig
	newForConfigFunc    = func(cfg *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(cfg)
	}
)

// initialBackoff is the first retry backoff interval in ensureWorkflowRunning.
// Exposed as a variable so tests can shorten it.
var initialBackoff = time.Second

// maxBackoff caps the retry backoff; a variable so tests can shorten it.
var maxBackoff = 30 * time.Second

var initLoggingOnce sync.Once

func initLogging() {
	initLoggingOnce.Do(func() {
		klog.InitFlags(nil)
		flag.Parse()
	})
}

func main() {
	os.Exit(run())
}

// loadPositiveInt reads a strictly positive integer from the environment,
// falling back to def when unset or invalid.
func loadPositiveInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		klog.Warningf("%s is not set, using default: %d", key, def)
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		klog.Warningf("Invalid %s value %q, using default: %d", key, val, def)
		return def
	}
	klog.Infof("%s=%d", key, n)
	return n
}

// loadStartTimeout reads the overall start timeout (e.g. "120s") from the environment.
func loadStartTimeout() time.Duration {
	val := os.Getenv("START_TIMEOUT")
	if val == "" {
		klog.Warningf("START_TIMEOUT is not set, using default: %s", defaultStartTimeout)
		return defaultStartTimeout
	}
	d, err := time.ParseDuration(val)
	if err != nil || d <= 0 {
		klog.Warningf("Invalid START_TIMEOUT value %q, using default: %s", val, defaultStartTimeout)
		return defaultStartTimeout
	}
	klog.Infof("START_TIMEOUT=%s", d)
	return d
}

// isRetryable reports whether the error from ExecuteWorkflow is transient and
// worth retrying (server temporarily unavailable, overloaded, or timed out).
func isRetryable(err error) bool {
	var (
		unavailable *serviceerror.Unavailable
		deadline    *serviceerror.DeadlineExceeded
		exhausted   *serviceerror.ResourceExhausted
	)
	return errors.As(err, &unavailable) ||
		errors.As(err, &deadline) ||
		errors.As(err, &exhausted)
}

// ensureWorkflowRunning idempotently ensures the workflow identified by
// workflowID is running. Idempotency is achieved without the (SDK-version
// dependent) conflict policy: an already-running workflow surfaces as a
// WorkflowExecutionAlreadyStarted error which is treated as success, while
// WorkflowIDReusePolicy=ALLOW_DUPLICATE lets a closed workflow be restarted.
// Transient failures are retried with exponential backoff up to maxAttempts.
func ensureWorkflowRunning(ctx context.Context, c temporalClient, workflowID string, maxAttempts int) error {
	backoff := initialBackoff
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
			ID:                    workflowID,
			TaskQueue:             taskQueue,
			WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		}, workflowType)
		if err == nil {
			klog.Infof("Workflow ensured running: %s (runID=%s)", we.GetID(), we.GetRunID())
			return nil
		}
		lastErr = err

		// An already-started workflow means it is running: treat as success.
		var alreadyStarted *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &alreadyStarted) {
			klog.Infof("Workflow %s already started, treating as success", workflowID)
			return nil
		}

		if !isRetryable(err) {
			return fmt.Errorf("permanent error starting %s: %w", workflowID, err)
		}

		klog.Warningf("Transient error starting workflow %s (attempt %d/%d): %v", workflowID, attempt, maxAttempts, err)
		if attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled while retrying %s: %w", workflowID, ctx.Err())
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return fmt.Errorf("exhausted %d attempt(s) starting %s: %w", maxAttempts, workflowID, lastErr)
}

func run() int {
	initLogging()
	flag.Parse()
	defer klog.Flush()

	shardCount := loadPositiveInt("SHARD_COUNT", defaultShardCount)
	maxAttempts := loadPositiveInt("MAX_START_ATTEMPTS", defaultMaxStartAttempts)
	startTimeout := loadStartTimeout()

	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		klog.Warning("TEMPORAL_ADDRESS environment variable is not set. Temporal client will use the default address: temporal-stack-frontend.temporal.svc.cluster.local:7233.")
		temporalAddress = "temporal-stack-frontend.temporal.svc.cluster.local:7233"
	}

	temporalNamespace := os.Getenv("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		klog.Warning("TEMPORAL_NAMESPACE environment variable is not set. Temporal client will use the default namespace: default.")
		temporalNamespace = "default"
	}

	// Connect to Temporal server
	c, err := temporalClientDial(client.Options{
		HostPort:  temporalAddress,
		Namespace: temporalNamespace,
	})
	if err != nil {
		klog.Errorf("unable to create Temporal client: %v", err)
		return 1
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	emitter := newEventEmitter(ctx)

	klog.Infof("Ensuring %d workflow shard(s) are running", shardCount)

	var failed []string
	for i := 0; i < shardCount; i++ {
		workflowID := fmt.Sprintf("monitor-workflow-%d", i)
		if err := ensureWorkflowRunning(ctx, c, workflowID, maxAttempts); err != nil {
			klog.Errorf("Failed to ensure workflow %s: %v", workflowID, err)
			emitter.event(corev1.EventTypeWarning, "WorkflowStartFailed",
				"failed to ensure workflow %s running after up to %d attempt(s): %v", workflowID, maxAttempts, err)
			failed = append(failed, workflowID)
		}
	}

	if len(failed) > 0 {
		emitter.event(corev1.EventTypeWarning, "WorkflowStartIncomplete",
			"%d/%d workflow shard(s) failed to start: %v", len(failed), shardCount, failed)
		klog.Errorf("%d/%d workflow shard(s) failed to start: %v", len(failed), shardCount, failed)
		return 1
	}

	emitter.event(corev1.EventTypeNormal, "WorkflowStarted",
		"all %d workflow shard(s) ensured running", shardCount)
	klog.Infof("Successfully ensured all %d workflow shard(s) running", shardCount)
	return 0
}

// eventEmitter emits Kubernetes Events attributed to the owning Job so that
// failures are visible via `kubectl describe job`. Events are created
// synchronously so they are guaranteed to be persisted before the short-lived
// Job process exits. It is nil-safe: when the Kubernetes client cannot be
// initialised, all methods fall back to logs.
type eventEmitter struct {
	events typedcorev1.EventInterface
	ref    *corev1.ObjectReference
	source corev1.EventSource
}

// newEventEmitter builds an emitter targeting the Job that owns this Pod.
// Returns nil (logs-only mode) when running outside a cluster or when the
// downward-API POD_NAME/POD_NAMESPACE variables are not provided.
func newEventEmitter(ctx context.Context) *eventEmitter {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podName == "" || podNamespace == "" {
		klog.Warning("POD_NAME/POD_NAMESPACE not set; Kubernetes Events disabled, using logs only")
		return nil
	}

	cfg, err := inClusterConfigFunc()
	if err != nil {
		klog.Warningf("in-cluster config unavailable; Kubernetes Events disabled: %v", err)
		return nil
	}
	clientset, err := newForConfigFunc(cfg)
	if err != nil {
		klog.Warningf("Kubernetes client init failed; Kubernetes Events disabled: %v", err)
		return nil
	}

	ref, err := resolveJobReference(ctx, clientset, podNamespace, podName)
	if err != nil {
		klog.Warningf("failed to resolve owning Job; Kubernetes Events disabled: %v", err)
		return nil
	}

	return &eventEmitter{
		events: clientset.CoreV1().Events(podNamespace),
		ref:    ref,
		source: corev1.EventSource{Component: "temporal-client", Host: podName},
	}
}

// resolveJobReference returns an ObjectReference to the Job that owns the given
// Pod (falling back to the Pod itself when no owning Job is found).
func resolveJobReference(ctx context.Context, clientset kubernetes.Interface, namespace, podName string) (*corev1.ObjectReference, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod %s/%s: %w", namespace, podName, err)
	}
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "Job" {
			return &corev1.ObjectReference{
				Kind:       "Job",
				APIVersion: owner.APIVersion,
				Namespace:  namespace,
				Name:       owner.Name,
				UID:        owner.UID,
			}, nil
		}
	}
	return &corev1.ObjectReference{
		Kind:       "Pod",
		APIVersion: "v1",
		Namespace:  namespace,
		Name:       pod.Name,
		UID:        pod.UID,
	}, nil
}

// event creates a Kubernetes Event synchronously, falling back to logs when the
// emitter is unavailable (nil receiver is safe) or when the API call fails.
// Synchronous creation guarantees the Event is persisted before the process exits.
func (e *eventEmitter) event(eventType, reason, messageFmt string, args ...any) {
	msg := fmt.Sprintf(messageFmt, args...)
	if e == nil || e.events == nil || e.ref == nil {
		klog.Infof("[event:%s/%s] %s", eventType, reason, msg)
		return
	}
	now := metav1.Now()
	ev := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: e.ref.Name + ".",
			Namespace:    e.ref.Namespace,
		},
		InvolvedObject: *e.ref,
		Reason:         reason,
		Message:        msg,
		Type:           eventType,
		Source:         e.source,
		FirstTimestamp: now,
		LastTimestamp:  now,
		Count:          1,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := e.events.Create(ctx, ev, metav1.CreateOptions{}); err != nil {
		klog.Warningf("failed to create Kubernetes Event (%s/%s): %v; message: %s", eventType, reason, err, msg)
	}
}
