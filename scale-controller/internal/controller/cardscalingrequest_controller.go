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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"
)

// CardScalingRequestReconciler reconciles a CardScalingRequest object
type CardScalingRequestReconciler struct {
	// client for workload cluster
	client.Client

	// scheme for workload cluster
	Scheme *runtime.Scheme

	// interface for event recorder
	Recorder record.EventRecorder

	// client for management cluster
	ClientHelper helperclient.ClientHelper

	// setting for enabling TTL
	enableTTLSecondsAfterFinished bool

	// setting for TTL seconds
	ttlSecondsAfterFinished int
}

// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardscalingrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardscalingrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.my.domain,resources=cardscalingrequests/finalizers,verbs=update

// Additional rbac
// +kubebuilder:rbac:groups=cluster.my.domain,resources=nodescalingrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// Reconcile check the finalizer for CardScalingRequest and call doFinalize() if there is a deletion request.
// Otherwise, call doReconcile() to reconcile the CardScalingRequest resource.
func (r *CardScalingRequestReconciler) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	// Get CardScalingRequest
	csr := &scalev1.CardScalingRequest{}
	if err := r.Get(ctx, req.NamespacedName, csr); apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get CardScalingRequest: %s: %w", req.String(), err)
	}

	// Check if the CardScalingRequest is being deleted then finalize it.
	if !csr.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(csr, scalev1.CardScalingRequestFinalizerName) {
			return ctrl.Result{}, nil // finalizer for this reconciler is not set, so we do nothing
		}
		rslt, err := r.doFinalize(ctx, csr)
		if err != nil {
			err = fmt.Errorf("failed to finalize CardScalingRequest: %w", err)
		}
		return rslt, err
	}

	// Reconcile the CardScalingRequest
	rslt, err := r.doReconcile(ctx, csr)
	if err != nil {
		err = fmt.Errorf("failed to reconcile CardScalingRequest: %w", err)
	}
	return rslt, err
}

// SetupWithManager sets up the controller with the Manager.
//   - It also reads TTL configuration for CardScalingRequest from environment variable.
//   - ClientHelper(field of CardScalingRequestReconciler) must be set before calling this function.
func (r *CardScalingRequestReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {
	if mgr == nil {
		// This case is only for failsafe
		return errors.New("expected mgr is not nil, but got nil")
	}
	logger := mgr.GetLogger().WithName("cardscalingrequest-controller-setup")

	// Get information from environment variable
	isEnabled, ttlSeconds, err := utils.GetTTLConfigFromEnv(cardScalingRequestTTLEnvVarName)
	if err != nil {
		return fmt.Errorf("failed to get TTL configuration from environment variable: %w", err)
	}
	r.enableTTLSecondsAfterFinished = isEnabled
	r.ttlSecondsAfterFinished = ttlSeconds

	// Print settings got from environment
	if r.enableTTLSecondsAfterFinished {
		logger.Info("TTL for CardScalingRequest resource", "ttlSeconds", r.ttlSecondsAfterFinished)
	} else {
		logger.Info("TTL for CardScalingRequest resource", "ttlSeconds", "disabled")
	}

	// Check client for management cluster
	if r.ClientHelper == nil {
		// This case is only for failsafe
		return errors.New("expected ClientHelper is not nil, but got nil")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&scalev1.CardScalingRequest{}).
		Owns(&scalev1.NodeScalingRequest{}).
		Named("cardscalingrequest").
		Complete(r)
}

// doReconcile is a function to reconcile CardScalingRequest.
// It calls phase-specific reconcile functions based on the current phase of the CardScalingRequest.
func (r *CardScalingRequestReconciler) doReconcile(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	// lastCSR is CSR that is stored in the etcd.
	// After calling Update() or Patch(), lastCSR should be updated.
	lastCSR := csr.DeepCopy()

	// Reconcile CardScalingRequest
	var rslt ctrl.Result
	var err error
	var eventReason = eventReasonUnexpectedError
	switch csr.Status.Phase {
	case scalev1.PhaseCardScaleInitial, scalev1.PhaseCardScalePending:
		rslt, err = r.reconcileCardScaleInitial(ctx, csr, lastCSR)
		eventReason = eventReasonInitializingFailed

	case scalev1.PhaseCardScaleInitializing:
		rslt, err = r.reconcileCardScaleInitializing(ctx, csr)
		eventReason = eventReasonInitializingFailed

	case scalev1.PhaseDCIMachineTemplateCreating:
		rslt, err = r.reconcileDCIMachineTemplateCreating(ctx, csr)
		eventReason = eventReasonPreparingResourceFailed

	case scalev1.PhaseMachineDeploymentCreating:
		rslt, err = r.reconcileMachineDeploymentCreating(ctx, csr)
		eventReason = eventReasonPreparingResourceFailed

	case scalev1.PhaseDCIMachineAnnotating:
		rslt, err = r.reconcileDCIMachineAnnotating(ctx, csr)
		eventReason = eventReasonPreparingResourceFailed

	case scalev1.PhaseCardScaleInitialized:
		rslt, err = r.reconcileCardScaleInitialized(ctx, csr)
		eventReason = eventReasonScalingDownFailed

	case scalev1.PhaseNodeScalingDown:
		rslt, err = r.reconcileNodeScalingDown(ctx, csr)
		eventReason = eventReasonScalingDownFailed

	case scalev1.PhaseNodeScaleDownDeleting:
		rslt, err = r.reconcileNodeScaleDownDeleting(ctx, csr)
		eventReason = eventReasonScalingDownFailed

	case scalev1.PhaseNodeScaleDownCompleted:
		rslt, err = r.reconcileNodeScaleDownCompleted(ctx, csr)
		eventReason = eventReasonScalingUpFailed

	case scalev1.PhaseNodeScalingUp:
		rslt, err = r.reconcileNodeScalingUp(ctx, csr)
		eventReason = eventReasonScalingUpFailed

	case scalev1.PhaseCardScaleCompleted, scalev1.PhaseNodeScaleDownFailed, scalev1.PhaseNodeScaleUpFailed:
		rslt, err = r.reconcileTerminalPhase(ctx, csr)
		eventReason = eventReasonTerminationFailed

	default:
		// This case is only for failsafe, but should not happen because crd validation is checked.
		err = fmt.Errorf("unexpected phase(%s) is detected in doReconcile", lastCSR.Status.Phase)
	}

	// Patch status of CardScalingRequest
	if statusErr := r.updateCardScalingRequestStatus(ctx, csr, lastCSR); statusErr != nil {
		statusErr = fmt.Errorf("failed to update CardScalingRequest status: %w", statusErr)
		rslt = ctrl.Result{} // override result to not requeue
		if err == nil {
			err = statusErr
		} else {
			err = fmt.Errorf("%w: %w", statusErr, err)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to execute Reconcile process for state '%s': %w", lastCSR.Status.Phase, err)
		r.Recorder.Event(csr, corev1.EventTypeWarning, eventReason, err.Error())
	}

	return rslt, err
}

// updateCardScalingRequestStatus is a function to update status of CardScalingRequest and synchronize
// cached resource version of CardScalingRequest.
func (r *CardScalingRequestReconciler) updateCardScalingRequestStatus(
	ctx context.Context, csr, lastCSR *scalev1.CardScalingRequest,
) error {
	// Patch status of CardScalingRequest
	if err := r.Status().Patch(ctx, csr, client.MergeFrom(lastCSR)); err != nil {
		return fmt.Errorf("failed to patch CardScalingRequest status: %w", err)
	}

	if syncErr := r.syncCachedResourceVersion(ctx, csr); syncErr != nil {
		return fmt.Errorf("failed to synchronize cached resource version of CardScalingRequest %s/%s: %w",
			csr.Namespace, csr.Name, syncErr)
	}

	return nil
}

// syncCachedResourceVersion is a function to synchronize cached resource version of CardScalingRequest
// after updating status of CardScalingRequest.
func (r *CardScalingRequestReconciler) syncCachedResourceVersion(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) error {
	return syncCachedResourceVersion(ctx, r.Client, csr, &scalev1.CardScalingRequest{})
}

// getProcessingCardScalingRequests is a function to get processing CardScalingRequests except csrExcept.
//
//nolint:dupl // similar code is used in getProcessingCardAlerts
func (r *CardScalingRequestReconciler) getProcessingCardScalingRequests(
	ctx context.Context, csrExcept *scalev1.CardScalingRequest,
) ([]scalev1.CardScalingRequest, error) {
	// List CardScalingRequests
	csrList := &scalev1.CardScalingRequestList{}
	if err := r.List(ctx, csrList); err != nil { // List CardScalingRequests from all namespaces
		return nil, fmt.Errorf("failed to list CardScalingRequests: %w", err)
	}

	// Filter out the current CardScalingRequest and those that are not in a running phase
	var processingCSRs []scalev1.CardScalingRequest
	for idx := range csrList.Items {
		csr := &csrList.Items[idx]
		if csrExcept != nil && csr.Name == csrExcept.Name && csr.Namespace == csrExcept.Namespace {
			continue // skip self resource
		}
		if csr.IsProcessing() {
			processingCSRs = append(processingCSRs, *csr)
		}
	}

	return processingCSRs, nil
}

// reconcileCardScaleInitial is a function to reconcile CardScalingRequest in PhaseCardScaleInitial.
// It adds finalizer and updates phase to PhaseCardScaleInitializing.
func (r *CardScalingRequestReconciler) reconcileCardScaleInitial(
	ctx context.Context, csr, lastCSR *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to initialize CardScalingRequest")

	// Add finalizer for CardScalingRequest
	if !controllerutil.ContainsFinalizer(csr, scalev1.CardScalingRequestFinalizerName) {
		controllerutil.AddFinalizer(csr, scalev1.CardScalingRequestFinalizerName)
		if err := r.Patch(ctx, csr, client.MergeFrom(lastCSR)); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer to CardScalingRequest: %w", err)
		}
		// Update lastCSR after Patch
		csr.DeepCopyInto(lastCSR)
	}

	// Check if there are any CardScalingRequests is processing
	processingCSRs, err := r.getProcessingCardScalingRequests(ctx, csr)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get other running CardScalingRequests: %w", err)
	}
	if len(processingCSRs) > 0 {
		// Other processing CardScalingRequests exist, update Phase to PhaseCardScalePending
		r.Recorder.Event(csr, corev1.EventTypeNormal, eventReasonPending,
			"Pending this CardScalingRequest until another one is completed")
		csr.Status.Phase = scalev1.PhaseCardScalePending
		return ctrl.Result{RequeueAfter: requeueIntervalCardScalingRequestPending}, nil
	}

	// Update Phase to PhaseCardScaleInitializing
	csr.Status.Phase = scalev1.PhaseCardScaleInitializing

	return ctrl.Result{}, nil
}

// reconcileCardScaleInitializing is a function to reconcile CardScalingRequest in PhaseCardScaleInitializing.
// It collects necessary information from workload cluster and management cluster and updates phase to PhaseDCIMachineTemplateCreating.
//
//nolint:unparam,gocyclo,funlen // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileCardScaleInitializing(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	// Get Node resource at workload cluster
	node, err := r.ClientHelper.GetNode(ctx, csr.Spec.NodeName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Node resource: %w", err)
	}

	// Get ClusterName and ClusterNamespace from Node resource
	clusterName, clusterNamespace, err := r.ClientHelper.GetClusterNameAndNamespaceFromNode(node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Cluster info from Node resource: %w", err)
	}

	// Get Machine at management cluster from Node resource
	machine, err := r.ClientHelper.GetMachineFromNode(ctx, node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Machine resource from Node resource: %w", err)
	}

	// Get MachineDeployment at management cluster from Machine resource
	md, err := r.ClientHelper.GetMachineDeploymentFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get MachineDeployment resource from Machine resource: %w", err)
	}

	// Get DCIMachineTemplate at management cluster from MachineDeployment resource
	dciMachineTemplate, err := r.ClientHelper.GetDCIMachineTemplateFromMachineDeployment(ctx, md, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachineTemplate resource from MachineDeployment resource: %w", err)
	}

	// Get DCIMachine at management cluster from Machine resource
	dciMachine, err := r.ClientHelper.GetDCIMachineFromMachine(ctx, machine, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachine resource from Machine resource: %w", err)
	}

	// Get LogicalServer at management cluster from DCIMachine resource
	if csr.Spec.ScalingParameter.UUIDs != nil {
		ls, lsErr := r.ClientHelper.GetLogicalServerFromDCIMachine(ctx, dciMachine)
		if lsErr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get LogicalServer resource from DCIMachine resource: %w", lsErr)
		}
		if !utils.HasAnyUUIDsInLogicalServer(csr.Spec.ScalingParameter.UUIDs, ls.Spec.ServerSpec.GPU) {
			csr.Status.Phase = scalev1.PhaseCardScaleCompleted
			logger := log.FromContext(ctx)
			logger.Info("Target GPU is already removed from LogicalServer, so skip scaling process and complete CardScalingRequest",
				"GPUSpec", ls.Spec.ServerSpec.GPU)
			r.Recorder.Event(csr, corev1.EventTypeNormal, eventReasonCompleted,
				"CardScalingRequest is completed because target GPU is already removed from LogicalServer")
			return ctrl.Result{}, nil
		}
	}

	// Check if DCIMachine is already annotated for card scaling
	if utils.HasDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations) {
		return ctrl.Result{}, errors.New("expected DCIMachine is not annotated for card scaling, but already annotated")
	}

	// Create DCIMachineTemplate after scaling
	expectedSpec, err := utils.GenerateDCIMachineSpecAfterScaling(&dciMachine.Spec, csr)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to generate expected DCIMachineSpec after scaling: %w", err)
	}

	// Get valid DCIMachineTemplate after scaling
	foundTemplates, generatedTemplateName, err := r.ClientHelper.GetDCIMachineTemplatesValid(ctx,
		expectedSpec.Location, &expectedSpec.ServerSpec, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get valid DCIMachineTemplate after scaling: %w", err)
	}

	// Get valid MachineDeployment after scaling
	foundMDs, generatedMDName, _, err := r.ClientHelper.GetMachineDeploymentsAllocatable(ctx,
		clusterNamespace, utils.MachineTemplateNameMap[utils.ProviderTypeDCI], clusterName, foundTemplates)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get valid MachineDeployment after scaling: %w", err)
	}

	// Create status info
	oldDCIMachineInfo := &scalev1.DCIMachineInfo{
		MachineDeploymentName:  md.Name,
		MachineDeploymentUID:   &md.UID,
		DCIMachineTemplateName: dciMachineTemplate.Name,
		DCIMachineTemplateUID:  &dciMachineTemplate.UID,
		DCIMachineName:         dciMachine.Name,
		DCIMachineUID:          &dciMachine.UID,
	}
	newDCIMachineInfo := &scalev1.DCIMachineInfo{}
	if len(foundMDs) > 0 {
		// Set found MachineDeployment info and its DCIMachineTemplate info
		foundMD := &foundMDs[0]
		newDCIMachineInfo.MachineDeploymentName = foundMD.Name
		newDCIMachineInfo.MachineDeploymentUID = &foundMD.UID
		templateName := foundMD.Spec.Template.Spec.InfrastructureRef.Name
		newDCIMachineInfo.DCIMachineTemplateName = templateName
		for _, template := range foundTemplates {
			if templateName == template.GetName() {
				newDCIMachineInfo.DCIMachineTemplateUID = ptr.To(template.GetUID())
				break
			}
		}
		// if valid template not found, regenerate template in next phase, so not set UID here.
	} else {
		// Set generated names
		newDCIMachineInfo.MachineDeploymentName = generatedMDName
		newDCIMachineInfo.MachineDeploymentUID = nil
		if len(foundTemplates) > 0 {
			// Set found DCIMachineTemplate info
			foundTemplate := foundTemplates[0]
			newDCIMachineInfo.DCIMachineTemplateName = foundTemplate.GetName()
			newDCIMachineInfo.DCIMachineTemplateUID = ptr.To(foundTemplate.GetUID())
		} else {
			// Set generated DCIMachineTemplate name
			newDCIMachineInfo.DCIMachineTemplateName = generatedTemplateName
			newDCIMachineInfo.DCIMachineTemplateUID = nil
		}
	}

	// Update Status and Phase to PhaseDCIMachineTemplateCreating
	csr.Status.ClusterName = clusterName
	csr.Status.ClusterNamespace = clusterNamespace
	csr.Status.OldDCIMachineInfo = oldDCIMachineInfo
	csr.Status.NewDCIMachineInfo = newDCIMachineInfo
	csr.Status.Phase = scalev1.PhaseDCIMachineTemplateCreating

	return ctrl.Result{}, nil
}

// reconcileDCIMachineTemplateCreating is a function to reconcile CardScalingRequest in PhaseDCIMachineTemplateCreating.
// It creates DCIMachineTemplate for scaling if needed and updates phase to PhaseMachineDeploymentCreating.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileDCIMachineTemplateCreating(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.OldDCIMachineInfo == nil || csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected OldDCIMachineInfo and NewDCIMachineInfo are not nil, but got nil")
	}

	// Check if DCIMachineTemplate is already existing
	if csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID != nil {
		// DCIMachineTemplate is already created, so we skip creating it.
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCreating
		return ctrl.Result{}, nil
	}

	oldTemplateName := csr.Status.OldDCIMachineInfo.DCIMachineTemplateName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get old DCIMachineTemplate resource from management cluster
	dciMachineTemplate, err := r.ClientHelper.GetDCIMachineTemplate(ctx, oldTemplateName, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get old DCIMachineTemplate resource from management cluster: %w", err)
	}

	// Generate expected DCIMachineSpec after scaling
	expectedSpec, err := utils.GenerateDCIMachineSpecAfterScaling(&dciMachineTemplate.Spec.Template.Spec, csr)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to generate expected DCIMachineSpec after scaling: %w", err)
	}

	// Create new DCIMachineTemplate resource instance
	newDCIMachineTemplate, err := utils.NewDCIMachineTemplateFromCardScalingRequest(csr, expectedSpec)
	if err != nil {
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("failed to create new DCIMachineTemplate instance: %w", err)
	}

	// Create DCIMachineTemplate resource in management cluster
	if err = r.ClientHelper.CreateDCIMachineTemplate(ctx, newDCIMachineTemplate); err == nil {
		// Update Phase to PhaseMachineDeploymentCreating
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID = &newDCIMachineTemplate.UID
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCreating
		return ctrl.Result{}, nil
	} else if !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("failed to create DCIMachineTemplate resource in management cluster: %w", err)
	} // else already exists

	// Maybe other process created it after creating this name, so we check the resource.
	validTemplates, generatedName, extraErr := r.ClientHelper.GetDCIMachineTemplatesValid(ctx,
		expectedSpec.Location, &expectedSpec.ServerSpec, clusterNamespace)
	if extraErr != nil {
		extraErr = fmt.Errorf("failed to create unique name for DCIMachineTemplate: %w", extraErr)
		err = fmt.Errorf("%w: %w", extraErr, err)
		return ctrl.Result{}, err
	}
	if len(validTemplates) != 0 {
		// Valid DCIMachineTemplate is found, so we use it.
		// This case happens when patching status of CardScalingRequest failed
		// after creating DCIMachineTemplate in previous reconciliation or
		// other process created it.
		foundTemplate := validTemplates[0]
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateName = foundTemplate.GetName()
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID = ptr.To(foundTemplate.GetUID())
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCreating
		err = nil // clear error
	} else {
		// Set generated name and try to create it again in next reconciliation.
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateName = generatedName
		csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID = nil
		err = fmt.Errorf("retry to create a new DCIMachineTemplate with new name(%s): %w", generatedName, err)
	}
	return ctrl.Result{}, err
}

// reconcileMachineDeploymentCreating is a function to reconcile CardScalingRequest in PhaseMachineDeploymentCreating.
// It creates MachineDeployment for scaling with zero replicas if needed and updates phase to PhaseDCIMachineAnnotating.
// The replicas will be updated later by NodeScalingRequest.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileMachineDeploymentCreating(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.OldDCIMachineInfo == nil || csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected OldDCIMachineInfo and NewDCIMachineInfo are not nil, but got nil")
	}

	oldMDName := csr.Status.OldDCIMachineInfo.MachineDeploymentName
	newMDName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
	newTemplateName := csr.Status.NewDCIMachineInfo.DCIMachineTemplateName
	clusterName := csr.Status.ClusterName
	clusterNamespace := csr.Status.ClusterNamespace
	infraRefName := utils.MachineTemplateNameMap[utils.ProviderTypeDCI]

	// Check if MachineDeployment is already existing
	if csr.Status.NewDCIMachineInfo.MachineDeploymentUID != nil {
		md, err := r.ClientHelper.GetMachineDeployment(ctx, newMDName, clusterNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get new MachineDeployment after scaled resource from management cluster: %w", err)
		}
		lastMD := md.DeepCopy()
		// Add annotation to indicate that card scaling is in progress
		md.Annotations = utils.AddCardScalingRequestInProgressUIDAnnotation(md.Annotations, csr.UID)
		if err = r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add annotation to new MachineDeployment after scaled resource: %w", err)
		}
		// MachineDeployment is already created, so we skip creating it.
		csr.Status.Phase = scalev1.PhaseDCIMachineAnnotating
		return ctrl.Result{}, nil
	}

	// Get old MachineDeployment resource from management cluster
	md, err := r.ClientHelper.GetMachineDeployment(ctx, oldMDName, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get old MachineDeployment resource from management cluster: %w", err)
	}

	// Create new MachineDeployment resource instance
	newMD, err := utils.NewMachineDeploymentFromCardScalingRequest(csr, md)
	if err != nil {
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("failed to create new MachineDeployment instance: %w", err)
	}
	// Add annotation to indicate that card scaling is in progress
	newMD.Annotations = utils.AddCardScalingRequestInProgressUIDAnnotation(newMD.Annotations, csr.UID)

	// Create MachineDeployment resource in management cluster
	if err = r.ClientHelper.CreateMachineDeployment(ctx, newMD); err == nil {
		// Update Phase to PhaseDCIMachineAnnotating
		csr.Status.NewDCIMachineInfo.MachineDeploymentUID = &newMD.UID
		csr.Status.Phase = scalev1.PhaseDCIMachineAnnotating
		return ctrl.Result{}, nil
	} else if !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("failed to create MachineDeployment resource in management cluster: %w", err)
	}

	// Maybe other process created it after creating this name, so we check the resource.
	dciMachineTemplate, extraErr := r.ClientHelper.GetDCIMachineTemplate(ctx, newTemplateName, clusterNamespace)
	if extraErr != nil {
		extraErr = fmt.Errorf("failed to get DCIMachineTemplate resource from management cluster: %w", extraErr)
		err = fmt.Errorf("%w: %w", extraErr, err)
		return ctrl.Result{}, err
	}
	// Get valid MachineDeployment after scaling, if not found we generate a unique name
	validMDs, generatedName, _, extraErr := r.ClientHelper.GetMachineDeploymentsAllocatable(ctx,
		clusterNamespace, infraRefName, clusterName, []client.Object{dciMachineTemplate})
	if extraErr != nil {
		extraErr = fmt.Errorf("failed to create unique name for MachineDeployment: %w", extraErr)
		err = fmt.Errorf("%w: %w", extraErr, err)
		return ctrl.Result{}, err
	}
	if len(validMDs) != 0 {
		// Valid MachineDeployment is found, so we use it.
		// This case will be happened when patching status of CardScalingRequest failed
		// after creating MachineDeployment in previous reconciliation or
		// other process created it.
		csr.Status.NewDCIMachineInfo.MachineDeploymentName = validMDs[0].Name
		csr.Status.NewDCIMachineInfo.MachineDeploymentUID = &validMDs[0].UID
		csr.Status.Phase = scalev1.PhaseDCIMachineAnnotating
		err = nil // clear error
	} else {
		// Set generated name and try to create it again in next reconciliation.
		csr.Status.NewDCIMachineInfo.MachineDeploymentName = generatedName
		csr.Status.NewDCIMachineInfo.MachineDeploymentUID = nil
		err = fmt.Errorf("retry to create a new MachineDeployment with new name(%s): %w", generatedName, err)
	}

	return ctrl.Result{}, err
}

// generateAnnotationCardScalingInfo generates annotation for card scaling info for DCIMachine resource.
func generateAnnotationCardScalingInfo(
	csr *scalev1.CardScalingRequest,
) (string, error) {
	// Validate input
	if csr == nil || csr.Status.NewDCIMachineInfo == nil {
		return "", errors.New("expected CardScalingRequest is not nil and ScalingParameter is not empty, but got invalid value")
	}

	// Generate annotation value
	mdName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
	targetParam := csr.Spec.ScalingParameter
	return utils.GenerateDCIMachineCardScalingInfoAnnotation(mdName, targetParam.Op, targetParam.UUIDs)
}

// reconcileDCIMachineAnnotating is a function to reconcile CardScalingRequest in PhaseDCIMachineAnnotating.
// It annotates DCIMachine for scaling information and updates phase to PhaseCardScaleInitialized.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileDCIMachineAnnotating(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.OldDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected OldDCIMachineInfo is not nil, but got nil")
	}

	oldDCIMachineName := csr.Status.OldDCIMachineInfo.DCIMachineName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get DCIMachine resource from management cluster
	dciMachine, err := r.ClientHelper.GetDCIMachine(ctx, oldDCIMachineName, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachine resource from management cluster: %w", err)
	}
	lastDCIMachine := dciMachine.DeepCopy()

	// Generate annotation info
	annotationValue, err := generateAnnotationCardScalingInfo(csr)
	if err != nil {
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("failed to generate annotation info for DCIMachine: %w", err)
	}

	// Check if DCIMachine is already annotated
	if utils.HasDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations) &&
		!utils.IsSameDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations, annotationValue) {
		// This case is only for failsafe because we already checked in reconcileCardScaleInitializing
		// DCIMachine is annotated with a different value, so we return error to requeue with Exponential backoff
		// Update Phase to PhaseCardScaleInitializing(resetting) to get the latest DCIMachine info again,
		// because DCIMachine will be deleted and recreated by other CardScalingRequest.
		csr.Status.Phase = scalev1.PhaseCardScaleInitializing
		return ctrl.Result{}, errors.New("expected the annotation of DCIMachine is not annotated, but annotated with unexpected value")
	}

	// Set annotation into DCIMachine resource
	dciMachine.Annotations = utils.AddDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations, annotationValue)

	// Patch DCIMachine resource in management cluster
	if err = r.ClientHelper.PatchDCIMachine(ctx, dciMachine, lastDCIMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch DCIMachine resource in management cluster: %w", err)
	}

	// Update Phase to PhaseCardScaleInitialized
	csr.Status.Phase = scalev1.PhaseCardScaleInitialized

	return ctrl.Result{}, nil
}

// reconcileCardScaleInitialized is a function to reconcile CardScalingRequest in PhaseCardScaleInitialized.
// It creates NodeScalingRequest for scaling down and updates phase to PhaseNodeScalingDown.
// The name and namespace of NodeScalingRequest are the same as CardScalingRequest.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileCardScaleInitialized(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to create NodeScalingRequest for scaling down")

	// Create NodeScalingRequest for scaling down
	nsr := &scalev1.NodeScalingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csr.Name,
			Namespace: csr.Namespace,
		},
		Spec: scalev1.NodeScalingRequestSpec{
			NodeRequestType: scalev1.TypeDelete,
			Node:            &csr.Spec.NodeName,
		},
	}
	// Set owner reference
	if err := controllerutil.SetControllerReference(csr, nsr, r.Client.Scheme()); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference to NodeScalingRequest: %w", err)
	}

	// Create NodeScalingRequest resource in workload cluster
	if err := r.Create(ctx, nsr); err == nil {
		// Update Phase to PhaseNodeScalingDown
		csr.Status.Phase = scalev1.PhaseNodeScalingDown
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonScalingDownStarted,
			"NodeScalingRequest %s/%s is created for scaling down", nsr.Namespace, nsr.Name)
		return ctrl.Result{}, nil
	} else if !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("failed to create NodeScalingRequest for scaling down: %w", err)
	} // else already exists

	// Check the owner of existing NodeScalingRequest
	nsrExisting := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsrExisting); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get existing NodeScalingRequest: %w", err)
	}
	if metav1.IsControlledBy(nsrExisting, csr) {
		// NodeScalingRequest is already created by this CardScalingRequest, so we proceed to the next phase.
		csr.Status.Phase = scalev1.PhaseNodeScalingDown
		return ctrl.Result{}, nil
	}
	return ctrl.Result{},
		fmt.Errorf("expected existing NodeScalingRequest(%s) is controlled by CardScalingRequest, but not controlled", nsrKey.String())
}

// reconcileNodeScalingDown is a function to reconcile CardScalingRequest in PhaseNodeScalingDown.
// It waits for completing NodeScalingRequest for scaling down and updates phase to PhaseNodeScaleDownDeleting.
// When scaling down fails, it updates the phase to PhaseNodeScaleDownFailed.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileNodeScalingDown(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Waiting for completing NodeScalingRequest for scaling down")

	// Get NodeScalingRequest resource
	nsr := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsr); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get NodeScalingRequest: %w", err)
	}

	// Check status of NodeScalingRequest
	switch nsr.Status.Phase {
	case scalev1.PhaseNodeDeleted:
		logger.Info("NodeScalingRequest for scaling down is completed successfully")
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonScalingDownCompleted,
			"NodeScalingRequest %s/%s is completed for scaling down", nsr.Namespace, nsr.Name)
		// Update Phase to PhaseNodeScaleDownDeleting
		csr.Status.Phase = scalev1.PhaseNodeScaleDownDeleting
		return ctrl.Result{}, nil

	case scalev1.PhaseNodeDeleteFailed:
		// Set LastTransitionTime
		csr.Status.LastTransitionTime = ptr.To(metav1.NewTime(time.Now()))

		// Update Phase to PhaseNodeScaleDownFailed
		csr.Status.Phase = scalev1.PhaseNodeScaleDownFailed
		return ctrl.Result{}, fmt.Errorf("failed to scale down, transition to %s", scalev1.PhaseNodeScaleDownFailed)

	default:
		// Not finished yet, need to requeue
		return ctrl.Result{}, nil // expect requeue by update event of NodeScalingRequest
	}
}

// reconcileNodeScaleDownDeleting is a function to reconcile CardScalingRequest in PhaseNodeScaleDownDeleting.
// It deletes NodeScalingRequest for scaling down and updates phase to PhaseNodeScaleDownCompleted.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileNodeScaleDownDeleting(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	// Get existing NodeScalingRequest for scaling down
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	nsrExisting := &scalev1.NodeScalingRequest{}
	if err := r.Get(ctx, nsrKey, nsrExisting); apierrors.IsNotFound(err) {
		// Already deleted
		csr.Status.Phase = scalev1.PhaseNodeScaleDownCompleted
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get existing NodeScalingRequest for scaling down: %w", err)
	} // else found

	// Delete NodeScalingRequest
	if err := r.Delete(ctx, nsrExisting); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to delete existing NodeScalingRequest for scaling down: %w", err)
	}

	// Update Phase to PhaseNodeScaleDownCompleted
	csr.Status.Phase = scalev1.PhaseNodeScaleDownCompleted

	return ctrl.Result{}, nil
}

// reconcileNodeScaleDownCompleted is a function to reconcile CardScalingRequest in PhaseNodeScaleDownCompleted.
// It creates NodeScalingRequest for scaling up and updates phase to PhaseNodeScalingUp.
// The name and namespace of NodeScalingRequest are the same as CardScalingRequest.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileNodeScaleDownCompleted(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start to create NodeScalingRequest for scaling up")

	if csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected NewDCIMachineInfo is not nil, but got nil")
	}

	newTemplateName := csr.Status.NewDCIMachineInfo.DCIMachineTemplateName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get DCIMachineTemplate resource from management cluster to get ServerSpec and Location
	dciMachineTemplate, err := r.ClientHelper.GetDCIMachineTemplate(ctx, newTemplateName, clusterNamespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachineTemplate resource from management cluster: %w", err)
	}

	// Create NodeScalingRequest for scaling up
	nsr := &scalev1.NodeScalingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csr.Name,
			Namespace: csr.Namespace,
		},
		Spec: scalev1.NodeScalingRequestSpec{
			NodeRequestType: scalev1.TypeCreate,
			Location:        dciMachineTemplate.Spec.Template.Spec.Location,
			ServerSpec:      &dciMachineTemplate.Spec.Template.Spec.ServerSpec,
		},
	}
	// Set owner reference
	if err := controllerutil.SetControllerReference(csr, nsr, r.Client.Scheme()); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference to NodeScalingRequest for scaling up: %w", err)
	}

	// Create NodeScalingRequest resource in workload cluster
	if err := r.Create(ctx, nsr); err == nil {
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonScalingUpStarted,
			"NodeScalingRequest %s/%s is created for scaling up", nsr.Namespace, nsr.Name)
		// Update Phase to PhaseNodeScalingUp
		csr.Status.Phase = scalev1.PhaseNodeScalingUp
		return ctrl.Result{}, nil
	} else if !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("failed to create NodeScalingRequest for scaling up: %w", err)
	} // else already exists

	// Check the owner of existing NodeScalingRequest
	nsrExisting := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsrExisting); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get existing NodeScalingRequest for scaling up: %w", err)
	}
	if nsrExisting.DeletionTimestamp != nil {
		// Wait until deletion is completed
		return ctrl.Result{}, nil
	}
	if metav1.IsControlledBy(nsrExisting, csr) {
		// NodeScalingRequest is already created by this CardScalingRequest, so we proceed to the next phase.
		csr.Status.Phase = scalev1.PhaseNodeScalingUp
		return ctrl.Result{}, nil
	}
	return ctrl.Result{},
		fmt.Errorf("expected existing NodeScalingRequest(%s) is controlled by CardScalingRequest, but not controlled", nsrKey.String())
}

// reconcileNodeScalingUp is a function to reconcile CardScalingRequest in PhaseNodeScalingUp.
// It waits for completing NodeScalingRequest for scaling up and updates phase to PhaseCardScaleCompleted.
// When scaling up fails, it updates the phase to PhaseNodeScaleUpFailed.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileNodeScalingUp(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Waiting for completing NodeScalingRequest for scaling up")

	if csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected NewDCIMachineInfo is not nil, but got nil")
	}

	// Get NodeScalingRequest resource
	nsr := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsr); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get NodeScalingRequest for scaling up: %w", err)
	}

	// Check status of NodeScalingRequest
	switch nsr.Status.Phase {
	case scalev1.PhaseNodeCreated:
		logger.Info("NodeScalingRequest for scaling up is completed successfully")
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonScalingUpCompleted,
			"NodeScalingRequest %s/%s is completed for scaling up", nsr.Namespace, nsr.Name)
		// Cleanup MachineDeployment annotation for card scaling
		newMDName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
		newMDNamespace := csr.Status.ClusterNamespace
		md, err := r.ClientHelper.GetMachineDeployment(ctx, newMDName, newMDNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get MachineDeployment annotation for card scaling: %w", err)
		}
		if utils.IsSameCardScalingRequestInProgressUIDAnnotation(md.Annotations, csr.UID) {
			lastMD := md.DeepCopy()
			utils.RemoveCardScalingRequestInProgressUIDAnnotation(md.Annotations)
			if err = r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove annotation from MachineDeployment after scaling: %w", err)
			}
		}

		// Get Node resource
		node, err := r.ClientHelper.GetNode(ctx, csr.Spec.NodeName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get Node resource: %w", err)
		}

		// Get Machine from Node resource
		machine, err := r.ClientHelper.GetMachineFromNode(ctx, node)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get Machine resource from Node resource: %w", err)
		}

		// Get DCIMachine from Machine resource
		dciMachine, err := r.ClientHelper.GetDCIMachineFromMachine(ctx, machine, csr.Status.ClusterNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get DCIMachine resource from Machine resource: %w", err)
		}

		// Set LastTransitionTime
		csr.Status.LastTransitionTime = ptr.To(metav1.NewTime(time.Now()))

		// Update Phase to PhaseCardScaleCompleted and set new DCIMachine info
		logger.Info("CardScalingRequest is completed successfully")
		csr.Status.Phase = scalev1.PhaseCardScaleCompleted
		csr.Status.NewDCIMachineInfo.DCIMachineName = dciMachine.Name
		csr.Status.NewDCIMachineInfo.DCIMachineUID = &dciMachine.UID

		return ctrl.Result{}, nil

	case scalev1.PhaseNodeCreateFailed:
		// Set LastTransitionTime
		csr.Status.LastTransitionTime = ptr.To(metav1.NewTime(time.Now()))

		// Update Phase to PhaseNodeScaleUpFailed
		csr.Status.Phase = scalev1.PhaseNodeScaleUpFailed
		return ctrl.Result{}, fmt.Errorf("failed to scale up, transition to %s", scalev1.PhaseNodeScaleUpFailed)

	default:
		// Not finished yet, so we requeue
		return ctrl.Result{}, nil // expect requeue by update event of NodeScalingRequest
	}
}

// reconcileTerminalPhase is a function to clean up expired CardScalingRequest.
// It deletes expired CardScalingRequest if below conditions are met:
//   - TTL is enabled
//   - No owner references
//   - CardScalingRequest is not in a failed state
func (r *CardScalingRequestReconciler) reconcileTerminalPhase(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if !r.enableTTLSecondsAfterFinished || len(csr.OwnerReferences) != 0 || csr.IsFailed() {
		// Skip this function if TTL is disabled, CardScalingRequest has owner references, or CardScalingRequest is in failed state.
		// It should not delete automatically CardScalingRequest with failure phases to notify users about the failure.
		return ctrl.Result{}, nil
	}

	if csr.Status.LastTransitionTime == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected LastTransitionTime is not nil in terminal phase, but got nil")
	}

	timeToExpire := csr.Status.LastTransitionTime.Add(time.Duration(r.ttlSecondsAfterFinished) * time.Second)

	// Check if TTL is expired
	now := time.Now()
	if !now.After(timeToExpire) {
		return ctrl.Result{RequeueAfter: timeToExpire.Sub(now)}, nil // recalculate remain seconds
	}

	// Delete expired CardScalingRequest
	logger := log.FromContext(ctx)
	if err := r.Delete(ctx, csr); apierrors.IsNotFound(err) {
		logger.Info("Expired CardScalingRequest is already deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete expired CardScalingRequest: %w", err)
	}

	logger.Info("Expired CardScalingRequest is deleted successfully")
	r.Recorder.Event(csr, corev1.EventTypeNormal, eventReasonExpired,
		"CardScalingRequest is terminated and deleted after TTL expiration")
	return ctrl.Result{}, nil
}

// doFinalize is a function to finalize CardScalingRequest.
// It calls phase-specific reconcile functions for finalization based on the current phase of the CardScalingRequest.
func (r *CardScalingRequestReconciler) doFinalize(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	// lastCSR is CSR that is stored in the etcd.
	// After calling Update() or Patch(), lastCSR should be updated.
	lastCSR := csr.DeepCopy()

	// Reconcile CardScalingRequest
	var rslt ctrl.Result
	var err error
	var eventReason = eventReasonFinalizerUnexpectedError
	switch csr.Status.Phase {
	case scalev1.PhaseNodeScaleDownDeleting, scalev1.PhaseNodeScalingDown, scalev1.PhaseCardScaleInitialized:
		rslt, err = r.reconcileWaitingForScaleDown(ctx, csr)
		eventReason = eventReasonFinalizerWaitingCompletionFailed

	case scalev1.PhaseNodeScalingUp, scalev1.PhaseNodeScaleDownCompleted:
		rslt, err = r.reconcileWaitingForScaleUp(ctx, csr)
		eventReason = eventReasonFinalizerWaitingCompletionFailed

	case scalev1.PhaseDCIMachineCleanup, // valid phase only when finalizing
		scalev1.PhaseDCIMachineAnnotating, scalev1.PhaseNodeScaleDownFailed:
		rslt, err = r.reconcileDCIMachineCleanup(ctx, csr)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.PhaseMachineDeploymentCleanup, // valid phase only when finalizing
		scalev1.PhaseMachineDeploymentCreating, scalev1.PhaseNodeScaleUpFailed:
		rslt, err = r.reconcileMachineDeploymentCleanup(ctx, csr)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.PhaseDCIMachineTemplateCleanup, // valid phase only when finalizing
		scalev1.PhaseDCIMachineTemplateCreating:
		rslt, err = r.reconcileDCIMachineTemplateCleanup(ctx, csr)
		eventReason = eventReasonFinalizerCleanupFailed

	case scalev1.PhaseCardScaleFinalized, // valid phase only when finalizing
		scalev1.PhaseCardScaleInitial, scalev1.PhaseCardScalePending, scalev1.PhaseCardScaleInitializing, scalev1.PhaseCardScaleCompleted:
		rslt, err = r.reconcileCardScaleFinalized(ctx, csr)
		// reconcileCardScaleFinalized removes finalizer from CardScalingRequest,
		// so Patch of status is not needed after calling it when err is nil.
		if err == nil {
			return rslt, nil
		}
		eventReason = eventReasonFinalizerRemovingFinalizerFailed

	default:
		// This case is only for failsafe, but should not happen because crd validation is checked.
		err = fmt.Errorf("unexpected phase(%s) is detected in doFinalize", lastCSR.Status.Phase)
	}

	// Patch status of CardScalingRequest
	if patchErr := r.updateCardScalingRequestStatus(ctx, csr, lastCSR); patchErr != nil {
		rslt = ctrl.Result{}
		patchErr = fmt.Errorf("failed to update CardScalingRequest status: %w", patchErr)
		if err == nil {
			err = patchErr
		} else {
			err = fmt.Errorf("%w: %w", patchErr, err)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to execute Finalizer process for state '%s': %w", lastCSR.Status.Phase, err)
		r.Recorder.Event(csr, corev1.EventTypeWarning, eventReason, err.Error())
	}

	return rslt, err
}

// reconcileWaitingForScaleUp is a function to reconcile CardScalingRequest waiting for scaling up to be completed.
// It waits for completing NodeScalingRequest for scaling up and updates phase to PhaseCardScaleFinalized or PhaseMachineDeploymentCleanup.
//
// reason for nolint:
//   - ctrl.Result is returned for consistency with other functions, but not used here.
//   - similar code is used in reconcileWaitingForScaleDown.
//
//nolint:unparam,dupl // see above comments
func (r *CardScalingRequestReconciler) reconcileWaitingForScaleUp(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Waiting for NodeScalingRequest for scaling up to be completed for finalization")

	// Get NodeScalingRequest resource
	nsr := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsr); apierrors.IsNotFound(err) {
		// Already deleted, so proceed to cleanup created resources.
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup
		return ctrl.Result{}, nil // expect requeue by update event
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get NodeScalingRequest for scaling up: %w", err)
	} // else found

	// Check status of NodeScalingRequest
	switch nsr.Status.Phase {
	case scalev1.PhaseNodeCreated:
		logger.Info("NodeScalingRequest for scaling up is completed successfully")
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonFinalizerScalingUpCompleted,
			"NodeScalingRequest %s/%s is completed for scaling up in Finalizer", nsr.Namespace, nsr.Name)
		// Scaling up is completed successfully,
		// so we need not to cleanup created resources.
		// And NodeScalingRequest is expected to be deleted by OwnerReference scheme.
		csr.Status.Phase = scalev1.PhaseCardScaleFinalized
		return ctrl.Result{}, nil // expect requeue by update event

	case scalev1.PhaseNodeCreateFailed:
		logger.Info("NodeScalingRequest for scaling up is completed with failure")
		r.Recorder.Eventf(csr, corev1.EventTypeWarning, eventReasonFinalizerScalingUpFailed,
			"NodeScalingRequest %s/%s is failed for scaling up in Finalizer", nsr.Namespace, nsr.Name)
		// Scaling up is completed unsuccessfully,
		// so we proceed to clean up created resources.
		// And NodeScalingRequest is expected to be deleted by OwnerReference scheme.
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup
		return ctrl.Result{}, nil // expect requeue by update event

	case scalev1.PhaseNodeCreating:
		// Not finished yet, need to requeue
		return ctrl.Result{}, nil // expect requeue by update event of NodeScalingRequest

	default:
		// This case is only for failsafe, but should not happen because NodeScalingRequest phase is validated.
		return ctrl.Result{}, fmt.Errorf("unexpected phase(%s) of NodeScalingRequest", nsr.Status.Phase)
	}
}

// reconcileWaitingForScaleDown is a function to reconcile CardScalingRequest waiting for scaling down to be completed.
// It waits for completing NodeScalingRequest for scaling down and updates phase to PhaseMachineDeploymentCleanup or PhaseDCIMachineCleanup.
//
// reason for nolint:
//   - ctrl.Result is returned for consistency with other functions, but not used here.
//   - similar code is used in reconcileWaitingForScaleUp.
//
//nolint:unparam,dupl // see above comments
func (r *CardScalingRequestReconciler) reconcileWaitingForScaleDown(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Waiting for NodeScalingRequest for scaling down to be completed for finalization")

	// Get NodeScalingRequest resource
	nsr := &scalev1.NodeScalingRequest{}
	nsrKey := client.ObjectKey{
		Name:      csr.Name,
		Namespace: csr.Namespace,
	}
	if err := r.Get(ctx, nsrKey, nsr); apierrors.IsNotFound(err) {
		// Already deleted, so proceed to cleanup created resources.
		csr.Status.Phase = scalev1.PhaseDCIMachineCleanup
		return ctrl.Result{}, nil // expect requeue by update event
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get NodeScalingRequest for scaling down: %w", err)
	} // else found

	// Check status of NodeScalingRequest
	switch nsr.Status.Phase {
	case scalev1.PhaseNodeDeleted:
		logger.Info("NodeScalingRequest for scaling down is completed successfully")
		r.Recorder.Eventf(csr, corev1.EventTypeNormal, eventReasonFinalizerScalingDownCompleted,
			"NodeScalingRequest %s/%s is completed for scaling down in Finalizer", nsr.Namespace, nsr.Name)
		// Scaling down is completed successfully,
		// so we skip to cleanup DCIMachine annotation, and proceed to cleanup MachineDeployment.
		// And NodeScalingRequest is expected to be deleted by OwnerReference scheme.
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup
		return ctrl.Result{}, nil // expect requeue by update event

	case scalev1.PhaseNodeDeleteFailed:
		logger.Info("NodeScalingRequest for scaling down is completed with failure")
		r.Recorder.Eventf(csr, corev1.EventTypeWarning, eventReasonFinalizerScalingDownFailed,
			"NodeScalingRequest %s/%s is failed for scaling down in Finalizer", nsr.Namespace, nsr.Name)
		// Scaling down is completed unsuccessfully,
		// so we proceed to clean up created resources.
		// And NodeScalingRequest is expected to be deleted by OwnerReference scheme.
		csr.Status.Phase = scalev1.PhaseDCIMachineCleanup
		return ctrl.Result{}, nil // expect requeue by update event

	case scalev1.PhaseNodeDeleting:
		// Not finished yet, need to requeue
		return ctrl.Result{}, nil // expect requeue by update event of NodeScalingRequest

	default:
		// This case is only for failsafe, but should not happen because NodeScalingRequest phase is validated.
		return ctrl.Result{}, fmt.Errorf("unexpected phase(%s) of NodeScalingRequest", nsr.Status.Phase)
	}
}

// reconcileDCIMachineCleanup is a function to reconcile CardScalingRequest in PhaseDCIMachineRemovingAnnotation.
// It removes annotation from old DCIMachine and updates phase to PhaseMachineDeploymentCleanup.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileDCIMachineCleanup(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.OldDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected OldDCIMachineInfo is not nil, but got nil")
	}

	oldDCIMachineName := csr.Status.OldDCIMachineInfo.DCIMachineName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get DCIMachine resource from management cluster
	dciMachine, err := r.ClientHelper.GetDCIMachine(ctx, oldDCIMachineName, clusterNamespace)
	if apierrors.IsNotFound(err) {
		// Already deleted, so we proceed to next phase
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DCIMachine resource from management cluster: %w", err)
	} // else found

	// Generate annotation info
	annotationValue, err := generateAnnotationCardScalingInfo(csr)
	if err != nil {
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("failed to generate annotation info for DCIMachine: %w", err)
	}

	if !utils.IsSameDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations, annotationValue) {
		// DCIMachine is annotated with a different value, so we skip removing the annotation.
		// Update Phase to PhaseMachineDeploymentCleanup
		csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup

		return ctrl.Result{}, nil
	} // annotation is found and matches

	// Remove annotation from DCIMachine resource
	lastDCIMachine := dciMachine.DeepCopy()
	utils.RemoveDCIMachineCardScalingInfoAnnotation(dciMachine.Annotations)

	// Patch DCIMachine resource in management cluster
	if err = r.ClientHelper.PatchDCIMachine(ctx, dciMachine, lastDCIMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch DCIMachine resource in management cluster: %w", err)
	}

	// Update Phase to PhaseMachineDeploymentCleanup
	csr.Status.Phase = scalev1.PhaseMachineDeploymentCleanup

	return ctrl.Result{}, nil
}

// reconcileMachineDeploymentCleanup is a function to reconcile CardScalingRequest in PhaseMachineDeploymentCleanup.
// It deletes created MachineDeployment if needed and updates phase to PhaseDCIMachineTemplateCleanup.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileMachineDeploymentCleanup(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected NewDCIMachineInfo is not nil, but got nil")
	}

	newMDName := csr.Status.NewDCIMachineInfo.MachineDeploymentName
	newMDUID := csr.Status.NewDCIMachineInfo.MachineDeploymentUID
	clusterName := csr.Status.ClusterName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get existing MachineDeployment
	md, err := r.ClientHelper.GetMachineDeployment(ctx, newMDName, clusterNamespace)
	if apierrors.IsNotFound(err) {
		// Already deleted, so we proceed to next phase
		csr.Status.Phase = scalev1.PhaseDCIMachineTemplateCleanup
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get existing MachineDeployment: %w", err)
	} // else found

	// Check if we can delete MachineDeployment
	if newMDUID != nil && md.UID != *newMDUID {
		// UID mismatches, so we skip deleting it and removing annotation.
		csr.Status.Phase = scalev1.PhaseDCIMachineTemplateCleanup
		return ctrl.Result{}, nil
	}
	isLast, err := r.ClientHelper.IsLastMachineDeployment(ctx, md, clusterNamespace, clusterName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if MachineDeployment is the last one in the cluster: %w", err)
	}
	if isLast || (md.Spec.Replicas != nil && *md.Spec.Replicas != 0) {
		// MachineDeployment is the last one in the cluster,
		// or MachineDeployment has replicas(=in use),
		// so we skip deleting it.
		// We check .Spec.Replicas instead of .Status.ReadyReplicas
		// because .Spec.Replicas reflects the desired number of replicas,
		// indicating whether there are pending scaling requests for this MachineDeployment.
		if utils.IsSameCardScalingRequestInProgressUIDAnnotation(md.Annotations, csr.UID) {
			// Remove annotation from MachineDeployment
			lastMD := md.DeepCopy()
			utils.RemoveCardScalingRequestInProgressUIDAnnotation(md.Annotations)
			if err = r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to patch MachineDeployment to remove annotation: %w", err)
			}
		}
		csr.Status.Phase = scalev1.PhaseDCIMachineTemplateCleanup
		return ctrl.Result{}, nil
	}

	// Delete MachineDeployment
	if err = r.ClientHelper.DeleteMachineDeployment(ctx, md); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete existing MachineDeployment: %w", err)
	}

	// Update Phase to PhaseDCIMachineTemplateCleanup
	csr.Status.Phase = scalev1.PhaseDCIMachineTemplateCleanup

	return ctrl.Result{}, nil
}

// reconcileDCIMachineTemplateCleanup is a function to reconcile CardScalingRequest in PhaseDCIMachineTemplateCleanup.
// It deletes created DCIMachineTemplate if needed and updates phase to PhaseCardScaleFinalized.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileDCIMachineTemplateCleanup(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	if csr.Status.NewDCIMachineInfo == nil {
		// This case is only for failsafe
		return ctrl.Result{}, errors.New("expected NewDCIMachineInfo is not nil, but got nil")
	}

	newTemplateName := csr.Status.NewDCIMachineInfo.DCIMachineTemplateName
	clusterNamespace := csr.Status.ClusterNamespace

	// Get existing DCIMachineTemplate
	dciMachineTemplate, err := r.ClientHelper.GetDCIMachineTemplate(ctx, newTemplateName, clusterNamespace)
	if apierrors.IsNotFound(err) {
		// Already deleted
		csr.Status.Phase = scalev1.PhaseCardScaleFinalized
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get existing DCIMachineTemplate: %w", err)
	} // else found

	// Check if we can delete DCIMachineTemplate
	if csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID != nil &&
		dciMachineTemplate.UID != *csr.Status.NewDCIMachineInfo.DCIMachineTemplateUID {
		// UID mismatches, so we skip deleting it.
		csr.Status.Phase = scalev1.PhaseCardScaleFinalized
		return ctrl.Result{}, nil
	}
	mds, err := r.ClientHelper.GetMachineDeploymentsFromTemplate(ctx, csr.Status.ClusterNamespace, nil, dciMachineTemplate)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get MachineDeployments using the DCIMachineTemplate: %w", err)
	}
	if len(mds) != 0 {
		// DCIMachineTemplate is still in use by existing MachineDeployments, so we skip deleting it.
		csr.Status.Phase = scalev1.PhaseCardScaleFinalized
		return ctrl.Result{}, nil
	}

	// Delete DCIMachineTemplate
	if err = r.ClientHelper.DeleteDCIMachineTemplate(ctx, dciMachineTemplate); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete existing DCIMachineTemplate: %w", err)
	}

	// Update Phase to PhaseCardScaleFinalized
	csr.Status.Phase = scalev1.PhaseCardScaleFinalized

	return ctrl.Result{}, nil
}

// reconcileCardScaleFinalized is a function to reconcile CardScalingRequest in PhaseCardScaleFinalized.
// It removes finalizer from CardScalingRequest to allow deletion.
//
//nolint:unparam // ctrl.Result is returned for consistency with other functions, but not used here.
func (r *CardScalingRequestReconciler) reconcileCardScaleFinalized(
	ctx context.Context, csr *scalev1.CardScalingRequest,
) (ctrl.Result, error) {
	// lastCSR is CSR that is stored in the etcd.
	lastCSR := csr.DeepCopy()

	// Remove finalizer from CardScalingRequest
	controllerutil.RemoveFinalizer(csr, scalev1.CardScalingRequestFinalizerName)

	// Patch CardScalingRequest to remove finalizer
	if err := r.Patch(ctx, csr, client.MergeFrom(lastCSR)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer from CardScalingRequest: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("CardScalingRequest is finalized successfully")

	return ctrl.Result{}, nil
}
