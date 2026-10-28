package controller

import (
	examplecomv1 "app-conn-info-maker-dci/api/v1"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

type PodEventPredicate struct {
	predicate.Funcs
}

const (
	DCIAppConnectionPodReconcilerInterval = 1
	NamespaceDefault                      = "default"
	PodRateLimiterMaxDelay                = 5 * time.Minute
	PodRateLimiterRPS                     = 10
	PodRateLimiterBurst                   = 100
)

const (
	DCIAppConnectionPodEventKindNone   = "None"
	DCIAppConnectionPodEventKindCreate = "Create"
	DCIAppConnectionPodEventKindUpdate = "Update"
	DCIAppConnectionPodEventKindDelete = "Delete"
)

const (
	DCIAppConnectionFromToSubjectNone = 0
	DCIAppConnectionFromToSubjectFrom = 1
	DCIAppConnectionFromToSubjectTo   = 2
)

// +kubebuilder:rbac:groups=example.com,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=dciappconnections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mellanox.com,resources=hostdevicenetworks,verbs=get;list;watch
// +kubebuilder:rbac:groups=mellanox.com,resources=hostdevicenetworks/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dcipodconnectinformation", req.Name)
	var err error
	stCtrlResult := ctrl.Result{}
	bProcFlag := true
	cFuncName := getFuncName()

	for range 1 {
		stFromPodInfos := examplecomv1.FromPodInfos{}
		stToPodInfos := examplecomv1.ToPodInfos{}
		stToPodInfos.ToPodInfos = make(map[int]examplecomv1.ToPods)

		pstCRData := &corev1.Pod{}
		err = r.Get(ctx, req.NamespacedName, pstCRData)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info(cFuncName + " could not find Pod resource." +
					" Name: " + req.Name +
					", Namespace: " + req.Namespace)
				err = nil
			} else {
				log.Error(err, cFuncName+" failed to get Pod resource."+
					" Name: "+req.Name+
					", Namespace: "+req.Namespace)
				bProcFlag = false
			}
			break
		}
		log.Info(cFuncName + " successfully fetched the Pod resource.")

		// Get 2nd Nic Address
		var c2ndNicAddress string
		var bFuncRet bool
		err, bFuncRet = get2ndNicAddress(
			ctx,
			r.Client,
			log,
			pstCRData,
			&c2ndNicAddress)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		pstCRDeployment := &appsv1.Deployment{}
		pstCRConnection := &examplecomv1.DCIAppConnection{}
		cFinalizerName := getFinalizerName(r.Client, pstCRConnection)
		iFromToSubject := DCIAppConnectionFromToSubjectNone

		// Get the event type
		cEventKind := getEventKind(pstCRData, cFinalizerName)

		// Get PodRelationship Info
		if DCIAppConnectionPodEventKindCreate == cEventKind ||
			DCIAppConnectionPodEventKindDelete == cEventKind {
			err, bFuncRet = r.getPodRelationshipInfo(
				ctx,
				log,
				pstCRData,
				&stFromPodInfos,
				&stToPodInfos,
				c2ndNicAddress,
				pstCRDeployment,
				pstCRConnection,
				cFinalizerName,
				&iFromToSubject,
				cEventKind)

			if err != nil {
				bProcFlag = false
				break
			}
			if !bFuncRet {
				break
			}
		}

		if DCIAppConnectionPodEventKindCreate == cEventKind {
			err, bFuncRet = r.reconcileCreate(
				ctx,
				log,
				pstCRData,
				c2ndNicAddress,
				cFinalizerName,
				pstCRConnection,
				iFromToSubject,
				&stFromPodInfos,
				&stToPodInfos)

			if err != nil {
				bProcFlag = false
				break
			}

			if !bFuncRet {
				bProcFlag = false
				break
			}
		} else if DCIAppConnectionPodEventKindUpdate == cEventKind {
			// do nothing
			log.Info(cFuncName + " has no action for the update event of Pod resource.")
		} else if DCIAppConnectionPodEventKindDelete == cEventKind {
			err = r.reconcileDelete(
				ctx,
				log,
				pstCRData,
				cFinalizerName,
				pstCRConnection,
				iFromToSubject,
				&stFromPodInfos,
				&stToPodInfos)

			if err != nil {
				bProcFlag = false
				break
			}
		} else if DCIAppConnectionPodEventKindNone == cEventKind {
			// do noting
			log.Info(cFuncName + " has no action of this type event for Pod resource.")
		}
	}

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
			time.Second * DCIAppConnectionPodReconcilerInterval
	}

	//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
	log.Info(cFuncName + " return." +
		" stCtrlResult.Requeue: " + strconv.FormatBool(stCtrlResult.Requeue) +
		", stCtrlResult.RequeueAfter: " + stCtrlResult.RequeueAfter.String() +
		", err: " + fmt.Sprint(err))
	return stCtrlResult, err
}

func (PodEventPredicate) Update(e event.UpdateEvent) bool {
	bProcFlag := true
	for range 1 {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			break
		}

		oldObj, okOld := e.ObjectOld.(*corev1.Pod)
		newObj, okNew := e.ObjectNew.(*corev1.Pod)

		if !okOld || !okNew {
			break
		}

		if !newObj.DeletionTimestamp.IsZero() {
			break
		}

		if !reflect.DeepEqual(newObj.GetAnnotations(), oldObj.GetAnnotations()) {
			break
		}

		bOldReadyFlag := getConditionReady(oldObj)
		bNewReadyFlag := getConditionReady(newObj)
		if bOldReadyFlag != bNewReadyFlag {
			break
		}
		bProcFlag = false
	}
	return bProcFlag
}

func (PodEventPredicate) Create(_ event.CreateEvent) bool {
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&corev1.Pod{}).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
					1*time.Second,
					PodRateLimiterMaxDelay,
				),
				&workqueue.TypedBucketRateLimiter[ctrl.Request]{
					Limiter: rate.NewLimiter(
						rate.Limit(PodRateLimiterRPS),
						PodRateLimiterBurst,
					),
				},
			),
		}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			PodEventPredicate{},
		)).
		Named("pod").
		Owns(&discoveryv1.EndpointSlice{}).
		Complete(r)
}

func getEventKind(
	pstCRData *corev1.Pod,
	cFinalizerName string) string {
	cEventKind := DCIAppConnectionPodEventKindNone

	if !controllerutil.ContainsFinalizer(pstCRData, cFinalizerName) {
		if !pstCRData.DeletionTimestamp.IsZero() {
			// do nothing
		} else {
			cEventKind = DCIAppConnectionPodEventKindCreate
		}
	} else {
		if !pstCRData.DeletionTimestamp.IsZero() {
			cEventKind = DCIAppConnectionPodEventKindDelete
		} else {
			cEventKind = DCIAppConnectionPodEventKindUpdate
		}
	}
	return cEventKind
}

func (r *PodReconciler) reconcileCreate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	c2ndNicAddress string,
	cFinalizerName string,
	pstCRConnection *examplecomv1.DCIAppConnection,
	iFromToSubject int,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) (error, bool) {
	var err error
	bProcFlag := false

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" c2ndNicAddress: " + c2ndNicAddress +
		", DCIAppConnectionRef: " + pstCRConnection.Name + "," + pstCRConnection.Namespace +
		", iFromToSubject: " + strconv.Itoa(iFromToSubject))

	for range 1 {
		bReadyFlag := getConditionReady(pstCRData)
		if !bReadyFlag {
			log.Info(cFuncName + " detected that Pod resource Status.Conditions is not Ready." +
				" Name: " + pstCRData.Name +
				", Namespace: " + pstCRData.Namespace)
			break
		}

		cLabels := KubernetesAppNameLabel
		pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())
		cVal := pstCRConnection.Name + "_" + pstCRConnection.Namespace
		gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRConnection)
		cKind := strings.ToLower(gvks[0].Kind)
		pstCRData.Labels[cKind] = cVal
		controllerutil.AddFinalizer(pstCRData, cFinalizerName)
		pstEndpointslice := &discoveryv1.EndpointSlice{}

		if DCIAppConnectionFromToSubjectTo == (iFromToSubject & DCIAppConnectionFromToSubjectTo) {
			for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
				if stToPodInfo.DeploymentRef.Name == pstCRData.Labels[cLabels] &&
					stToPodInfo.DeploymentRef.Namespace == pstCRData.Namespace {
					bPodHitFlag := false
					for _, stToPod := range stToPodInfo.ToPods {
						if stToPod.PodR.Name != pstCRData.Name ||
							stToPod.PodR.Namespace != pstCRData.Namespace {
							continue
						}
						bPodHitFlag = true

						pstEndpointslice = stToPodInfo.EndpointSliceR
						if pstEndpointslice != nil && c2ndNicAddress != "" {
							err = r.addAddressEndpointSlice(
								ctx,
								log,
								pstCRData,
								pstEndpointslice,
								c2ndNicAddress)
							if err != nil {
								break
							}
						}
						log.Info(cFuncName + " added a finalizer to Pod resource.")
						err = r.Patch(ctx, pstCRData, pstBaseCRDData)
						if err != nil {
							log.Error(err, cFuncName+" failed to patch Pod resource."+
								" Name: "+pstCRData.Name+
								", Namespace: "+pstCRData.Namespace)

							if pstEndpointslice != nil &&
								stToPod.Address != "" {
								_ = r.delAddressEndpointSlice(
									ctx,
									log,
									pstCRData,
									pstEndpointslice)
							}
							break
						}
						log.Info(cFuncName + " patched Pod resource successfully." +
							" Name: " + pstCRData.Name +
							", Namespace: " + pstCRData.Namespace)
						break
					}
					if bPodHitFlag {
						break
					}
				}
			}
			if err != nil {
				break
			}
		}

		for index := range pstFromPodInfos.FromPods {
			stFromPod := &pstFromPodInfos.FromPods[index]
			if stFromPod.PodR == nil {
				continue
			}
			err = makeConnectEndpoint(
				ctx,
				r.Client,
				log,
				r.Scheme,
				pstCRConnection,
				stFromPod,
				pstToPodInfos)
			if err != nil {
				for ind := range index {
					if pstFromPodInfos.FromPods[ind].PodR == nil {
						continue
					}
					if pstFromPodInfos.FromPods[ind].PodR.Name == pstCRData.Name &&
						pstFromPodInfos.FromPods[ind].PodR.Namespace == pstCRData.Namespace {
						_ = deleteConnectEndpoint(
							ctx,
							r.Client,
							log,
							pstFromPodInfos.FromPods[ind].ConnectEndpointCR)
					} else {
						for _, depindex := range pstFromPodInfos.FromPods[ind].DeploymentIndexs {
							for podInd, stToPod := range pstToPodInfos.ToPodInfos[depindex].ToPods {
								if stToPod.PodR.Name == pstCRData.Name &&
									stToPod.PodR.Namespace == pstCRData.Namespace {
									pstToPodInfos.ToPodInfos[depindex].ToPods[podInd].DeleteFlag = true
								}
							}
						}
						_ = makeConnectEndpoint(
							ctx,
							r.Client,
							log,
							r.Scheme,
							pstCRConnection,
							&pstFromPodInfos.FromPods[ind],
							pstToPodInfos)
					}
				}

				delete(pstCRData.Labels, cKind)
				controllerutil.RemoveFinalizer(pstCRData, cFinalizerName)
				_ = r.Patch(ctx, pstCRData, pstBaseCRDData)

				if DCIAppConnectionFromToSubjectTo == (iFromToSubject & DCIAppConnectionFromToSubjectTo) {
					for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
						if stToPodInfo.DeploymentRef.Name == pstCRData.Labels[cLabels] &&
							stToPodInfo.DeploymentRef.Namespace == pstCRData.Namespace &&
							pstEndpointslice != nil &&
							stToPodInfo.ToPods[0].Address != "" {
							_ = r.delAddressEndpointSlice(
								ctx,
								log,
								pstCRData,
								pstEndpointslice)
						}
					}
				}
				break
			} else if stFromPod.PodR.Name == pstCRData.Name &&
				stFromPod.PodR.Namespace == pstCRData.Namespace {
				if DCIAppConnectionFromToSubjectTo !=
					(iFromToSubject & DCIAppConnectionFromToSubjectTo) {
					err = r.Patch(ctx, pstCRData, pstBaseCRDData)
					if err != nil {
						log.Error(err, cFuncName+" failed to patch Pod resource."+
							" Name: "+pstCRData.Name+
							", Namespace: "+pstCRData.Namespace)
					}
				}
			}
		}
		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *PodReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	cFinalizerName string,
	pstCRConnection *examplecomv1.DCIAppConnection,
	iFromToSubject int,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) error {
	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DCIAppConnectionRef: " + pstCRConnection.Name + "," + pstCRConnection.Namespace +
		", iFromToSubject: " + strconv.Itoa(iFromToSubject))

	for range 1 {
		cLabels := KubernetesAppNameLabel
		if DCIAppConnectionFromToSubjectTo == (iFromToSubject & DCIAppConnectionFromToSubjectTo) {
			for _, stToPodInfo := range pstToPodInfos.ToPodInfos {
				if stToPodInfo.DeploymentRef.Name == pstCRData.Labels[cLabels] &&
					stToPodInfo.DeploymentRef.Namespace == pstCRData.Namespace {
					if stToPodInfo.EndpointSliceR != nil {
						for _, stToPod := range stToPodInfo.ToPods {
							if stToPod.Address != "" {
								_ = r.delAddressEndpointSlice(
									ctx,
									log,
									pstCRData,
									stToPodInfo.EndpointSliceR)
							}
						}
					}
					break
				}
			}

			for i := range pstFromPodInfos.FromPods {
				stFromPod := &pstFromPodInfos.FromPods[i]
				if stFromPod.PodR == nil {
					continue
				}
				if stFromPod.PodR.Name != pstCRData.Name ||
					stFromPod.PodR.Namespace != pstCRData.Namespace {
					err = makeConnectEndpoint(
						ctx,
						r.Client,
						log,
						r.Scheme,
						pstCRConnection,
						stFromPod,
						pstToPodInfos)
					if err != nil {
						break
					}
				}
			}
			if err != nil {
				break
			}
		}

		if DCIAppConnectionFromToSubjectFrom == (iFromToSubject & DCIAppConnectionFromToSubjectFrom) {
			for i := range pstFromPodInfos.FromPods {
				stFromPod := &pstFromPodInfos.FromPods[i]
				if stFromPod.PodR == nil {
					continue
				}
				if stFromPod.PodR.Name == pstCRData.Name &&
					stFromPod.PodR.Namespace == pstCRData.Namespace {
					err = deleteConnectEndpoint(
						ctx,
						r.Client,
						log,
						stFromPod.ConnectEndpointCR)
					if err != nil {
						break
					}
				}
			}
			if err != nil {
				break
			}
		}

		pstBaseCRData := client.MergeFrom(pstCRData.DeepCopy())

		controllerutil.RemoveFinalizer(pstCRData, cFinalizerName)
		log.Info(cFuncName + " removed the finalizer from Pod resource.")

		err = r.Patch(ctx, pstCRData, pstBaseCRData)
		if err != nil {
			log.Error(err, cFuncName+" failed to patch Pod resource."+
				" Name: "+pstCRData.Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}
		log.Info(cFuncName + " patched Pod resource successfully." +
			" Name: " + pstCRData.Name +
			", Namespace: " + pstCRData.Namespace)
	}

	log.Info(cFuncName + " End." +
		", err: " + fmt.Sprint(err))
	return err
}

func (r *PodReconciler) getPodConnection(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	pstCRDeployment *appsv1.Deployment,
	pstCRConnection *examplecomv1.DCIAppConnection) (error, bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")
	var err error
	bProcFlag := false

	for range 1 {
		// Get ReplicaSet Resource
		pstCRReplicaSet := &appsv1.ReplicaSet{}
		err = r.Get(ctx,
			client.ObjectKey{
				Name:      pstCRData.OwnerReferences[0].Name,
				Namespace: pstCRData.Namespace,
			},
			pstCRReplicaSet)
		if errors.IsNotFound(err) {
			log.Info(cFuncName + " could not find ReplicaSet resource." +
				" Name: " + pstCRData.OwnerReferences[0].Name +
				", Namespace: " + pstCRData.Namespace)
			err = nil
			break
		} else if err != nil {
			log.Error(err, cFuncName+" failed to get ReplicaSet resource."+
				" Name: "+pstCRData.OwnerReferences[0].Name+
				", Namespace: "+pstCRData.Namespace)
			break
		}
		if !pstCRReplicaSet.DeletionTimestamp.IsZero() {
			log.Info(cFuncName + " detected that ReplicaSet resource has a DeletionTimestamp." +
				" Name: " + pstCRReplicaSet.Name +
				", Namespace: " + pstCRReplicaSet.Namespace)
			break
		}

		// Get Deployment Resource
		err = r.Get(ctx,
			client.ObjectKey{
				Name:      pstCRReplicaSet.OwnerReferences[0].Name,
				Namespace: pstCRReplicaSet.Namespace,
			},
			pstCRDeployment)
		if errors.IsNotFound(err) {
			log.Info(cFuncName + " could not find Deployment resource." +
				" Name: " + pstCRReplicaSet.OwnerReferences[0].Name +
				", Namespace: " + pstCRReplicaSet.Namespace)
			err = nil
			break
		} else if err != nil {
			log.Error(err, cFuncName+" failed to get Deployment resource."+
				" Name: "+pstCRReplicaSet.OwnerReferences[0].Name+
				", Namespace: "+pstCRReplicaSet.Namespace)
			break
		}
		if !pstCRDeployment.DeletionTimestamp.IsZero() {
			log.Info(cFuncName + " detected that Deployment resource has a DeletionTimestamp." +
				" Name: " + pstCRDeployment.Name +
				", Namespace: " + pstCRDeployment.Namespace)
			break
		}

		if pstCRDeployment.Labels == nil {
			log.Info(cFuncName + " could not find Deployment resource Labels." +
				" Name: " + pstCRDeployment.Name +
				", Namespace: " + pstCRDeployment.Namespace)
			break
		}

		gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRConnection)
		cKind := strings.ToLower(gvks[0].Kind)

		cLabel, exists := pstCRDeployment.Labels[cKind]
		if !exists {
			log.Info(cFuncName + " could not find Deployment resource Label for DCIAppConnectionCR." +
				" Labels Key: " + cKind +
				", Name: " + pstCRReplicaSet.OwnerReferences[0].Name +
				", Namespace: " + pstCRReplicaSet.Namespace)
			break
		}
		if cLabel == "" {
			log.Info(cFuncName + " detected that Deployment resource Label value length is Zero." +
				" Labels Key: " + cKind +
				", Name: " + pstCRReplicaSet.OwnerReferences[0].Name +
				", Namespace: " + pstCRReplicaSet.Namespace)
			break
		}
		log.Info(cFuncName + " detected that Deployment resource Label for DCIAppConnectionCR." +
			" Labels Key: " + cKind +
			", Val: " + cLabel +
			", Name: " + pstCRReplicaSet.OwnerReferences[0].Name +
			", Namespace: " + pstCRReplicaSet.Namespace)

		cNameNamespace := strings.Split(cLabel, "_")
		if len(cNameNamespace) != 0 {
			cNameNamespace = append(cNameNamespace, NamespaceDefault)
		}

		// Get DCIAppConnectionCR
		err = r.Get(ctx,
			client.ObjectKey{
				Name:      cNameNamespace[0],
				Namespace: cNameNamespace[1],
			},
			pstCRConnection)
		if errors.IsNotFound(err) {
			log.Info(cFuncName + " could not find DCIAppConnectionCR." +
				" Name: " + cNameNamespace[0] +
				", Namespace: " + cNameNamespace[1])
			err = nil
			break
		} else if err != nil {
			log.Error(err, cFuncName+" failed to get DCIAppConnectionCR."+
				" Name: "+cNameNamespace[0]+
				", Namespace: "+cNameNamespace[1])
			break
		}
		if !pstCRConnection.DeletionTimestamp.IsZero() {
			log.Info(cFuncName + " detected that DCIAppConnectionCR has a DeletionTimestamp." +
				" Name: " + pstCRConnection.Name +
				", Namespace: " + pstCRConnection.Namespace)
			break
		}
		if pstCRConnection.Status.Phase != examplecomv1.StatusPhaseCreated {
			log.Info(cFuncName + " detected that DCIAppConnectionCR is Not Created." +
				" Name: " + pstCRConnection.Name +
				", Namespace: " + pstCRConnection.Namespace)
			break
		}

		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *PodReconciler) getPodConnectRelationshipInOut(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	c2ndNicAddress string,
	iIOMode int,
	pstCRDeployment *appsv1.Deployment,
	pstCRConnection *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	piFromToSubject *int) (error, bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" c2ndNicAddress: " + c2ndNicAddress +
		", iIOMode: " + strconv.Itoa(iIOMode) +
		", DeploymentRef: " + pstCRDeployment.Name + "," + pstCRDeployment.Namespace +
		", DCIAppConnectionRef: " + pstCRConnection.Name + "," + pstCRConnection.Namespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", iFromToSubject: " + strconv.Itoa(*piFromToSubject))
	bProcFlag := true
	var err error

	for range 1 {
		bFromPodInfoExistFlag := false
		var iFromIdx int
		for i := range pstFromPodInfos.FromPods {
			iFromIdx = i
			if pstCRDeployment.Name == pstFromPodInfos.FromPods[iFromIdx].DeploymentRef.Name &&
				pstCRDeployment.Namespace == pstFromPodInfos.FromPods[iFromIdx].DeploymentRef.Namespace {
				bFromPodInfoExistFlag = true
				break
			}
		}

		var arInputSubsets []string
		for iInIdx := range pstCRConnection.Status.InputEndpoints {
			if pstCRDeployment.Name == pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Name &&
				pstCRDeployment.Namespace == pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Namespace {
				for iSubIdx := range pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets {
					bSubsetExists := false
					for iFromSubIdx := range arInputSubsets {
						if pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
							arInputSubsets[iFromSubIdx] {
							bSubsetExists = true
							break
						}
					}
					if !bSubsetExists {
						arInputSubsets = append(arInputSubsets,
							pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
					}
				}
			}
		}
		var arOutputSubsets []string
		for iOutIdx := range pstCRConnection.Status.OutputEndpoints {
			if pstCRDeployment.Name == pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Name &&
				pstCRDeployment.Namespace == pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Namespace {
				for iSubIdx := range pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets {
					bSubsetExists := false
					for iFromSubIdx := range arOutputSubsets {
						if pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
							arOutputSubsets[iFromSubIdx] {
							bSubsetExists = true
							break
						}
					}
					if !bSubsetExists {
						arOutputSubsets = append(arOutputSubsets,
							pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
					}
				}
			}
		}

		if bFromPodInfoExistFlag {
			pstFromPodInfos.FromPods[iFromIdx].InputSubsets = arInputSubsets
			pstFromPodInfos.FromPods[iFromIdx].OutputSubsets = arOutputSubsets
			pstFromPodInfos.FromPods[iFromIdx].IOMode = iIOMode
		} else {
			stFromPod := examplecomv1.FromPodInfo{}
			stFromPod.DeploymentRef.Name = pstCRDeployment.Name
			stFromPod.DeploymentRef.Namespace = pstCRDeployment.Namespace
			stFromPod.DeploymentR = pstCRDeployment
			stFromPod.InputSubsets = arInputSubsets
			stFromPod.OutputSubsets = arOutputSubsets
			stFromPod.PodR = pstCRData
			stFromPod.Address = c2ndNicAddress
			stFromPod.IOMode = iIOMode
			if !pstCRData.DeletionTimestamp.IsZero() {
				stFromPod.DeleteFlag = true
			}
			// Get ConnectEndpoint
			stConnectEndpoint := &examplecomv1.ConnectEndpoint{}
			err = r.Get(ctx,
				client.ObjectKey{
					Name:      pstCRData.Name,
					Namespace: pstCRConnection.Namespace,
				},
				stConnectEndpoint)

			if err != nil {
				if !errors.IsNotFound(err) {
					log.Error(err, cFuncName+" failed to get ConnectEndpointCR."+
						" Name: "+pstCRData.Name+
						", Namespace: "+pstCRConnection.Namespace)
					bProcFlag = false
					break
				}
				err = nil
			} else if !stConnectEndpoint.DeletionTimestamp.IsZero() {
				if bAdditionalProcFlag {
					log.Error(err, cFuncName+
						" will requeue because ConnectEndpointCR has a DeletionTimestamp set. And it is a request for additional processing.")
					bProcFlag = false
					break
				}
			} else {
				stFromPod.ConnectEndpointCR = stConnectEndpoint
			}
			pstFromPodInfos.FromPods = append(pstFromPodInfos.FromPods, stFromPod)
			*piFromToSubject |= DCIAppConnectionFromToSubjectFrom
		}
	}
	log.Info(cFuncName + " End.")
	return err, bProcFlag
}

func (r *PodReconciler) getPodConnectRelationshipConn(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	c2ndNicAddress string,
	iIOMode int,
	pstCRDeployment *appsv1.Deployment,
	pstCRConnection *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	piFromToSubject *int) (error, bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" c2ndNicAddress: " + c2ndNicAddress +
		", iIOMode: " + strconv.Itoa(iIOMode) +
		", DeploymentRef: " + pstCRDeployment.Name + "," + pstCRDeployment.Namespace +
		", DCIAppConnectionRef: " + pstCRConnection.Name + "," + pstCRConnection.Namespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", iFromToSubject: " + strconv.Itoa(*piFromToSubject))
	var err error
	bProcFlag := false

	for iDeployIdx := range pstCRConnection.Status.Connections {
		stConnection := pstCRConnection.Status.Connections[iDeployIdx]
		stFromPod := examplecomv1.FromPodInfo{}

		if (*piFromToSubject & DCIAppConnectionFromToSubjectFrom) != DCIAppConnectionFromToSubjectFrom {
			// Set From Pod Information Equeal From DeploymentRef
			if pstCRDeployment.Name ==
				stConnection.From.DeploymentRef.Name &&
				pstCRDeployment.Namespace ==
					stConnection.From.DeploymentRef.Namespace {
				stFromPod.DeploymentRef.Name = pstCRDeployment.Name
				stFromPod.DeploymentRef.Namespace = pstCRDeployment.Namespace
				stFromPod.DeploymentR = pstCRDeployment
				stFromPod.PodR = pstCRData
				stFromPod.Address = c2ndNicAddress
				// Do not set for IOMode. Leave the default value]
				if !pstCRData.DeletionTimestamp.IsZero() {
					stFromPod.DeleteFlag = true
				}
				stFromPod.IOMode = iIOMode

				// Get ConnectEndpoint
				stConnectEndpoint := &examplecomv1.ConnectEndpoint{}
				err = r.Get(ctx,
					client.ObjectKey{
						Name:      pstCRData.Name,
						Namespace: pstCRConnection.Namespace,
					},
					stConnectEndpoint)

				if err != nil {
					if !errors.IsNotFound(err) {
						log.Error(err, cFuncName+" failed to get ConnectEndpointCR."+
							" Name: "+pstCRData.Name+
							", Namespace: "+pstCRConnection.Namespace)
						break
					}
					err = nil
				} else if !stConnectEndpoint.DeletionTimestamp.IsZero() {
					if bAdditionalProcFlag {
						log.Error(err, cFuncName+" will requeue because ConnectEndpointCR has a DeletionTimestamp set. "+
							" And it is a request for additional processing.")
						break
					}
				} else {
					stFromPod.ConnectEndpointCR = stConnectEndpoint
				}

				pstFromPodInfos.FromPods = append(
					pstFromPodInfos.FromPods,
					stFromPod)

				*piFromToSubject |= DCIAppConnectionFromToSubjectFrom
			}
		}

		if (*piFromToSubject & DCIAppConnectionFromToSubjectTo) != DCIAppConnectionFromToSubjectTo {
			// Search From Pod Information Equeal To DeploymentRef
			if pstCRDeployment.Name ==
				stConnection.To.DeploymentRef.Name &&
				pstCRDeployment.Namespace ==
					stConnection.To.DeploymentRef.Namespace {
				iToIOMode := examplecomv1.IOModeNONE
				var arInputSubsets []string
				for iInIdx := range pstCRConnection.Status.InputEndpoints {
					if stConnection.From.DeploymentRef.Name == pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Name &&
						stConnection.From.DeploymentRef.Namespace == pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Namespace {
						iToIOMode |= examplecomv1.IOModeIN
						for iSubIdx := range pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets {
							bSubsetExists := false
							for iFromSubIdx := range arInputSubsets {
								if pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets[iSubIdx] ==
									arInputSubsets[iFromSubIdx] {
									bSubsetExists = true
									break
								}
							}
							if !bSubsetExists {
								arInputSubsets = append(arInputSubsets,
									pstCRConnection.Status.InputEndpoints[iInIdx].ExternalSubsets[iSubIdx])
							}
						}
					}
				}

				var arOutputSubsets []string
				for iOutIdx := range pstCRConnection.Status.OutputEndpoints {
					if stConnection.From.DeploymentRef.Name == pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Name &&
						stConnection.From.DeploymentRef.Namespace == pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Namespace {
						iToIOMode |= examplecomv1.IOModeOUT
						for iSubIdx := range pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets {
							bSubsetExists := false
							for iFromSubIdx := range arOutputSubsets {
								if pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx] ==
									arOutputSubsets[iFromSubIdx] {
									bSubsetExists = true
									break
								}
							}
							if !bSubsetExists {
								arOutputSubsets = append(arOutputSubsets,
									pstCRConnection.Status.OutputEndpoints[iOutIdx].ExternalSubsets[iSubIdx])
							}
						}
					}
				}

				var bFuncRet bool
				err, bFuncRet = getFromConnectionInformation(
					ctx,
					r.Client,
					log,
					stConnection.From.DeploymentRef.Name,
					stConnection.From.DeploymentRef.Namespace,
					pstCRConnection.Name,
					pstCRConnection.Namespace,
					arInputSubsets,
					arOutputSubsets,
					bAdditionalProcFlag,
					true,
					pstFromPodInfos,
					-1,
					iToIOMode)

				if err != nil {
					break
				}
				if !bFuncRet {
					break
				}
				*piFromToSubject |= DCIAppConnectionFromToSubjectTo
			}
		}

		if *piFromToSubject != DCIAppConnectionFromToSubjectNone {
			bProcFlag = true
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *PodReconciler) getPodConnectRelationship(
	ctx context.Context,
	log logr.Logger,
	pstCRConnection *examplecomv1.DCIAppConnection,
	bAdditionalProcFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos) (error, bool) {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DCIAppConnectionRef: " + pstCRConnection.Name + "," + pstCRConnection.Namespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag))
	var err error
	bProcFlag := true

	// Set From Pod Information to Deployment Relationship
	for iConnIdx := range pstCRConnection.Status.Connections {
		stConnection := pstCRConnection.Status.Connections[iConnIdx]
		bFromActiveFlag := false
		log.Info("debug: " + cFuncName + " get Connections.From.Ref: " + stConnection.From.DeploymentRef.Name)

		for iFromIdx := range pstFromPodInfos.FromPods {
			stFromPod := pstFromPodInfos.FromPods[iFromIdx]

			log.Info("debug: " + cFuncName + " get stFromPod: " + stFromPod.DeploymentRef.Name)

			if stFromPod.DeploymentRef.Name ==
				stConnection.From.DeploymentRef.Name &&
				stFromPod.DeploymentRef.Namespace ==
					stConnection.From.DeploymentRef.Namespace {
				bFromActiveFlag = true
				pstFromPodInfos.FromPods[iFromIdx].DeploymentIndexs = append(
					pstFromPodInfos.FromPods[iFromIdx].DeploymentIndexs,
					iConnIdx)
			}
		}

		log.Info("debug: " + cFuncName + " get deployment.To.Ref: " + stConnection.To.DeploymentRef.Name +
			", bFromActiveFlag: " + strconv.FormatBool(bFromActiveFlag))

		if bFromActiveFlag {
			// Set To Pod Information
			var bFuncRet bool
			err, bFuncRet = getToConnectionInformation(
				ctx,
				r.Client,
				log,
				stConnection.To.DeploymentRef.Name,
				stConnection.To.DeploymentRef.Namespace,
				pstCRConnection.Name,
				pstCRConnection.Namespace,
				bAdditionalProcFlag,
				true,
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
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func (r *PodReconciler) addAddressEndpointSlice(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	pstCREndpointSlice *discoveryv1.EndpointSlice,
	c2ndNicAddress string) error {
	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" EndpointSliceRef: " + pstCREndpointSlice.Name + "," + pstCREndpointSlice.Namespace +
		", c2ndNicAddress: " + c2ndNicAddress)

	for range 1 {
		pstBaseCREndpointSlice := client.MergeFrom(pstCREndpointSlice.DeepCopy())

		getAppendEndpointSlice(
			ctx,
			r.Client,
			log,
			pstCRData,
			pstCREndpointSlice,
			c2ndNicAddress)

		err = r.Patch(ctx, pstCREndpointSlice, pstBaseCREndpointSlice)
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

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *PodReconciler) delAddressEndpointSlice(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	pstCREndpointSlice *discoveryv1.EndpointSlice) error {
	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" EndpointSliceRef: " + pstCREndpointSlice.Name + "," + pstCREndpointSlice.Namespace)
	var err error

	for range 1 {
		bIpExistFlag := false
		pstBaseCREndpointSlice := client.MergeFrom(pstCREndpointSlice.DeepCopy())

		gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRData)
		cKind := strings.ToLower(gvks[0].Kind)

		stEndpoints := []discoveryv1.Endpoint{}
		for iEndpointIdx := range pstCREndpointSlice.Endpoints {
			stEndpoint := pstCREndpointSlice.Endpoints[iEndpointIdx]

			if cKind != stEndpoint.TargetRef.Kind ||
				pstCRData.Name != stEndpoint.TargetRef.Name ||
				pstCRData.Namespace != stEndpoint.TargetRef.Namespace {
				stEndpoints = append(stEndpoints, stEndpoint)
				continue
			}
			bIpExistFlag = true
		}

		if bIpExistFlag {
			pstCREndpointSlice.Endpoints = stEndpoints
			err = r.Patch(ctx, pstCREndpointSlice, pstBaseCREndpointSlice)
			if err != nil {
				log.Error(err, cFuncName+" failed to patch EndpointSlice resource."+
					" Name: "+pstCREndpointSlice.Name+
					", Namespace: "+pstCREndpointSlice.Namespace)
				break
			}
			log.Info(cFuncName + " patched EndpointSlice resource successfully." +
				" Name: " + pstCREndpointSlice.Name +
				", Namespace: " + pstCREndpointSlice.Namespace)
			break
		}
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *PodReconciler) getPodRelationshipInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *corev1.Pod,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	pstToPodInfos *examplecomv1.ToPodInfos,
	c2ndNicAddress string,
	pstCRDeployment *appsv1.Deployment,
	pstCRConnection *examplecomv1.DCIAppConnection,
	cFinalizerName string,
	iFromToSubject *int,
	cEventKind string) (error, bool) {
	var err error
	bProcFlag := false

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" c2ndNicAddress: " + c2ndNicAddress +
		", cEventKind: " + cEventKind)

	for range 1 {
		var bFuncRet bool
		err, bFuncRet = r.getPodConnection(
			ctx,
			log,
			pstCRData,
			pstCRDeployment,
			pstCRConnection)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		} else {
			if len(pstCRData.Labels) == 0 {
				if DCIAppConnectionPodEventKindDelete == cEventKind {
					bFuncRet = false
				}
			}

			// Check Pod Labels
			gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRConnection)
			cLabels := pstCRData.Labels[strings.ToLower(gvks[0].Kind)]

			if cLabels == "" {
				if DCIAppConnectionPodEventKindDelete == cEventKind {
					bFuncRet = false
				}
			} else {
				name, namespace, _ := strings.Cut(cLabels, "_")
				if namespace == "" {
					namespace = NamespaceDefault
				}

				if name != pstCRConnection.Name || namespace != pstCRConnection.Namespace {
					bFuncRet = false
				}
			}

			if !bFuncRet {
				if DCIAppConnectionPodEventKindDelete == cEventKind {
					_ = unsetDciPodConnectionRelationshipFunc(
						ctx,
						r.Client,
						log,
						"",
						cFinalizerName,
						pstCRData)
				}
				break
			}
		}

		iIOMode := examplecomv1.IOModeNONE

		for iInIdx := range pstCRConnection.Status.InputEndpoints {
			if pstCRDeployment.Name ==
				pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Name &&
				pstCRDeployment.Namespace ==
					pstCRConnection.Status.InputEndpoints[iInIdx].Internal.DeploymentRef.Namespace {
				// IOMode is Input Only
				iIOMode = examplecomv1.IOModeIN
				break
			}
		}

		for iOutIdx := range pstCRConnection.Status.OutputEndpoints {
			if pstCRDeployment.Name ==
				pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Name &&
				pstCRDeployment.Namespace ==
					pstCRConnection.Status.OutputEndpoints[iOutIdx].Internal.DeploymentRef.Namespace {
				// IOMode is Input-Output or Output Only
				iIOMode += examplecomv1.IOModeOUT
				break
			}
		}

		bAdditionalProcFlag := true
		if DCIAppConnectionPodEventKindDelete == cEventKind {
			bAdditionalProcFlag = false
		}

		err, bFuncRet = r.getPodConnectRelationshipConn(
			ctx,
			log,
			pstCRData,
			c2ndNicAddress,
			iIOMode,
			pstCRDeployment,
			pstCRConnection,
			bAdditionalProcFlag,
			pstFromPodInfos,
			iFromToSubject)
		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		if iIOMode != examplecomv1.IOModeNONE {
			err, bFuncRet = r.getPodConnectRelationshipInOut(
				ctx,
				log,
				pstCRData,
				c2ndNicAddress,
				iIOMode,
				pstCRDeployment,
				pstCRConnection,
				bAdditionalProcFlag,
				pstFromPodInfos,
				iFromToSubject)
			if err != nil {
				break
			}
			if !bFuncRet {
				break
			}
		}

		err, bFuncRet = r.getPodConnectRelationship(
			ctx,
			log,
			pstCRConnection,
			bAdditionalProcFlag,
			pstFromPodInfos,
			pstToPodInfos)
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
