package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"os"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"
)

// ConnectEndpointReconciler reconciles a ConnectEndpoint object
type ConnectEndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

type ConnectEndpointEventPredicate struct {
	predicate.Funcs
}

// +kubebuilder:rbac:groups=example.com,resources=connectendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints/finalizers,verbs=update
// +kubebuilder:rbac:groups=example.com,resources=dciappstates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=dciappstates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConnectEndpoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *ConnectEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var resultErr error

	// local variable
	log := r.Log.WithValues("connectendpoint", req.Name)
	cMyFuncName := getFuncName()
	stCtrlResult := ctrl.Result{}
	bProcFlag := false
	bExclusiveFlag := false

	for range 1 {
		var cEventKind string // Create, Update, Delete, None
		pstCRData := &examplecomv1.ConnectEndpoint{}

		// Fetch DCIConnectInformation resources
		err = r.Get(ctx, req.NamespacedName, pstCRData)
		if err != nil {
			if apierrors.IsNotFound(err) {
				bProcFlag = true
				err = nil
				log.Error(err, cMyFuncName+" could not find ConnectEndpointCR, but it will not requeue.")
			} else {
				log.Error(err, cMyFuncName+" is unable to fetch the ConnectEndpointCR.")
			}
			break
		}
		log.Info(cMyFuncName + " successfully fetched the ConnectEndpointCR.")

		// Collect information about your own node
		myPodName := os.Getenv(examplecomv1.MyPodName)
		myPodNamespace := os.Getenv(examplecomv1.MyPodNamespace)

		if myPodName != pstCRData.Spec.From.FromPod.PodRef.Name {
			log.Error(err, cMyFuncName+" failed to get my pod name from the environment variable, so it will not requeue.")
			bProcFlag = true
			break
		}

		if myPodNamespace != pstCRData.Spec.From.FromPod.PodRef.Namespace {
			log.Error(err, cMyFuncName+" failed to get my pod namespace from the environment variable, so it will not requeue.")
			bProcFlag = true
			break
		}

		var cFilePath string
		var iFd int
		var eFdErr error
		cFilePath, iFd, eFdErr = fileExclusive(myPodName, myPodNamespace)
		if eFdErr != nil {
			log.Info(cMyFuncName + " could not continue processing due to exclusive state, so it will requeue.")
			bExclusiveFlag = true
			break
		}

		cFinalizerName := r.getFinalizerName(pstCRData)

		// Get the event type
		cEventKind = getEventKind(pstCRData, cFinalizerName)

		bFuncRetFlag := false
		if examplecomv1.Create == cEventKind {
			bFuncRetFlag, err = r.reconcileCreate(
				ctx,
				log,
				pstCRData,
				cFinalizerName)
		} else if examplecomv1.Update == cEventKind {
			bFuncRetFlag, err = r.reconcileUpdate(
				ctx,
				log,
				pstCRData)
		} else if examplecomv1.Delete == cEventKind {
			bFuncRetFlag, err = r.reconcileDelete(
				ctx,
				log,
				pstCRData,
				cFinalizerName)
		} else {
			bFuncRetFlag = true
			// do noting
			log.Info(cMyFuncName + " has no action of this type event for ConnectEndpointCR.")
		}
		if bFuncRetFlag {
			bProcFlag = true
		}

		deferExclusive(log, iFd, cFilePath)
	}

	if err != nil {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = false
		resultErr = err
	} else if bProcFlag {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = false
	} else {
		//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
		stCtrlResult.Requeue = true

		if bExclusiveFlag {
			stCtrlResult.RequeueAfter =
				time.Millisecond * examplecomv1.ExclusiveInterval
		} else {
			stCtrlResult.RequeueAfter =
				time.Millisecond * examplecomv1.ConnectEndpointInterval
		}
	}

	//nolint:staticcheck // Intentionally kept to preserve existing requeue behavior and compatibility with current tests.
	log.Info(cMyFuncName + " return." +
		" stCtrlResult.Requeue: " + strconv.FormatBool(stCtrlResult.Requeue) +
		", stCtrlResult.RequeueAfter: " + stCtrlResult.RequeueAfter.String() +
		", err: " + fmt.Sprint(err))
	return stCtrlResult, resultErr
}

func (ConnectEndpointEventPredicate) Update(e event.UpdateEvent) bool {
	bProcFlag := true

	for range 1 {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			break
		}

		oldObj, okOld := e.ObjectOld.(*examplecomv1.ConnectEndpoint)
		newObj, okNew := e.ObjectNew.(*examplecomv1.ConnectEndpoint)

		if !okOld || !okNew {
			break
		}
		if !newObj.DeletionTimestamp.IsZero() {
			break
		}
		if newObj.Status.Condition != oldObj.Status.Condition {
			break
		}

		bProcFlag = false
	}
	return bProcFlag
}

func (ConnectEndpointEventPredicate) Create(_ event.CreateEvent) bool {
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1.ConnectEndpoint{}).
		WithOptions(NewCustomControllerOptions()).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			ConnectEndpointEventPredicate{},
		)).
		Named("ConnectEndpoint").
		Complete(r)
}

func getEventKind(
	pstCRData *examplecomv1.ConnectEndpoint,
	cFinalizerName string) string {
	eventKind := examplecomv1.None

	if !controllerutil.ContainsFinalizer(pstCRData, cFinalizerName) {
		if !pstCRData.DeletionTimestamp.IsZero() {
			// do nothing
		} else {
			eventKind = examplecomv1.Create
		}
	} else {
		if !pstCRData.DeletionTimestamp.IsZero() {
			eventKind = examplecomv1.Delete
		} else {
			eventKind = examplecomv1.Update
		}
	}
	return eventKind
}

func (r *ConnectEndpointReconciler) getFinalizerName(
	pstCRData *examplecomv1.ConnectEndpoint) string {
	// Value to set in the finalizer
	gvks, _, _ := r.Client.Scheme().ObjectKinds(pstCRData)
	return strings.ToLower(gvks[0].Kind) + ".finalizers." +
		strings.ToLower(gvks[0].Group+"/"+gvks[0].Version)
}

func (r *ConnectEndpointReconciler) getDCIAppStateInfo(
	ctx context.Context,
	log logr.Logger,
	name string,
	namespace string) (examplecomv1.DCIAppState, error) {
	var err error
	var dciAppStateCR examplecomv1.DCIAppState

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" DCIAppStateName: " + name +
		", DCIAppStateNamespace: " + namespace)

	for range 1 {
		dciAppStateCRData := &examplecomv1.DCIAppState{}
		err = r.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name}, dciAppStateCRData)
		if apierrors.IsNotFound(err) {
			log.Error(nil, cMyFuncName+" could not find DCIAppStateCR."+
				" Name: "+name+
				", Namespace: "+namespace)
		} else if err != nil {
			log.Error(err, cMyFuncName+" is unable to fetch DCIAppStateCR."+
				" Name: "+name+
				", Namespace: "+namespace)
			break
		} else {
			dciAppStateCR = *dciAppStateCRData.DeepCopy()
		}
	}
	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return dciAppStateCR, err
}

func checkStatusToPodUpdate(
	log logr.Logger,
	bfrToPodData []examplecomv1.ToPodStatus,
	aftToPodData []examplecomv1.ToPodStatus) bool {
	changeFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		if len(bfrToPodData) != len(aftToPodData) {
			log.Info(cMyFuncName + " detected a change in the number of elements in the ToPod area array" +
				" of the Status of the ConnectEndpointCR.")
			changeFlag = true
			break
		}
		for _, bfrPodValue := range bfrToPodData {
			ExistPodFlag := false
			for _, aftPodValue := range aftToPodData {
				if bfrPodValue.PodRef != aftPodValue.PodRef {
					continue
				}
				ExistPodFlag = true

				if bfrPodValue.Address != aftPodValue.Address {
					changeFlag = true
					log.Info(cMyFuncName + " detected a change in the address field in the ToPod area" +
						" of the Status of the ConnectEndpointCR." +
						" name: " + bfrPodValue.PodRef.Name +
						", namespace: " + bfrPodValue.PodRef.Namespace)
					break
				}

				if bfrPodValue.AppStatus.Raw == nil &&
					aftPodValue.AppStatus.Raw == nil {
					break
				}
				if bfrPodValue.AppStatus.Raw != nil &&
					aftPodValue.AppStatus.Raw != nil {
					break
				}
				changeFlag = true
				log.Info(cMyFuncName + " detected a change in either of the appStatuses in the ToPod area" +
					" of the Status of the ConnectEndpointCR." +
					" name: " + bfrPodValue.PodRef.Name +
					", namespace: " + bfrPodValue.PodRef.Namespace)
				break
			}
			if !ExistPodFlag {
				changeFlag = true
				log.Info(cMyFuncName + " determined that it does not match the podRef.name in the To area" +
					" of the Status of the ConnectEndpointCR.")
			}
		}
		break
	}
	log.Info(cMyFuncName + " End." +
		" changeFlag: " + strconv.FormatBool(changeFlag))
	return changeFlag
}

func checkStatusToUpdate(
	log logr.Logger,
	bfrToData []examplecomv1.ToDeploymentStatus,
	aftToData []examplecomv1.ToDeploymentStatus) bool {
	changeFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		if len(bfrToData) != len(aftToData) {
			changeFlag = true
			log.Info(cMyFuncName + " detected a change in the number of elements in the To area array" +
				" of the Status of the ConnectEndpointCR.")
			break
		}

		ExistFlag := false
		for _, bfrValue := range bfrToData {
			for _, aftValue := range aftToData {
				if bfrValue.DeploymentRef != aftValue.DeploymentRef {
					continue
				}
				ExistFlag = true

				changePodFlag := checkStatusToPodUpdate(
					log,
					bfrValue.ToPod,
					aftValue.ToPod)
				if changePodFlag {
					changeFlag = true
				}
				break
			}
		}
		if !ExistFlag {
			changeFlag = true
			log.Info(cMyFuncName + " determined that it does not match the deploymentRef.name in the To area" +
				" of the Status of the ConnectEndpointCR.")
		}
		break
	}
	log.Info(cMyFuncName + " End." +
		" changeFlag: " + strconv.FormatBool(changeFlag))
	return changeFlag
}

func checkExternalEndpointUpdate(
	log logr.Logger,
	bfrData *examplecomv1.ExternalEndpoint,
	aftData *examplecomv1.ExternalEndpoint) bool {
	var changeFlag bool

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	if bfrData == nil && aftData == nil {
		// do nothing

	} else if bfrData != nil && aftData != nil {
		if len(bfrData.ExternalSubsets) != len(aftData.ExternalSubsets) {
			changeFlag = true
			log.Info(cMyFuncName + " detected a change in the number of elements in the ExternalSubsets area array" +
				" of the Status of the ConnectEndpointCR.")
		} else {
			ExistFlag := false
			for _, bfrValue := range bfrData.ExternalSubsets {
				for _, aftValue := range aftData.ExternalSubsets {
					if bfrValue == aftValue {
						ExistFlag = true
						break
					}
				}
				if ExistFlag {
					break
				}
			}

			if !ExistFlag {
				changeFlag = true
				log.Info(cMyFuncName + " detected a change in the ExternalSubsets area" +
					" of the Status of the ConnectEndpointCR.")
			}
		}
	} else {
		changeFlag = true
	}

	log.Info(cMyFuncName + " End." +
		" changeFlag: " + strconv.FormatBool(changeFlag))
	return changeFlag
}

func checkStatusUpdate(
	log logr.Logger,
	bfrData *examplecomv1.ConnectEndpoint,
	aftData *examplecomv1.ConnectEndpoint) bool {
	var changeFlag bool

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	if bfrData.Status.From != aftData.Status.From {
		changeFlag = true
		log.Info(cMyFuncName + " detected a change in the From area of the Status of the ConnectEndpointCR.")
	}

	if !changeFlag {
		changeFlag = checkStatusToUpdate(
			log,
			bfrData.Status.To,
			aftData.Status.To)
	}

	if !changeFlag {
		changeFlag = checkExternalEndpointUpdate(
			log,
			bfrData.Status.InputEndpoint,
			aftData.Status.InputEndpoint)
	}

	if !changeFlag {
		changeFlag = checkExternalEndpointUpdate(
			log,
			bfrData.Status.OutputEndpoint,
			aftData.Status.OutputEndpoint)
	}

	log.Info(cMyFuncName + " End." +
		" changeFlag: " + strconv.FormatBool(changeFlag))
	return changeFlag
}

func (r *ConnectEndpointReconciler) addChkDCIAppStateInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	cPodName string,
	cPodNamespace string) (examplecomv1.DCIAppState, bool, error) {
	bNotFoundFlag := false
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" DCIAppStateName: " + cPodName +
		", DCIAppStateNamespace: " + cPodNamespace)

	var updateDCIAppStateCR examplecomv1.DCIAppState

	updateDCIAppStateCR, err = r.getDCIAppStateInfo(
		ctx,
		log,
		cPodName,
		cPodNamespace)
	if apierrors.IsNotFound(err) {
		bNotFoundFlag = true
		err = nil
	} else if err == nil {
		if !updateDCIAppStateCR.DeletionTimestamp.IsZero() {
			bNotFoundFlag = true
			log.Error(err, cMyFuncName+" found a DeletionTimestamp on the DCIAppStateCR."+
				" Name: "+cPodName+
				", Namespace: "+cPodNamespace)
		} else if updateDCIAppStateCR.Labels != nil {
			_, exists := updateDCIAppStateCR.Labels[pstCRData.Name]
			if exists {
				if updateDCIAppStateCR.Labels[pstCRData.Name] !=
					pstCRData.Namespace {
					bNotFoundFlag = true
					log.Error(err, cMyFuncName+" found that the DCIAppStateCR. labels did not match."+
						" Labels["+pstCRData.Name+"]: "+
						updateDCIAppStateCR.Labels[pstCRData.Name]+
						", CheckLabel: "+pstCRData.Namespace+
						", Name: "+cPodName+
						", Namespace: "+cPodNamespace)
				} else {
					log.Info(cMyFuncName + " found that the DCIAppStateCR labels matched, it is normal." +
						" Labels[" + pstCRData.Name + "]: " +
						updateDCIAppStateCR.Labels[pstCRData.Name] +
						", Name: " + cPodName +
						", Namespace: " + cPodNamespace)
				}
			} else {
				log.Info(cMyFuncName + " could not find key that the DCIAppStateCR labels, it is normal." +
					" LabelsKey: " + pstCRData.Name +
					", Name: " + cPodName +
					", Namespace: " + cPodNamespace)
			}
		} else {
			log.Info(cMyFuncName + " found that the DCIAppStateCR had no labels, it is normal." +
				" Name: " + cPodName +
				", Namespace: " + cPodNamespace)
		}
	}

	log.Info(cMyFuncName + " End." +
		" bNotFoundFlag: " + strconv.FormatBool(bNotFoundFlag) +
		", err: " + fmt.Sprint(err))
	return updateDCIAppStateCR, bNotFoundFlag, err
}

func (r *ConnectEndpointReconciler) getChangePodInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	specTo *examplecomv1.ToDeployment,
	deleteIndexSlice *[][]int,
	deleteIndex int,
	podDeleteIndexSlice *[]int,
	updateDCIAppStateCRList *[]examplecomv1.DCIAppState) ([]examplecomv1.ToPodStatus, bool, error) {
	notFoundFlag := false
	var err error
	var stToPodStatusSlice []examplecomv1.ToPodStatus

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for _, specToPod := range specTo.ToPod {
		log.Info(cMyFuncName + " Pod Spec Check." +
			" specToPod.PodRef: " + specToPod.PodRef.Name +
			"," + specToPod.PodRef.Namespace)

		var addToPodStatus examplecomv1.ToPodStatus
		addToPodStatus.PodRef = specToPod.PodRef
		addToPodStatus.Address = specToPod.Address

		for deletePodIndex, statusPodIndex := range *podDeleteIndexSlice {
			log.Info(cMyFuncName + " Pod Status Check." +
				" statusToPod.PodRef: " +
				pstCRData.Status.To[deleteIndex].ToPod[statusPodIndex].PodRef.Name +
				"," + pstCRData.Status.To[deleteIndex].ToPod[statusPodIndex].PodRef.Namespace)

			if specToPod.PodRef ==
				pstCRData.Status.To[deleteIndex].ToPod[statusPodIndex].PodRef {
				(*deleteIndexSlice)[deleteIndex] =
					append(((*deleteIndexSlice)[deleteIndex])[:deletePodIndex],
						((*deleteIndexSlice)[deleteIndex])[deletePodIndex+1:]...)
				if pstCRData.Status.To[deleteIndex].ToPod[statusPodIndex].AppStatus.Raw != nil {
					addToPodStatus.AppStatus =
						pstCRData.Status.To[deleteIndex].ToPod[statusPodIndex].AppStatus
				}
				log.Info(cMyFuncName + " retrieves the already configured AppState." +
					" AppState: " + string(addToPodStatus.AppStatus.Raw))
				break
			}
		}

		if addToPodStatus.AppStatus.Raw == nil {
			var updateDCIAppStateCR examplecomv1.DCIAppState
			var bNotFoundFlag bool
			updateDCIAppStateCR, bNotFoundFlag, err = r.addChkDCIAppStateInfo(
				ctx,
				log,
				pstCRData,
				specToPod.PodRef.Name,
				specToPod.PodRef.Namespace)
			if err != nil {
				break
			} else if bNotFoundFlag {
				notFoundFlag = true
			} else {
				*updateDCIAppStateCRList =
					append(*updateDCIAppStateCRList, updateDCIAppStateCR)
				log.Info(cMyFuncName + " retrieves AppState from the new resource." +
					" AppState: " + string(updateDCIAppStateCR.Status.AppStatus.Raw))
				if updateDCIAppStateCR.Status.AppStatus.Raw == nil {
					notFoundFlag = true
				} else {
					addToPodStatus.AppStatus = updateDCIAppStateCR.Status.AppStatus
				}
			}
		}
		stToPodStatusSlice = append(stToPodStatusSlice, addToPodStatus)
	}
	log.Info(cMyFuncName + " End." +
		" notFoundFlag: " + strconv.FormatBool(notFoundFlag) +
		", err: " + fmt.Sprint(err))
	return stToPodStatusSlice, notFoundFlag, err
}

func (r *ConnectEndpointReconciler) getChangeDeploymentInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	specTo *examplecomv1.ToDeployment,
	updateDCIAppStateCRList *[]examplecomv1.DCIAppState,
	updConnectEndpointStatus *examplecomv1.ConnectEndpointStatus) (bool, error) {
	notFoundFlag := false
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	var addToStatus examplecomv1.ToDeploymentStatus
	addToStatus.DeploymentRef = specTo.DeploymentRef

	for _, specToPod := range specTo.ToPod {
		log.Info(cMyFuncName + " Pod Spec Check." +
			" specToPod.PodRef: " + specToPod.PodRef.Name +
			"," + specToPod.PodRef.Namespace)

		var addToPodStatus examplecomv1.ToPodStatus
		addToPodStatus.PodRef = specToPod.PodRef
		addToPodStatus.Address = specToPod.Address

		var updateDCIAppStateCR examplecomv1.DCIAppState
		var bNotFoundFlag bool
		updateDCIAppStateCR, bNotFoundFlag, err = r.addChkDCIAppStateInfo(
			ctx,
			log,
			pstCRData,
			specToPod.PodRef.Name,
			specToPod.PodRef.Namespace)
		if err != nil {
			break
		} else if bNotFoundFlag {
			notFoundFlag = true
		} else {
			*updateDCIAppStateCRList =
				append(*updateDCIAppStateCRList, updateDCIAppStateCR)
			log.Info(cMyFuncName + " retrieves the already configured AppState." +
				" AppState: " + string(updateDCIAppStateCR.Status.AppStatus.Raw))
			if updateDCIAppStateCR.Status.AppStatus.Raw == nil {
				notFoundFlag = true
			} else {
				addToPodStatus.AppStatus = updateDCIAppStateCR.Status.AppStatus
			}
		}
		addToStatus.ToPod =
			append(addToStatus.ToPod, addToPodStatus)
	}

	updConnectEndpointStatus.To =
		append(updConnectEndpointStatus.To,
			addToStatus)

	log.Info(cMyFuncName + " End." +
		" notFoundFlag: " + strconv.FormatBool(notFoundFlag) +
		", err: " + fmt.Sprint(err))
	return notFoundFlag, err
}

func (r *ConnectEndpointReconciler) getChangeToInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	deleteIndexSlice *[][]int,
	updateDCIAppStateCRList *[]examplecomv1.DCIAppState,
	updConnectEndpointStatus *examplecomv1.ConnectEndpointStatus) (bool, error) {
	notFoundFlag := false
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		for _, specTo := range pstCRData.Spec.To {
			log.Info(cMyFuncName + " Deployment Spec Check." +
				" specTo.DeploymentRef: " + specTo.DeploymentRef.Name +
				"," + specTo.DeploymentRef.Namespace)

			deploymentMatchFlag := false
			for deleteIndex, podDeleteIndexSlice := range *deleteIndexSlice {
				log.Info(cMyFuncName + " Deployment Status Check." +
					" statusTo.DeploymentRef: " +
					pstCRData.Status.To[deleteIndex].DeploymentRef.Name +
					"," + pstCRData.Status.To[deleteIndex].DeploymentRef.Namespace)

				if specTo.DeploymentRef ==
					pstCRData.Status.To[deleteIndex].DeploymentRef {
					deploymentMatchFlag = true

					var stToPodStatusSlice []examplecomv1.ToPodStatus
					var bNotFoundFlag bool
					stToPodStatusSlice, bNotFoundFlag, err =
						r.getChangePodInfo(
							ctx,
							log,
							pstCRData,
							&specTo,
							deleteIndexSlice,
							deleteIndex,
							&podDeleteIndexSlice,
							updateDCIAppStateCRList)
					if err != nil {
						break
					} else if bNotFoundFlag {
						notFoundFlag = true
					}
					var stToStatus examplecomv1.ToDeploymentStatus
					stToStatus.DeploymentRef = specTo.DeploymentRef
					stToStatus.ToPod = stToPodStatusSlice
					updConnectEndpointStatus.To =
						append(updConnectEndpointStatus.To, stToStatus)
				}
			}
			if err != nil {
				break
			}
			if !deploymentMatchFlag {
				log.Info(cMyFuncName + " could not find a configured Deployment that matches the new Deployment. A new Deployment will be added.")

				var bNotFoundFlag bool
				bNotFoundFlag, err = r.getChangeDeploymentInfo(
					ctx,
					log,
					pstCRData,
					&specTo,
					updateDCIAppStateCRList,
					updConnectEndpointStatus)
				if err != nil {
					break
				} else if bNotFoundFlag {
					notFoundFlag = true
				}
			}
		}
	}

	log.Info(cMyFuncName + " End." +
		" notFoundFlag: " + strconv.FormatBool(notFoundFlag) +
		", err: " + fmt.Sprint(err))
	return notFoundFlag, err
}

func (r *ConnectEndpointReconciler) getChangeInfo(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	updateDCIAppStateCRList *[]examplecomv1.DCIAppState,
	deleteDCIAppStateCRList *[]examplecomv1.DCIAppState,
	updConnectEndpointStatus *examplecomv1.ConnectEndpointStatus) (bProcFlag, notFoundFlag bool, err error) {
	bProcFlag = true
	notFoundFlag = false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	// local variable
	var deleteIndexSlice [][]int

	for _, to := range pstCRData.Status.To {
		var podDeleteIndexSlice []int
		for indexPod := range to.ToPod {
			podDeleteIndexSlice = append(podDeleteIndexSlice, indexPod)
		}
		deleteIndexSlice = append(deleteIndexSlice, podDeleteIndexSlice)
	}

	var bNotFoundFlag bool
	bNotFoundFlag, err = r.getChangeToInfo(
		ctx,
		log,
		pstCRData,
		&deleteIndexSlice,
		updateDCIAppStateCRList,
		updConnectEndpointStatus)
	if err != nil {
		bProcFlag = false
	} else if bNotFoundFlag {
		notFoundFlag = true
	}

	if bProcFlag {
		for deleteIndex, value := range deleteIndexSlice {
			for _, podvalue := range value {
				var deleteDCIAppStateCR examplecomv1.DCIAppState

				deleteDCIAppStateCR, err = r.getDCIAppStateInfo(
					ctx,
					log,
					pstCRData.Status.To[deleteIndex].ToPod[podvalue].PodRef.Name,
					pstCRData.Status.To[deleteIndex].ToPod[podvalue].PodRef.Namespace)
				if apierrors.IsNotFound(err) {
					err = nil
				} else if err != nil {
					bProcFlag = false
					break
				} else {
					*deleteDCIAppStateCRList =
						append(*deleteDCIAppStateCRList, deleteDCIAppStateCR)
				}
			}
		}
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", notFoundFlag: " + strconv.FormatBool(notFoundFlag) +
		", err: " + fmt.Sprint(err))
	return bProcFlag, notFoundFlag, err
}

func (r *ConnectEndpointReconciler) delAppStateCRLabel(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	pstDCIAppStateCRData *examplecomv1.DCIAppState) error {
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	if pstDCIAppStateCRData.Labels != nil {
		_, exists := pstDCIAppStateCRData.Labels[pstCRData.Name]
		if exists {
			if pstDCIAppStateCRData.Labels[pstCRData.Name] ==
				pstCRData.Namespace {
				pstBaseDciAppStateCRDData :=
					client.MergeFrom(pstDCIAppStateCRData.DeepCopy())

				log.Info(cMyFuncName + " delete a key from the labels of DCIAppStateCR." +
					" Name: " + pstDCIAppStateCRData.Name +
					", Namespace: " + pstDCIAppStateCRData.Namespace)
				delete(pstDCIAppStateCRData.Labels, pstCRData.Name)

				err = r.Patch(ctx, pstDCIAppStateCRData, pstBaseDciAppStateCRDData)
				if err != nil {
					log.Error(err, cMyFuncName+" failed to update the DCIAppStateCR."+
						" Name: "+pstDCIAppStateCRData.Name+
						", Namespace: "+pstDCIAppStateCRData.Namespace)
				}
			} else {
				log.Error(err, cMyFuncName+" found that the DCIAppStateCR. labels did not match."+
					" Labels["+pstCRData.Name+"]: "+
					pstDCIAppStateCRData.Labels[pstCRData.Name]+
					", CheckLabel: "+pstCRData.Namespace+
					", Name: "+pstDCIAppStateCRData.Name+
					", Namespace: "+pstDCIAppStateCRData.Namespace)
			}
		} else {
			log.Info(cMyFuncName + " could not find key that the DCIAppStateCR labels, it is normal." +
				" LabelsKey: " + pstCRData.Name +
				", Name: " + pstDCIAppStateCRData.Name +
				", Namespace: " + pstDCIAppStateCRData.Namespace)
		}
	} else {
		log.Info(cMyFuncName + " found that the DCIAppStateCR had no labels, it is normal." +
			" Name: " + pstDCIAppStateCRData.Name +
			", Namespace: " + pstDCIAppStateCRData.Namespace)
	}

	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func (r *ConnectEndpointReconciler) updAppStateCRList(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	pstBeforeCRData *examplecomv1.ConnectEndpoint,
	updateDCIAppStateCRList []examplecomv1.DCIAppState,
	deleteDCIAppStateCRList []examplecomv1.DCIAppState) error {
	var errSave error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")
	// local variable
	var err error

	for updateIndex := range updateDCIAppStateCRList {
		dciAppStateCRData := &updateDCIAppStateCRList[updateIndex]

		if dciAppStateCRData.DeletionTimestamp.IsZero() {
			pstBaseDciAppStateCRDData :=
				client.MergeFrom(dciAppStateCRData.DeepCopy())

			if dciAppStateCRData.Labels == nil {
				dciAppStateCRData.Labels = make(map[string]string)
			}
			_, exists := dciAppStateCRData.Labels[pstCRData.Name]
			if !exists {
				dciAppStateCRData.Labels[pstCRData.Name] =
					pstCRData.Namespace

				log.Info(cMyFuncName + " added label to the DCIAppStateCR." +
					" Labels[" + pstCRData.Name + "]: " + pstCRData.Namespace +
					", Name: " + dciAppStateCRData.Name +
					", Namespace: " + dciAppStateCRData.Namespace)

				err = r.Patch(
					ctx,
					dciAppStateCRData,
					pstBaseDciAppStateCRDData)
				if err != nil {
					if errSave == nil {
						errSave = err
					}
					log.Error(err, cMyFuncName+" failed to update the DCIAppStateCR."+
						" Name: "+dciAppStateCRData.Name+
						", Namespace: "+dciAppStateCRData.Namespace)
				}
			} else {
				log.Info(cMyFuncName + " found that the label already existed on the DCIAppStateCR." +
					" Label[" + pstCRData.Name + "]: " +
					dciAppStateCRData.Labels[pstCRData.Name] +
					", Name: " + dciAppStateCRData.Name +
					", Namespace: " + dciAppStateCRData.Namespace)
			}
		}
	}

	for deleteIndex := range deleteDCIAppStateCRList {
		dciAppStateCRData := &deleteDCIAppStateCRList[deleteIndex]

		err = r.delAppStateCRLabel(
			ctx,
			log,
			pstCRData,
			dciAppStateCRData)
		if err != nil {
			if errSave == nil {
				errSave = err
			}

			// Restoration of deletion status information
			stStatusTo := examplecomv1.ToDeploymentStatus{}
			for iBfrToIndex := range pstBeforeCRData.Status.To {
				for iBfrToPodIndex := range pstBeforeCRData.Status.To[iBfrToIndex].ToPod {
					stToPod := pstBeforeCRData.Status.To[iBfrToIndex].ToPod[iBfrToPodIndex]
					if stToPod.PodRef.Name == dciAppStateCRData.Name &&
						stToPod.PodRef.Namespace == dciAppStateCRData.Namespace {
						stStatusTo.DeploymentRef =
							pstBeforeCRData.Status.To[iBfrToIndex].DeploymentRef
						stStatusTo.ToPod = append(stStatusTo.ToPod, stToPod)

						bDeploymentHitFlag := false
						for iaftToIndex := range pstCRData.Status.To {
							if pstCRData.Status.To[iaftToIndex].DeploymentRef ==
								stStatusTo.DeploymentRef {
								pstCRData.Status.To[iaftToIndex].ToPod =
									append(pstCRData.Status.To[iaftToIndex].ToPod,
										stStatusTo.ToPod[0])
								bDeploymentHitFlag = true
								break
							}
						}
						if !bDeploymentHitFlag {
							pstCRData.Status.To =
								append(pstCRData.Status.To, stStatusTo)
						}
						break
					}
				}
			}
		}
	}

	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(errSave))
	return errSave
}

func (r *ConnectEndpointReconciler) delAppStateCRList(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint) error {
	var errSave error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for _, to := range pstCRData.Status.To {
		dciAppStateCRData := &examplecomv1.DCIAppState{}
		for _, toPod := range to.ToPod {
			err := r.Get(ctx, client.ObjectKey{
				Namespace: toPod.PodRef.Namespace,
				Name:      toPod.PodRef.Name}, dciAppStateCRData)
			if apierrors.IsNotFound(err) {
				log.Info(cMyFuncName + " could not find DCIAppStateCR and failed to retrieve it." +
					" Name: " + toPod.PodRef.Name +
					", Namespace: " + toPod.PodRef.Namespace)
			} else if err != nil {
				log.Error(err, cMyFuncName+" is unable to fetch DCIAppStateCR.")
				if errSave == nil {
					errSave = err
				}
			} else {
				err = r.delAppStateCRLabel(
					ctx,
					log,
					pstCRData,
					dciAppStateCRData)
				if errSave == nil {
					errSave = err
				}
			}
		}
	}
	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(errSave))
	return errSave
}

func (r *ConnectEndpointReconciler) updateProcess(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	pstBeforeCRData *examplecomv1.ConnectEndpoint,
	updateDCIAppStateCRList []examplecomv1.DCIAppState,
	deleteDCIAppStateCRList []examplecomv1.DCIAppState,
	notFoundFlag bool) (bool, error) {
	var errSave error
	bProcFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	errSave = r.updAppStateCRList(
		ctx,
		log,
		pstCRData,
		pstBeforeCRData,
		updateDCIAppStateCRList,
		deleteDCIAppStateCRList)

	if examplecomv1.Configured == pstCRData.Status.Condition ||
		(examplecomv1.NotConfigured == pstCRData.Status.Condition &&
			!notFoundFlag) {
		var bFuncRetFlag bool
		var err error

		if errSave == nil {
			bFuncRetFlag, err = createConfigFileFunc(
				log,
				pstCRData,
				examplecomv1.Create)
		} else {
			bFuncRetFlag, err = createConfigFileFunc(
				log,
				pstCRData,
				examplecomv1.Delete)
		}
		if err != nil {
			if errSave == nil {
				errSave = err
			}
		} else if bFuncRetFlag {
			bFuncRetFlag, err = sendSignalFunc(log)
			if errSave == nil {
				if err != nil {
					errSave = err
				} else if bFuncRetFlag {
					bProcFlag = true
				}
			}
		}
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(errSave))
	return bProcFlag, errSave
}

func (r *ConnectEndpointReconciler) reconcileCreate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	cFinalizerName string) (bool, error) {
	var err error
	bProcFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	pstBaseCRDData := client.MergeFrom(pstCRData.DeepCopy())

	controllerutil.AddFinalizer(pstCRData, cFinalizerName)
	log.Info(cMyFuncName + " added a finalizer to the ConnectEndpointCR.")

	pstCRData.Status.Condition = examplecomv1.NotConfigured
	log.Info(cMyFuncName + " changed the condition in the status area of the ConnectEndpointCR to NotConfigured.")

	err = r.Patch(ctx, pstCRData, pstBaseCRDData)
	if err != nil {
		log.Error(err, cMyFuncName+" failed to update the ConnectEndpointCR.")
	} else {
		bProcFlag = true
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return bProcFlag, err
}

func (r *ConnectEndpointReconciler) reconcileUpdate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint) (bool, error) {
	var errSave error
	bProcFlag := false

	// local variable
	notFoundFlag := false
	var updateDCIAppStateCRList []examplecomv1.DCIAppState
	var deleteDCIAppStateCRList []examplecomv1.DCIAppState
	var updConnectEndpointStatus examplecomv1.ConnectEndpointStatus
	var bFuncRetFlag bool
	var statusChangeFlag bool

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		bFuncRetFlag, notFoundFlag, errSave = r.getChangeInfo(
			ctx,
			log,
			pstCRData,
			&updateDCIAppStateCRList,
			&deleteDCIAppStateCRList,
			&updConnectEndpointStatus)
		if !bFuncRetFlag {
			break
		}

		stBeforeCRData := pstCRData.DeepCopy()

		pstCRData.Status.From = pstCRData.Spec.From
		pstCRData.Status.To = updConnectEndpointStatus.To
		pstCRData.Status.InputEndpoint = pstCRData.Spec.InputEndpoint
		pstCRData.Status.OutputEndpoint = pstCRData.Spec.OutputEndpoint

		statusChangeFlag = checkStatusUpdate(
			log,
			stBeforeCRData,
			pstCRData)

		if !statusChangeFlag {
			if examplecomv1.Configured == pstCRData.Status.Condition {
				if !notFoundFlag {
					bProcFlag = true
				}
				break
			}
		}

		var err error
		var bFuncRetFlag bool
		bFuncRetFlag, err = r.updateProcess(
			ctx,
			log,
			pstCRData,
			stBeforeCRData,
			updateDCIAppStateCRList,
			deleteDCIAppStateCRList,
			notFoundFlag)
		if errSave == nil {
			errSave = err
		}

		if bFuncRetFlag {
			if examplecomv1.Configured != pstCRData.Status.Condition {
				pstCRData.Status.Condition = examplecomv1.Configured
				statusChangeFlag = true
			}
		} else {
			if examplecomv1.Configured == pstCRData.Status.Condition {
				pstCRData.Status.Condition = examplecomv1.NotConfigured
				statusChangeFlag = true
			}
		}

		if statusChangeFlag {
			pstBaseCRDData := client.MergeFrom(stBeforeCRData.DeepCopy())
			err = r.Patch(ctx, pstCRData, pstBaseCRDData)
			if err != nil {
				log.Error(err, cMyFuncName+" failed to update the ConnectEndpointCR.")
			}
			if errSave == nil {
				errSave = err
			}
		}

		if errSave == nil {
			if examplecomv1.Configured != pstCRData.Status.Condition &&
				!notFoundFlag {
				bProcFlag = true
				log.Info(cMyFuncName + " found that condition in the status area of the ConnectEndpointCR was NotConfigured.")
			} else if examplecomv1.Configured == pstCRData.Status.Condition &&
				!notFoundFlag {
				log.Info(cMyFuncName + " found that condition in the status area of the ConnectEndpointCR was Configured.")
				bProcFlag = true
			}
		}
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(errSave))
	return bProcFlag, errSave
}

func (r *ConnectEndpointReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	cFinalizerName string) (bool, error) {
	var err error
	bProcFlag := false

	// local variable
	var errSave error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		var pstBaseCRDData client.Patch
		bConfiguredFlag := false
		var bFuncRetFlag bool
		if examplecomv1.Configured == pstCRData.Status.Condition {
			bConfiguredFlag = true
			bFuncRetFlag, err = createConfigFileFunc(
				log,
				pstCRData,
				examplecomv1.Delete)

			if !bFuncRetFlag {
				errSave = err
				break
			}

			_, _ = sendSignalFunc(log)

			pstBaseCRDData = client.MergeFrom(pstCRData.DeepCopy())

			pstCRData.Status.Condition = examplecomv1.NotConfigured
			log.Info(cMyFuncName + " changed the condition in the status area of the ConnectEndpointCR to NotConfigured.")
		}
		if examplecomv1.NotConfigured == pstCRData.Status.Condition {
			err = r.delAppStateCRList(ctx, log, pstCRData)
			if err != nil {
				if errSave == nil {
					errSave = err
				}
				break
			}

			if !bConfiguredFlag {
				pstBaseCRDData = client.MergeFrom(pstCRData.DeepCopy())
			}

			controllerutil.RemoveFinalizer(pstCRData, cFinalizerName)
			log.Info(cMyFuncName + " deletes the finalizer from ConnectEndpointCR.")

			err = r.Patch(ctx, pstCRData, pstBaseCRDData)
			if err != nil {
				errSave = err
				log.Error(err, cMyFuncName+" failed to update the ConnectEndpointCR.")
			} else {
				bProcFlag = true
			}
		}
	}

	err = errSave
	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return bProcFlag, err
}
