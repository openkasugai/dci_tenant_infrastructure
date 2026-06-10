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
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"hash/fnv"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"
	"sync"
	"temporal/shared"
	"time"
)

const maxRequestSize int64 = 10 * 1024 * 1024 // 10 MB
const defaultShardCount = 3

const defaultMaxSignalBufSize = 600

func loadMaxSignalBufSize() int {
	val := os.Getenv("MAX_SIGNAL_BUF_SIZE")
	if val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			klog.Infof("MAX_SIGNAL_BUF_SIZE=%d", n)
			return n
		}
		klog.Warningf("Invalid MAX_SIGNAL_BUF_SIZE value %q, using default: %d", val, defaultMaxSignalBufSize)
	} else {
		klog.Warningf("MAX_SIGNAL_BUF_SIZE is not set, using default: %d", defaultMaxSignalBufSize)
	}
	return defaultMaxSignalBufSize
}

var maxSignalBufSize = loadMaxSignalBufSize()

func loadShardCount() uint32 {
	val := os.Getenv("SHARD_COUNT")
	if val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			klog.Infof("SHARD_COUNT=%d", n)
			return uint32(n)
		}
		klog.Warningf("Invalid SHARD_COUNT value %q, using default: %d", val, defaultShardCount)
	} else {
		klog.Warningf("SHARD_COUNT is not set, using default: %d", defaultShardCount)
	}
	return defaultShardCount
}

var shardCount = loadShardCount()

// shardedWorkflowID returns the workflow ID for the given pod name based on FNV hash.
func shardedWorkflowID(podName string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(podName))
	shard := h.Sum32() % shardCount
	return fmt.Sprintf("monitor-workflow-%d", shard)
}

// temporalClient is a narrow interface used by run for easier testing.
type temporalClient interface {
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
	QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error)
	Close()
}

// bufferedSignal holds a signal that could not be sent because the target
// workflow was in the draining (ContinueAsNew-pending) state.
type bufferedSignal struct {
	workflowID string
	signalName string
	arg        interface{}
}

var (
	signalBufMu sync.Mutex
	signalBuf   []bufferedSignal
)

// drainLoopInterval controls how often the drain loop retries buffered signals.
// Exposed as a variable for testing.
var drainLoopInterval = 200 * time.Millisecond

// queryIsDraining queries the workflow for the "is-draining" state.
// Returns true (draining) on any error so the caller buffers the signal safely.
func queryIsDraining(ctx context.Context, c temporalClient, workflowID string) bool {
	resp, err := c.QueryWorkflow(ctx, workflowID, "", "is-draining")
	if err != nil {
		klog.Warningf("QueryWorkflow 'is-draining' failed for %s (treating as draining): %v", workflowID, err)
		return true
	}
	var isDraining bool
	if err := resp.Get(&isDraining); err != nil {
		klog.Warningf("Failed to decode 'is-draining' response for %s (treating as draining): %v", workflowID, err)
		return true
	}
	return isDraining
}

// bufferOrSignal sends the signal immediately when the workflow is not draining.
// If the workflow is draining (or the query fails), the signal is placed in the
// in-memory buffer and will be delivered by the drain loop once CaN is complete.
func bufferOrSignal(ctx context.Context, c temporalClient, workflowID, signalName string, arg interface{}) error {
	if queryIsDraining(ctx, c, workflowID) {
		signalBufMu.Lock()
		if len(signalBuf) >= maxSignalBufSize {
			dropped := signalBuf[0]
			signalBuf = append(signalBuf[1:], bufferedSignal{workflowID, signalName, arg})
			signalBufMu.Unlock()
			klog.Errorf("signalBuf overflow (size=%d), dropped oldest signal: workflowID=%s signalName=%s",
				maxSignalBufSize, dropped.workflowID, dropped.signalName)
			return nil
		}
		klog.Infof("Workflow %s is draining, buffering signal %s", workflowID, signalName)
		signalBuf = append(signalBuf, bufferedSignal{workflowID, signalName, arg})
		signalBufMu.Unlock()
		return nil
	}
	return c.SignalWorkflow(ctx, workflowID, "", signalName, arg)
}

// flushSignalBuffer attempts to deliver every buffered signal.
// Signals whose target workflow is still draining are kept in the buffer.
func flushSignalBuffer(ctx context.Context, c temporalClient) {
	signalBufMu.Lock()
	if len(signalBuf) == 0 {
		signalBufMu.Unlock()
		return
	}
	todo := signalBuf
	signalBuf = nil
	signalBufMu.Unlock()

	var remaining []bufferedSignal
	for _, sig := range todo {
		if queryIsDraining(ctx, c, sig.workflowID) {
			remaining = append(remaining, sig)
			continue
		}
		if err := c.SignalWorkflow(ctx, sig.workflowID, "", sig.signalName, sig.arg); err != nil {
			klog.Warningf("Failed to flush buffered signal to %s: %v, re-buffering", sig.workflowID, err)
			remaining = append(remaining, sig)
		} else {
			klog.Infof("Flushed buffered signal to workflow %s: %s", sig.workflowID, sig.signalName)
		}
	}

	if len(remaining) > 0 {
		signalBufMu.Lock()
		signalBuf = append(remaining, signalBuf...)
		signalBufMu.Unlock()
	}
}

// startDrainLoop starts a background goroutine that periodically calls
// flushSignalBuffer until ctx is cancelled.
func startDrainLoop(ctx context.Context, c temporalClient) {
	go func() {
		ticker := time.NewTicker(drainLoopInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				flushSignalBuffer(ctx, c)
			}
		}
	}()
}

// temporalClientDial is an indirection for client.Dial so tests can replace it.
var temporalClientDial = func(options client.Options) (temporalClient, error) {
	return client.Dial(options)
}

// httpListenAndServe is an indirection for http.ListenAndServe so tests can stub the server.
var httpListenAndServe = func(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}

var inClusterConfigFunc = rest.InClusterConfig

var newForConfigFunc = func(cfg *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(cfg)
}

var podExistsFunc = podExists

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

func podExists(namespace, podName string) (bool, error) {
	config, err := inClusterConfigFunc()
	if err != nil {
		return false, err
	}

	clientset, err := newForConfigFunc(config)
	if err != nil {
		return false, err
	}

	_, err = clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func run() {
	initLogging()
	flag.Parse()
	defer klog.Flush()

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
		return
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startDrainLoop(ctx, c)

	http.HandleFunc("/vector-signal", func(w http.ResponseWriter, r *http.Request) {
		//debug
		klog.Infof("Received HTTP request for /vector-signal. Method: %s, RemoteAddr: %s", r.Method, r.RemoteAddr)

		// Read the body from r.Body
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxRequestSize))
		if err != nil {
			klog.Errorf("Error reading vector-signal request body: %v from %s", err, r.RemoteAddr)
			return
		} else if int64(len(bodyBytes)) >= maxRequestSize {
			klog.Errorf("Request body size (%d bytes) exceeds maxRequestSize (10MB) for request %s %s from %s", len(bodyBytes), r.Method, r.URL.Path, r.RemoteAddr)
			return
		}

		// Convert the read body to a string and log it
		bodyString := string(bodyBytes)
		klog.Infof("Received body: %s", bodyString)

		// Note: r.Body is a stream that becomes empty after being read once.
		// If you need to read the body again later (e.g., in another middleware),
		// you must reset r.Body with the content you just read.
		// Otherwise, subsequent reads will return nothing.
		// By using io.NopCloser and bytes.NewBuffer,
		// we convert the content of bodyBytes into a new readable io.ReadCloser.
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the body
		// Decode into the outer struct
		var outer shared.OuterStruct
		err = json.NewDecoder(r.Body).Decode(&outer)
		if err != nil {
			klog.Errorf("Failed to decode outer JSON: %v. Request URL: %s, RemoteAddr: %s", err, r.URL.Path, r.RemoteAddr)
			return
		}
		// Decode the string in outer.Metadata into the inner struct
		var inner shared.InnerMetadata
		err = json.Unmarshal([]byte(outer.Metadata), &inner)
		if err != nil {
			klog.Errorf("Failed to decode inner JSON from metadata string: %v. Request URL: %s, RemoteAddr: %s", err, r.URL.Path, r.RemoteAddr)
			return
		}

		// Pod existence check (if namespace and pod name are not empty)
		ns := inner.Metadata.Namespace
		pod := inner.Metadata.Name
		if ns == "" || pod == "" {
			klog.Errorf("namespace/pod label is empty in vector-signal: namespace=%q, pod=%q", ns, pod)
			return
		}
		exists, err := podExistsFunc(ns, pod)
		if err != nil {
			klog.Errorf("failed to check pod existence in vector-signal: namespace=%s, pod=%s, err=%v", ns, pod, err)
			return
		}
		if !exists {
			klog.Errorf("pod not found in vector-signal: namespace=%s, pod=%s", ns, pod)
			return
		}

		// At this point, the `inner` variable contains all the data
		klog.Infof("Successfully decoded: %+v\n", inner)
		// Accessing specific fields
		klog.Infof("Message: %s\n", inner.Message)
		klog.Infof("App Status: %+v\n", inner.Status.AppStatus)
		klog.Infof("App Status.status: %s\n", string(inner.Status.AppStatus.Status))
		//Format JSON for sending to ArgoEvents. Convert InnerMetadata to PodUpdateStatus
		var update shared.PodUpdateStatus

		update.Metadata = inner.Metadata
		update.Spec = inner.Spec
		update.Status = inner.Status
		klog.Infof("Update App Status: %+v\n", update.Status.AppStatus)
		klog.Infof("Update App Status.status: %s\n", string(update.Status.AppStatus.Status))

		workflowID := shardedWorkflowID(pod)
		klog.Infof("Routing vector-signal to workflow: %s (pod: %s)", workflowID, pod)
		err = bufferOrSignal(context.Background(), c, workflowID, "vector-signal", update)
		if err != nil {
			klog.Errorf("Failed to signal workflow '%s': %v. Signal Name: 'vector-signal'", workflowID, err)
			return
		}

		n, writeErr := w.Write([]byte("Signal sent."))
		if writeErr != nil {
			klog.Errorf("Error writing response body: %v, bytes written: %d. Request URL: %s, RemoteAddr: %s", writeErr, n, r.URL.Path, r.RemoteAddr)
			return
		}

		klog.Info("Successfully signal workflow execute.")
	})

	http.HandleFunc("/alertmanager-signal", func(w http.ResponseWriter, r *http.Request) {
		//debug
		klog.Infof("Received HTTP request for /alertmanager-signal. Method: %s, RemoteAddr: %s", r.Method, r.RemoteAddr)

		// Read the body from r.Body
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxRequestSize))
		if err != nil {
			klog.Errorf("failed to read alertmanager-signal request body: %v", err)
			return
		} else if int64(len(bodyBytes)) >= maxRequestSize {
			klog.Errorf("Request body size (%d bytes) exceeds maxRequestSize (10MB) for request %s %s from %s", len(bodyBytes), r.Method, r.URL.Path, r.RemoteAddr)
			return
		}

		// Convert the read body to a string and log it
		bodyString := string(bodyBytes)
		klog.Infof("Received body: %s", bodyString)

		// Note: r.Body is a stream that becomes empty after being read once.
		// If you need to read the body again later,
		// you must reset r.Body with the content you just read.
		// Otherwise, subsequent reads will return nothing.
		// By using io.NopCloser and bytes.NewBuffer,
		// we convert the content of bodyBytes into a new readable io.ReadCloser.
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the body
		// Decode into the outer struct
		var outer shared.OuterStruct
		err = json.NewDecoder(r.Body).Decode(&outer)
		if err != nil {
			klog.Errorf("Failed to decode outer JSON: %v. Request URL: %s, RemoteAddr: %s", err, r.URL.Path, r.RemoteAddr)
			return
		}
		// Decode the string in outer.Metadata into the alerts struct
		var alerts shared.PodAlerts
		err = json.Unmarshal([]byte(outer.Metadata), &alerts)
		if err != nil {
			klog.Errorf("Failed to decode inner JSON from metadata string: %v. Request URL: %s, RemoteAddr: %s", err, r.URL.Path, r.RemoteAddr)
			return
		}

		// `alerts` variable contains all the data
		klog.Infof("Successfully decoded: %+v\n", alerts)

		for _, alert := range alerts.Alerts {
			ns := alert.Labels.GetNamespace()
			pod := alert.Labels.GetPod()

			if ns == "" || pod == "" {
				klog.Errorf("namespace/pod label is empty in alertmanager-signal: namespace=%q, pod=%q", ns, pod)
				return
			}

			exists, err := podExistsFunc(ns, pod)
			if err != nil {
				klog.Errorf("failed to check pod existence in alertmanager-signal: namespace=%s, pod=%s, err=%v", ns, pod, err)
				return
			}

			if !exists {
				klog.Warningf("pod not found in alertmanager-signal: namespace=%s, pod=%s", ns, pod)
				return
			}
		}

		var alertWorkflowID string
		if len(alerts.Alerts) > 0 {
			firstPod := alerts.Alerts[0].Labels.GetPod()
			alertWorkflowID = shardedWorkflowID(firstPod)
		} else {
			alertWorkflowID = "monitor-workflow-0"
		}
		klog.Infof("Routing alertmanager-signal to workflow: %s", alertWorkflowID)
		err = bufferOrSignal(context.Background(), c, alertWorkflowID, "alertmanager-signal", alerts)
		if err != nil {
			klog.Errorf("Failed to signal workflow '%s': %v. Signal Name: 'alertmanager-signal'", alertWorkflowID, err)
			return
		}

		n, writeErr := w.Write([]byte("Signal sent."))
		if writeErr != nil {
			klog.Errorf("Error writing response body: %v, bytes written: %d. Request URL: %s, RemoteAddr: %s", writeErr, n, r.URL.Path, r.RemoteAddr)
			return
		}
		klog.Info("Successfully signal workflow execute.")
	})

	klog.Info("HTTP Gateway listening on :8080")
	if err := httpListenAndServe(":8080", nil); err != nil {
		klog.Errorf("Failed to start HTTP server: %v", err)
		return
	}

}
