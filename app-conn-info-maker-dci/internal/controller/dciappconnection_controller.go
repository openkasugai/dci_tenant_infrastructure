package controller

import (
	examplecomv1 "app-conn-info-maker-dci/api/v1"
	"context"
	"fmt"
	"slices"
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
	DCIAppConnectionReconcilerInterval  = 1
	StringNone                          = "None"
	ServiceKey                          = "kubernetes.io/service-name"
	DCIAppConnectionRateLimiterMaxDelay = 5 * time.Minute
	DCIAppConnectionRateLimiterRPS      = 10
	DCIAppConnectionRateLimiterBurst    = 100
)

const (
	DCIAppConnectionEventKindNone   = StringNone
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *DCIAppConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	stCtrlResult := ctrl.Result{}
	var err error

	log := r.Log.WithValues("dciappconnect", req.Name)
	bProcFlag := false
	cFuncName := getFuncName()

	for range 1 {
		pstCRData := &examplecomv1.DCIAppConnection{}
		err = r.Get(ctx, req.NamespacedName, pstCRData)
		if err != nil {
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
		stFromDeploymentInfos := examplecomv1.FromDeploymentInfos{}
		stToDeploymentInfos := examplecomv1.ToDeploymentInfos{}
		stToDeploymentInfos.ToDeploymentInfos = make(map[int]examplecomv1.ToDeployments)

		// Get the event type
		cFinalizerName := getFinalizerName(r.Client, pstCRData)
		cEventKind := r.getEventKind(pstCRData, cFinalizerName)
		var bFuncRet bool

		if DCIAppConnectionEventKindUpdate == cEventKind {
			bFuncRet = chkConnectionRelation(
				ctx,
				log,
				pstCRData)
			if bFuncRet {
				if pstCRData.Status.Phase == examplecomv1.StatusPhasePending {
					pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
					pstCRData.Status.Phase = examplecomv1.StatusPhaseCreated
					log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Created.")

					err = r.Patch(ctx, pstCRData, pstBaseCRDData)
					if err != nil {
						log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
							" Name: "+pstCRData.Name+
							", Namespace: "+pstCRData.Namespace)
						break
					}
					log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
						" Name: " + pstCRData.Name +
						", Namespace: " + pstCRData.Namespace)
				}
				bProcFlag = true
				break
			}
		}

		if DCIAppConnectionEventKindCreate == cEventKind {
			bAdditionalProcFlag := true
			err, bFuncRet = r.getConnectionRelation(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				&stFromPodInfos,
				&stToPodInfos)
			if err != nil {
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
				if err != nil {
					break
				}
			}

			if bFuncRet {
				bProcFlag = true
			} else if pstCRData.Status.Phase == "" {
				pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
				pstCRData.Status.Phase = examplecomv1.StatusPhasePending
				log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Pending.")

				err = r.Patch(ctx, pstCRData, pstBaseCRDData)
				if err != nil {
					log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
						" Name: "+pstCRData.Name+
						", Namespace: "+pstCRData.Namespace)
					break
				}
				log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
					" Name: " + pstCRData.Name +
					", Namespace: " + pstCRData.Namespace)
			}
		} else if DCIAppConnectionEventKindUpdate == cEventKind {
			bAdditionalProcFlag := true
			err, bFuncRet = r.getConnectionRelation(
				ctx,
				log,
				pstCRData,
				bAdditionalProcFlag,
				&stFromPodInfos,
				&stToPodInfos)
			if err != nil {
				break
			}

			if bFuncRet {
				getStatusRelation(
					ctx,
					log,
					pstCRData,
					&stFromDeploymentInfos,
					&stToDeploymentInfos)
				err = r.reconcileUpdate(
					ctx,
					log,
					pstCRData,
					cFinalizerName,
					&stFromPodInfos,
					&stToPodInfos,
					&stFromDeploymentInfos,
					&stToDeploymentInfos)
				if err != nil {
					break
				}
			}

			if bFuncRet {
				bProcFlag = true
			} else if pstCRData.Status.Phase != examplecomv1.StatusPhasePending {
				pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
				pstCRData.Status.Phase = examplecomv1.StatusPhasePending
				log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Pending.")

				err = r.Patch(ctx, pstCRData, pstBaseCRDData)
				if err != nil {
					log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
						" Name: "+pstCRData.Name+
						", Namespace: "+pstCRData.Namespace)
					break
				}
				log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
					" Name: " + pstCRData.Name +
					", Namespace: " + pstCRData.Namespace)
			}
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
			if err != nil {
				break
			}

			err = r.reconcileDelete(
				ctx,
				log,
				pstCRData,
				cFinalizerName,
				&stFromPodInfos,
				&stToPodInfos)
			if err != nil {
				break
			}
			bProcFlag = true
		} else {
			bProcFlag = true

			// do noting
			log.Info(cFuncName + " has no action of this type event for DCIAppConnectionCR.")
		}
	} // SA4008: This loop is intentionally only a one-time loop

	if err != nil {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = false
	} else if bProcFlag {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = false
	} else {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = true
		stCtrlResult.RequeueAfter =
			time.Second * DCIAppConnectionReconcilerInterval
	}

	//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
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

func (DCIAppConnectionEventPredicate) Create(_ event.CreateEvent) bool {
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
					DCIAppConnectionRateLimiterMaxDelay,
				),
				&workqueue.TypedBucketRateLimiter[ctrl.Request]{
					Limiter: rate.NewLimiter(
						rate.Limit(DCIAppConnectionRateLimiterRPS),
						DCIAppConnectionRateLimiterBurst,
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

func (*DCIAppConnectionReconciler) getEventKind(
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

	for range 1 {
		err, bFuncRet = r.getConnectionRelationship(
			ctx,
			log,
			pstCRData,
			bAdditionalProcFlag,
			pstFromPodInfos,
			pstToPodInfos)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		err, bFuncRet = r.getConnectionRelationshipInput(
			ctx,
			log,
			pstCRData,
			bAdditionalProcFlag,
			pstFromPodInfos)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		err, bFuncRet = r.getConnectionRelationshipOutput(
			ctx,
			log,
			pstCRData,
			bAdditionalProcFlag,
			pstFromPodInfos)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *DCIAppConnectionReconciler) addToConnection(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	cLabelVal string,
	pstToPodInfos *examplecomv1.ToPodInfos,
	iConnIdx int,
	bFromExistFlag bool) (sDeployments, sPods, sServices, sEndpointSlices []types.NamespacedName, err error) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bFromExistFlag: " + strconv.FormatBool(bFromExistFlag))

	for range 1 {
		stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]
		bExistFlag := false
		if stToPodInfo.DeploymentR != nil {
			for iConnChkIndex := range iConnIdx {
				if pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Name ==
					stToPodInfo.DeploymentR.Name &&
					pstToPodInfos.ToPodInfos[iConnChkIndex].DeploymentRef.Namespace ==
						stToPodInfo.DeploymentR.Namespace {
					bExistFlag = true
					break
				}
			}
			if !bExistFlag {
				err = r.registDciService(
					ctx,
					log,
					pstCRData,
					iConnIdx,
					pstToPodInfos,
					&sServices,
					&sEndpointSlices)
				if err != nil {
					break
				}

				if !bFromExistFlag {
					err = setDciDeploymentConnectionRelationship(
						ctx,
						r.Client,
						log,
						cLabelKey,
						cLabelVal,
						cFinalizerName,
						stToPodInfo.DeploymentR)
					if err != nil {
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
								cLabelKey,
								cLabelVal,
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
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return sDeployments, sPods, sServices, sEndpointSlices, err
}

func (r *DCIAppConnectionReconciler) addFromConnection(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	cLabelVal string,
	pstFromPod *examplecomv1.FromPodInfo,
	pstToPodInfos *examplecomv1.ToPodInfos,
	bToExistFlag bool) (sDeployments, sPods, sConnectEndpoints []types.NamespacedName, err error) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bToExistFlag: " + strconv.FormatBool(bToExistFlag))

	cDeploymentName := pstFromPod.DeploymentRef.Name
	cDeploymentNamespace := pstFromPod.DeploymentRef.Namespace

	for range 1 {
		if !bToExistFlag {
			bExistFlag := false
			for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
				if stToPodInfo.DeploymentRef.Name == cDeploymentName &&
					stToPodInfo.DeploymentRef.Namespace == cDeploymentNamespace {
					bExistFlag = true
					break
				}
			}

			if !bExistFlag {
				if pstFromPod.MstPodFlag {
					if pstFromPod.DeploymentR != nil {
						err = setDciDeploymentConnectionRelationship(
							ctx,
							r.Client,
							log,
							cLabelKey,
							cLabelVal,
							cFinalizerName,
							pstFromPod.DeploymentR)
						if err != nil {
							break
						}
						sDeployments = append(sDeployments, types.NamespacedName{
							Namespace: pstFromPod.DeploymentR.Namespace,
							Name:      pstFromPod.DeploymentR.Name,
						})
					}
				}

				if !pstFromPod.DeleteFlag && pstFromPod.PodR != nil {
					err = setDciPodConnectionRelationship(
						ctx,
						r.Client,
						log,
						cLabelKey,
						cLabelVal,
						cFinalizerName,
						pstFromPod.PodR)
					if err != nil {
						break
					}
					sPods = append(sPods, types.NamespacedName{
						Namespace: pstFromPod.PodR.Namespace,
						Name:      pstFromPod.PodR.Name,
					})
				}
			}
		}

		if pstFromPod.PodR != nil {
			err = makeConnectEndpoint(
				ctx,
				r.Client,
				log,
				r.Scheme,
				pstCRData,
				pstFromPod,
				pstToPodInfos)
			if err != nil {
				break
			}
			sConnectEndpoints =
				append(sConnectEndpoints, types.NamespacedName{
					Namespace: pstCRData.Namespace,
					Name:      pstFromPod.PodR.Name,
				})
		}
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return sDeployments, sPods, sConnectEndpoints, err
}

func (r *DCIAppConnectionReconciler) UpdateTriggeredCreate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	cLabelVal string,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) error {
	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	var sDeployments, sPods, sServices, sEndpointSlices, sConnectEndpoints []types.NamespacedName
	var sUpdateConnectEndpointCRs []*examplecomv1.ConnectEndpoint

	for range 1 {
		for iConnIdx := range len(pstToPodInfos.ToPodInfos) {
			stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]

			bExistToFlag := false
			for iStatusIdx := range len(pstToDeploymentInfos.ToDeploymentInfos) {
				if pstToDeploymentInfos.ToDeploymentInfos[iStatusIdx].DeploymentRef ==
					stToPodInfo.DeploymentRef {
					bExistToFlag = true
					stToDeploymentInfo := pstToDeploymentInfos.ToDeploymentInfos[iStatusIdx]
					stToDeploymentInfo.DeleteFlag = false
					pstToDeploymentInfos.ToDeploymentInfos[iStatusIdx] = stToDeploymentInfo
				}
			}

			bExistFromFlag := false
			for iFromStatusIdx := range len(pstFromDeploymentInfos.FromDeployments) {
				if pstFromDeploymentInfos.FromDeployments[iFromStatusIdx].DeploymentRef ==
					stToPodInfo.DeploymentRef {
					bExistFromFlag = true
				}
			}
			if !bExistToFlag {
				var sToDeployments, sToPods, sToServices, sToEndpointSlices []types.NamespacedName
				sToDeployments, sToPods, sToServices, sToEndpointSlices, err = r.addToConnection(
					ctx,
					log,
					pstCRData,
					cFinalizerName,
					cLabelKey,
					cLabelVal,
					pstToPodInfos,
					iConnIdx,
					bExistFromFlag)
				if err != nil {
					sDeployments = sToDeployments
					sPods = sToPods
					sServices = sToServices
					sEndpointSlices = sToEndpointSlices
					break
				}
				sDeployments = append(sDeployments, sToDeployments...)
				sPods = append(sPods, sToPods...)
				sServices = append(sServices, sToServices...)
				sEndpointSlices = append(sEndpointSlices, sToEndpointSlices...)
			}
		}

		if err != nil {
			break
		}

		for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
			stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]

			bExistToFlag := false
			for iStatusIdx := range len(pstToDeploymentInfos.ToDeploymentInfos) {
				if pstToDeploymentInfos.ToDeploymentInfos[iStatusIdx].DeploymentRef ==
					stFromPod.DeploymentRef {
					bExistToFlag = true
				}
			}
			bFromNotChangeFlag := true
			bExistFromFlag := false
			for iFromStatusIdx := range len(pstFromDeploymentInfos.FromDeployments) {
				if pstFromDeploymentInfos.FromDeployments[iFromStatusIdx].DeploymentRef !=
					stFromPod.DeploymentRef {
					continue
				}

				bExistFromFlag = true
				pstFromDeploymentInfos.FromDeployments[iFromStatusIdx].DeleteFlag = false

				log.Info(cFuncName + " Status.Connections.From exist." +
					" Name: " + stFromPod.DeploymentRef.Name +
					", Namespace: " + stFromPod.DeploymentRef.Namespace)

				stFromDeployment := pstFromDeploymentInfos.FromDeployments[iFromStatusIdx]
				if len(stFromDeployment.DeploymentIndexs) ==
					len(stFromPod.DeploymentIndexs) {
					var iFromChkIdx int
					for iFromChkIdx = 0; len(stFromPod.DeploymentIndexs) > iFromChkIdx; iFromChkIdx++ {
						iSpecToIdx := stFromPod.DeploymentIndexs[iFromChkIdx]
						iStatusToIdx := stFromDeployment.DeploymentIndexs[iFromChkIdx]
						if pstToPodInfos.ToPodInfos[iSpecToIdx].DeploymentRef !=
							pstToDeploymentInfos.ToDeploymentInfos[iStatusToIdx].DeploymentRef {
							break
						}
					}
					if len(stFromPod.DeploymentIndexs) != iFromChkIdx {
						bFromNotChangeFlag = false
						log.Info(cFuncName + " Status.Connections.To change." +
							" Name: " + stFromPod.DeploymentRef.Name +
							", Namespace: " + stFromPod.DeploymentRef.Namespace)
					}
					if !slices.Equal(stFromPod.InputSubsets, stFromDeployment.InputSubsets) {
						bFromNotChangeFlag = false
						log.Info(cFuncName + " Status.InputEndpoint.ExternalSubsets change." +
							" Name: " + stFromPod.DeploymentRef.Name +
							", Namespace: " + stFromPod.DeploymentRef.Namespace)
					}
					if !slices.Equal(stFromPod.OutputSubsets, stFromDeployment.OutputSubsets) {
						bFromNotChangeFlag = false
						log.Info(cFuncName + " Status.OutputEndpoint.ExternalSubsets change." +
							" Name: " + stFromPod.DeploymentRef.Name +
							", Namespace: " + stFromPod.DeploymentRef.Namespace)
					}
				} else {
					bFromNotChangeFlag = false
					log.Info(cFuncName + " Status.Connections.To len change." +
						" Name: " + stFromPod.DeploymentRef.Name +
						", Namespace: " + stFromPod.DeploymentRef.Namespace)
				}
				if stFromDeployment.IOMode != stFromPod.IOMode {
					bFromNotChangeFlag = false
					log.Info(cFuncName + " Status.Connections.From IOMode change." +
						" Name: " + stFromPod.DeploymentRef.Name +
						", Namespace: " + stFromPod.DeploymentRef.Namespace +
						", IOMode: " + strconv.Itoa(stFromDeployment.IOMode) +
						" -> " + strconv.Itoa(stFromPod.IOMode))
				}
			}
			if !bExistFromFlag {
				var sFromDeployments, sFromPods, sFromConnectEndpoints []types.NamespacedName
				sFromDeployments, sFromPods, sFromConnectEndpoints, err = r.addFromConnection(
					ctx,
					log,
					pstCRData,
					cFinalizerName,
					cLabelKey,
					cLabelVal,
					&stFromPod,
					pstToPodInfos,
					bExistToFlag)
				sDeployments = append(sDeployments, sFromDeployments...)
				sPods = append(sPods, sFromPods...)
				if len(sFromConnectEndpoints) > 0 {
					sConnectEndpoints = append(
						sConnectEndpoints, sFromConnectEndpoints...)
				}
				if err != nil {
					break
				}
			} else if !bFromNotChangeFlag {
				bExistToFlag = true

				var sFromConnectEndpoints []types.NamespacedName
				_, _, sFromConnectEndpoints, err = r.addFromConnection(
					ctx,
					log,
					pstCRData,
					cFinalizerName,
					cLabelKey,
					cLabelVal,
					&stFromPod,
					pstToPodInfos,
					bExistToFlag)
				if len(sFromConnectEndpoints) > 0 {
					if stFromPod.ConnectEndpointCR == nil {
						sConnectEndpoints = append(
							sConnectEndpoints, sFromConnectEndpoints...)
					} else {
						sUpdateConnectEndpointCRs = append(
							sUpdateConnectEndpointCRs,
							stFromPod.ConnectEndpointCR)
					}
				}
				if err != nil {
					break
				}
			}
		}

		if err != nil {
			break
		}

		pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
		pstCRData.Status.Connections = pstCRData.Spec.Connections
		pstCRData.Status.InputEndpoints = pstCRData.Spec.InputEndpoints
		pstCRData.Status.OutputEndpoints = pstCRData.Spec.OutputEndpoints

		pstCRData.Status.Phase = examplecomv1.StatusPhaseCreated
		log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Created.")

		err = r.Patch(ctx, pstCRData, pstBaseCRDData)
		if err != nil {
			log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
				" Name: "+pstCRData.Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}

		log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
			" Name: " + pstCRData.Name +
			", Namespace: " + pstCRData.Namespace)
	}

	if err != nil {
		r.updateRollBack(
			ctx,
			log,
			sUpdateConnectEndpointCRs)
		r.createRollBack(
			ctx,
			log,
			cLabelKey,
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

func (r *DCIAppConnectionReconciler) delToConnection(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos,
	iConnIdx int,
	bExistFromFlag bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bExistFromFlag: " + strconv.FormatBool(bExistFromFlag))

	stToDeploymentInfo := pstToDeploymentInfos.ToDeploymentInfos[iConnIdx]
	bExistFlag := false
	for iConnChkIndex := range iConnIdx {
		if pstToDeploymentInfos.ToDeploymentInfos[iConnChkIndex].DeploymentRef ==
			stToDeploymentInfo.DeploymentRef {
			bExistFlag = true
			break
		}
	}
	if !bExistFlag {
		r.deleteDciService(
			ctx,
			log,
			pstCRData,
			iConnIdx,
			pstToDeploymentInfos)
		if !bExistFromFlag {
			bAdditionalProcFlag := false
			bPodControllerFlag := false
			bFuncRet, pstCRDeployment, pstCRPods, err := getDeploymentList(
				ctx,
				r.Client,
				log,
				stToDeploymentInfo.DeploymentRef.Name,
				stToDeploymentInfo.DeploymentRef.Namespace,
				bAdditionalProcFlag,
				bPodControllerFlag)
			if err == nil && bFuncRet {
				log.Info(cFuncName + " set ToPod Info. Pod Count: " +
					strconv.Itoa(len(pstCRPods.Items)))
				for iPodIdx := range len(pstCRPods.Items) {
					stCRPod := pstCRPods.Items[iPodIdx]
					_ = unsetDciPodConnectionRelationship(
						ctx,
						r.Client,
						log,
						cLabelKey,
						cFinalizerName,
						&stCRPod)
				}
			}

			if pstCRDeployment != nil {
				_ = unsetDciDeploymentConnectionRelationship(
					ctx,
					r.Client,
					log,
					cLabelKey,
					cFinalizerName,
					pstCRDeployment)
			}
		}
	}

	log.Info(cFuncName + " End.")
}

func (r *DCIAppConnectionReconciler) delFromConnection(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	pstFromDeployment *examplecomv1.FromDeploymentInfo,
	_ *examplecomv1.ToDeploymentInfos,
	bExistToFlag bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" bExistToFlag: " + strconv.FormatBool(bExistToFlag))

	for range 1 {
		bAdditionalProcFlag := false
		bPodControllerFlag := false
		bFuncRet, pstCRDeployment, pstCRPods, err := getDeploymentList(
			ctx,
			r.Client,
			log,
			pstFromDeployment.DeploymentRef.Name,
			pstFromDeployment.DeploymentRef.Namespace,
			bAdditionalProcFlag,
			bPodControllerFlag)
		if err == nil && bFuncRet {
			log.Info(cFuncName + " set FromPod Info. Pod Count: " +
				strconv.Itoa(len(pstCRPods.Items)))

			for iPodIdx := range len(pstCRPods.Items) {
				stCRPod := pstCRPods.Items[iPodIdx]

				pstConnectEndpoint := &examplecomv1.ConnectEndpoint{}
				err = r.Get(ctx,
					client.ObjectKey{
						Name:      stCRPod.Name,
						Namespace: pstCRData.Namespace,
					}, pstConnectEndpoint)
				if err != nil {
					if !errors.IsNotFound(err) {
						log.Error(err, cFuncName+" failed to get ConnectEndpointCR.")
					}
				} else {
					if !pstConnectEndpoint.DeletionTimestamp.IsZero() {
						log.Info(cFuncName +
							" detected that ConnectEndpointCR has a DeletionTimestamp.")
					} else {
						_ = deleteConnectEndpoint(
							ctx,
							r.Client,
							log,
							pstConnectEndpoint)
					}
				}

				if !bExistToFlag {
					_ = unsetDciPodConnectionRelationship(
						ctx,
						r.Client,
						log,
						cLabelKey,
						cFinalizerName,
						&stCRPod)
				}
			}

			if !bExistToFlag {
				if pstCRDeployment != nil {
					_ = unsetDciDeploymentConnectionRelationship(
						ctx,
						r.Client,
						log,
						cLabelKey,
						cFinalizerName,
						pstCRDeployment)
				}
			}
		}
	}

	log.Info(cFuncName + " End.")
}

func (r *DCIAppConnectionReconciler) UpdateTriggeredDelete(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	cLabelKey string,
	_ string,
	_ *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for iConnIdx, stToDeploymentInfo := range pstToDeploymentInfos.ToDeploymentInfos {
		if !stToDeploymentInfo.DeleteFlag {
			continue
		}

		bExistFromFlag := false
		for iFromStatusIdx := range len(pstFromDeploymentInfos.FromDeployments) {
			if pstFromDeploymentInfos.FromDeployments[iFromStatusIdx].DeploymentRef ==
				stToDeploymentInfo.DeploymentRef {
				if !pstFromDeploymentInfos.FromDeployments[iFromStatusIdx].DeleteFlag {
					bExistFromFlag = true
				}
			}
		}

		r.delToConnection(
			ctx,
			log,
			pstCRData,
			cFinalizerName,
			cLabelKey,
			pstToDeploymentInfos,
			iConnIdx,
			bExistFromFlag)
	}

	for iFromDeploymentIdx := range len(pstFromDeploymentInfos.FromDeployments) {
		stFromDeployment :=
			pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx]
		if !stFromDeployment.DeleteFlag {
			continue
		}

		bExistToFlag := false
		for iStatusIdx := range len(pstToPodInfos.ToPodInfos) {
			if pstToPodInfos.ToPodInfos[iStatusIdx].DeploymentRef.Name ==
				stFromDeployment.DeploymentRef.Name &&
				pstToPodInfos.ToPodInfos[iStatusIdx].DeploymentRef.Namespace ==
					stFromDeployment.DeploymentRef.Namespace {
				bExistToFlag = true
			}
		}

		r.delFromConnection(
			ctx,
			log,
			pstCRData,
			cFinalizerName,
			cLabelKey,
			&stFromDeployment,
			pstToDeploymentInfos,
			bExistToFlag)
	}

	log.Info(cFuncName + " End.")
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

	for range 1 {
		for iConnIdx := range len(pstToPodInfos.ToPodInfos) {
			var sToDeployments, sToPods, sToServices, sToEndpointSlices []types.NamespacedName
			sToDeployments, sToPods, sToServices, sToEndpointSlices, err = r.addToConnection(
				ctx,
				log,
				pstCRData,
				cFinalizerName,
				cKind,
				cVal,
				pstToPodInfos,
				iConnIdx,
				false)
			if err != nil {
				sDeployments = sToDeployments
				sPods = sToPods
				sServices = sToServices
				sEndpointSlices = sToEndpointSlices
				break
			}
			sDeployments = append(sDeployments, sToDeployments...)
			sPods = append(sPods, sToPods...)
			sServices = append(sServices, sToServices...)
			sEndpointSlices = append(sEndpointSlices, sToEndpointSlices...)
		}

		if err != nil {
			break
		}

		for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
			stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]
			var sFromDeployments, sFromPods, sFromConnectEndpoints []types.NamespacedName
			sFromDeployments, sFromPods, sFromConnectEndpoints, err = r.addFromConnection(
				ctx,
				log,
				pstCRData,
				cFinalizerName,
				cKind,
				cVal,
				&stFromPod,
				pstToPodInfos,
				false)
			sDeployments = append(sDeployments, sFromDeployments...)
			sPods = append(sPods, sFromPods...)
			sConnectEndpoints = append(sConnectEndpoints, sFromConnectEndpoints...)
			if err != nil {
				break
			}
		}

		if err != nil {
			break
		}

		pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
		pstCRData.Status.Connections = pstCRData.Spec.Connections
		pstCRData.Status.InputEndpoints = pstCRData.Spec.InputEndpoints
		pstCRData.Status.OutputEndpoints = pstCRData.Spec.OutputEndpoints

		pstCRData.Status.Phase = examplecomv1.StatusPhaseCreated
		log.Info(cFuncName + " sets DCIAppConnectionCR Phase to Created.")

		controllerutil.AddFinalizer(pstCRData, cFinalizerName)
		log.Info(cFuncName + " added a finalizer to DCIAppConnectionCR.")

		err = r.Patch(ctx, pstCRData, pstBaseCRDData)
		if err != nil {
			log.Error(err, cFuncName+" failed to patch DCIAppConnectionCR."+
				" Name: "+pstCRData.Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}

		log.Info(cFuncName + " patched DCIAppConnectionCR successfully." +
			" Name: " + pstCRData.Name +
			", Namespace: " + pstCRData.Namespace)
	}

	if err != nil {
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

func (r *DCIAppConnectionReconciler) reconcileUpdate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) error {
	var err error

	gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRData)
	cKind := strings.ToLower(gvks[0].Kind)
	cVal := pstCRData.Name + "_" + pstCRData.Namespace

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	err = r.UpdateTriggeredCreate(
		ctx,
		log,
		pstCRData,
		cFinalizerName,
		cKind,
		cVal,
		pstFromPodInfos,
		pstToPodInfos,
		pstFromDeploymentInfos,
		pstToDeploymentInfos)

	if err == nil {
		r.UpdateTriggeredDelete(
			ctx,
			log,
			pstCRData,
			cFinalizerName,
			cKind,
			cVal,
			pstFromPodInfos,
			pstToPodInfos,
			pstFromDeploymentInfos,
			pstToDeploymentInfos)
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

	for range 1 {
		for iConnIdx := range len(pstToPodInfos.ToPodInfos) {
			stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]
			iConnExistFlag := false

			if stToPodInfo.DeploymentR != nil {
				for iConnChkIndex := range iConnIdx {
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

		for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
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
		if err != nil {
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

	arConnections := pstCRData.Spec.Connections
	if !bAdditionalProcFlag {
		arConnections = pstCRData.Status.Connections
	}

	// Connections Relation
	for iConnIdx := range arConnections {
		pstConnection := arConnections[iConnIdx]
		bDupFlag := false

		for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
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
			var arInputSubsets []string
			var arOutputSubsets []string
			var bFuncRet bool
			err, bFuncRet = getFromConnectionInformation(
				ctx,
				r.Client,
				log,
				pstConnection.From.DeploymentRef.Name,
				pstConnection.From.DeploymentRef.Namespace,
				pstCRData.Name,
				pstCRData.Namespace,
				arInputSubsets,
				arOutputSubsets,
				bAdditionalProcFlag,
				false,
				pstFromPodInfos,
				iConnIdx,
				examplecomv1.IOModeNONE)
			if err != nil {
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
		if err != nil {
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

	for range 1 {
		arInputEndpoints := pstCRData.Spec.InputEndpoints
		if !bAdditionalProcFlag {
			arInputEndpoints = pstCRData.Status.InputEndpoints
		}

		for iInIdx := range arInputEndpoints {
			inDeploymentRef :=
				arInputEndpoints[iInIdx].Internal.DeploymentRef

			bConnExistFlag := false
			for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
				stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]
				if inDeploymentRef.Name == stFromPod.DeploymentRef.Name &&
					inDeploymentRef.Namespace == stFromPod.DeploymentRef.Namespace {
					bConnExistFlag = true
					arInputSubsets := stFromPod.InputSubsets
					for iSubIdx := range len(arInputEndpoints[iInIdx].ExternalSubsets) {
						var iFromSubIdx int
						for iFromSubIdx = 0; len(arInputSubsets) > iFromSubIdx; iFromSubIdx++ {
							if arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
								arInputSubsets[iFromSubIdx] {
								break
							}
						}
						if len(arInputSubsets) == iFromSubIdx {
							arInputSubsets = append(arInputSubsets,
								arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
						}
					}
					pstFromPodInfos.FromPods[iFromPodIdx].InputSubsets = arInputSubsets
					pstFromPodInfos.FromPods[iFromPodIdx].IOMode |= examplecomv1.IOModeIN
					log.Info(cFuncName + " set the InputSubsets Info in the FromPod." +
						" InputSubsets len: " +
						strconv.Itoa(len(pstFromPodInfos.FromPods[iFromPodIdx].InputSubsets)) +
						", Name: " + inDeploymentRef.Name)
				}
			}

			if !bConnExistFlag {
				var arInputSubsets []string
				for iSubIdx := range len(arInputEndpoints[iInIdx].ExternalSubsets) {
					var iFromSubIdx int
					for iFromSubIdx = 0; len(arInputSubsets) > iFromSubIdx; iFromSubIdx++ {
						if arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
							arInputSubsets[iFromSubIdx] {
							break
						}
					}
					if len(arInputSubsets) == iFromSubIdx {
						arInputSubsets = append(arInputSubsets,
							arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
					}
				}
				var arOutputSubsets []string

				var bFuncRet bool
				err, bFuncRet = getFromConnectionInformation(
					ctx,
					r.Client,
					log,
					inDeploymentRef.Name,
					inDeploymentRef.Namespace,
					pstCRData.Name,
					pstCRData.Namespace,
					arInputSubsets,
					arOutputSubsets,
					bAdditionalProcFlag,
					false,
					pstFromPodInfos,
					-1,
					examplecomv1.IOModeIN)
				if err != nil {
					bProcFlag = false
					break
				}
				if !bFuncRet {
					bProcFlag = false
					break
				}
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

	for range 1 {
		arOutputEndpoints := pstCRData.Spec.OutputEndpoints
		if !bAdditionalProcFlag {
			arOutputEndpoints = pstCRData.Status.OutputEndpoints
		}

		for iOutIdx := range arOutputEndpoints {
			outDeploymentRef :=
				arOutputEndpoints[iOutIdx].Internal.DeploymentRef

			bFromExistFlag := false
			for iFromPodIdx := range len(pstFromPodInfos.FromPods) {
				stFromPod := pstFromPodInfos.FromPods[iFromPodIdx]
				if outDeploymentRef.Name == stFromPod.DeploymentRef.Name &&
					outDeploymentRef.Namespace == stFromPod.DeploymentRef.Namespace {
					bFromExistFlag = true
					arOutputSubsets := stFromPod.OutputSubsets
					for iSubIdx := range len(arOutputEndpoints[iOutIdx].ExternalSubsets) {
						var iFromSubIdx int
						for iFromSubIdx = 0; len(arOutputSubsets) > iFromSubIdx; iFromSubIdx++ {
							if arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
								arOutputSubsets[iFromSubIdx] {
								break
							}
						}
						if len(arOutputSubsets) == iFromSubIdx {
							arOutputSubsets = append(arOutputSubsets,
								arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
						}
					}
					pstFromPodInfos.FromPods[iFromPodIdx].OutputSubsets = arOutputSubsets
					pstFromPodInfos.FromPods[iFromPodIdx].IOMode |= examplecomv1.IOModeOUT
					log.Info(cFuncName + " set the OutputSubsets Info in the FromPod." +
						" OutputSubsets len: " +
						strconv.Itoa(len(pstFromPodInfos.FromPods[iFromPodIdx].OutputSubsets)) +
						", Name: " + outDeploymentRef.Name)
				}
			}

			if !bFromExistFlag {
				var arOutputSubsets []string
				for iSubIdx := range len(arOutputEndpoints[iOutIdx].ExternalSubsets) {
					var iFromSubIdx int
					for iFromSubIdx = 0; len(arOutputSubsets) > iFromSubIdx; iFromSubIdx++ {
						if arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
							arOutputSubsets[iFromSubIdx] {
							break
						}
					}
					if len(arOutputSubsets) == iFromSubIdx {
						arOutputSubsets = append(arOutputSubsets,
							arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
					}
				}
				var arInputSubsets []string

				var bFuncRet bool
				err, bFuncRet = getFromConnectionInformation(
					ctx,
					r.Client,
					log,
					outDeploymentRef.Name,
					outDeploymentRef.Namespace,
					pstCRData.Name,
					pstCRData.Namespace,
					arInputSubsets,
					arOutputSubsets,
					bAdditionalProcFlag,
					false,
					pstFromPodInfos,
					-1,
					examplecomv1.IOModeOUT)
				if err != nil {
					bProcFlag = false
					break
				}
				if !bFuncRet {
					bProcFlag = false
					break
				}
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

	for range 1 {
		stToPodInfo := pstToPodInfos.ToPodInfos[iConnIdx]
		cServiceName := stToPodInfo.DeploymentRef.Name
		cServiceNamespace := pstCRData.Namespace

		b2ndNicFlag := false
		for _, stToPod := range stToPodInfo.ToPods {
			if stToPod.Address != "" {
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
				ClusterIP: StringNone,
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
							ServiceKey: cServiceName,
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

				for iPodIdx := range len(stToPodInfo.ToPods) {
					if stToPodInfo.ToPods[iPodIdx].DeleteFlag {
						continue
					}
					if stToPodInfo.ToPods[iPodIdx].Address == "" {
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
			if err != nil {
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

func (r *DCIAppConnectionReconciler) deleteDciService(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	iConnIdx int,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) {
	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" iConnIdx: " + strconv.Itoa(iConnIdx))

	for range 1 {
		stToDeploymentInfo := pstToDeploymentInfos.ToDeploymentInfos[iConnIdx]
		cServiceName := stToDeploymentInfo.DeploymentRef.Name
		cServiceNamespace := pstCRData.Namespace

		bAdditionalProcFlag := false
		bPodControllerFlag := false
		var pstCRService *corev1.Service
		_, pstCRService, _, err = getToConnectionInformationService(
			ctx,
			r.Client,
			log,
			cServiceName,
			cServiceNamespace,
			pstCRData.Name,
			pstCRData.Namespace,
			bAdditionalProcFlag,
			bPodControllerFlag)
		if err != nil {
			break
		}
		if pstCRService == nil {
			break
		}
		if !pstCRService.DeletionTimestamp.IsZero() {
			break
		}

		var iOwnerIdx int
		for iOwnerIdx = 0; len(pstCRService.OwnerReferences) > iOwnerIdx; iOwnerIdx++ {
			if pstCRService.OwnerReferences[iOwnerIdx].Name == pstCRData.Name {
				break
			}
		}
		if len(pstCRService.OwnerReferences) == iOwnerIdx {
			log.Info(cFuncName + " Service not authorized as an owner." +
				" Name: " + pstCRService.Name +
				", Namespace: " + pstCRService.Namespace)
			break
		}

		err = r.Delete(ctx, pstCRService)
		if err != nil {
			log.Error(err, cFuncName+" failed to delete Service resource."+
				" Name: "+pstCRService.Name+
				", Namespace: "+pstCRService.Namespace)
			break
		}
		log.Info(cFuncName + " deleted Service resource successfully." +
			" Name: " + pstCRService.Name +
			", Namespace: " + pstCRService.Namespace)
	}

	log.Info(cFuncName + " End.")
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

	for range 1 {
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
}

func (r *DCIAppConnectionReconciler) updateRollBack(
	ctx context.Context,
	log logr.Logger,
	sUpdateConnectEndpointCRs []*examplecomv1.ConnectEndpoint) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	pstCRConnectEndpoint := &examplecomv1.ConnectEndpoint{}

	for range 1 {
		// RollBack ConnectEndpoint
		for _, cep := range sUpdateConnectEndpointCRs {
			err := r.Get(ctx,
				client.ObjectKey{
					Name:      cep.Name,
					Namespace: cep.Namespace,
				}, pstCRConnectEndpoint)
			if err != nil {
				log.Error(err, cFuncName+" failed to get ConnectEndpointCR."+
					" Name: "+cep.Name+
					", Namespace: "+cep.Namespace)
			} else {
				pstBaseCRConnectEndpoint := client.MergeFrom(
					pstCRConnectEndpoint.DeepCopy())
				pstCRConnectEndpoint.Spec.InputEndpoint = cep.Spec.InputEndpoint
				pstCRConnectEndpoint.Spec.OutputEndpoint = cep.Spec.OutputEndpoint
				pstCRConnectEndpoint.Spec.To = cep.Spec.To

				err = r.Patch(ctx, pstCRConnectEndpoint, pstBaseCRConnectEndpoint)
				if err != nil {
					log.Error(err, cFuncName+
						" failed to patch ConnectEndpointCR, and the rollback could not be completed."+
						" Name: "+cep.Name+
						", Namespace: "+cep.Namespace)
				} else {
					log.Info(cFuncName + " rollback patched ConnectEndpointCR successfully." +
						" Name: " + cep.Name +
						", Namespace: " + cep.Namespace)
				}
			}
		}
	}
	log.Info(cFuncName + " End.")
}

func chkConnectionRelation(
	_ context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection) bool {
	bProcFlag := false

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")
	var bFuncRet bool

	for range 1 {
		bFuncRet = chkConnectionRelationship(
			log,
			pstCRData)

		if bFuncRet {
			bFuncRet = chkConnectionRelationshipInput(
				log,
				pstCRData)
		}

		if bFuncRet {
			bFuncRet = chkConnectionRelationshipOutput(
				log,
				pstCRData)
		}

		if bFuncRet {
			bProcFlag = true
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag))
	return bProcFlag
}

func chkConnectionRelationship(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection) bool {
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for range 1 {
		// Connections Relation
		if len(pstCRData.Spec.Connections) != len(pstCRData.Status.Connections) {
			bProcFlag = false
			break
		}
		for iConnIdx := range len(pstCRData.Spec.Connections) {
			if pstCRData.Spec.Connections[iConnIdx].From.DeploymentRef !=
				pstCRData.Status.Connections[iConnIdx].From.DeploymentRef {
				bProcFlag = false
				break
			}
			if pstCRData.Spec.Connections[iConnIdx].To.DeploymentRef !=
				pstCRData.Status.Connections[iConnIdx].To.DeploymentRef {
				bProcFlag = false
				break
			}
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag))
	return bProcFlag
}

func chkConnectionRelationshipInput(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection) bool {
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for range 1 {
		// InputEndpoint Relation
		if len(pstCRData.Spec.InputEndpoints) != len(pstCRData.Status.InputEndpoints) {
			bProcFlag = false
			break
		}
		for iInIdx := range len(pstCRData.Spec.InputEndpoints) {
			if pstCRData.Spec.InputEndpoints[iInIdx].Internal.DeploymentRef !=
				pstCRData.Status.InputEndpoints[iInIdx].Internal.DeploymentRef {
				bProcFlag = false
				break
			}
			if !slices.Equal(pstCRData.Spec.InputEndpoints[iInIdx].ExternalSubsets,
				pstCRData.Status.InputEndpoints[iInIdx].ExternalSubsets) {
				bProcFlag = false
				break
			}
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag))
	return bProcFlag
}

func chkConnectionRelationshipOutput(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection) bool {
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for range 1 {
		// OutputEndpoints Relation
		if len(pstCRData.Spec.OutputEndpoints) != len(pstCRData.Status.OutputEndpoints) {
			bProcFlag = false
			break
		}
		for iOutIdx := range len(pstCRData.Spec.OutputEndpoints) {
			if pstCRData.Spec.OutputEndpoints[iOutIdx].Internal.DeploymentRef !=
				pstCRData.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef {
				bProcFlag = false
				break
			}
			if !slices.Equal(pstCRData.Spec.OutputEndpoints[iOutIdx].ExternalSubsets,
				pstCRData.Status.OutputEndpoints[iOutIdx].ExternalSubsets) {
				bProcFlag = false
				break
			}
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag))
	return bProcFlag
}

func getStatusRelation(
	_ context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	getStatusRelationship(
		log,
		pstCRData,
		pstFromDeploymentInfos,
		pstToDeploymentInfos)

	getStatusRelationshipInput(
		log,
		pstCRData,
		pstFromDeploymentInfos)

	getStatusRelationshipOutput(
		log,
		pstCRData,
		pstFromDeploymentInfos)

	log.Info(cFuncName + " End.")
}

func getStatusRelationship(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos,
	pstToDeploymentInfos *examplecomv1.ToDeploymentInfos) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	arConnections := pstCRData.Status.Connections

	// Connections Relation
	for iConnIdx := range arConnections {
		pstConnection := arConnections[iConnIdx]
		bDupFlag := false

		for iFromDeploymentIdx := range len(pstFromDeploymentInfos.FromDeployments) {
			stFromDeployment := pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx]

			if stFromDeployment.DeploymentRef.Name == pstConnection.From.DeploymentRef.Name &&
				stFromDeployment.DeploymentRef.Namespace == pstConnection.From.DeploymentRef.Namespace {
				stFromDeployment.DeploymentIndexs = append(
					stFromDeployment.DeploymentIndexs,
					iConnIdx)

				pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx] = stFromDeployment
				bDupFlag = true
				log.Info(cFuncName + " set the DeploymentIndexs Info in the FromDeployment." +
					" DeploymentIndexs len: " +
					strconv.Itoa(len(pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].DeploymentIndexs)) +
					", Name: " + pstConnection.From.DeploymentRef.Name)
			}
		}

		if !bDupFlag {
			log.Info(cFuncName + " set FromDeployment Info." +
				" Index: " + strconv.Itoa(iConnIdx) +
				", Name: " + pstConnection.From.DeploymentRef.Name)

			iIOMode := examplecomv1.IOModeNONE

			var stFromDeployment examplecomv1.FromDeploymentInfo
			stFromDeployment.DeploymentRef.Name = pstConnection.From.DeploymentRef.Name
			stFromDeployment.DeploymentRef.Namespace = pstConnection.From.DeploymentRef.Namespace
			stFromDeployment.IOMode = iIOMode
			stFromDeployment.DeleteFlag = true
			stFromDeployment.DeploymentIndexs = append(
				stFromDeployment.DeploymentIndexs,
				iConnIdx)
			pstFromDeploymentInfos.FromDeployments = append(
				pstFromDeploymentInfos.FromDeployments,
				stFromDeployment)
		}

		log.Info(cFuncName + " set ToDeployment Info. Name: " + pstConnection.To.DeploymentRef.Name)
		stToDeployments := pstToDeploymentInfos.ToDeploymentInfos[iConnIdx]
		stToDeployments.DeploymentRef.Name = pstConnection.To.DeploymentRef.Name
		stToDeployments.DeploymentRef.Namespace = pstConnection.To.DeploymentRef.Namespace
		stToDeployments.DeleteFlag = true
		pstToDeploymentInfos.ToDeploymentInfos[iConnIdx] = stToDeployments
	}

	log.Info(cFuncName + " End.")
}

func getStatusRelationshipInput(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	arInputEndpoints := pstCRData.Status.InputEndpoints

	for iInIdx := range arInputEndpoints {
		inDeploymentRef :=
			arInputEndpoints[iInIdx].Internal.DeploymentRef

		bConnExistFlag := false
		for iFromDeploymentIdx := range len(pstFromDeploymentInfos.FromDeployments) {
			stFromDeployment := pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx]
			if inDeploymentRef.Name == stFromDeployment.DeploymentRef.Name &&
				inDeploymentRef.Namespace == stFromDeployment.DeploymentRef.Namespace {
				bConnExistFlag = true
				arInputSubsets := stFromDeployment.InputSubsets
				for iSubIdx := range len(arInputEndpoints[iInIdx].ExternalSubsets) {
					var iFromSubIdx int
					for iFromSubIdx = 0; len(arInputSubsets) > iFromSubIdx; iFromSubIdx++ {
						if arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
							arInputSubsets[iFromSubIdx] {
							break
						}
					}
					if len(arInputSubsets) == iFromSubIdx {
						arInputSubsets = append(arInputSubsets,
							arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
					}
				}
				pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].InputSubsets = arInputSubsets
				pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].IOMode |= examplecomv1.IOModeIN
				log.Info(cFuncName + " set the InputSubsets Info in the FromDeployment." +
					" InputSubsets len: " +
					strconv.Itoa(len(pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].InputSubsets)) +
					", Name: " + inDeploymentRef.Name)
			}
		}

		if !bConnExistFlag {
			log.Info(cFuncName + " set FromDeployment Info." +
				" Index: " + strconv.Itoa(iInIdx) +
				", Name: " + inDeploymentRef.Name)

			var arInputSubsets []string
			for iSubIdx := range len(arInputEndpoints[iInIdx].ExternalSubsets) {
				var iFromSubIdx int
				for iFromSubIdx = 0; len(arInputSubsets) > iFromSubIdx; iFromSubIdx++ {
					if arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
						arInputSubsets[iFromSubIdx] {
						break
					}
				}
				if len(arInputSubsets) == iFromSubIdx {
					arInputSubsets = append(arInputSubsets,
						arInputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
				}
			}

			var stFromDeployment examplecomv1.FromDeploymentInfo
			stFromDeployment.DeploymentRef.Name = inDeploymentRef.Name
			stFromDeployment.DeploymentRef.Namespace = inDeploymentRef.Namespace
			stFromDeployment.InputSubsets = arInputSubsets
			stFromDeployment.IOMode = examplecomv1.IOModeIN
			stFromDeployment.DeleteFlag = true
			pstFromDeploymentInfos.FromDeployments = append(
				pstFromDeploymentInfos.FromDeployments,
				stFromDeployment)
		}
	}

	log.Info(cFuncName + " End.")
}

func getStatusRelationshipOutput(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppConnection,
	pstFromDeploymentInfos *examplecomv1.FromDeploymentInfos) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	arOutputEndpoints := pstCRData.Status.OutputEndpoints

	for iOutIdx := range arOutputEndpoints {
		outDeploymentRef :=
			arOutputEndpoints[iOutIdx].Internal.DeploymentRef

		bFromExistFlag := false
		for iFromDeploymentIdx := range len(pstFromDeploymentInfos.FromDeployments) {
			stFromDeployment := pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx]
			if outDeploymentRef.Name == stFromDeployment.DeploymentRef.Name &&
				outDeploymentRef.Namespace == stFromDeployment.DeploymentRef.Namespace {
				bFromExistFlag = true
				arOutputSubsets := stFromDeployment.OutputSubsets
				for iSubIdx := range len(arOutputEndpoints[iOutIdx].ExternalSubsets) {
					var iFromSubIdx int
					for iFromSubIdx = 0; len(arOutputSubsets) > iFromSubIdx; iFromSubIdx++ {
						if arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
							arOutputSubsets[iFromSubIdx] {
							break
						}
					}
					if len(arOutputSubsets) == iFromSubIdx {
						arOutputSubsets = append(arOutputSubsets,
							arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
					}
				}
				pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].OutputSubsets = arOutputSubsets
				pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].IOMode |= examplecomv1.IOModeOUT
				log.Info(cFuncName + " set the OutputSubsets Info in the FromDeployment." +
					" OutputSubsets len: " +
					strconv.Itoa(len(pstFromDeploymentInfos.FromDeployments[iFromDeploymentIdx].OutputSubsets)) +
					", Name: " + outDeploymentRef.Name)
			}
		}

		if !bFromExistFlag {
			log.Info(cFuncName + " set FromDeployment Info." +
				" Index: " + strconv.Itoa(iOutIdx) +
				", Name: " + outDeploymentRef.Name)

			var arOutputSubsets []string
			for iSubIdx := range len(arOutputEndpoints[iOutIdx].ExternalSubsets) {
				var iFromSubIdx int
				for iFromSubIdx = 0; len(arOutputSubsets) > iFromSubIdx; iFromSubIdx++ {
					if arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
						arOutputSubsets[iFromSubIdx] {
						break
					}
				}
				if len(arOutputSubsets) == iFromSubIdx {
					arOutputSubsets = append(arOutputSubsets,
						arOutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
				}
			}
			var stFromDeployment examplecomv1.FromDeploymentInfo
			stFromDeployment.DeploymentRef.Name = outDeploymentRef.Name
			stFromDeployment.DeploymentRef.Namespace = outDeploymentRef.Namespace
			stFromDeployment.OutputSubsets = arOutputSubsets
			stFromDeployment.IOMode = examplecomv1.IOModeOUT
			stFromDeployment.DeleteFlag = true
			pstFromDeploymentInfos.FromDeployments = append(
				pstFromDeploymentInfos.FromDeployments,
				stFromDeployment)
		}
	}

	log.Info(cFuncName + " End.")
}
