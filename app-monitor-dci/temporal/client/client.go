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
	"flag"
	"fmt"
	"go.temporal.io/sdk/client"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"sync"
)

// temporalClient is a narrow interface used by run for easier testing.
type temporalClient interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)
	Close()
}

// temporalClientDial is an indirection for client.Dial so tests can replace it.
var temporalClientDial = func(options client.Options) (temporalClient, error) {
	return client.Dial(options)
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

	shardCount := 3
	if val := os.Getenv("SHARD_COUNT"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			shardCount = n
		} else {
			klog.Warningf("Invalid SHARD_COUNT value %q, using default: %d", val, shardCount)
		}
	} else {
		klog.Warningf("SHARD_COUNT is not set, using default: %d", shardCount)
	}
	klog.Infof("Starting %d workflow shard(s)", shardCount)

	// Start shardCount workflow instances (monitor-workflow-0, 1, ...)
	successCount := 0
	for i := range shardCount {
		workflowID := fmt.Sprintf("monitor-workflow-%d", i)
		we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "monitor-task-queue",
		}, "MonitorWorkflow")
		if err != nil {
			klog.Errorf("Failed to start workflow %s: %v", workflowID, err)
			continue
		}
		klog.Infof("Started workflow: %s", we.GetID())
		successCount++
	}
	if successCount == 0 {
		klog.Errorf("All %d workflow(s) failed to start", shardCount)
	} else {
		klog.Infof("Successfully started %d/%d workflow shard(s)", successCount, shardCount)
	}
}
