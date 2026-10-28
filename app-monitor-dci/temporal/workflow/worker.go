package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"sync"
	examplecomv1 "temporal/api/v1"
	"temporal/shared"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// temporalClient is a narrow interface used by run for easier testing.
type temporalClient interface {
	Close()
}

// temporalWorker abstracts the Temporal worker for testing.
type temporalWorker interface {
	RegisterWorkflow(w any)
	RegisterActivity(a any)
	RegisterActivityWithOptions(a any, options activity.RegisterOptions)
	Run(<-chan any) error
}

// temporalClientDial is an indirection for client.Dial so tests can replace it.
var temporalClientDial = func(options client.Options) (temporalClient, error) {
	return client.Dial(options)
}

var getConfigOrDie = ctrl.GetConfigOrDie

var k8sClientNew = k8sclient.New

// workerNew is an indirection for worker.New so tests can stub worker behavior.
var workerNew = func(c temporalClient, taskQueue string, options worker.Options) temporalWorker {
	cli, ok := c.(client.Client)
	if !ok {
		panic("temporal client does not implement client.Client")
	}
	return worker.New(cli, taskQueue, options)
}

type Activities struct {
	K8sClient k8sclient.Client
	Scheme    *runtime.Scheme
}

func (acts *Activities) VectorSignalActivity(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
	return VectorSignalActivity(ctx, acts.K8sClient, podState, acts.Scheme)
}
func (acts *Activities) AlertManagerSignalActivity(ctx context.Context, podState shared.PodUpdateStatus, alertNamespace string, alertPod string) (*shared.PodUpdateStatus, error) {
	return AlertManagerSignalActivity(ctx, acts.K8sClient, podState, acts.Scheme, alertNamespace, alertPod)
}

var initLoggingOnce sync.Once

func initLogging() {
	initLoggingOnce.Do(func() {
		klog.InitFlags(nil)
		flag.Parse()
	})
}

func main() {
	run()
}

func run() {
	initLogging()
	flag.Parse()
	defer klog.Flush()

	cfg := getConfigOrDie()
	cfg.QPS = 50
	cfg.Burst = 100
	var acts Activities
	var err error
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = examplecomv1.AddToScheme(sch)
	acts.K8sClient, err = k8sClientNew(cfg, k8sclient.Options{Scheme: sch})
	if err != nil {
		klog.Errorf("k8sclient.New(): %v", err)
		return
	}
	acts.Scheme = sch

	// Get Temporal server address from environment variable
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		klog.Warning("TEMPORAL_ADDRESS environment variable is not set. Temporal client will use the default address: temporal-stack-frontend.temporal.svc.cluster.local:7233.")
		// Use default Temporal server address if not specified
		temporalAddress = "temporal-stack-frontend.temporal.svc.cluster.local:7233"
	}
	// Get Temporal namespace from environment variable
	temporalNamespace := os.Getenv("TEMPORAL_NAMESPACE")
	if temporalNamespace == "" {
		klog.Warning("TEMPORAL_NAMESPACE environment variable is not set. Temporal client will use the default namespace: default.")
		temporalNamespace = "default"
	}
	// Create Temporal client
	c, err := temporalClientDial(client.Options{
		HostPort:  temporalAddress,
		Namespace: temporalNamespace,
	})
	if err != nil {
		klog.Errorf("unable to create Temporal client: %v", err)
		return
	}
	defer c.Close()
	// Create worker for the monitor-task-queue
	w := workerNew(c, "monitor-task-queue", worker.Options{})
	// Register workflow and activities
	w.RegisterWorkflow(MonitorWorkflow)

	var activityCounter sync.WaitGroup
	w.RegisterActivityWithOptions(
		func(ctx context.Context, podState shared.PodUpdateStatus) (*shared.PodUpdateStatus, error) {
			activityCounter.Add(1)
			defer activityCounter.Done()
			return acts.VectorSignalActivity(ctx, podState)
		},
		activity.RegisterOptions{Name: "VectorSignalActivity"},
	)
	w.RegisterActivityWithOptions(
		func(ctx context.Context, podState shared.PodUpdateStatus, alertNamespace string, alertPod string) (*shared.PodUpdateStatus, error) {
			activityCounter.Add(1)
			defer activityCounter.Done()
			return acts.AlertManagerSignalActivity(ctx, podState, alertNamespace, alertPod)
		},
		activity.RegisterOptions{Name: "AlertManagerSignalActivity"},
	)

	// Start the worker and listen for tasks
	klog.Info("Worker started (with Activity), listening on monitor-task-queue")
	if err := w.Run(worker.InterruptCh()); err != nil {
		if err == context.Canceled || strings.Contains(err.Error(), "context canceled") || strings.Contains(err.Error(), "worker stopped") {
			klog.Warningf("Worker stopped gracefully: %v", err)
			return
		}
		klog.Errorf("Temporal worker failed unexpectedly: %v", err)
		return
	}
	klog.Info("Temporal worker stop requested. Waiting in-flight")

	activityCounter.Wait()
	klog.Info("Temporal worker process finished normally.")
}
