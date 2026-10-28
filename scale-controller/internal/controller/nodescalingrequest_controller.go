// Package controller provides functions to handle NodeScalingRequest.
package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"k8s.io/utils/ptr"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	helperclient "scale-controller/internal/helper/client"
	"scale-controller/internal/helper/utils"
)

// Types for metav1.conditions[].
const (
	Initialized string = "Initialized"
	Requested   string = "Requested"
	Completed   string = "Completed"
)

// progressCondition represents the state of progress for a MachineDeployment.
type progressCondition struct {
	state progressState
	uid   string
}

type progressState int

const (
	notInProgress progressState = iota
	inProgressByThisNSR
	inProgressByOtherNSR
)

// scaleDownResources represents the resources to be handled in scale down process
type scaleDownResources struct {
	machineDeployment *capiv1.MachineDeployment
	machine           *capiv1.Machine
}

// NodeScalingRequestReconciler reconciles a NodeScalingRequest object.
type NodeScalingRequestReconciler struct {
	// client for workload cluster
	client.Client

	// scheme for workload cluster
	Scheme *runtime.Scheme

	// interface for event recorder
	Recorder events.EventRecorder

	// helper client for management cluster
	ClientHelper helperclient.ClientHelper

	// setting for provider type: DCI, Docker
	providerType utils.ProviderType

	// setting for enabling TTL
	enableTTLSecondsAfterFinished bool

	// setting for TTL seconds
	ttlSecondsAfterFinished int
}

// +kubebuilder:rbac:groups=cluster.my.domain,resources=nodescalingrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.my.domain,resources=nodescalingrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.my.domain,resources=nodescalingrequests/finalizers,verbs=update

// Additional rbac
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;update;patch

// Reconcile is a function to move the current state of the cluster closer to the desired state,
// as part of the main kubernetes reconciliation loop.
//
// This function updates the cluster to the desired state according to the trigger set by SetupWithManager().
//
// Node Addition
//  1. Extract the target MachineDeployment resource from the Management Cluster.
//  2. If a matched MachineDeployment resource is found, scale up its replicas to +1.
//  3. If a matched MachineDeployment resource is NOT found, create a new MachineDeployment
//     and scale up its replicas and scale up to 1.
//  4. Watch Node resources of the Workload Cluster and update to "Created"
//     when ready replicas of the MachineDeployment resource is greater than or equal to desiredReplicas.
//
// Node Deletion
//  1. Get the target Node resource to be deleted in the Workload Cluster.
//  2. Get the MachineDeployment resource to be scaled down or to be deleted in the Management Cluster.
//  3. If there are other MachineDeployments and replicas of the MachineDeployment is 1,
//     delete the MachineDeployment and MachineTemplate if needed, and skip to 8.
//  4. Get the MachineSet resource to be deleted from the UID of the MachineDeployment resource.
//  5. Set "true" to annotation for pausing in the MachineSet resource.
//  6. Scale down its replicas to -1.
//  7. Delete the Machine resource binding the target Node resource in the Management Cluster.
//  8. Wait for deletion of the target Node resources with watching it.
//  9. Delete the annotation set in the MachineSet resource for pausing.
//  10. Update the value of Phase to "Deleted" when ready replicas of the MachineDeployment resource is
//     less than or equal to desiredReplicas.
//
//nolint:gocyclo,funlen // This is acceptable to be long function because this is main Reconcile function.
func (r *NodeScalingRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get NodeScalingRequest
	nsr := &scalev1.NodeScalingRequest{}
	if err := r.Get(ctx, req.NamespacedName, nsr); apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get NodeScalingRequest: %s: %w", req.String(), err)
	}
	// lastNSR is NodeScalingRequest that is stored in the etcd.
	// After calling Update() or Patch(), lastNSR should be updated.
	lastNSR := nsr.DeepCopy()

	// Check if the NodeScalingRequest is being deleted then finalize it.
	if finalized, err := r.doFinalize(ctx, nsr, lastNSR); err != nil {
		r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonFinalizerFailed, eventActionReconciling, err.Error())
		return ctrl.Result{}, fmt.Errorf("failed to finalize NodeScalingRequest: %w", err)
	} else if finalized {
		// If NodeScalingRequest is being deleted, we should not proceed with reconciliation.
		logger.Info("finalized")
		return ctrl.Result{}, nil
	} // else continue to reconcile

	var rslt = ctrl.Result{}
	var err = error(nil)
	var conditions = &nsr.Status.Conditions

	switch nsr.Status.Phase {
	case scalev1.PhaseNodeScaleInitial, scalev1.PhaseNodeScalePending:
		// initialize Reconcile
		if rslt, err = r.initializeReconcile(ctx, nsr, lastNSR); err != nil {
			err = fmt.Errorf("failed to initialize reconcile: %w", err)
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonInitializingFailed, eventActionReconciling, err.Error())
			break
		} else if !rslt.IsZero() {
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonPending, eventActionReconciling,
				"Pending this NodeScalingRequest until another one is completed")
			break
		}

		// Main process
		switch nsr.Spec.NodeRequestType {
		case scalev1.TypeCreate:
			rslt, err = r.startScaleUp(ctx, nsr, lastNSR)
			if err != nil {
				err = fmt.Errorf("failed to scale up node: %w", err)
				r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonScalingUpFailed, eventActionReconciling, err.Error())
			}

		case scalev1.TypeDelete:
			rslt, err = r.startScaleDown(ctx, nsr, lastNSR)
			if err != nil {
				err = fmt.Errorf("failed to scale down node: %s: %w", *nsr.Spec.Node, err)
				r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonScalingDownFailed, eventActionReconciling, err.Error())
			}

		default:
			// This case is only for failsafe, but should not happen because crd validation is checked.
			nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
			err = fmt.Errorf("nodeRequestType is invalid: NodeRequestType is %s", nsr.Spec.NodeRequestType)
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonInvalidType, eventActionReconciling, err.Error())
			updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "InvalidRequestType", err.Error())
		}

	case scalev1.PhaseNodeCreating:
		rslt, err = r.proceedScaleUp(ctx, nsr)
		if err != nil {
			err = fmt.Errorf("failed to proceed scale up node: %w", err)
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonScalingUpFailed, eventActionReconciling, err.Error())
		}

	case scalev1.PhaseNodeDeleting:
		rslt, err = r.proceedScaleDown(ctx, nsr)
		if err != nil {
			err = fmt.Errorf("failed to proceed scale down node: %s: %w", *nsr.Spec.Node, err)
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonScalingDownFailed, eventActionReconciling, err.Error())
		}

	case scalev1.PhaseNodeCreated, scalev1.PhaseNodeCreateFailed, scalev1.PhaseNodeDeleted, scalev1.PhaseNodeDeleteFailed:
		rslt, err = r.cleanUpAfterFinished(ctx, nsr)
		if err != nil {
			err = fmt.Errorf("failed to delete expired NodeScalingRequest: %w", err)
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonTerminationFailed, eventActionReconciling, err.Error())
		}
		return rslt, err // should return without patching

	default:
		// This case is only for failsafe, but should not happen because crd validation is checked.
		err = fmt.Errorf("phase is invalid: %s", nsr.Status.Phase)
		r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonInvalidPhase, eventActionReconciling, err.Error())
		updateStatusConditions(conditions, Initialized, metav1.ConditionFalse, "InvalidPhase", err.Error())
	}

	// patch NodeScalingRequest
	if statusErr := r.updateNodeScalingRequestStatus(ctx, nsr, lastNSR); statusErr != nil {
		rslt = ctrl.Result{} // override result to not requeue
		if err == nil {
			err = statusErr
		} else {
			err = fmt.Errorf("%w :%w", statusErr, err)
		}
		err = fmt.Errorf("failed to patch NodeScalingRequest status: %w", err)
		r.Recorder.Eventf(nsr, nil, corev1.EventTypeWarning, eventReasonUpdatingStatusFailed, eventActionReconciling, err.Error())
	}

	return rslt, err
}

// updateNodeScalingRequestStatus is a function to update status of NodeScalingRequest and synchronize
// cached resource version of NodeScalingRequest.
func (r *NodeScalingRequestReconciler) updateNodeScalingRequestStatus(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
) error {
	// Patch status of NodeScalingRequest
	if err := r.Status().Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
		return fmt.Errorf("failed to patch NodeScalingRequest status: %w", err)
	}

	if syncErr := r.syncCachedResourceVersion(ctx, nsr); syncErr != nil {
		return fmt.Errorf("failed to synchronize cached resource version of NodeScalingRequest %s/%s: %w",
			nsr.Namespace, nsr.Name, syncErr)
	}

	return nil
}

// syncCachedResourceVersion is a function to synchronize cached resource version of NodeScalingRequest
// after updating status of NodeScalingRequest.
func (r *NodeScalingRequestReconciler) syncCachedResourceVersion(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) error {
	return syncCachedResourceVersion(ctx, r.Client, nsr, &scalev1.NodeScalingRequest{})
}

// SetupWithManager is a function to set up the controller with the Manager.
//
// This controller will queue the appropriate custom resources (NodeScalingRequest) when:
//   - the custom resources (NodeScalingRequest) in the self cluster are updated.
//   - the node resources in the self cluster are updated.
//
// Environment variables to be set:
//   - PROVIDER_NAME: supporting provider name, available providers(DCI, Docker).
//   - NSR_TTL_SECONDS_AFTER_FINISHED: TTL seconds for NodeScalingRequest after its task is finished.
//   - MANAGEMENT_KUBECONFIG: path for config of management cluster kubeconfig file.
func (r *NodeScalingRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if mgr == nil {
		return errors.New("expected manager is not nil, but got nil")
	}
	// get information from environment variables
	logger := mgr.GetLogger()

	provider, err := utils.GetProviderTypeFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get provider type from environment variable: %w", err)
	}
	r.providerType = provider

	isEnabled, ttlSeconds, err := utils.GetTTLConfigFromEnv(nodeScalingRequestTTLEnvVarName)
	if err != nil {
		return fmt.Errorf("failed to get TTL configuration from environment variable: %w", err)
	}
	r.enableTTLSecondsAfterFinished = isEnabled
	r.ttlSecondsAfterFinished = ttlSeconds

	// print settings got from environment
	logger.Info("Supporting provider", "providerType", r.providerType)
	if r.enableTTLSecondsAfterFinished {
		logger.Info("TTL for NodeScalingRequest resource", "ttlSeconds", r.ttlSecondsAfterFinished)
	} else {
		logger.Info("TTL for NodeScalingRequest resource", "ttlSeconds", "disabled")
	}

	// check client for management cluster
	if r.ClientHelper == nil {
		return errors.New("expected ManagementClusterClient is not nil, but got nil")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&scalev1.NodeScalingRequest{}).
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(r.mapperNodeToNodeScalingRequest),
			// We need only create and delete events
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(_ event.CreateEvent) bool {
					return true
				},
				DeleteFunc: func(_ event.DeleteEvent) bool {
					return true
				},
				UpdateFunc: func(_ event.UpdateEvent) bool {
					return false
				},
				GenericFunc: func(_ event.GenericEvent) bool {
					return false
				},
			}),
		).
		Named("nodescalingrequest").
		Complete(r)
}

// mapperNodeToNodeScalingRequest is a function to map Node events to NodeScalingRequest reconcile requests.
func (r *NodeScalingRequestReconciler) mapperNodeToNodeScalingRequest(ctx context.Context, _ client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// Write error to log if NodeScalingRequest cannot be listed
	nodeScalingRequestList := &scalev1.NodeScalingRequestList{}
	if err := r.List(ctx, nodeScalingRequestList); err != nil {
		r.Recorder.Eventf(nodeScalingRequestList, nil, corev1.EventTypeWarning, eventReasonFetchingFailed, eventActionReconciling, err.Error())
		logger.Error(err, "failed to list NodeScalingRequests")
		return nil
	}

	requests := make([]reconcile.Request, 0, len(nodeScalingRequestList.Items))

	// Add all NodeScalingRequests which are not finished to the Reconcile queue
	for idx := range nodeScalingRequestList.Items {
		nodeScalingRequest := &nodeScalingRequestList.Items[idx]
		if nodeScalingRequest.IsFinished() {
			continue
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      nodeScalingRequest.Name,
				Namespace: nodeScalingRequest.Namespace,
			},
		})
	}
	return requests
}

// initializeReconcile is a function to initialize reconcile.
// NOTE:
//
//	r.Patch() to adding finalizer could cause update nsr of the 2nd parameter, nsr.Status will be nullified.
//	To avoid this, change nsr.Status after r.Patch().
func (r *NodeScalingRequestReconciler) initializeReconcile(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	// Check if processingNSRs NodeScalingRequest is running.
	processingNSRs, err := r.getProcessingNodeScalingRequests(ctx, nsr)
	if err != nil {
		err = fmt.Errorf("failed to get other NodeScalingRequest is running: %w", err)
		updateStatusConditions(&nsr.Status.Conditions, Initialized, metav1.ConditionFalse, "CheckingAnotherRequestFailed", err.Error())
		return ctrl.Result{}, err
	}
	if len(processingNSRs) > 0 {
		// Requeue later because another NodeScalingRequest is in progress.
		if nsr.Status.Phase != scalev1.PhaseNodeScalePending {
			nsr.Status.Phase = scalev1.PhaseNodeScalePending
			updateStatusConditions(&nsr.Status.Conditions, Initialized, metav1.ConditionFalse, "Pending", "")
		}
		return ctrl.Result{RequeueAfter: requeueIntervalNodeScalingRequestPending}, nil
	}

	// Get clusterNamespace and clusterName from Node resources.
	clusterName, clusterNamespace, err := r.ClientHelper.GetClusterNameAndNamespace(ctx)
	if err != nil {
		err := fmt.Errorf("failed to get cluster namespace: no node found with annotations %s and %s",
			capiv1.ClusterNameAnnotation, capiv1.ClusterNamespaceAnnotation)
		updateStatusConditions(&nsr.Status.Conditions, Initialized, metav1.ConditionFalse, "GettingClusterInfoFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Add finalizer if not present.
	if nsr.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(nsr, scalev1.NodeScalingRequestFinalizerName) {
		controllerutil.AddFinalizer(nsr, scalev1.NodeScalingRequestFinalizerName)
		// Update here to add finalizer in the metadata.
		// Note: r.Patch() will nullify nsr.Status of the 2nd parameter.
		if err := r.Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
			err = fmt.Errorf("failed to add finalizer: %w", err)
			updateStatusConditions(&nsr.Status.Conditions, Initialized, metav1.ConditionFalse, "AddingFinalizerFailed", err.Error())
			return ctrl.Result{}, err
		}
		// Update ObjectMeta of lastNSR to reflect the finalizer addition.
		nsr.ObjectMeta.DeepCopyInto(&lastNSR.ObjectMeta)
	}
	// -- change nsr.Status below here --
	nsr.Status.ClusterName = clusterName
	nsr.Status.ClusterNamespace = clusterNamespace
	updateStatusConditions(&nsr.Status.Conditions, Initialized, metav1.ConditionTrue, "InitializationSucceeded", "")
	return ctrl.Result{}, nil
}

// getProcessingNodeScalingRequests is a function to get processing NodeScalingRequests except nsrExcept.
func (r *NodeScalingRequestReconciler) getProcessingNodeScalingRequests(
	ctx context.Context, nsrExcept *scalev1.NodeScalingRequest,
) ([]scalev1.NodeScalingRequest, error) {
	nsrList := &scalev1.NodeScalingRequestList{}
	if err := r.List(ctx, nsrList); err != nil { // List NodeScalingRequests from all namespaces
		return nil, fmt.Errorf("failed to list NodeScalingRequests: %w", err)
	}
	processingNSRs := []scalev1.NodeScalingRequest{}
	for idx := range nsrList.Items {
		nsr := &nsrList.Items[idx]
		if nsrExcept != nil && nsr.Name == nsrExcept.Name && nsr.Namespace == nsrExcept.Namespace {
			continue // skip self resource
		}
		if nsr.IsProcessing() {
			processingNSRs = append(processingNSRs, *nsr)
		}
		if nsr.IsNotStarted() && meta.FindStatusCondition(nsr.Status.Conditions, Requested) != nil {
			// This case is assumed that the request to the management cluster is about to be made,
			// whether the request was made or not before the phase becomes "Creating" or "Deleting"
			// including "pre-request" status.
			processingNSRs = append(processingNSRs, *nsr)
		}
	}
	return processingNSRs, nil
}

// startScaleUp is a function to create a node which matches requested resources.
//
//nolint:funlen // This function is acceptable to be long because it has multiple steps.
func (r *NodeScalingRequestReconciler) startScaleUp(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	location := nsr.Spec.Location
	serverSpec := nsr.Spec.ServerSpec
	clusterNamespace := nsr.Status.ClusterNamespace
	clusterName := nsr.Status.ClusterName
	conditions := &nsr.Status.Conditions

	logger := log.FromContext(ctx)
	logger.Info("startScaleUp",
		"location", location,
		"serverSpec", serverSpec,
		"clusterNamespace", clusterNamespace,
		"clusterName", clusterName)

	// check parameter
	if serverSpec == nil || (r.providerType == utils.ProviderTypeDocker && location == nil) {
		// This case is only for failsafe when DCI Provider mode is enabled
		err := fmt.Errorf("parameter is nil: .spec.location(%v) .spec.serverSpec(%v)", location, serverSpec)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidationFailed", err.Error())
		nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
		return ctrl.Result{}, err
	}

	// get MachineTemplate list in the target clusterNamespace
	machineTemplates, machineTemplateName, err := r.getInfraMachineTemplateAllocatable(ctx,
		location, serverSpec, clusterNamespace, conditions)
	if err != nil {
		// error is already set in getInfraMachineTemplateAllocatable()
		return ctrl.Result{}, err
	}

	// Get Allocatable MachineDeployments
	mdsAllocatable, mdName, mdBase, err := r.ClientHelper.GetMachineDeploymentsAllocatable(ctx,
		clusterNamespace, utils.MachineTemplateNameMap[r.providerType], clusterName, machineTemplates)
	if err != nil {
		err = fmt.Errorf("failed to get valid MachineDeployment list to scale up: %w", err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "TargetResourceNotFound", err.Error())
		return ctrl.Result{}, err
	}
	if len(mdsAllocatable) == 0 {
		// not found MachineDeployment with satisfying request and create a new resource
		return r.startScaleUpWithCreatingResource(ctx, nsr, lastNSR, mdBase, machineTemplates, mdName, machineTemplateName)
	}

	// select the first MachineDeployment from allocatable list
	var mdReplicas int32
	md := &mdsAllocatable[0]
	if md.Spec.Replicas == nil {
		mdReplicas = 0
	} else {
		mdReplicas = *md.Spec.Replicas
	}

	// found MachineDeployment with satisfying request and scale up its replicas
	// Get progress condition of the MachineDeployment
	var replicas int32
	switch progress := getProgressMachineDeployment(md, nsr.GetUID()); progress.state {
	case notInProgress:
		replicas = mdReplicas + 1
	case inProgressByThisNSR:
		// If MachineDeployment is in progress by this NodeScalingRequest, it is assumed that we already requested scale
		// but status of NodeScalingRequest was failed to update or access to management cluster resources was failed.
		logger.Info("MachineDeployment is in progress by this NodeScalingRequest", "MachineDeployment", md.Name)
		// Current replicas of MD should be desired replicas and continue to scale up.
		replicas = mdReplicas
	case inProgressByOtherNSR:
		// If MachineDeployment is in progress by other NodeScalingRequest, we will only requeue by returning error.
		err := fmt.Errorf("MachineDeployment (%s) is in progress by other NodeScalingRequest (UID=%s)", md.Name, progress.uid)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ScalingRequestFailed", err.Error())
		return ctrl.Result{}, err
	default:
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("progress condition (%v) is invalid", progress)
	}

	// If MachineDeployment is not in progress or in progress by this NodeScalingRequest, we will scale up its replicas.

	// Patch pre-request status
	// This is needed to cleanup the MachineDeployment by finalizer.
	updateStatusConditions(conditions, Requested, metav1.ConditionUnknown, "Requesting", "")
	nsr.Status.DesiredReplicas = &replicas
	nsr.Status.MachineDeploymentName = md.Name
	if err := r.Status().Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
		err = fmt.Errorf("failed to patch pre-request status: %w", err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "PatchingStatusFailed", err.Error())
		return ctrl.Result{}, err
	}
	nsr.Status.DeepCopyInto(&lastNSR.Status)

	// Update replicas and set in-progress label of MachineDeployment
	lastMD := md.DeepCopy()
	md.Spec.Replicas = &replicas
	md.Labels = utils.AddNodeScalingRequestInProgressUIDLabel(md.Labels, nsr.GetUID())
	if err := r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
		err = fmt.Errorf("valid MachineDeployment(%s) found, but failed to scale up: clusterName(%s), clusterNamespace(%s): %w",
			md.Name, clusterName, clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ScalingRequestFailed", err.Error())
		return ctrl.Result{}, err
	}

	// succeed to request to management cluster, now update phase and conditions
	updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", "")
	nsr.Status.Phase = scalev1.PhaseNodeCreating
	r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingUpStarted, eventActionReconciling,
		"MachineDeployment(%s) is requested to scale up to %d replicas", md.Name, *md.Spec.Replicas)
	return ctrl.Result{}, nil // Success
}

// getInfraMachineTemplateAllocatable is a function to get allocatable InfraMachineTemplate list in the target clusterNamespace.
// Supporting providers: DCI, Docker.
func (r *NodeScalingRequestReconciler) getInfraMachineTemplateAllocatable(
	ctx context.Context, location *string, serverSpec *rpmv1.ServerSpec, clusterNamespace string, conditions *[]metav1.Condition,
) (machineTemplates []client.Object, machineTemplateName string, err error) {
	// get MachineTemplate list in the target clusterNamespace
	switch r.providerType {
	case utils.ProviderTypeDCI:
		machineTemplates, machineTemplateName, err = r.ClientHelper.GetDCIMachineTemplatesAllocatable(ctx, location, serverSpec, clusterNamespace)
		if err != nil {
			err = fmt.Errorf("failed to request DCIMachineTemplate list in clusterNamespace(%s) through client: %w",
				clusterNamespace, err)
			updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		}
	case utils.ProviderTypeDocker:
		machineTemplates, machineTemplateName, err = r.ClientHelper.GetDockerMachineTemplatesAllocatable(ctx, *location, clusterNamespace)
		if err != nil {
			err = fmt.Errorf("failed to request DockerMachineTemplate list in clusterNamespace(%s) through client: %w",
				clusterNamespace, err)
			updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		}
	default:
		// This case is only for failsafe, but should not happen because provider type is checked at SetupWithManager().
		err = fmt.Errorf("provider type(%s) is invalid", r.providerType)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "InvalidProviderType", err.Error())
	}
	return machineTemplates, machineTemplateName, err
}

// startScaleUpWithCreatingResource is a function to create new resource for request.
//
// This function should be called only from startScaleUp().
func (r *NodeScalingRequestReconciler) startScaleUpWithCreatingResource(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
	mdBase *capiv1.MachineDeployment, dmts []client.Object, mdName, dmtName string,
) (ctrl.Result, error) {
	clusterNamespace := nsr.Status.ClusterNamespace
	conditions := &nsr.Status.Conditions

	logger := log.FromContext(ctx)

	if mdBase == nil {
		// This case is only for failsafe
		err := fmt.Errorf("parameter is nil: mdBase(%v)", mdBase)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidationFailed", err.Error())
		nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
		return ctrl.Result{}, err
	}

	// Patch pre-request status
	// This is needed to cleanup the MachineDeployment by finalizer in case of failure to update status of NodeScalingRequest.
	updateStatusConditions(conditions, Requested, metav1.ConditionUnknown, "Requesting", "")
	nsr.Status.MachineDeploymentName = mdName
	nsr.Status.DesiredReplicas = ptr.To(int32(1)) // set 1 replica for new MachineDeployment
	if err := r.Status().Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
		err = fmt.Errorf("failed to patch pre-request status: %w", err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "PatchingStatusFailed", err.Error())
		return ctrl.Result{}, err
	}
	nsr.Status.DeepCopyInto(&lastNSR.Status)

	// Create a new MachineTemplate if needed
	if len(dmts) == 0 {
		// Create a new MachineTemplate
		switch r.providerType {
		case utils.ProviderTypeDCI:
			if err := r.ClientHelper.CreateDCIMachineTemplateFromNodeScalingRequest(ctx, dmtName, clusterNamespace, nsr); err != nil {
				err = fmt.Errorf("failed to create DCIMachineTemplate(%s) in clusterNamespace(%s): %w", dmtName, clusterNamespace, err)
				updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "CreatingResourceFailed", err.Error())
				return ctrl.Result{}, err
			}
		case utils.ProviderTypeDocker:
			if err := r.ClientHelper.CreateDockerMachineTemplateFromNodeScalingRequest(ctx, dmtName, clusterNamespace, nsr); err != nil {
				err = fmt.Errorf("failed to create DockerMachineTemplate(%s) in clusterNamespace(%s): %w", dmtName, clusterNamespace, err)
				updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "CreatingResourceFailed", err.Error())
				return ctrl.Result{}, err
			}
		default:
			// This case is only for failsafe, but should not happen because provider type is checked at SetupWithManager().
			err := fmt.Errorf("provider type(%s) is invalid", r.providerType)
			updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "InvalidProviderType", err.Error())
			return ctrl.Result{}, err
		}
	} else {
		dmtName = dmts[0].GetName()
	}

	// create a new MachineDeployment spec from NodeScalingRequest
	if err := r.ClientHelper.CreateMachineDeploymentFromNodeScalingRequest(ctx, mdName, clusterNamespace, dmtName, nsr, mdBase); err != nil {
		err = fmt.Errorf("failed to create MachineDeployment(%s) in clusterNamespace(%s): %w", mdName, clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "CreatingResourceFailed", err.Error())
		return ctrl.Result{}, err
	}

	// succeed to request to management cluster, now update phase and conditions
	logger.Info("MachineDeployment is requested to scale up", "name", mdName, "desiredReplicas", 1)
	updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "CreatingResourceSucceeded", "")
	nsr.Status.Phase = scalev1.PhaseNodeCreating
	r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingUpStarted, eventActionReconciling,
		"MachineDeployment(%s) is requested to scale up to %d replicas", mdName, 1)
	return ctrl.Result{}, nil
}

// startScaleDown is a function to delete the node from node name.
func (r *NodeScalingRequestReconciler) startScaleDown(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	nodeName := nsr.Spec.Node
	clusterNamespace := nsr.Status.ClusterNamespace
	clusterName := nsr.Status.ClusterName
	conditions := &nsr.Status.Conditions

	logger := log.FromContext(ctx)
	logger.Info("startScaleDown", "nodeName", nodeName, "clusterNamespace", clusterNamespace, "clusterName", clusterName)

	params, err := r.getScaleDownResources(ctx, nsr)
	if err != nil {
		err = fmt.Errorf("failed to get Information to scale down: %w", err)
		return ctrl.Result{}, err
	}
	if params == nil {
		// already deleted, and phase is updated in getScaleDownResources()
		return ctrl.Result{}, nil
	}
	md := params.machineDeployment

	// Get progress condition of the MachineDeployment
	var replicas int32
	switch progress := getProgressMachineDeployment(md, nsr.GetUID()); progress.state {
	case notInProgress:
		replicas = *md.Spec.Replicas - 1
	case inProgressByThisNSR:
		// If MachineDeployment is in progress by this NodeScalingRequest, it is assumed that we already requested scale
		// but status of NodeScalingRequest was failed to update or access to management cluster resources was failed.
		logger.Info("MachineDeployment is in progress by this NodeScalingRequest", "MachineDeployment", md.Name)
		// Current replicas of MD should be desired replicas and continue to scale down.
		replicas = *md.Spec.Replicas
	case inProgressByOtherNSR:
		// If MachineDeployment is in progress by other NodeScalingRequest, we will only requeue by returning error.
		err = fmt.Errorf("MachineDeployment (%s) is in progress by other NodeScalingRequest (UID=%s)", md.Name, progress.uid)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ScalingRequestFailed", err.Error())
		return ctrl.Result{}, err
	default:
		// This case is only for failsafe
		return ctrl.Result{}, fmt.Errorf("progress condition (%v) is invalid", progress)
	}

	if replicas == 0 && !utils.HasCardScalingRequestInProgressUIDAnnotation(md.Annotations) {
		// check if MachineDeployment is the last one or not
		var isLast bool
		if isLast, err = r.ClientHelper.IsLastMachineDeployment(ctx, md, clusterNamespace, clusterName); err != nil {
			err = fmt.Errorf("failed to check if valid MachineDeployment exist or not: %w", err)
			updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ValidResourceNotFound", err.Error())
			return ctrl.Result{}, err
		} else if !isLast {
			// delete MachineDeployment (and MachineTemplate) if needed
			if err = r.deleteMachineDeploymentAndMachineTemplate(ctx, md); err != nil {
				err = fmt.Errorf("failed to scale down by deleting MachineDeployment(%s) in Namespace(%s): %w",
					md.Name, md.Namespace, err)
				updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ScalingRequestFailed", err.Error())
				return ctrl.Result{}, err
			}
			updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", "")
			nsr.Status.Phase = scalev1.PhaseNodeDeleting
			nsr.Status.DesiredReplicas = ptr.To(int32(0))
			nsr.Status.MachineDeploymentName = md.Name
			r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingDownStarted, eventActionReconciling,
				"MachineDeployment(%s) is requested to be deleted", md.Name)
			return ctrl.Result{}, nil // success
		} // exist only 1 MachineDeployment
	}

	// If MachineDeployment is not in progress or in progress by this NodeScalingRequest, we will scale down its replicas.
	return r.deleteMachine(ctx, nsr, lastNSR, params, &replicas)
}

// getScaleDownResources is a function to get information needed to scale down.
func (r *NodeScalingRequestReconciler) getScaleDownResources(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (*scaleDownResources, error) {
	nodeName := nsr.Spec.Node
	clusterNamespace := nsr.Status.ClusterNamespace
	clusterName := nsr.Status.ClusterName
	conditions := &nsr.Status.Conditions

	logger := log.FromContext(ctx)

	// check parameter
	if nodeName == nil {
		// This case is only for failsafe
		err := fmt.Errorf("parameter is nil: .spec.node(%v)", nodeName)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidationFailed", err.Error())
		nsr.Status.Phase = scalev1.PhaseNodeDeleteFailed
		return nil, err
	}

	// Get the target node in workload cluster
	node, err := r.ClientHelper.GetNode(ctx, *nodeName)
	if err != nil {
		err = fmt.Errorf("failed to request Node list to WorkloadCluster: %w", err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return nil, err
	}

	// Get the target Machine resource from the target node
	machine, err := r.ClientHelper.GetMachineFromNode(ctx, node)
	if apierrors.IsNotFound(err) {
		// the Machine is already deleted
		// we can't determine the MachineDeployment to scale down
		logger.Info("Machine already deleted")
		updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "MachineAlreadyDeleted", "")
		nsr.Status.Phase = scalev1.PhaseNodeDeleting
		return nil, nil
	} else if err != nil {
		err = fmt.Errorf("failed to request Machine resource in clusterNamespace(%s) through client: %w", clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return nil, err
	}

	// Get the target MachineDeployment from the target Machine
	md, err := r.ClientHelper.GetMachineDeploymentFromMachine(ctx, machine, clusterNamespace)
	if apierrors.IsNotFound(err) {
		// the MachineDeployment is already deleted
		// nothing to do for scaling down
		logger.Info("MachineDeployment already deleted", "ClusterNamespace", clusterNamespace)
		updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "MachineDeploymentAlreadyDeleted", "")
		nsr.Status.Phase = scalev1.PhaseNodeDeleting
		return nil, nil
	} else if err != nil {
		err = fmt.Errorf("failed to request MachineDeployment in clusterNamespace(%s) through client: %w", clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return nil, err
	}

	// Validate the target MachineDeployment
	if md.Spec.ClusterName != clusterName {
		err = fmt.Errorf("not found valid clusterName(%s) for MachineDeployment(%s)", clusterName, md.Name)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ValidResourceNotFound", err.Error())
		return nil, err
	}
	if md.Spec.Replicas == nil {
		err = fmt.Errorf("not found valid Replicas for MachineDeployment(%s): Replicas is nil", md.Name)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ValidResourceNotFound", err.Error())
		return nil, err
	}
	if *md.Spec.Replicas <= 0 {
		err = fmt.Errorf("not found valid Replicas for MachineDeployment(%s), Replicas=%d", md.Name, *md.Spec.Replicas)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ValidResourceNotFound", err.Error())
		return nil, err
	}

	return &scaleDownResources{
		machineDeployment: md,
		machine:           machine,
	}, nil
}

// deleteMachine is a function to scale down by deleting Machine and updating MachineDeployment.
func (r *NodeScalingRequestReconciler) deleteMachine(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest, resources *scaleDownResources, replicas *int32,
) (ctrl.Result, error) {
	clusterNamespace := nsr.Status.ClusterNamespace
	conditions := &nsr.Status.Conditions
	md := resources.machineDeployment
	machine := resources.machine

	logger := log.FromContext(ctx)

	// Patch pre-request status
	// This is needed to cleanup the MachineDeployment and MachineSet by finalizer.
	updateStatusConditions(conditions, Requested, metav1.ConditionUnknown, "Requesting", "")
	nsr.Status.DesiredReplicas = replicas
	nsr.Status.MachineDeploymentName = md.Name
	nsr.Status.MachineDeploymentUID = &md.UID
	if err := r.Status().Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
		err = fmt.Errorf("failed to patch pre-request status: %w", err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "PatchingStatusFailed", err.Error())
		return ctrl.Result{}, err
	}
	nsr.Status.DeepCopyInto(&lastNSR.Status)

	// get the target MachineSet and set "paused" annotation to it
	ms, err := r.ClientHelper.GetMachineSetFromUID(ctx, clusterNamespace, md.UID)
	if err != nil {
		err = fmt.Errorf("failed to get valid MachineSet with MachineDeployment UID(%v) in namespace(%s): %w", md.UID, clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "TargetResourceNotFound", err.Error())
		return ctrl.Result{}, err
	}
	if ms == nil {
		logger.Info("not found target MachineSet", "MachineDeploymentUID", md.UID, "namespace", clusterNamespace)
		// continue to scale down MachineDeployment
	} else {
		// set "paused" annotation
		lastMS := ms.DeepCopy()
		ms.Annotations = utils.AddClusterAPIPausedAnnotation(ms.Annotations)
		if err := r.ClientHelper.PatchMachineSet(ctx, ms, lastMS); err != nil {
			err = fmt.Errorf("failed to patch annotation(%s) to true through client: %w", capiv1.PausedAnnotation, err)
			updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
			return ctrl.Result{}, err
		}
	}

	// Update replicas and set in-progress label of MachineDeployment
	lastMD := md.DeepCopy()
	md.Spec.Replicas = replicas
	md.Labels = utils.AddNodeScalingRequestInProgressUIDLabel(md.Labels, nsr.GetUID())
	if err := r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
		err = fmt.Errorf("failed to scale down MachineDeployment(%s) in Namespace(%s): %w", md.Name, clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "ScalingRequestFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Delete the target Machine
	if err := r.ClientHelper.DeleteMachine(ctx, machine); err != nil {
		err = fmt.Errorf("failed to delete Machine(%s) in Namespace(%s): %w", machine.Name, clusterNamespace, err)
		updateStatusConditions(conditions, Requested, metav1.ConditionFalse, "DeletingMachineFailed", err.Error())
		return ctrl.Result{}, err
	}

	// succeed to request to management cluster, now update phase and conditions
	updateStatusConditions(conditions, Requested, metav1.ConditionTrue, "ScalingRequestSucceeded", "")
	nsr.Status.Phase = scalev1.PhaseNodeDeleting
	r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingDownStarted, eventActionReconciling,
		"MachineDeployment(%s) is requested to scale down to %d replicas", md.Name, *md.Spec.Replicas)
	return ctrl.Result{}, nil
}

// proceedScaleUp is a function to update status if creating node is finished or not.
func (r *NodeScalingRequestReconciler) proceedScaleUp(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	replicas := nsr.Status.DesiredReplicas
	machineDeploymentName := nsr.Status.MachineDeploymentName
	clusterNamespace := nsr.Status.ClusterNamespace
	clusterName := nsr.Status.ClusterName
	conditions := &nsr.Status.Conditions

	logger := log.FromContext(ctx)
	logger.Info("proceedScaleUp", "replicas", replicas, "machineDeploymentName", machineDeploymentName,
		"clusterNamespace", clusterNamespace, "clusterName", clusterName)

	// check parameter
	if replicas == nil {
		// This case is only for failsafe
		err := fmt.Errorf("parameter is nil: .status.desiredReplicas(%v)", replicas)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidationFailed", err.Error())
		nsr.Status.Phase = scalev1.PhaseNodeCreateFailed
		return ctrl.Result{}, err
	}

	// Get MachineDeployment in the target clusterNamespace
	md, err := r.ClientHelper.GetMachineDeployment(ctx, machineDeploymentName, clusterNamespace)
	if err != nil {
		err = fmt.Errorf("not found target MachineDeployment(%s) in clusterNamespace(%s): %w", machineDeploymentName, clusterNamespace, err)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return ctrl.Result{}, err
	}

	// check if ReadyReplicas of MD becomes desired value
	if md.Status.ReadyReplicas == nil || *replicas > *md.Status.ReadyReplicas {
		// requeue with const interval when ReadyReplicas is not expected value yet
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "Requeueing", "")
		return ctrl.Result{RequeueAfter: requeueIntervalNodeScalingRequestWaitingScaling}, nil
	}

	// remove in-progress labels from MachineDeployment
	lastMD := md.DeepCopy()
	utils.RemoveNodeScalingRequestInProgressUIDLabel(md.Labels)
	if err := r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
		err = fmt.Errorf("failed to remove in-progress labels from MachineDeployment: %s: %w", md.Name, err)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return ctrl.Result{}, err
	}

	logger.Info("succeed to scale up", "machineDeploymentName", machineDeploymentName)
	updateStatusConditions(conditions, Completed, metav1.ConditionTrue, "CompletionSucceeded", "")
	nsr.Status.Phase = scalev1.PhaseNodeCreated
	r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingUpCompleted, eventActionReconciling,
		"MachineDeployment(%s) is scaled up to %d replicas", md.Name, *md.Spec.Replicas)
	return ctrl.Result{RequeueAfter: (time.Duration(r.ttlSecondsAfterFinished)) * time.Second}, nil
}

// proceedScaleDown is a function to update status if deleting node is finished or not.
//
//nolint:gocyclo,funlen // This is acceptable to be long function because this is a part of main Reconcile function.
func (r *NodeScalingRequestReconciler) proceedScaleDown(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	replicas := nsr.Status.DesiredReplicas
	nodeName := nsr.Spec.Node
	machineDeploymentName := nsr.Status.MachineDeploymentName
	clusterNamespace := nsr.Status.ClusterNamespace
	clusterName := nsr.Status.ClusterName
	conditions := &nsr.Status.Conditions
	mdUID := nsr.Status.MachineDeploymentUID

	logger := log.FromContext(ctx)
	logger.Info("proceedScaleDown", "replicas", replicas, "nodeName", nodeName, "machineDeploymentName", machineDeploymentName,
		"clusterNamespace", clusterNamespace, "clusterName", clusterName)

	// check parameter
	if nodeName == nil {
		// This case is only for failsafe
		err := errors.New("spec.node is nil")
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidationFailed", err.Error())
		nsr.Status.Phase = scalev1.PhaseNodeDeleteFailed
		return ctrl.Result{}, err
	}

	// get MachineDeployment in the target clusterNamespace
	md := &capiv1.MachineDeployment{}
	if machineDeploymentName == "" {
		// Machine or MachineDeployment that is target of scale down is already deleted
		md = nil
	} else {
		var err error
		md, err = r.ClientHelper.GetMachineDeployment(ctx, machineDeploymentName, clusterNamespace)
		if apierrors.IsNotFound(err) {
			md = nil
		} else if err != nil {
			err = fmt.Errorf("not found target MachineDeployment(%s) in clusterNamespace(%s): %w", machineDeploymentName, clusterNamespace, err)
			updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
			return ctrl.Result{}, err
		}
	}
	// validate MachineDeployment in case
	if md != nil && md.Spec.ClusterName != clusterName {
		err := fmt.Errorf("not match clusterName(%s) with clusterName(%s) in MachineDeployment(%s)",
			clusterName, md.Spec.ClusterName, machineDeploymentName)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "ValidResourceNotFound", err.Error())
		return ctrl.Result{}, err
	}

	// get Node in workload cluster
	node, err := r.ClientHelper.GetNode(ctx, *nodeName)
	if apierrors.IsNotFound(err) {
		node = nil
	} else if err != nil {
		// error happened when get nodes
		err = fmt.Errorf("failed to get Node(%s) for reasons other than 'not found': %w", *nodeName, err)
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
		return ctrl.Result{}, err
	}

	if node != nil {
		// target node is still alive
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "Requeueing", "")
		return ctrl.Result{}, nil // expecting event of node deletion
	}

	// remove paused annotation from MachineSet
	if mdUID != nil {
		ms, err := r.ClientHelper.GetMachineSetFromUID(ctx, clusterNamespace, *mdUID)
		if err != nil {
			err = fmt.Errorf("failed to get valid MachineSet with MachineDeployment UID(%v) in namespace(%s): %w", *mdUID, clusterNamespace, err)
			updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
			return ctrl.Result{}, err
		}
		if ms != nil {
			lastMS := ms.DeepCopy()
			utils.RemoveClusterAPIPausedAnnotation(ms.Annotations)
			if err := r.ClientHelper.PatchMachineSet(ctx, ms, lastMS); err != nil {
				err = fmt.Errorf("failed to remove paused annotation from MachineSet binding MachineDeployment UID(%v) in namespace(%s): %w",
					*mdUID, clusterNamespace, err)
				updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
				return ctrl.Result{}, err
			}
		}
	}

	// check if ReadyReplicas of MD becomes desired value
	if md != nil && replicas != nil && md.Status.ReadyReplicas != nil && *replicas < *md.Status.ReadyReplicas {
		// requeue with const interval
		updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "Requeueing", "")
		return ctrl.Result{RequeueAfter: requeueIntervalNodeScalingRequestWaitingScaling}, nil
	}

	// remove in-progress labels from MachineDeployment
	if md != nil {
		lastMD := md.DeepCopy()
		utils.RemoveNodeScalingRequestInProgressUIDLabel(md.Labels)
		if err := r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
			err = fmt.Errorf("failed to remove in-progress labels from MachineDeployment: %s: %w", lastMD.Name, err)
			updateStatusConditions(conditions, Completed, metav1.ConditionFalse, "RequestingClientFailed", err.Error())
			return ctrl.Result{}, err
		}
	}

	// the target Node is deleted successfully
	logger.Info("succeed to scale down the target node", "nodeName", *nodeName)
	updateStatusConditions(conditions, Completed, metav1.ConditionTrue, "CompletionSucceeded", "")
	nsr.Status.Phase = scalev1.PhaseNodeDeleted
	r.Recorder.Eventf(nsr, nil, corev1.EventTypeNormal, eventReasonScalingDownCompleted, eventActionReconciling,
		"Node(%s) is deleted successfully", *nodeName)
	return ctrl.Result{RequeueAfter: (time.Duration(r.ttlSecondsAfterFinished)) * time.Second}, nil
}

// deleteMachineDeploymentAndMachineTemplate is a function to delete MachineDeployment and its MachineTemplate if needed.
func (r *NodeScalingRequestReconciler) deleteMachineDeploymentAndMachineTemplate(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	// delete MachineTemplate which is not used by other MachineDeployment
	if err := r.deleteMachineTemplateUnused(ctx, md); err != nil {
		return fmt.Errorf("failed to delete %sMachineTemplate for MachineDeployment(%s) in Namespace(%s): %w",
			r.providerType, md.Name, md.Namespace, err)
	}

	// delete MachineDeployment
	if err := r.ClientHelper.DeleteMachineDeployment(ctx, md); apierrors.IsNotFound(err) {
		return nil // not exist already
	} else if err != nil {
		return fmt.Errorf("failed to delete MachineDeployment(%s) in Namespace(%s): %w", md.Name, md.Namespace, err)
	}
	return nil // succeed to delete
}

// deleteMachineTemplateUnused is a function to delete MachineTemplate if it is not referred by other MachineDeployment.
func (r *NodeScalingRequestReconciler) deleteMachineTemplateUnused(
	ctx context.Context, md *capiv1.MachineDeployment,
) error {
	switch r.providerType {
	case utils.ProviderTypeDCI:
		return r.ClientHelper.DeleteDCIMachineTemplateUnused(ctx, md)
	case utils.ProviderTypeDocker:
		return r.ClientHelper.DeleteDockerMachineTemplateUnused(ctx, md)
	default:
		// This case is only for failsafe, but should not happen because provider type is checked at SetupWithManager().
		return fmt.Errorf("provider type(%s) is invalid", r.providerType)
	}
}

// cleanUpAfterFinished is a function to delete NodeScalingRequest after finished.
//
// This function will do nothing if IsEnableTTLAfterFinished() return false.
func (r *NodeScalingRequestReconciler) cleanUpAfterFinished(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) (ctrl.Result, error) {
	if !r.enableTTLSecondsAfterFinished || len(nsr.OwnerReferences) != 0 {
		// skip this function if TTL is disabled, or nsr has owner references
		return ctrl.Result{}, nil
	}

	logger := log.FromContext(ctx)

	// get last transition time
	condition := meta.FindStatusCondition(nsr.Status.Conditions, Completed)
	if condition == nil {
		// This case is only for failsafe, condition always should not be nil when finalized
		return ctrl.Result{}, errors.New("invalid NodeScalingRequest: should contain Condition with 'Completed'")
	}
	timeToExpire := condition.LastTransitionTime.Add(time.Duration(r.ttlSecondsAfterFinished) * time.Second)

	// check if ttl is expired
	now := time.Now()
	if !now.After(timeToExpire) {
		return ctrl.Result{RequeueAfter: timeToExpire.Sub(now)}, nil // recalculate remain seconds
	}
	if err := r.Delete(ctx, nsr); apierrors.IsNotFound(err) {
		logger.Info("Expired NodeScalingRequest is already deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete expired NodeScalingRequest: %w", err)
	}

	logger.Info("Expired NodeScalingRequest is deleted successfully")
	return ctrl.Result{}, nil
}

// doFinalize cleanups related resources and removes finalizer if NodeScalingRequest is being deleted
// and returns whether the NodeScalingRequest is being deleted or not.
func (r *NodeScalingRequestReconciler) doFinalize(
	ctx context.Context, nsr, lastNSR *scalev1.NodeScalingRequest,
) (bool, error) {
	// Check if the NodeScalingRequest is being deleted.
	if nsr.DeletionTimestamp.IsZero() {
		return false, nil // not being deleted
	}
	// Check if finalizer is set by this controller
	if !controllerutil.ContainsFinalizer(nsr, scalev1.NodeScalingRequestFinalizerName) {
		return true, nil // finalizer is not set by this controller, so skip finalization
	}
	// Clean up related resources before removing finalizer.
	if err := r.cleanupOnFinalize(ctx, nsr); err != nil {
		return true, fmt.Errorf("failed to clean up on finalize: %w", err)
	}
	// Remove finalizer.
	controllerutil.RemoveFinalizer(nsr, scalev1.NodeScalingRequestFinalizerName)
	if err := r.Patch(ctx, nsr, client.MergeFrom(lastNSR)); err != nil {
		return true, fmt.Errorf("failed to remove finalizer: %w", err)
	}
	// Update ObjectMeta of lastNSR to reflect the finalizer removal.
	nsr.ObjectMeta.DeepCopyInto(&lastNSR.ObjectMeta)
	return true, nil // succeed to finalize
}

// clean up related resources when finalizing NodeScalingRequest.
func (r *NodeScalingRequestReconciler) cleanupOnFinalize(
	ctx context.Context, nsr *scalev1.NodeScalingRequest,
) error {
	logger := log.FromContext(ctx)
	if nsr.IsFinished() {
		// Nothing to clean up
		return nil
	}
	clusterNamespace := nsr.Status.ClusterNamespace
	if clusterNamespace == "" {
		return nil
	}

	// Clean up MachineSet
	if mdUID := nsr.Status.MachineDeploymentUID; mdUID != nil {
		logger.Info("cleaning up MachineSet", "mdUID", *mdUID, "Namespace", clusterNamespace)
		if err := r.cleanupMachineSet(ctx, mdUID, clusterNamespace); err != nil {
			return fmt.Errorf("failed to clean up MachineSet: %w", err)
		}
	}

	// Clean up MachineDeployment
	if mdName := nsr.Status.MachineDeploymentName; mdName != "" {
		logger.Info("cleaning up MachineDeployment", "MachineDeploymentName", mdName, "Namespace", clusterNamespace)
		if err := r.cleanupMachineDeployment(ctx, mdName, clusterNamespace, nsr.GetUID()); err != nil {
			return fmt.Errorf("failed to clean up MachineDeployment: %w", err)
		}
	}

	return nil
}

// clean up MachineDeployment
func (r *NodeScalingRequestReconciler) cleanupMachineDeployment(
	ctx context.Context, mdName, mdNamespace string, nsrUID types.UID,
) error {
	// Get MachineDeployment
	md, err := r.ClientHelper.GetMachineDeployment(ctx, mdName, mdNamespace)
	if apierrors.IsNotFound(err) {
		return nil // MachineDeployment not found, nothing to clean up
	} else if err != nil {
		return fmt.Errorf("failed to get MachineDeployment (%s) in namespace (%s): %w", mdName, mdNamespace, err)
	}

	// Check progress state
	if progress := getProgressMachineDeployment(md, nsrUID); progress.state != inProgressByThisNSR {
		return nil // MachineDeployment is not in progress by this NodeScalingRequest, nothing to clean up
	}

	// Remove in-progress labels from MD if in progress by this NodeScalingRequest
	lastMD := md.DeepCopy()
	utils.RemoveNodeScalingRequestInProgressUIDLabel(md.Labels)
	if err := r.ClientHelper.PatchMachineDeployment(ctx, md, lastMD); err != nil {
		return fmt.Errorf("failed to patch MachineDeployment (%s) in namespace (%s): %w", mdName, mdNamespace, err)
	}

	return nil
}

// clean up MachineSet
func (r *NodeScalingRequestReconciler) cleanupMachineSet(
	ctx context.Context, mdUID *types.UID, msNamespace string,
) error {
	// Get MachineSet owned by specified MachineDeployment
	ms, err := r.ClientHelper.GetMachineSetFromUID(ctx, msNamespace, *mdUID)
	if err != nil {
		return fmt.Errorf("failed to get MachineSet binding MachineDeployment UID (%v) in namespace (%s): %w", *mdUID, msNamespace, err)
	}
	if ms == nil {
		return nil // MachineSet not found, nothing to clean up
	}

	// Remove paused annotation from MachineSet
	lastMS := ms.DeepCopy()
	utils.RemoveClusterAPIPausedAnnotation(ms.Annotations)
	if err := r.ClientHelper.PatchMachineSet(ctx, ms, lastMS); err != nil {
		return fmt.Errorf("failed to patch MachineSet (%s) in namespace (%s): %w", ms.Name, msNamespace, err)
	}
	return nil
}

// getProgressMachineDeployment is a function to return the progress state of the MachineDeployment.
func getProgressMachineDeployment(
	md *capiv1.MachineDeployment, nsrUID types.UID,
) progressCondition {
	uid, ok := md.Labels[scalev1.InProgressUIDLabel]
	if !ok {
		return progressCondition{notInProgress, ""}
	}
	if uid == string(nsrUID) {
		return progressCondition{inProgressByThisNSR, uid}
	}
	return progressCondition{inProgressByOtherNSR, uid}
}

// updateStatusConditions is a function to update condition.
func updateStatusConditions(
	conditions *[]metav1.Condition, condType string, condStatus metav1.ConditionStatus, condReason, condMessage string,
) {
	condition := metav1.Condition{
		Type:    condType,
		Status:  condStatus,
		Reason:  condReason,
		Message: condMessage,
	}
	meta.SetStatusCondition(conditions, condition)
}
