package controller

import (
	"time"
)

// Definitions for wait time
const (
	// requeue interval for pending CardAlert
	requeueIntervalCardAlertPending = 15 * time.Second

	// requeue interval for pending CardScalingRequest
	requeueIntervalCardScalingRequestPending = 15 * time.Second

	// requeue interval for pending NodeScalingRequest
	requeueIntervalNodeScalingRequestPending = 15 * time.Second

	// requeue interval for waiting for scaling MachineDeployment of NodeScalingRequest
	requeueIntervalNodeScalingRequestWaitingScaling = 1 * time.Minute

	// polling interval to sync resource version for custom resource after updating status of the resource
	pollingIntervalSyncResourceVersion = 100 * time.Millisecond

	// polling interval to check the completion of MachineAlert in auto replacement flow
	pollingIntervalCheckMachineAlertCompletion = 15 * time.Second
)

// Definitions for environment variable names
const (
	// NSR_TTL_SECONDS_AFTER_FINISHED is the environment variable
	// to get the TTL setting for NodeScalingRequest resource after it is completed.
	// If the value is a positive integer, TTL is enabled and the value is used as seconds for TTL.
	// If the value is "0" or a negative integer, setting will fail.
	// If the value is not set, TTL is disabled.
	nodeScalingRequestTTLEnvVarName = "NSR_TTL_SECONDS_AFTER_FINISHED"

	// CSR_TTL_SECONDS_AFTER_FINISHED is the environment variable
	// to get the TTL setting for CardScalingRequest resource after it is completed.
	// If the value is a positive integer, TTL is enabled and the value is used as seconds for TTL.
	// If the value is "0" or a negative integer, setting will fail.
	// If the value is not set, TTL is disabled.
	cardScalingRequestTTLEnvVarName = "CSR_TTL_SECONDS_AFTER_FINISHED"

	// CARD_ALERT_TTL_SECONDS_AFTER_FINISHED is the environment variable
	// to get the TTL setting for CardAlert resource after it is completed.
	// If the value is a positive integer, TTL is enabled and the value is used as seconds for TTL.
	// If the value is "0" or a negative integer, setting will fail.
	// If the value is not set, TTL is disabled.
	cardAlertTTLEnvVarName = "CARD_ALERT_TTL_SECONDS_AFTER_FINISHED"

	// ENV_CARD_ALERT_AUTO_REPLACE_ENABLED is the environment variable
	// to enable or disable auto replacement of nodes when a card alert occurs.
	// If set "false", auto replacement is disabled, otherwise enabled(default).
	autoReplaceEnabledEnvVarName = "ENV_CARD_ALERT_AUTO_REPLACE_ENABLED"

	// ENV_CARD_ALERT_CARD_TAINT_POLICY is the environment variable
	// to specify the policy to create DeviceTaintRule.
	// Valid values are "NoExecute" and "NoSchedule". If not set, the default is "NoExecute".
	// This variable is used only when auto replacement is disabled,
	// so the value is D.C. when auto replacement is enabled.
	deviceTaintRulePolicyEnvVarName = "ENV_CARD_ALERT_CARD_TAINT_POLICY"

	// MACHINE_ALERT_NAMESPACE is the environment variable
	// to specify the namespace in management cluster where MachineAlert exists.
	// This variable should be set when auto replacement is enabled,
	// so the value is D.C. when auto replacement is disabled.
	machineAlertNamespaceEnvVarName = "MACHINE_ALERT_NAMESPACE"
)

// Event reasons for Reconcile
const (
	eventReasonUnexpectedError          = "UnexpectedError"
	eventReasonFetchingFailed           = "FetchingFailed"
	eventReasonInvalidType              = "InvalidType"
	eventReasonInvalidPhase             = "InvalidPhase"
	eventReasonInitializingFailed       = "InitializingFailed"
	eventReasonPreparingResourceFailed  = "PreparingResourceFailed"
	eventReasonScalingDownRequestFailed = "ScalingDownRequestFailed"
	eventReasonScalingDownFailed        = "ScalingDownFailed"
	eventReasonScalingUpRequestFailed   = "ScalingUpRequestFailed"
	eventReasonScalingUpFailed          = "ScalingUpFailed"
	eventReasonVMTypeNotSupported       = "VMTypeNotSupported"
	eventReasonCardScalingRequestFailed = "CardScalingRequestFailed"
	eventReasonAutoReplacingFailed      = "AutoReplacingFailed"
	eventReasonDeviceTaintingFailed     = "DeviceTaintingFailed"
	eventReasonUpdatingStatusFailed     = "UpdatingStatusFailed"
	eventReasonSyncingCacheFailed       = "SyncingCacheFailed"
	eventReasonTerminationFailed        = "TerminationFailed"

	eventReasonPending                   = "Pending"
	eventReasonCardScalingRequestCreated = "CardScalingRequestCreated"
	eventReasonMachineAlertCreated       = "MachineAlertCreated"
	eventReasonScalingDownStarted        = "ScalingDownStarted"
	eventReasonScalingDownCompleted      = "ScalingDownCompleted"
	eventReasonScalingUpStarted          = "ScalingUpStarted"
	eventReasonScalingUpCompleted        = "ScalingUpCompleted"
	eventReasonCompleted                 = "Completed"
	eventReasonExpired                   = "Expired"
	eventReasonNoSelectorsBuilt          = "NoSelectorsBuilt"
)

// Event reasons for Finalizer
const (
	eventReasonFinalizerUnexpectedError         = "FinalizerUnexpectedError"
	eventReasonFinalizerWaitingCompletionFailed = "FinalizerWaitingCompletionFailed"
	eventReasonFinalizerCleanupFailed           = "FinalizerCleanupFailed"
	eventReasonFinalizerScalingUpFailed         = "FinalizerScalingUpFailed"
	eventReasonFinalizerScalingDownFailed       = "FinalizerScalingDownFailed"
	eventReasonFinalizerRemovingFinalizerFailed = "FinalizerRemovingFinalizerFailed"
	eventReasonFinalizerFailed                  = "FinalizerFailed"

	eventReasonFinalizerScalingUpCompleted   = "FinalizerScalingUpCompleted"
	eventReasonFinalizerScalingDownCompleted = "FinalizerScalingDownCompleted"
)

// Event actions for Reconcile
// TODO: With recent Kubernetes upgrades(from v1.34.x to v1.36.x), Event action can be set; for now,
// we keep it as `Reconciling` to avoid relying on reason as a temporary workaround.
// In the future, this should be updated to proper values, including reason.
const (
	eventActionReconciling = "Reconciling"
)

// Definitions for DeviceTaintRule
const (
	deviceTaintRuleKeyUnhealthy      = "dra.example.com/unhealthy"
	deviceTaintRuleValueBroken       = "Broken"
	deviceTaintRuleCELPathUUID       = "uuid"
	deviceTaintRuleCELPathParentUUID = "parentUUID"
	deviceTaintRuleCELLabelGPUNvidia = "gpu.nvidia.com"
	deviceTaintRuleDevice            = "dra"
)
