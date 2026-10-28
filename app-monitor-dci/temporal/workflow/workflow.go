package main

import (
	"fmt"
	"os"
	"strconv"
	"temporal/shared"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"k8s.io/klog/v2"
)

// MonitorWorkflowState is the workflow state to be passed with Continue-As-New
type MonitorWorkflowState struct {
	State map[string]shared.PodUpdateStatus
}

const VectorSignalName = "vector-signal"
const AlertmanagerSignalName = "alertmanager-signal"

const DefaultHistoryEventThreshold = 10000

func loadHistoryEventThreshold() int {
	val := os.Getenv("HISTORY_EVENT_THRESHOLD")
	if val != "" {
		if threshold, err := strconv.Atoi(val); err == nil && threshold > 0 {
			klog.Infof("HISTORY_EVENT_THRESHOLD=%d", threshold)
			return threshold
		}
		klog.Warningf("Invalid HISTORY_EVENT_THRESHOLD value %q, using default: %d", val, DefaultHistoryEventThreshold)
	} else {
		klog.Warningf("HISTORY_EVENT_THRESHOLD is not set, using default: %d", DefaultHistoryEventThreshold)
	}
	return DefaultHistoryEventThreshold
}

var HistoryEventThreshold = loadHistoryEventThreshold()

// indirection variables for activities to allow testing without calling real external services
var (
	vectorSignalActivityFn       = VectorSignalActivity
	alertManagerSignalActivityFn = AlertManagerSignalActivity
	currentHistoryLengthFn       = func(ctx workflow.Context) int {
		return int(workflow.GetInfo(ctx).GetCurrentHistoryLength())
	}
	setQueryHandlerFn = func(ctx workflow.Context, queryType string, handler any) error {
		return workflow.SetQueryHandler(ctx, queryType, handler)
	}
)

// Persistent monitor workflow
func MonitorWorkflow(ctx workflow.Context, initialState *MonitorWorkflowState) error {
	// Initialize or resume state
	state := make(map[string]shared.PodUpdateStatus)
	// Get logger
	logger := workflow.GetLogger(ctx)
	// Log start message
	if initialState != nil {
		// Resume from previous state
		state = initialState.State
		klog.Infof("MonitorWorkflow resumed from previous state with %d items", len(state))
	} else {
		// Start fresh state
		klog.Info("MonitorWorkflow started fresh")
		logger.Info("MonitorWorkflow started fresh")
	}
	// isDraining is set to true just before ContinueAsNew so the gateway
	// can detect the draining window via the "is-draining" Query Handler and
	// buffer incoming signals instead of sending them during the transition.
	isDraining := false
	if err := setQueryHandlerFn(ctx, "is-draining", func() (bool, error) {
		return isDraining, nil
	}); err != nil {
		return err
	}

	// Get signal channels
	vectorSignalCh := workflow.GetSignalChannel(ctx, VectorSignalName)
	alertmanagerSignalCh := workflow.GetSignalChannel(ctx, AlertmanagerSignalName)

	// Set activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 2,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second * 3,
			MaximumAttempts: 3,
		},
	}

	// Apply activity options to context
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger.Info("MonitorWorkflow started(with activity), waiting for signals...")
	klog.Info("MonitorWorkflow started(with activity), waiting for signals...")

	// Main event loop
	for {
		// Check history event count for Continue-As-New
		currentHistoryLength := currentHistoryLengthFn(ctx)

		// Trigger drain mode on first threshold breach: notify the Gateway (via
		// the "is-draining" query) so it buffers new signals, while the workflow
		// continues to consume signals that are already in the Temporal queue.
		if !isDraining && currentHistoryLength >= HistoryEventThreshold {
			isDraining = true
			logger.Info("History event count approaching limit, draining signals before CaN...",
				"current_events", currentHistoryLength,
				"vector_queue", vectorSignalCh.Len(),
				"alert_queue", alertmanagerSignalCh.Len())
			klog.Infof("History event count approaching limit (%d events), draining %d vector + %d alert signals before CaN",
				currentHistoryLength, vectorSignalCh.Len(), alertmanagerSignalCh.Len())
		}

		// In drain mode: perform CaN only once all queued signals have been processed
		if isDraining && vectorSignalCh.Len() == 0 && alertmanagerSignalCh.Len() == 0 {
			klog.Infof("Signal drain complete, continuing as new workflow")
			logger.Info("Signal drain complete, continuing as new workflow")
			return workflow.NewContinueAsNewError(ctx, MonitorWorkflow, &MonitorWorkflowState{State: state})
		}

		canceled := false
		var update shared.PodUpdateStatus
		// Create selector to handle multiple channels
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(ctx.Done(), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			canceled = true
		})

		// Handle vector signal
		selector.AddReceive(vectorSignalCh, func(c workflow.ReceiveChannel, more bool) {
			// Receive update from channel
			c.Receive(ctx, &update)
			// Update state with received update
			podKey := fmt.Sprintf("%s/%s", update.Metadata.Namespace, update.Metadata.Name)
			state[podKey] = update
			logger.Info("Received vector signal!", "update status", update.Metadata.Namespace, update.Metadata.Name)
			klog.Infof("Received vector signal from namespace=%s, pod=%s", update.Metadata.Namespace, update.Metadata.Name)

			// Call VectorSignalActivity
			var vectorResult *shared.PodUpdateStatus
			// Execute activity and get result
			err := workflow.ExecuteActivity(ctx, vectorSignalActivityFn, update).Get(ctx, &vectorResult)
			if err != nil {
				logger.Error("Activity failed", "err", err)
				klog.Errorf("VectorSignalActivity failed with other error: %v", err)
			} else {
				// Update state with activity result
				state[podKey] = *vectorResult
				klog.Infof("vector result:%+v", vectorResult)
				logger.Info("Activity success")
				klog.Info("VectorSignalActivity completed successfully")
			}
		})

		// Handle alertmanager signal
		selector.AddReceive(alertmanagerSignalCh, func(c workflow.ReceiveChannel, more bool) {
			var alerts shared.PodAlerts
			// Receive alerts from channel
			c.Receive(ctx, &alerts)

			logger.Info("Received alertmanager signal!")
			klog.Infof("Received alertmanager signal! (alerts count: %d)", len(alerts.Alerts))

			// Process each alert
			for _, alert := range alerts.Alerts {
				// Debug: Log all label fields
				klog.Infof("Alert Labels - Namespace:%q Pod:%q ExportedNamespace:%q ExportedPod:%q",
					alert.Labels.Namespace, alert.Labels.Pod,
					alert.Labels.ExportedNamespace, alert.Labels.ExportedPod)

				ns := alert.Labels.GetNamespace()
				pod := alert.Labels.GetPod()
				klog.Infof("Resolved - Namespace:%s Pod:%s", ns, pod)
				key := fmt.Sprintf("%s/%s", ns, pod)
				// Check if pod exists in state
				if alertPodState, ok := state[key]; ok {
					var alertResult *shared.PodUpdateStatus
					// Execute AlertManagerSignalActivity
					err := workflow.ExecuteActivity(ctx, alertManagerSignalActivityFn, alertPodState, ns, pod).Get(ctx, &alertResult)
					if err != nil {
						logger.Error("Activity failed", "err", err)
						klog.Errorf("AlertManagerSignalActivity failed with other error: %v", err)

					} else {
						// Update state with activity result
						state[key] = *alertResult
						klog.Infof("alert result:%+v", alertResult)
						logger.Info("Activity success")
						klog.Info("AlertManagerSignalActivity completed successfully")
					}
				} else {
					logger.Info("Pod not found in state", "pod", pod)
					klog.Infof("Pod not found in state: namespace=%s, pod=%s", ns, pod)
				}
			}
		})
		selector.Select(ctx)
		if canceled {
			return workflow.ErrCanceled
		}
	}
}
