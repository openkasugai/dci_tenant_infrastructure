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
	examplecomv1 "app-conn-info-maker-dci/api/v1"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DCIAppConnectionReconciler reconciles a DCIAppConnection object
type DCIAppConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

const (
	DCIAppConnectionReconcilerInterval = 1
)

const (
	DCIAppConnectionEventKindNone   = "None"
	DCIAppConnectionEventKindCreate = "Create"
	DCIAppConnectionEventKindUpdate = "Update"
	DCIAppConnectionEventKindDelete = "Delete"
)

type DCIAppConnectionEventPredicate struct {
	predicate.Funcs
}

// +kubebuilder:rbac:groups=example.com,resources=dciappconnections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=dciappconnections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=dciappconnections/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=replicasets/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mellanox.com,resources=hostdevicenetworks,verbs=get;list;watch
// +kubebuilder:rbac:groups=mellanox.com,resources=hostdevicenetworks/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DCIAppConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DCIAppConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	stCtrlResult := ctrl.Result{}
	var err error

	log := r.Log.WithValues("dciappconnect", req.Name)
	bProcFlag := false
	cFuncName := getFuncName()

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		pstCRData := &examplecomv1.DCIAppConnection{}
		err = r.Get(ctx, req.NamespacedName, pstCRData)
		if nil != err {
			if errors.IsNotFound(err) {
				log.Info(cFuncName + " could not find DCIAppConnectionCR." +
					" Name: " + req.Name +
					", Namespace: " + req.Namespace)
				err = nil
				bProcFlag = true
			} else {
				log.Error(err, cFuncName+" failed to get DCIAppConnectionCR."+
					" Name: "+req.Name+
					", Namespace: "+req.Namespace)
			}
			break
		}
		log.Info(cFuncName + " successfully fetched the DCIAppConnectionCR.")

		stFromPodInfos := examplecomv1.FromPodInfos{}
		stToPodInfos := examplecomv1.ToPodInfos{}
		stToPodInfos.ToPodInfos = make(map[int]examplecomv1.ToPods)

		// Get the event type
		cFinalizerName := getFinalizerName(r.Client, pstCRData)
		cEventKind := r.getEventKind(pstCRData, cFinalizerName)
		var bFuncRet bool

		if DCIAppConnectionEventKindCreate == cEventKind {
			bAdditionalProcFlag := true
			err, bFuncRet = r.getConnectionRelation(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				&stFromPodInfos,
				&stToPodInfos)
			if nil != err {
				break
			}
			if bFuncRet {
				err = r.reconcileCreate(
					ctx,
					log,
					pstCRData,
					cFinalizerName,
					&stFromPodInfos,
					&stToPodInfos)
				if nil != err {
					break
				}
			}

			if bFuncRet {
				bProcFlag = true
			} else {
				if len(pstCRData.Status.Phase) == 0 {
					pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
					pstCRData.Status.Phase = examplecomv1.StatusPhasePending
					log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Pending.")

					err = r.Patch(ctx, pstCRData, pstBaseCRDData)
					if nil != err {
						log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
							" Name: "+pstCRData.Name+
							", Namespace: "+pstCRData.Namespace)
						break
					}
					log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
						" Name: " + pstCRData.Name +
						", Namespace: " + pstCRData.Namespace)
				}
			}
		} else if DCIAppConnectionEventKindUpdate == cEventKind {
			bProcFlag = true

			// do nothing
			log.Info(cFuncName + " has no action for the update event of DCIAppConnectionCR.")

		} else if DCIAppConnectionEventKindDelete == cEventKind {
			bAdditionalProcFlag := false
			// The second return value cannot be false when deleting,
			// so there is no need to check the return value.
			err, _ = r.getConnectionRelation(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				&stFromPodInfos,
				&stToPodInfos)
			if nil != err {
				break
			}

			err = r.reconcileDelete(
				ctx,
				log,
				pstCRData,
				cFinalizerName,
				&stFromPodInfos,
				&stToPodInfos)
			if nil != err {
				break
			}
			bProcFlag = true
		} else {
			bProcFlag = true

			// do noting
			log.Info(cFuncName + " has no action of this type event for DCIAppConnectionCR.")
		}
	} // SA4008: This loop is intentionally only a one-time loop

	if nil != err {
		stCtrlResult.Requeue = false
	} else if bProcFlag {
		stCtrlResult.Requeue = false
	} else {
		stCtrlResult.Requeue = true
		stCtrlResult.RequeueAfter =
			time.Second * DCIAppConnectionReconcilerInterval
	}

	log.Info(cFuncName + " return." +
		" stCtrlResult.Requeue: " + strconv.FormatBool(stCtrlResult.Requeue) +
		", stCtrlResult.RequeueAfter: " + stCtrlResult.RequeueAfter.String() +
		", err: " + fmt.Sprint(err))
	return stCtrlResult, err
}

func (DCIAppConnectionEventPredicate) Update(e event.UpdateEvent) bool {
	bProcFlag := true

	for range 1 {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			break
		}

		newObj, okNew := e.ObjectNew.(*examplecomv1.DCIAppConnection)

		if !okNew {
			break
		}
		if !newObj.DeletionTimestamp.IsZero() {
			break
		}

		bProcFlag = false
	}
	return bProcFlag
}

func (DCIAppConnectionEventPredicate) Create(e event.CreateEvent) bool {
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *DCIAppConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1.DCIAppConnection{}).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
					1*time.Second,
					5*time.Minute,
				),
				&workqueue.TypedBucketRateLimiter[ctrl.Request]{
					Limiter: rate.NewLimiter(
						rate.Limit(10),
						100,
					),
				},
			),
		}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			DCIAppConnectionEventPredicate{},
		)).
		Named("dciappconnection").
		Complete(r)
}

func (r *DCIAppConnectionReconciler) getEventKind(
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string) string {

	cEventKind := DCIAppConnectionEventKindNone

	if !controllerutil.ContainsFinalizer(pstCRData, cFinalizerName) {

		if !pstCRData.DeletionTimestamp.IsZero() {
			// do nothing
		} else {
			cEventKind = DCIAppConnectionEventKindCreate
		}
	} else {
		if !pstCRData.DeletionTimestamp.IsZero() {
			cEventKind = DCIAppConnectionEventKindDelete
		} else {
			cEventKind = DCIAppConnectionEventKindUpdate
		}
	}
	return cEventKind
}

func (r *DCIAppConnectionReconciler) getConnectionRelation(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) (error, bool) {
	var err error
	bProcFlag := false

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag))
	var bFuncRet bool

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop
		err, bFuncRet = r.getConnectionRelationship(
			ctx,
			log,
			pstCRData,
			bAdditionalProcFlag,
			pstFromPodInfos,
			pstToPodInfos)
		if nil != err {
			break
		}
		if !bFuncRet {
			break
		}

		if pstCRData.Spec.InputEndpoint != nil {
			err, bFuncRet = r.getConnectionRelationshipInput(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				pstFromPodInfos)
			if nil != err {
				break
			}
			if !bFuncRet {
				break
			}
		}

		if pstCRData.Spec.OutputEndpoint != nil {
			err, bFuncRet = r.getConnectionRelationshipOutput(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				pstFromPodInfos)
			if nil != err {
				break
			}
			if !bFuncRet {
				break
			}
		}

		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *DCIAppConnectionReconciler) reconcileCreate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) error {
	var err error

	var sDeployments, sPods, sServices, sEndpointSlices, sConnectEndpoints []types.NamespacedName
	gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRData)
	cKind := strings.ToLower(gvks[0].Kind)
	cVal := pstCRData.Name + "_" + pstCRData.Namespace
	cFuncName := getFuncName()

	log.Info(cFuncName + " Start.")

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop
		for iConnIdx := 0; len(pstToPodInfos.ToPodInfos) > iConnIdx; iConnIdx++ {
			stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]

			err = r.registDciService(
				ctx,
				log,
				pstCRData,
				iConnIdx,
				pstToPodInfos,
				&sServices,
				&sEndpointSlices)
			if nil != err {
				break
			}

			iConnExistFlag := false
			if nil != stToPodInfo.DeploymentR {
				for iConnChkIndex := 0; iConnIdx > iConnChkIndex; iConnChkIndex++ {
					if pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Name ==
						stToPodInfo.DeploymentR.Name &&
						pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Namespace ==
							stToPodInfo.DeploymentR.Namespace {
						iConnExistFlag = true
						break
					}
				}
				if !iConnExistFlag {
					err = setDciDeploymentConnectionRelationship(
						ctx,
						r.Client,
						log,
						cKind,
						cVal,
						cFinalizerName,
						stToPodInfo.DeploymentR)
					if nil != err {
						break
					}

					sDeployments = append(sDeployments, types.NamespacedName{
						Namespace: stToPodInfo.DeploymentR.Namespace,
						Name:      stToPodInfo.DeploymentR.Name,
					})

					for _, stToPod := range stToPodInfo.ToPods {
						if !stToPod.DeleteFlag {
							err = setDciPodConnectionRelationship(
								ctx,
								r.Client,
								log,
								cKind,
								cVal,
								cFinalizerName,
								stToPod.PodR)
							if err != nil {
								break
							}

							sPods = append(sPods, types.NamespacedName{
								Namespace: stToPod.PodR.Namespace,
								Name:      stToPod.PodR.Name,
							})
						}
					}
					if err != nil {
						break
					}
				}
			}
		}

		if err != nil {
			break
		}

		for iFromPodIdx := 0; len(pstFromPodInfos.FromPods) > iFromPodIdx; iFromPodIdx++ {
			stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]
			cDeploymentName := stFromPod.DeploymentRef.Name
			cDeploymentNamespace := stFromPod.DeploymentRef.Namespace

			bExistFlag := false
			for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
				if stToPodInfo.DeploymentRef.Name == cDeploymentName &&
					stToPodInfo.DeploymentRef.Namespace == cDeploymentNamespace {
					bExistFlag = true
					break
				}
			}

			if !bExistFlag {
				if stFromPod.MstPodFlag {
					if nil != stFromPod.DeploymentR {
						err = setDciDeploymentConnectionRelationship(
							ctx,
							r.Client,
							log,
							cKind,
							cVal,
							cFinalizerName,
							stFromPod.DeploymentR)
						if nil != err {
							break
						}
						sDeployments = append(sDeployments, types.NamespacedName{
							Namespace: stFromPod.DeploymentR.Namespace,
							Name:      stFromPod.DeploymentR.Name,
						})
					}
				}

				if !stFromPod.DeleteFlag && stFromPod.PodR != nil {
					err = setDciPodConnectionRelationship(
						ctx,
						r.Client,
						log,
						cKind,
						cVal,
						cFinalizerName,
						stFromPod.PodR)
					if nil != err {
						break
					}
					sPods = append(sPods, types.NamespacedName{
						Namespace: stFromPod.PodR.Namespace,
						Name:      stFromPod.PodR.Name,
					})
				}
			}

			if stFromPod.PodR != nil {
				err = makeConnectEndpoint(
					ctx,
					r.Client,
					log,
					r.Scheme,
					pstCRData,
					&stFromPod,
					pstToPodInfos)
				if nil != err {
					break
				}
			}

			if stFromPod.ConnectEndpointCR != nil {
				sConnectEndpoints = append(sConnectEndpoints, types.NamespacedName{
					Namespace: stFromPod.ConnectEndpointCR.Namespace,
					Name:      stFromPod.ConnectEndpointCR.Name,
				})
			}
		}

		if nil != err {
			break
		}

		pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
		pstCRData.Status.Phase = examplecomv1.StatusPhaseCreated
		log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Created.")

		controllerutil.AddFinalizer(pstCRData, cFinalizerName)
		log.Info(cFuncName + " added a finalizer to DCIAppConnectionCR.")

		err = r.Patch(ctx, pstCRData, pstBaseCRDData)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
				" Name: "+pstCRData.Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}

		log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
			" Name: " + pstCRData.Name +
			", Namespace: " + pstCRData.Namespace)
	}

	if nil != err {
		r.createRollBack(
			ctx,
			log,
			cKind,
			cFinalizerName,
			sDeployments,
			sPods,
			sServices,
			sEndpointSlices,
			sConnectEndpoints)
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *DCIAppConnectionReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) error {
	var err error

	gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRData)
	cKind := strings.ToLower(gvks[0].Kind)
	cFuncName := getFuncName()

	log.Info(cFuncName + " Start.")

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		for iConnIdx := 0; len(pstToPodInfos.ToPodInfos) > iConnIdx; iConnIdx++ {
			stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]
			iConnExistFlag := false

			if stToPodInfo.DeploymentR != nil {
				for iConnChkIndex := 0; iConnIdx > iConnChkIndex; iConnChkIndex++ {
					if pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Name ==
						stToPodInfo.DeploymentR.Name &&
						pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Namespace ==
							stToPodInfo.DeploymentR.Namespace {
						iConnExistFlag = true
						break
					}
				}
				if !iConnExistFlag {
					for _, stToPod := range stToPodInfo.ToPods {
						_ = unsetDciPodConnectionRelationship(
							ctx,
							r.Client,
							log,
							cKind,
							cFinalizerName,
							stToPod.PodR)
					}

					_ = unsetDciDeploymentConnectionRelationship(
						ctx,
						r.Client,
						log,
						cKind,
						cFinalizerName,
						stToPodInfo.DeploymentR)
				}
			}
		}

		for iFromPodIdx := 0; len(pstFromPodInfos.FromPods) > iFromPodIdx; iFromPodIdx++ {
			stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]

			cDeploymentName := stFromPod.DeploymentRef.Name
			cDeploymentNamespace := stFromPod.DeploymentRef.Namespace

			bExistFlag := false
			for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
				if stToPodInfo.DeploymentRef.Name == cDeploymentName &&
					stToPodInfo.DeploymentRef.Namespace == cDeploymentNamespace {
					bExistFlag = true
					break
				}
			}
			if bExistFlag {
				continue
			}

			if stFromPod.PodR != nil {
				_ = unsetDciPodConnectionRelationship(
					ctx,
					r.Client,
					log,
					cKind,
					cFinalizerName,
					stFromPod.PodR)
			}

			if stFromPod.MstPodFlag {
				if stFromPod.DeploymentR != nil {
					_ = unsetDciDeploymentConnectionRelationship(
						ctx,
						r.Client,
						log,
						cKind,
						cFinalizerName,
						stFromPod.DeploymentR)
				}
			}
		}

		pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())

		controllerutil.RemoveFinalizer(pstCRData, cFinalizerName)
		log.Info(cFuncName + " removed the finalizer from DCIAppConnectionCR.")

		err = r.Patch(ctx, pstCRData, pstBaseCRDData)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
				" Name: "+pstCRData.Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}
		log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
			" Name: " + pstCRData.Name +
			", Namespace: " + pstCRData.Namespace)
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *DCIAppConnectionReconciler) getConnectionRelationship(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) (error, bool) {

	var err error
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag))

	var outDeploymentRef examplecomv1.NamespaceName
	var inDeploymentRef examplecomv1.NamespaceName
	if pstCRData.Spec.OutputEndpoint != nil {
		outDeploymentRef =
			pstCRData.Spec.OutputEndpoint.Internal.DeploymentRef
	}
	if pstCRData.Spec.InputEndpoint != nil {
		inDeploymentRef =
			pstCRData.Spec.InputEndpoint.Internal.DeploymentRef
	}

	// Connections Relation
	for iConnIdx := 0; len(pstCRData.Spec.Connections) > iConnIdx; iConnIdx++ {

		pstConnection := pstCRData.Spec.Connections[iConnIdx]
		bDupFlag := false

		for iFromPodIdx := 0; len(pstFromPodInfos.FromPods) > iFromPodIdx; iFromPodIdx++ {

			stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]
			cDeploymentName := stFromPod.DeploymentRef.Name
			cDeploymentNamespace := stFromPod.DeploymentRef.Namespace

			if cDeploymentName == pstConnection.From.DeploymentRef.Name &&
				cDeploymentNamespace == pstConnection.From.DeploymentRef.Namespace {

				stFromPod.DeploymentIndexs = append(
					stFromPod.DeploymentIndexs,
					iConnIdx)

				pstFromPodInfos.FromPods[iFromPodIdx] = stFromPod
				bDupFlag = true
				log.Info(cFuncName + " set the DeploymentIndexs Info in the FromPod." +
					" DeploymentIndexs len: " +
					strconv.Itoa(len(pstFromPodInfos.FromPods[iFromPodIdx].DeploymentIndexs)) +
					", Name: " + pstConnection.From.DeploymentRef.Name)
			}
		}

		if !bDupFlag {
			iIOMode := examplecomv1.IOModeNONE

			if pstConnection.From.DeploymentRef.Name ==
				inDeploymentRef.Name &&
				pstConnection.From.DeploymentRef.Namespace ==
					inDeploymentRef.Namespace {
				iIOMode |= examplecomv1.IOModeIN
			}

			if pstConnection.From.DeploymentRef.Name ==
				outDeploymentRef.Name &&
				pstConnection.From.DeploymentRef.Namespace ==
					outDeploymentRef.Namespace {
				iIOMode |= examplecomv1.IOModeOUT
			}

			log.Info(cFuncName + " set FromPod Info." +
				" Index: " + strconv.Itoa(iConnIdx) +
				", Name: " + pstConnection.From.DeploymentRef.Name)

			var bFuncRet bool
			err, bFuncRet = getFromConnectionInformation(
				ctx,
				r.Client,
				log,
				pstConnection.From.DeploymentRef.Name,
				pstConnection.From.DeploymentRef.Namespace,
				pstCRData.Name,
				pstCRData.Namespace,
				bAdditionalProcFlag,
				false,
				pstFromPodInfos,
				iConnIdx,
				iIOMode)
			if nil != err {
				break
			}
			if !bFuncRet {
				bProcFlag = false
				break
			}

			for i := range pstFromPodInfos.FromPods {
				fp := &pstFromPodInfos.FromPods[i]
				if fp.DeploymentRef.Name != pstConnection.From.DeploymentRef.Name ||
					fp.DeploymentRef.Namespace != pstConnection.From.DeploymentRef.Namespace {
					continue
				}
				fp.DeploymentIndexs = append(fp.DeploymentIndexs, iConnIdx)
			}
		}

		log.Info(cFuncName + " set ToPod Info. Name: " + pstConnection.To.DeploymentRef.Name)
		var bFuncRet bool
		err, bFuncRet = getToConnectionInformation(
			ctx,
			r.Client,
			log,
			pstConnection.To.DeploymentRef.Name,
			pstConnection.To.DeploymentRef.Namespace,
			pstCRData.Name,
			pstCRData.Namespace,
			bAdditionalProcFlag,
			false,
			pstToPodInfos,
			iConnIdx)
		if nil != err {
			bProcFlag = false
			break
		}
		if !bFuncRet {
			bProcFlag = false
			break
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *DCIAppConnectionReconciler) getConnectionRelationshipInput(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos) (error, bool) {

	var err error
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag))

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		inDeploymentRef :=
			pstCRData.Spec.InputEndpoint.Internal.DeploymentRef

		var outDeploymentRef examplecomv1.NamespaceName
		if pstCRData.Spec.OutputEndpoint != nil {
			outDeploymentRef =
				pstCRData.Spec.OutputEndpoint.Internal.DeploymentRef
		}

		bConnExistFlag := false
		for _, connections := range pstCRData.Spec.Connections {

			if inDeploymentRef.Name == connections.From.DeploymentRef.Name &&
				inDeploymentRef.Namespace == connections.From.DeploymentRef.Namespace {
				bConnExistFlag = true
				break
			}
		}

		if !bConnExistFlag {
			// InputEndpoint Relation
			iIOMode := examplecomv1.IOModeIN
			if inDeploymentRef.Name == outDeploymentRef.Name &&
				inDeploymentRef.Namespace == outDeploymentRef.Namespace {
				iIOMode = examplecomv1.IOModeINOUT
			}

			var bFuncRet bool
			err, bFuncRet = getFromConnectionInformation(
				ctx,
				r.Client,
				log,
				inDeploymentRef.Name,
				inDeploymentRef.Namespace,
				pstCRData.Name,
				pstCRData.Namespace,
				bAdditionalProcFlag,
				false,
				pstFromPodInfos,
				-1,
				iIOMode)
			if nil != err {
				bProcFlag = false
				break
			}
			if !bFuncRet {
				bProcFlag = false
				break
			}
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *DCIAppConnectionReconciler) getConnectionRelationshipOutput(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos) (error, bool) {

	var err error
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag))

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		var inDeploymentRef examplecomv1.NamespaceName
		if pstCRData.Spec.InputEndpoint != nil {
			inDeploymentRef =
				pstCRData.Spec.InputEndpoint.Internal.DeploymentRef
		}
		outDeploymentRef :=
			pstCRData.Spec.OutputEndpoint.Internal.DeploymentRef

		// OutputEndpoint Relation
		if inDeploymentRef.Name == outDeploymentRef.Name &&
			inDeploymentRef.Namespace == outDeploymentRef.Namespace {
			break
		}

		bFromExistFlag := false

		for iConnIdx := 0; len(pstCRData.Spec.Connections) > iConnIdx; iConnIdx++ {
			pstConnection := pstCRData.Spec.Connections[iConnIdx]
			if outDeploymentRef.Name ==
				pstConnection.From.DeploymentRef.Name &&
				outDeploymentRef.Namespace ==
					pstConnection.From.DeploymentRef.Namespace {

				bFromExistFlag = true
				break
			}
		}
		if !bFromExistFlag {
			var bFuncRet bool
			err, bFuncRet = getFromConnectionInformation(
				ctx,
				r.Client,
				log,
				outDeploymentRef.Name,
				outDeploymentRef.Namespace,
				pstCRData.Name,
				pstCRData.Namespace,
				bAdditionalProcFlag,
				false,
				pstFromPodInfos,
				-1,
				examplecomv1.IOModeOUT)
			if nil != err {
				bProcFlag = false
				break
			}
			if !bFuncRet {
				bProcFlag = false
				break
			}
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *DCIAppConnectionReconciler) registDciService(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	iConnIdx int,
	pstToPodInfos *examplecomv1.ToPodInfos,
	pServices, pEndpointSlices *[]types.NamespacedName) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" iConnIdx: " + strconv.Itoa(iConnIdx))

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]
		cServiceName := stToPodInfo.DeploymentRef.Name
		cServiceNamespace := pstCRData.Namespace

		if 0 < iConnIdx {
			var iPrevIdx int
			for iPrevIdx = 0; iConnIdx > iPrevIdx; iPrevIdx++ {
				stToPodInfo, exist := pstToPodInfos.ToPodInfos[iPrevIdx]
				if exist {
					if stToPodInfo.DeploymentRef.Name == cServiceName {
						break
					}
				}
			}
			if iConnIdx > iPrevIdx {
				break
			}
		}

		b2ndNicFlag := false
		for _, stToPod := range stToPodInfo.ToPods {
			if len(stToPod.Address) != 0 {
				b2ndNicFlag = true
			}
		}

		if b2ndNicFlag {
			// Get Service Resource
			bServiceAddFlag := false
			pstCRService := stToPodInfo.ServiceR
			var stBaseCRService client.Patch

			if pstCRService == nil {
				bServiceAddFlag = true
				pstCRService = &corev1.Service{}
				pstCRService.Name = cServiceName
				pstCRService.Namespace = cServiceNamespace
			} else {
				stBaseCRService = client.MergeFrom(pstCRService.DeepCopy())
			}

			// Create Service Resource
			var i32Port int32 = 80
			pstCRService.Spec = corev1.ServiceSpec{
				ClusterIP: "None",
				Ports: []corev1.ServicePort{
					{
						Port: i32Port,
					},
				},
			}

			err = ctrlSetControllerReferenceFunc(
				pstCRData,
				pstCRService,
				r.Scheme)
			if err != nil {
				break
			}

			if bServiceAddFlag {
				err = r.Create(ctx, pstCRService)
				if err != nil {
					log.Error(err, cFuncName+" failed to create Service resource."+
						" Name: "+pstCRService.Name+
						", Namespace: "+pstCRService.Namespace)
					break
				}
				log.Info(cFuncName + " created Service resource successfully." +
					" Name: " + pstCRService.Name +
					", Namespace: " + pstCRService.Namespace)
			} else {
				err = r.Patch(ctx, pstCRService, stBaseCRService)
				if err != nil {
					log.Error(err, cFuncName+" failed to patch Service resource. "+
						" Name: "+pstCRService.Name+
						", Namespace: "+pstCRService.Namespace)
					break
				}
				log.Info(cFuncName + " patched Service resource successfully." +
					" Name: " + pstCRService.Name +
					", Namespace: " + pstCRService.Namespace)
			}
			*pServices = append(*pServices, types.NamespacedName{
				Namespace: pstCRService.Namespace,
				Name:      pstCRService.Name,
			})

			// Get EndpointSlice Resource
			bEndpointSliceAddFlag := false
			pstCREndpointSlice := stToPodInfo.EndpointSliceR
			var stBaseCREndpointSlice client.Patch

			if pstCREndpointSlice == nil {
				bEndpointSliceAddFlag = true
				pstCREndpointSlice = &discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cServiceName,
						Namespace: cServiceNamespace,
						Labels: map[string]string{
							"kubernetes.io/service-name": cServiceName,
						},
					},
				}
			} else {
				stBaseCREndpointSlice = client.MergeFrom(pstCREndpointSlice.DeepCopy())
			}

			// Create EndpointSlice Resource
			stToPodInfo, exists := pstToPodInfos.ToPodInfos[iConnIdx]
			if exists {
				pstCREndpointSlice.AddressType = discoveryv1.AddressTypeIPv4

				for iPodIdx := 0; len(stToPodInfo.ToPods) > iPodIdx; iPodIdx++ {
					if stToPodInfo.ToPods[iPodIdx].DeleteFlag {
						continue
					}
					if len(stToPodInfo.ToPods[iPodIdx].Address) == 0 {
						continue
					}

					getAppendEndpointSlice(
						ctx,
						r.Client,
						log,
						stToPodInfo.ToPods[iPodIdx].PodR,
						pstCREndpointSlice,
						stToPodInfo.ToPods[iPodIdx].Address)
				}
			}

			err = ctrlSetControllerReferenceFunc(
				pstCRService,
				pstCREndpointSlice,
				r.Scheme)
			if nil != err {
				log.Error(err,
					cFuncName+" failed to set OwnerReference for EndpointSlice resource."+
						" Name: "+pstCREndpointSlice.Name+
						", Namespace: "+pstCREndpointSlice.Namespace)
				break
			}

			if bEndpointSliceAddFlag {
				err = r.Create(ctx, pstCREndpointSlice)
				if err != nil {
					log.Error(err, cFuncName+" failed to create EndpointSlice resource."+
						" Name: "+pstCREndpointSlice.Name+
						", Namespace: "+pstCREndpointSlice.Namespace)
					break
				}
				log.Info(cFuncName + " created EndpointSlice resource successfully." +
					" Name: " + pstCREndpointSlice.Name +
					", Namespace: " + pstCREndpointSlice.Namespace)
			} else {
				err = r.Patch(ctx, pstCREndpointSlice, stBaseCREndpointSlice)
				if err != nil {
					log.Error(err, cFuncName+" failed to patch EndpointSlice resource."+
						" Name: "+pstCREndpointSlice.Name+
						", Namespace: "+pstCREndpointSlice.Namespace)
					break
				}
				log.Info(cFuncName + " patched EndpointSlice resource successfully." +
					" Name: " + pstCREndpointSlice.Name +
					", Namespace: " + pstCREndpointSlice.Namespace)
			}

			*pEndpointSlices = append(*pEndpointSlices, types.NamespacedName{
				Namespace: pstCREndpointSlice.Namespace,
				Name:      pstCREndpointSlice.Name,
			})
		}
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *DCIAppConnectionReconciler) createRollBack(
	ctx context.Context,
	log logr.Logger,
	labelsKey, cFinalizerName string,
	sDeployments, sPods, sServices, sEndpointSlices, sConnectEndpoints []types.NamespacedName) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	stConnendpoint := examplecomv1.ConnectEndpoint{}
	stEndpointSlice := discoveryv1.EndpointSlice{}
	stService := corev1.Service{}

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop
		// RollBack ConnectEndpoint
		for _, cep := range sConnectEndpoints {
			err := r.Get(ctx, cep, &stConnendpoint)
			if err != nil {
				log.Error(err, cFuncName+" failed to get ConnectEndpointCR."+
					" Name: "+cep.Name+
					", Namespace: "+cep.Namespace)
			} else {
				err := r.Delete(ctx, &stConnendpoint)
				if err != nil {
					log.Error(err, cFuncName+" failed to delete ConnectEndpointCR."+
						" Name: "+cep.Name+
						", Namespace: "+cep.Namespace)
				} else {
					log.Info(cFuncName + " deleted ConnectEndpointCR successfully." +
						" Name: " + cep.Name +
						", Namespace: " + cep.Namespace)
				}
			}
		}

		// RollBack Pod
		for _, pod := range sPods {
			stCRPod := corev1.Pod{}
			err := r.Get(ctx, pod, &stCRPod)
			if err != nil {
				log.Error(err, cFuncName+" failed to get Pod resource."+
					" Name: "+pod.Name+
					", Namespace: "+pod.Namespace)
			} else {
				_ = unsetDciPodConnectionRelationship(
					ctx,
					r.Client,
					log,
					labelsKey,
					cFinalizerName,
					&stCRPod)
			}
		}

		// RollBack Deployment
		for _, deployment := range sDeployments {
			stCRDeployment := appsv1.Deployment{}
			err := r.Get(ctx, deployment, &stCRDeployment)
			if err != nil {
				log.Error(err, cFuncName+" failed to get Deployment resource."+
					" Name: "+deployment.Name+
					", Namespace: "+deployment.Namespace)
			} else {
				_ = unsetDciDeploymentConnectionRelationship(
					ctx,
					r.Client,
					log,
					labelsKey,
					cFinalizerName,
					&stCRDeployment)

			}
		}

		// RollBack EndpointSlice
		for _, eps := range sEndpointSlices {
			err := r.Get(ctx, eps, &stEndpointSlice)
			if err != nil {
				log.Error(err, cFuncName+" failed to get EndpointSlice resource."+
					" Name: "+eps.Name+
					", Namespace: "+eps.Namespace)
			} else {
				err := r.Delete(ctx, &stEndpointSlice)
				if err != nil {
					log.Error(err, cFuncName+" failed to delete EndpointSlice resource."+
						" Name: "+eps.Name+
						", Namespace: "+eps.Namespace)
				} else {
					log.Info(cFuncName + " deleted to EndpointSlice resource successfully." +
						" Name: " + eps.Name +
						", Namespace: " + eps.Namespace)
				}
			}
		}

		// RollBack Service
		for _, service := range sServices {
			err := r.Get(ctx, service, &stService)
			if err != nil {
				log.Error(err, cFuncName+" failed to get Service resource."+
					" Name: "+service.Name+
					", Namespace: "+service.Namespace)
			} else {
				err := r.Delete(ctx, &stService)
				if err != nil {
					log.Error(err, cFuncName+" failed to delete Service resource."+
						" Name: "+service.Name+
						", Namespace: "+service.Namespace)
				} else {
					log.Info(cFuncName + " deleted to Service resource successfully." +
						" Name: " + service.Name +
						", Namespace: " + service.Namespace)
				}
			}
		}
	}
	log.Info(cFuncName + " End.")
	return
}
