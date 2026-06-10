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

package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1alpha3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"
)

// CardAlertReconciler reconciles a CardAlert object.
type CardAlertReconciler struct {
	// client for workload cluster
	client.Client

	// scheme for workload cluster
	Scheme *runtime.Scheme

	// interface for event recorder
	Recorder record.EventRecorder

	// client for management cluster
	ClientHelper helperclient.ClientHelper

	// namespace for management cluster where MachineAlert exists
	MachineAlertNamespace string

	// enableAutoReplace indicates whether auto replacement is enabled or not
	enableAutoReplace bool

	// cardTaintPolicy indicates the pod processing policy when auto replacement is disabled
	cardTaintPolicy resourcev1.DeviceTaintEffect

	// setting for enabling TTL
	enableTTLSecondsAfterFinished bool

	// setting for TTL seconds
	ttlSecondsAfterFinished int
}

// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardalerts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardalerts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardalerts/finalizers,verbs=update

// Additional rbac
// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardscalingrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch
// +kubebuilder:rbac:groups=resource.k8s.io,resources=devicetaintrules,verbs=get;list;watch;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// Reconcile check the finalizer for CardAlert and call doFinalize() if there is a deletion request.
// Otherwise, call doReconcile() to reconcile the CardAlert resource.
func (r *CardAlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Get CardAlert.
	ca := &scalev1.CardAlert{}
	if err := r.Get(ctx, req.NamespacedName, ca); apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get CardAlert: %w", err)
	}

	// Check if the CardAlert is being deleted then finalize it.
	if !ca.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ca, scalev1.CardAlertFinalizerName) {
			return ctrl.Result{}, nil // finalizer for this reconciler is not set, so we do nothing
		}
		rslt, err := r.doFinalize(ctx, ca)
		if err != nil {
			err = fmt.Errorf("failed to finalize CardAlert: %w", err)
		}
		return rslt, err
	}

	// Reconcile the CardAlert.
	rslt, err := r.doReconcile(ctx, ca)
	if err != nil {
		err = fmt.Errorf("failed to reconcile CardAlert: %w", err)
	}
	return rslt, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CardAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if mgr == nil {
		// This case is only for failsafe.
		return errors.New("expected mgr is not nil, but got nil")
	}
	logger := mgr.GetLogger().WithName("cardalert-controller-setup")

	// Get auto-replace-config from environment variables.
	autoReplaceEnabled, found := os.LookupEnv(autoReplaceEnabledEnvVarName)
	if found {
		switch strings.ToLower(autoReplaceEnabled) {
		case "true":
			r.enableAutoReplace = true
		case "false":
			r.enableAutoReplace = false
		default:
			return fmt.Errorf("expected true or false, but got %s for %s", autoReplaceEnabled, autoReplaceEnabledEnvVarName)
		}
	} else {
		r.enableAutoReplace = true // default is enabled
	}

	if r.enableAutoReplace {
		// Get namespace for MachineAlert in management cluster from environment variables.
		maNamespace := os.Getenv(machineAlertNamespaceEnvVarName)
		if maNamespace == "" {
			return fmt.Errorf("expected %s to be set, but got empty value", machineAlertNamespaceEnvVarName)
		}
		r.MachineAlertNamespace = maNamespace
	} else {
		// Get device-taint-policy from environment variables.
		taintPolicy, found := os.LookupEnv(deviceTaintRulePolicyEnvVarName)
		if found {
			switch strings.ToLower(taintPolicy) {
			case strings.ToLower(string(resourcev1.DeviceTaintEffectNoExecute)):
				r.cardTaintPolicy = resourcev1.DeviceTaintEffectNoExecute
			case strings.ToLower(string(resourcev1.DeviceTaintEffectNoSchedule)):
				r.cardTaintPolicy = resourcev1.DeviceTaintEffectNoSchedule
			default:
				return fmt.Errorf("expected NoExecute or NoSchedule, but got %s for %s", taintPolicy, deviceTaintRulePolicyEnvVarName)
			}
		} else {
			r.cardTaintPolicy = resourcev1.DeviceTaintEffectNoExecute // default is NoExecute
		}
	}

	// Get TTL settings from environment variable.
	isEnabled, ttlSeconds, err := utils.GetTTLConfigFromEnv(cardAlertTTLEnvVarName)
	if err != nil {
		return fmt.Errorf("failed to get TTL configuration from environment variable: %w", err)
	}
	r.enableTTLSecondsAfterFinished = isEnabled
	r.ttlSecondsAfterFinished = ttlSeconds

	// Print settings got from environment.
	if r.enableAutoReplace {
		logger.Info("Auto replacement is enabled", "machineAlertNamespace", r.MachineAlertNamespace)
	} else {
		logger.Info("Auto replacement is disabled", "taintPolicy", r.cardTaintPolicy)
	}
	if r.enableTTLSecondsAfterFinished {
		logger.Info("TTL for CardAlert resource", "ttlSeconds", r.ttlSecondsAfterFinished)
	} else {
		logger.Info("TTL for CardAlert resource", "ttlSeconds", "disabled")
	}

	// Check client for management cluster.
	if r.ClientHelper == nil {
		// This case is only for failsafe.
		return errors.New("expected ClientHelper is not nil, but got nil")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&scalev1.CardAlert{}).
		Named("cardalert").
		Owns(&scalev1.CardScalingRequest{}).
		Complete(r)
}

// doReconcile is a function to reconcile CardAlert.
// It calls state-specific reconcile functions based on the current state of the CardAlert.
func (r *CardAlertReconciler) doReconcile(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// lastCA is CardAlert that is stored in the etcd.
	// After calling Update() or Patch(), lastCA should be updated.
	lastCA := ca.DeepCopy()

	// Reconcile CardAlert.
	var rslt ctrl.Result
	var err error
	var eventReason = eventReasonUnexpectedError
	switch ca.Status.State {
	case scalev1.CardAlertInitialState, scalev1.CardAlertPendingState:
		rslt, err = r.reconcileInitial(ctx, ca, lastCA)
		eventReason = eventReasonInitializingFailed

	case scalev1.CardAlertInitializingState:
		rslt, err = r.reconcileInitializing(ctx, ca)
		eventReason = eventReasonInitializingFailed

	case scalev1.CardAlertCreatingReplaceResourceState:
		rslt, err = r.reconcileCreatingReplaceResource(ctx, ca)
		eventReason = eventReasonAutoReplacingFailed

	case scalev1.CardAlertWaitingCardScalingRequestState:
		rslt, err = r.reconcileWaitingCardScalingRequest(ctx, ca)
		eventReason = eventReasonAutoReplacingFailed

	case scalev1.CardAlertWaitingMachineAlertState:
		rslt, err = r.reconcileWaitingMachineAlert(ctx, ca)
		eventReason = eventReasonAutoReplacingFailed

	case scalev1.CardAlertDeletingMachineAlertState:
		rslt, err = r.reconcileDeletingMachineAlert(ctx, ca)
		eventReason = eventReasonAutoReplacingFailed

	case scalev1.CardAlertCreatingTaintResourceState:
		rslt, err = r.reconcileCreatingTaintResource(ctx, ca)
		eventReason = eventReasonDeviceTaintingFailed

	case scalev1.CardAlertCompletedState, scalev1.CardAlertFailedState:
		rslt, err = r.reconcileCompleted(ctx, ca)
		eventReason = eventReasonTerminationFailed

	default:
		// This case is only for failsafe, but should not happen because crd validation is checked.
		err = fmt.Errorf("unexpected state(%s) is detected in doReconcile", lastCA.Status.State)
	}

	// Patch status of CardAlert.
	if statusErr := r.updateCardAlertStatus(ctx, ca, lastCA); statusErr != nil {
		rslt = ctrl.Result{} // if error exists, ctrl.Result should be zero value
		statusErr = fmt.Errorf("failed to update CardAlert status: %w", statusErr)
		if err == nil {
			err = statusErr
		} else {
			err = fmt.Errorf("%w: %w", statusErr, err)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to execute Reconcile process for state '%s': %w", lastCA.Status.State, err)
		r.Recorder.Event(ca, corev1.EventTypeWarning, eventReason, err.Error())
	}

	return rslt, err
}

// updateCardAlertStatus is a function to update status of CardAlert and synchronize
// cached resource version of CardAlert.
func (r *CardAlertReconciler) updateCardAlertStatus(
	ctx context.Context, ca, lastCA *scalev1.CardAlert,
) error {
	// Patch status of CardAlert
	if err := r.Status().Patch(ctx, ca, client.MergeFrom(lastCA)); err != nil {
		return fmt.Errorf("failed to patch CardAlert status: %w", err)
	}

	if syncErr := r.syncCachedResourceVersion(ctx, ca); syncErr != nil {
		return fmt.Errorf("failed to synchronize cached resource version of CardAlert %s/%s: %w",
			ca.Namespace, ca.Name, syncErr)
	}

	return nil
}

// syncCachedResourceVersion is a function to synchronize cached resource version of CardAlert
// after updating status of CardAlert.
func (r *CardAlertReconciler) syncCachedResourceVersion(
	ctx context.Context, ca *scalev1.CardAlert,
) error {
	return syncCachedResourceVersion(ctx, r.Client, ca, &scalev1.CardAlert{})
}

// reconcileInitial is a function to reconcile CardAlert in Initial.
// It adds finalizer and updates state to Initializing.
func (r *CardAlertReconciler) reconcileInitial(
	ctx context.Context, ca *scalev1.CardAlert, lastCA *scalev1.CardAlert,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to handle CardAlert")

	// Add finalizer for CardAlert.
	if !controllerutil.ContainsFinalizer(ca, scalev1.CardAlertFinalizerName) {
		controllerutil.AddFinalizer(ca, scalev1.CardAlertFinalizerName)
		if err := r.Patch(ctx, ca, client.MergeFrom(lastCA)); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer to CardAlert: %w", err)
		}
		// Update lastCA after Patch.
		ca.DeepCopyInto(lastCA)
	}

	// Check if there are any CardAlerts is processing.
	processingCAs, err := r.getProcessingCardAlerts(ctx, ca)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get other running CardAlerts: %w", err)
	}
	if len(processingCAs) > 0 {
		// Other processing CardAlerts exist, update state to Pending.
		ca.Status.State = scalev1.CardAlertPendingState
		r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonPending,
			"Pending this CardAlert until another one is completed")
		return ctrl.Result{RequeueAfter: requeueIntervalCardAlertPending}, nil
	}

	// Update state to Initializing.
	ca.Status.State = scalev1.CardAlertInitializingState

	return ctrl.Result{}, nil
}

// getProcessingCardAlerts is a function to get processing CardAlerts except caExcept.
//
//nolint:dupl // similar code is used in getProcessingCardScalingRequests
func (r *CardAlertReconciler) getProcessingCardAlerts(
	ctx context.Context, caExcept *scalev1.CardAlert,
) ([]scalev1.CardAlert, error) {
	// List CardAlerts from all namespaces.
	caList := &scalev1.CardAlertList{}
	if err := r.List(ctx, caList); err != nil {
		return nil, fmt.Errorf("failed to list CardAlerts: %w", err)
	}

	// Filter out the current CardAlert and those that are not in a running state.
	var processingCAs []scalev1.CardAlert
	for idx := range caList.Items {
		ca := &caList.Items[idx]
		if caExcept != nil && ca.Name == caExcept.Name && ca.Namespace == caExcept.Namespace {
			continue // skip self resource
		}
		if ca.IsProcessing() {
			processingCAs = append(processingCAs, *ca)
		}
	}

	return processingCAs, nil
}

// reconcileInitializing is a function to reconcile CardAlert in Initializing.
// It checks whether another CardAlert is being processed and updates state to
// CreatingReplaceResource or Pending.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) reconcileInitializing(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// Get Node.
	node, err := r.ClientHelper.GetNode(ctx, ca.Spec.NodeName)
	if apierrors.IsNotFound(err) {
		ca.Status.CompletedTime = metav1.NewTime(time.Now())
		ca.Status.State = scalev1.CardAlertCompletedState
		logger := log.FromContext(ctx)
		logger.Info("CardAlert is completed because there is already no target Node", "NodeName", ca.Spec.NodeName)
		r.Recorder.Eventf(ca, corev1.EventTypeNormal, eventReasonCompleted,
			"CardAlert is completed because there is already no target Node(%s)", ca.Spec.NodeName)
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Node: %w", err)
	}

	// Get namespace for cluster.
	_, clusterNamespace, err := r.ClientHelper.GetClusterNameAndNamespaceFromNode(node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Cluster Namespace from Node: %w", err)
	}

	// Get Machine from Node.
	machine, err := r.ClientHelper.GetMachineFromNode(ctx, node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Machine from Node: %w", err)
	}

	// Get DCIMachine from Machine.
	dciMachine, err := r.ClientHelper.GetDCIMachineFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachine from Machine: %w", err)
	}

	// Get LogicalServer from DCIMachine.
	lsName, lsNamespace, err := r.ClientHelper.GetLogicalServerNameAndNamespaceFromDCIMachine(dciMachine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get LogicalServer name and namespace from DCIMachine: %w", err)
	}

	// Set LogicalServer reference to status of CardAlert.
	ca.Status.LogicalServerRef = corev1.ObjectReference{
		APIVersion: rpmv1.GroupVersion.String(),
		Kind:       "LogicalServer",
		Name:       lsName,
		Namespace:  lsNamespace,
	}

	// Switch next state based on auto replacement setting.
	if r.enableAutoReplace {
		ca.Status.State = scalev1.CardAlertCreatingReplaceResourceState
	} else {
		ca.Status.State = scalev1.CardAlertCreatingTaintResourceState
	}

	return ctrl.Result{}, nil
}

// reconcileCreatingReplaceResource is a function to reconcile CardAlert in CreatingReplaceResource.
// It checks whether the target Node has the target GPUs and creates CardScalingRequest or MachineAlert
// for auto replacement based on the type of LogicalServer. If there is no target GPU in the Node,
// it updates the state to Completed without creating any resource for replacement.
// And updates the state to WaitingCardScalingRequest or WaitingMachineAlert after creating the resource for replacement.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) reconcileCreatingReplaceResource(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to create replace resource for CardAlert")

	// Get LogicalServer from CardAlert.
	lsRef := &ca.Status.LogicalServerRef
	ls, err := r.ClientHelper.GetLogicalServer(ctx, lsRef.Name, lsRef.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get LogicalServer: %w", err)
	}
	lsRef.UID = ls.UID

	// Check GPUs in the LogicalServer status.
	targets, skip := classifyCards(ca.Spec.Cards, ls)
	ca.Status.TargetCards = targets
	ca.Status.SkipCards = skip
	ca.Status.ServerType = ls.Status.AppliedSpec.Type

	// If there is no GPU in the LogicalServer status, it means that the card issue is resolved during the processing,.
	// so we can complete the CardAlert without creating any resource for replacement.
	if len(ca.Status.TargetCards) == 0 {
		// Update state to Completed.
		ca.Status.CompletedTime = metav1.NewTime(time.Now())
		ca.Status.State = scalev1.CardAlertCompletedState
		logger.Info("CardAlert is completed because there is no target card in LogicalServer")
		r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonCompleted,
			"CardAlert is completed because there is no target card in LogicalServer")
		return ctrl.Result{}, nil
	}
	ca.Status.CardModel = targets[0].Model // assume all target cards have the same model

	// Switch next state based on the type of the LogicalServer.
	switch ca.Status.ServerType {
	case rpmv1.AppliedServerTypeComposed:
		logger.Info("Start to create CardScalingRequest for auto replacement")
		// Create CardScalingRequest.
		if err := r.createCardScalingRequest(ctx, ca); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create CardScalingRequest: %w", err)
		}
		// Update state to WaitingCardScalingRequest.
		ca.Status.State = scalev1.CardAlertWaitingCardScalingRequestState
		r.Recorder.Eventf(ca, corev1.EventTypeNormal, eventReasonCardScalingRequestCreated,
			"Created CardScalingRequest(%s/%s) for auto replacement", ca.Status.ResourceRef.Namespace, ca.Status.ResourceRef.Name)
		return ctrl.Result{}, nil

	case rpmv1.AppliedServerTypeCOTS:
		logger.Info("Start to create MachineAlert for auto replacement")
		// Create MachineAlert in management cluster to trigger auto replacement by external controller.
		if err := r.createMachineAlert(ctx, ca); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create MachineAlert: %w", err)
		}
		// Update state to WaitingMachineAlert.
		ca.Status.State = scalev1.CardAlertWaitingMachineAlertState
		r.Recorder.Eventf(ca, corev1.EventTypeNormal, eventReasonMachineAlertCreated,
			"Created MachineAlert(%s/%s) for auto replacement", ca.Status.ResourceRef.Namespace, ca.Status.ResourceRef.Name)
		return ctrl.Result{}, nil

	case rpmv1.AppliedServerTypeVM:
		// Update state to Failed.
		err := errors.New("node created by VM is not supported for auto replacement")
		ca.Status.State = scalev1.CardAlertFailedState
		r.Recorder.Event(ca, corev1.EventTypeWarning, eventReasonVMTypeNotSupported,
			"CardAlert is failed because VM type LogicalServer is not supported for auto replacement")
		return ctrl.Result{}, err

	default:
		return ctrl.Result{}, fmt.Errorf("unexpected server type in LogicalServer: %s", ca.Status.ServerType)
	}
}

func classifyCards(
	cards []scalev1.CardInfo, ls *rpmv1.LogicalServer,
) (targets, skips []scalev1.CardInfo) {
	for _, gpu := range cards {
		if slices.ContainsFunc(ls.Spec.ServerSpec.GPU.Assignments, func(assign rpmv1.GPUAssignment) bool {
			return assign.UUID == gpu.GPUUUID
		}) {
			// attempt to create CardScalingRequest by these UUIDs
			targets = append(targets, gpu)
		} else {
			// skip uuids not in the target Node(LogicalServer)
			skips = append(skips, gpu)
		}
	}
	return
}

func (r *CardAlertReconciler) createCardScalingRequest(
	ctx context.Context, ca *scalev1.CardAlert,
) error {
	// List existing CardScalingRequest.
	csrList := &scalev1.CardScalingRequestList{}
	if err := r.List(ctx, csrList, client.InNamespace(ca.Namespace)); err != nil {
		return fmt.Errorf("failed to list CardScalingRequests: %w", err)
	}

	// Check if there is already a CardScalingRequest for the CardAlert.
	var csrExisting *scalev1.CardScalingRequest
	for idx := range csrList.Items {
		csr := &csrList.Items[idx]
		if metav1.IsControlledBy(csr, ca) {
			csrExisting = csr
			break
		}
	}

	if csrExisting == nil {
		// create uuids string slice from target cards in CardAlert status
		uuids := make([]string, len(ca.Status.TargetCards))
		for i, card := range ca.Status.TargetCards {
			uuids[i] = card.GPUUUID
		}
		// Create CardScalingRequest.
		csr := &scalev1.CardScalingRequest{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", ca.Name),
				Namespace:    ca.Namespace,
			},
			Spec: scalev1.CardScalingRequestSpec{
				NodeName: ca.Spec.NodeName,
				ScalingParameter: scalev1.ScalingParameter{
					Op:    capdv1.DCIMachineAnnotationCardScalingInfoOpReplace,
					Model: ca.Status.CardModel,
					UUIDs: uuids, // if nil or empty, all GPUs in the Node will be replaced
				},
			},
		}
		// Set owner reference to the CardAlert.
		if err := controllerutil.SetControllerReference(ca, csr, r.Client.Scheme()); err != nil {
			return fmt.Errorf("failed to set controller reference to CardScalingRequest: %w", err)
		}
		// Create CardScalingRequest in this cluster.
		if err := r.Create(ctx, csr); err != nil {
			return fmt.Errorf("failed to create CardScalingRequest: %w", err)
		}
		// Set reference of the newly created CardScalingRequest to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: scalev1.GroupVersion.String(),
			Kind:       "CardScalingRequest",
			Name:       csr.Name,
			Namespace:  csr.Namespace,
			UID:        csr.UID,
		}
	} else {
		// Set reference of the existing CardScalingRequest to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: scalev1.GroupVersion.String(),
			Kind:       "CardScalingRequest",
			Name:       csrExisting.Name,
			Namespace:  csrExisting.Namespace,
			UID:        csrExisting.UID,
		}
	}

	return nil
}

func (r *CardAlertReconciler) createMachineAlert(
	ctx context.Context, ca *scalev1.CardAlert,
) error {
	// List existing MachineAlerts.
	maList, err := r.ClientHelper.ListMachineAlert(ctx, r.MachineAlertNamespace)
	if err != nil {
		return fmt.Errorf("failed to list MachineAlerts: %w", err)
	}

	// Check if there is already a MachineAlert for the CardAlert.
	var maExisting *rpmv1.MachineAlert
	for idx := range maList.Items {
		ma := &maList.Items[idx]
		if utils.IsSameCardAlertCreatedByAnnotations(ma.Annotations, ca.Name, ca.Namespace) {
			maExisting = ma
			break
		}
	}

	if maExisting == nil {
		// Create MachineAlert.
		ma := &rpmv1.MachineAlert{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", ca.Name),
				Namespace:    r.MachineAlertNamespace,
				Annotations: map[string]string{
					scalev1.CreatedByNameLabel:      ca.Name,
					scalev1.CreatedByNamespaceLabel: ca.Namespace,
				},
			},
			Spec: rpmv1.MachineAlertSpec{
				ServerRef: corev1.ObjectReference{
					APIVersion: rpmv1.GroupVersion.String(),
					Kind:       "LogicalServer",
					Name:       ca.Status.LogicalServerRef.Name,
					Namespace:  ca.Status.LogicalServerRef.Namespace,
					UID:        ca.Status.LogicalServerRef.UID,
				},
				Source: rpmv1.MachineAlertSourceCardAlertOnTenant,
				Reason: ca.Spec.Reason,
			},
		}
		// Create MachineAlert in this cluster.
		if err := r.ClientHelper.CreateMachineAlert(ctx, ma); err != nil {
			return fmt.Errorf("failed to create MachineAlert: %w", err)
		}
		// Set reference of the newly created MachineAlert to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: rpmv1.GroupVersion.String(),
			Kind:       "MachineAlert",
			Name:       ma.Name,
			Namespace:  ma.Namespace,
			UID:        ma.UID,
		}
	} else {
		// Set reference of the existing MachineAlert to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: rpmv1.GroupVersion.String(),
			Kind:       "MachineAlert",
			Name:       maExisting.Name,
			Namespace:  maExisting.Namespace,
			UID:        maExisting.UID,
		}
	}

	return nil
}

// reconcileWaitingCardScalingRequest is a function to reconcile CardAlert in WaitingCardScalingRequest.
// It checks whether the CardScalingRequest is completed or failed by checking the status of the CardScalingRequest,
// and updates the state of CardAlert to Completed or Failed accordingly.
// If the CardScalingRequest is still in progress, it does nothing and expects to be requeued by the event of
// CardScalingRequest status update.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) reconcileWaitingCardScalingRequest(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get CardScalingRequest for the CardAlert.
	csrRef := &ca.Status.ResourceRef
	csr := &scalev1.CardScalingRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: csrRef.Name, Namespace: csrRef.Namespace}, csr); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get CardScalingRequest: %w", err)
	}

	switch csr.Status.Phase {
	case scalev1.PhaseCardScaleCompleted:
		// Update state to Completed.
		ca.Status.CompletedTime = metav1.NewTime(time.Now())
		ca.Status.State = scalev1.CardAlertCompletedState
		logger.Info("CardAlert is completed with successful auto replacement by CardScalingRequest")
		r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonCompleted,
			"CardAlert is completed with successful auto replacement by CardScalingRequest")
		return ctrl.Result{}, nil

	case scalev1.PhaseNodeScaleDownFailed, scalev1.PhaseNodeScaleUpFailed:
		// Update state to Failed.
		err := fmt.Errorf("failed CardAlert because auto replacement by CardScalingRequest failed with phase: %s", csr.Status.Phase)
		ca.Status.State = scalev1.CardAlertFailedState
		r.Recorder.Event(ca, corev1.EventTypeWarning, eventReasonCardScalingRequestFailed, err.Error())
		return ctrl.Result{}, err

	default:
		// do nothing, expect requeue by event of CardScalingRequest status update
		return ctrl.Result{}, nil
	}
}

// reconcileWaitingMachineAlert is a function to reconcile CardAlert in WaitingMachineAlert.
// It checks whether the MachineAlert is completed or not, and updates the state of CardAlert accordingly.
func (r *CardAlertReconciler) reconcileWaitingMachineAlert(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// Get MachineAlert for the CardAlert.
	maRef := &ca.Status.ResourceRef
	ma, err := r.ClientHelper.GetMachineAlert(ctx, maRef.Name, maRef.Namespace)
	if apierrors.IsNotFound(err) {
		// It may be already completed and deleted by TTL, so we can complete the CardAlert,
		// and we will pass DeletingMachineAlert state to simplify the flow.
		ca.Status.State = scalev1.CardAlertDeletingMachineAlertState
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get MachineAlert: %w", err)
	}

	switch ma.Status.State {
	case rpmv1.MachineAlertStateCompleted:
		// Update state to DeletingMachineAlert.
		ca.Status.State = scalev1.CardAlertDeletingMachineAlertState
		return ctrl.Result{}, nil

	default:
		// Wait for the MachineAlert to be completed, requeue after a certain interval to check the status again,
		// and failed state does not exist in MachineAlert, so we do not check the failed state here.
		return ctrl.Result{RequeueAfter: pollingIntervalCheckMachineAlertCompletion}, nil // requeue by requeueAfter
	}
}

// reconcileDeletingMachineAlert is a function to reconcile CardAlert in DeletingMachineAlert.
// It checks whether the MachineAlert is deleted or not, and updates the state of CardAlert accordingly.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) reconcileDeletingMachineAlert(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get MachineAlert for the CardAlert.
	maRef := &ca.Status.ResourceRef
	ma, err := r.ClientHelper.GetMachineAlert(ctx, maRef.Name, maRef.Namespace)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to get MachineAlert for deletion: %w", err)
	}

	// Delete MachineAlert.
	if !apierrors.IsNotFound(err) {
		if err := r.ClientHelper.DeleteMachineAlert(ctx, ma); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to delete MachineAlert: %w", err)
		}
	}

	// Update state to Completed.
	ca.Status.CompletedTime = metav1.NewTime(time.Now())
	ca.Status.State = scalev1.CardAlertCompletedState
	logger.Info("CardAlert is completed with successful creation and deletion of MachineAlert for auto replacement")
	r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonCompleted,
		"CardAlert is completed with successful creation and deletion of MachineAlert for auto replacement")

	return ctrl.Result{}, nil
}

// reconcileCreatingTaintResource is a function to reconcile CardAlert in CreatingTaintResource.
// It checks whether there is cards to be tainted in the Node(LogicalServer), and creates DeviceTaintRule
// to taint the cards if there are. If there is no card to be tainted, it updates the state to Completed without
// creating any resource for replacement. And updates the state to Completed after creating the DeviceTaintRule.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) reconcileCreatingTaintResource(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to create DeviceTaintRule for alerted cards")

	// Get LogicalServer from CardAlert.
	lsRef := &ca.Status.LogicalServerRef
	ls, err := r.ClientHelper.GetLogicalServer(ctx, lsRef.Name, lsRef.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get LogicalServer: %w", err)
	}
	lsRef.UID = ls.UID

	// Check GPUs in the LogicalServer status.
	targets, skip := classifyCards(ca.Spec.Cards, ls)
	ca.Status.TargetCards = targets
	ca.Status.SkipCards = skip
	ca.Status.ServerType = ls.Status.AppliedSpec.Type

	// If there is no GPU in the LogicalServer status, it means that the card issue is resolved during the processing,
	// so we can complete the CardAlert without creating any resource for replacement.
	if len(ca.Status.TargetCards) == 0 {
		// Update state to Completed.
		ca.Status.CompletedTime = metav1.NewTime(time.Now())
		ca.Status.State = scalev1.CardAlertCompletedState
		logger.Info("CardAlert is completed because there is no target card in LogicalServer")
		r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonCompleted,
			"CardAlert is completed because there is no target card in LogicalServer")
		return ctrl.Result{}, nil
	}
	ca.Status.CardModel = targets[0].Model // assume all target cards have the same model

	// Create DeviceTaintRule to taint the target Node(LogicalServer) when auto replacement is disabled.
	if err := r.createDeviceTaintRule(ctx, ca); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create DeviceTaintRule: %w", err)
	}

	// Update state to Completed.
	ca.Status.CompletedTime = metav1.NewTime(time.Now())
	ca.Status.State = scalev1.CardAlertCompletedState
	logger.Info("CardAlert is completed with successful creation of DeviceTaintRule")
	r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonCompleted,
		"CardAlert is completed with successful creation of DeviceTaintRule")

	return ctrl.Result{}, nil
}

// createDeviceTaintRule is a function to create DeviceTaintRule to taint the target Cards.
func (r *CardAlertReconciler) createDeviceTaintRule(
	ctx context.Context, ca *scalev1.CardAlert,
) error {
	// List existing DeviceTaintRules.
	ruleList, err := r.ClientHelper.ListDeviceTaintRule(ctx, ca.Namespace)
	if err != nil {
		return fmt.Errorf("failed to list DeviceTaintRules: %w", err)
	}

	// Check if there is already a DeviceTaintRule for the CardAlert.
	var ruleExisting *resourcev1.DeviceTaintRule
	for idx := range ruleList.Items {
		rule := &ruleList.Items[idx]
		if utils.IsSameCardAlertCreatedByAnnotations(rule.Annotations, ca.Name, ca.Namespace) {
			ruleExisting = rule
			break
		}
	}

	if ruleExisting == nil {
		// Create DeviceTaintRule.
		rule := &resourcev1.DeviceTaintRule{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", ca.Name),
				Annotations: map[string]string{
					scalev1.CreatedByNameLabel:      ca.Name,
					scalev1.CreatedByNamespaceLabel: ca.Namespace,
				},
			},
			Spec: resourcev1.DeviceTaintRuleSpec{
				DeviceSelector: &resourcev1.DeviceTaintSelector{
					Selectors: []resourcev1.DeviceSelector{
						{
							CEL: &resourcev1.CELDeviceSelector{
								Expression: createCELExpression(ca.Status.TargetCards),
							},
						},
					},
				},
				Taint: resourcev1.DeviceTaint{
					Key:    deviceTaintRuleKeyUnhealthy,
					Value:  deviceTaintRuleValueBroken,
					Effect: r.cardTaintPolicy,
				},
			},
		}
		if err := r.ClientHelper.CreateDeviceTaintRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to create DeviceTaintRule: %w", err)
		}

		// Set reference of the created DeviceTaintRule to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: resourcev1.SchemeGroupVersion.String(),
			Kind:       "DeviceTaintRule",
			Name:       rule.Name,
			Namespace:  rule.Namespace,
			UID:        rule.UID,
		}
	} else {
		// Set reference of the existing DeviceTaintRule to status of CardAlert.
		ca.Status.ResourceRef = corev1.ObjectReference{
			APIVersion: resourcev1.SchemeGroupVersion.String(),
			Kind:       "DeviceTaintRule",
			Name:       ruleExisting.Name,
			Namespace:  ruleExisting.Namespace,
			UID:        ruleExisting.UID,
		}
	}

	return nil
}

// createCELExpression is a function to create CEL expression for DeviceSelector in DeviceTaintRule
// based on the target cards in CardAlert status.
//
// See https://github.com/kubernetes/api/blob/v0.34.2/resource/v1alpha3/types.go#L33.
func createCELExpression(
	cards []scalev1.CardInfo,
) string {
	var cel string

	for i, card := range cards {
		if len(card.MIGUUIDs) > 0 {
			cel += fmt.Sprintf("%s.%s=='%s'", deviceTaintRuleDevice, deviceTaintRuleCELPathParentUUID, card.GPUUUID)
		} else {
			cel += fmt.Sprintf("%s.%s=='%s'", deviceTaintRuleDevice, deviceTaintRuleCELPathUUID, card.GPUUUID)
		}
		if i != len(cards)-1 {
			cel += "||"
		}
	}

	if cel == "" {
		return ""
	}

	return fmt.Sprintf("cel.bind(%s, device.attributes['%s'], %s)", deviceTaintRuleDevice, deviceTaintRuleCELLabelGPUNvidia, cel)
}

// reconcileCompleted is a function to reconcile CardAlert in Completed.
// It checks whether the CardAlert should be deleted based on the TTL setting, and deletes the CardAlert if it is expired.
func (r *CardAlertReconciler) reconcileCompleted(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	if !r.enableTTLSecondsAfterFinished || len(ca.OwnerReferences) != 0 || ca.IsFailed() {
		// Skip this function if TTL is disabled, CardAlert has owner references, or CardAlert is in failed state.
		// It should not delete automatically CardAlert with failure states to notify users about the failure.
		return ctrl.Result{}, nil
	}

	if ca.Status.CompletedTime.IsZero() {
		// CompletedTime should be set when state is updated to Completed in correct flow,
		// so return error if CompletedTime is not set to avoid deleting CardAlert.
		return ctrl.Result{}, fmt.Errorf("expected .status.completedTime is set when state is Completed, but not set")
	}
	timeToExpire := ca.Status.CompletedTime.Add(time.Duration(r.ttlSecondsAfterFinished) * time.Second)

	// Check if TTL is expired.
	now := time.Now()
	if !now.After(timeToExpire) {
		return ctrl.Result{RequeueAfter: timeToExpire.Sub(now)}, nil // recalculate remain seconds
	}

	// Delete expired CardAlert.
	logger := log.FromContext(ctx)
	if err := r.Delete(ctx, ca); apierrors.IsNotFound(err) {
		logger.Info("Expired CardAlert is already deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete expired CardAlert: %w", err)
	}

	logger.Info("Expired CardAlert is deleted successfully")
	r.Recorder.Event(ca, corev1.EventTypeNormal, eventReasonExpired, "CardAlert is deleted after TTL expiration")
	return ctrl.Result{}, nil
}

// doFinalize is a function to finalize CardAlert.
// It calls state-specific finalize functions based on the current state of the CardAlert.
func (r *CardAlertReconciler) doFinalize(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// lastCA is CardAlert that is stored in the etcd.
	// After calling Update() or Patch(), lastCA should be updated.
	lastCA := ca.DeepCopy()

	// Finalize CardAlert.
	var rslt ctrl.Result
	var err error
	var eventReason = eventReasonUnexpectedError
	switch ca.Status.State {
	case scalev1.CardAlertInitialState, scalev1.CardAlertInitializingState, scalev1.CardAlertPendingState,
		scalev1.CardAlertWaitingCardScalingRequestState, scalev1.CardAlertCreatingTaintResourceState,
		scalev1.CardAlertCompletedState, scalev1.CardAlertFailedState:
		// Update to Finalized.
		rslt, err = r.updateFinalized(ca)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.CardAlertWaitingMachineAlertState, scalev1.CardAlertDeletingMachineAlertState:
		// Clean up MachineAlert and update to Finalized.
		rslt, err = r.cleanupMachineAlert(ctx, ca)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.CardAlertCreatingReplaceResourceState:
		// Clean up MachineAlert when server is COTS server and update to Finalized.
		rslt, err = r.finalizeCreatingReplaceResource(ctx, ca)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.CardAlertFinalizedState:
		rslt, err = r.finalizeFinalized(ctx, ca)
		// finalizeFinalized removes finalizer from CardAlert,
		// so Patch of status is not needed after calling it when err is nil.
		if err == nil {
			return rslt, nil
		}
		eventReason = eventReasonFinalizerRemovingFinalizerFailed

	default:
		// This case is only for failsafe, but should not happen because crd validation is checked.
		err = fmt.Errorf("unexpected state(%s) is detected in doFinalize", lastCA.Status.State)
	}

	// Patch status of CardAlert.
	if patchErr := r.Status().Patch(ctx, ca, client.MergeFrom(lastCA)); patchErr != nil {
		rslt = ctrl.Result{} // if error exists, ctrl.Result should be zero value
		patchErr = fmt.Errorf("failed to patch CardAlert status: %w", patchErr)
		if err == nil {
			err = patchErr
		} else {
			err = fmt.Errorf("%w: %w", patchErr, err)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to execute Finalizer process for state '%s': %w", lastCA.Status.State, err)
		r.Recorder.Event(ca, corev1.EventTypeWarning, eventReason, err.Error())
	}

	return rslt, err
}

// updateFinalized is a function to update the state of CardAlert to Finalized.
// It updates state to Finalized.
func (*CardAlertReconciler) updateFinalized(
	ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// Update state to Finalized.
	ca.Status.State = scalev1.CardAlertFinalizedState

	return ctrl.Result{}, nil
}

// cleanupMachineAlert is a function to finalize CardAlert.
// It deletes the associated MachineAlert if it exists and updates the state to Finalized.
func (r *CardAlertReconciler) cleanupMachineAlert(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	var maToDelete *rpmv1.MachineAlert
	if ca.Status.ResourceRef.Kind == "MachineAlert" {
		ma, err := r.ClientHelper.GetMachineAlert(ctx, ca.Status.ResourceRef.Name, ca.Status.ResourceRef.Namespace)
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to get MachineAlert: %w", err)
		}
		if !apierrors.IsNotFound(err) {
			maToDelete = ma
		}
	} else {
		// This case if for failsafe when MachineAlert is created but reference is not set
		// in status of CardAlert due to any error.
		maList, err := r.ClientHelper.ListMachineAlert(ctx, r.MachineAlertNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to list MachineAlerts: %w", err)
		}
		for idx := range maList.Items {
			ma := &maList.Items[idx]
			if utils.IsSameCardAlertCreatedByAnnotations(ma.Annotations, ca.Name, ca.Namespace) {
				maToDelete = ma
				break
			}
		}
	}

	// Delete MachineAlert if exists.
	if maToDelete != nil {
		if err := r.ClientHelper.DeleteMachineAlert(ctx, maToDelete); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to delete MachineAlert: %w", err)
		}
	}

	return r.updateFinalized(ca)
}

// finalizeCreatingReplaceResource is a function to finalize CardAlert in CreatingReplaceResource.
// It checks the type of server by the LogicalServer, and if it is COTS server, it deletes the associated MachineAlert if it exists,
// and updates the state to Finalized. If it is not COTS server, it just updates the state to Finalized.
func (r *CardAlertReconciler) finalizeCreatingReplaceResource(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	lsRef := &ca.Status.LogicalServerRef
	ls, err := r.ClientHelper.GetLogicalServer(ctx, lsRef.Name, lsRef.Namespace)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to get LogicalServer: %w", err)
	}

	if ls == nil || ls.Status.AppliedSpec == nil || ls.Status.AppliedSpec.Type != rpmv1.AppliedServerTypeCOTS {
		// do nothing if the server type is not COTS server or the server does not exist
		return r.updateFinalized(ca)
	}

	// Cleanup MachineAlert if the server is COTS server because it should be created in this state.
	return r.cleanupMachineAlert(ctx, ca)
}

// finalizeFinalized is a function to finalize CardAlert in Finalized.
// It removes finalizer from CardAlert.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardAlertReconciler) finalizeFinalized(
	ctx context.Context, ca *scalev1.CardAlert,
) (ctrl.Result, error) {
	// lastCA is CardAlert that is stored in the etcd.
	lastCA := ca.DeepCopy()

	// Remove finalizer from CardAlert.
	controllerutil.RemoveFinalizer(ca, scalev1.CardAlertFinalizerName)
	// Patch CardAlert to remove finalizer.
	if err := r.Patch(ctx, ca, client.MergeFrom(lastCA)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from CardAlert: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("CardAlert is finalized successfully")

	return ctrl.Result{}, nil
}
