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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"bytes"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"
)

// DCIAppStateReconciler reconciles a DCIAppState object
type DCIAppStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

type DCIAppStateEventPredicate struct {
	predicate.Funcs
}

// GetDCIAppStateEventKind is a package-level variable so tests can override
// the behavior of determining the event kind. It defaults to the real
// getDCIAppStateEventKind function.
var GetDCIAppStateEventKind = getDCIAppStateEventKind

// +kubebuilder:rbac:groups=example.com,resources=dciappstates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=dciappstates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=dciappstates/finalizers,verbs=update
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=connectendpoints/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DCIAppState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DCIAppStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	stCtrlResult := ctrl.Result{}
	var resultErr error
	var err error

	// local variable
	log := r.Log.WithValues("dciappstate", req.Name)
	cMyFuncName := getFuncName()
	bExclusiveFlag := false

	for range 1 {
		var cEventKind string // Create, Update, Delete, None
		pstCRData := &examplecomv1.DCIAppState{}

		// Fetch DCIAppState resources
		err = r.Get(ctx, req.NamespacedName, pstCRData)
		if err != nil {
			if apierrors.IsNotFound(err) {
				err = nil
				log.Error(err, cMyFuncName+" could not find DCIAppStateCR, but it will not requeue.")
			} else {
				log.Error(err, cMyFuncName+" is unable to fetch the DCIAppStateCR.")
			}
			break
		}
		log.Info(cMyFuncName + " successfully fetched the DCIAppStateCR.")

		// Collect information about your own node
		myPodName := os.Getenv(examplecomv1.MyPodName)
		myPodNamespace := os.Getenv(examplecomv1.MyPodNamespace)

		_, exists := pstCRData.Labels[myPodName]
		if !exists {
			log.Info(cMyFuncName + " could not find key that the DCIAppStateCR labels, it is normal." +
				" LabelsKey: " + myPodName)
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

		// Get the event type (overridable in tests via GetDCIAppStateEventKind)
		cEventKind = GetDCIAppStateEventKind(pstCRData)

		if examplecomv1.Update == cEventKind {
			err = r.reconcileUpdate(
				ctx,
				log,
				pstCRData,
				myPodName)
		} else {
			// do noting
			log.Info(cMyFuncName + " has no action of this type event for DCIAppStateCR.")
		}
		deferExclusive(log, iFd, cFilePath)
	}

	if err != nil {
		stCtrlResult.Requeue = false
		resultErr = err
	} else if bExclusiveFlag {
		stCtrlResult.Requeue = true
		stCtrlResult.RequeueAfter =
			time.Millisecond * examplecomv1.ExclusiveInterval
	} else {
		stCtrlResult.Requeue = false
	}

	log.Info(cMyFuncName + " return." +
		" stCtrlResult.Requeue: " + strconv.FormatBool(stCtrlResult.Requeue) +
		", stCtrlResult.RequeueAfter: " + stCtrlResult.RequeueAfter.String() +
		", err: " + fmt.Sprint(err))
	return stCtrlResult, resultErr
}

func (DCIAppStateEventPredicate) Update(e event.UpdateEvent) bool {
	bProcFlag := true

	for range 1 {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			break
		}

		oldObj, okOld := e.ObjectOld.(*examplecomv1.DCIAppState)
		newObj, okNew := e.ObjectNew.(*examplecomv1.DCIAppState)

		if !okOld || !okNew {
			break
		}
		// treat nil vs empty slice as different (statusChangeCheck treats nil vs non-nil as change)
		oldRaw := oldObj.Status.AppStatus.Raw
		newRaw := newObj.Status.AppStatus.Raw

		equal := false
		if oldRaw == nil && newRaw == nil {
			equal = true
		} else if oldRaw != nil && newRaw != nil && bytes.Equal(newRaw, oldRaw) {
			equal = true
		}

		if !equal {
			// not equal -> process the event
			break
		}

		// equal -> no processing needed
		bProcFlag = false
	}
	return bProcFlag
}

// SetupWithManager sets up the controller with the Manager.
func (r *DCIAppStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1.DCIAppState{}).
		WithOptions(NewCustomControllerOptions()).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			DCIAppStateEventPredicate{},
		)).
		Named("DCIAppState").
		Complete(r)
}

func getDCIAppStateEventKind(pstCRData *examplecomv1.DCIAppState) string {
	var eventKind string

	if !pstCRData.DeletionTimestamp.IsZero() {
		eventKind = examplecomv1.Delete
	} else {
		eventKind = examplecomv1.Update
	}
	return eventKind
}

func statusChangeCheck(
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppState,
	connectEndpointCRData *examplecomv1.ConnectEndpoint) (changeFlag bool, updateIndex, updateIndexToPod int) {
	updateIndex = 0
	updateIndexToPod = 0
	changeFlag = false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" ConnectEndpoint: " + connectEndpointCRData.Name + ", " + connectEndpointCRData.Namespace)

	if examplecomv1.Configured == connectEndpointCRData.Status.Condition {
		for index, to := range connectEndpointCRData.Status.To {
			checkFlag := false
			for indexToPod, toPod := range to.ToPod {
				if pstCRData.Name == toPod.PodRef.Name &&
					pstCRData.Namespace == toPod.PodRef.Namespace {
					checkFlag = true

					if pstCRData.Status.AppStatus.Raw == nil {
						if toPod.AppStatus.Raw != nil {
							changeFlag = true
							log.Info(cMyFuncName + " detected that the appStatus in the status area of DCIAppStateCR was deleted.")
						}
					} else if toPod.AppStatus.Raw == nil {
						changeFlag = true
						log.Info(cMyFuncName + " detected that the appStatus was added to the status area of the DCIAppStateCR.")
					} else if !bytes.Equal(toPod.AppStatus.Raw, pstCRData.Status.AppStatus.Raw) {
						changeFlag = true
						log.Info(cMyFuncName + " detected that the appStatus in the status area of the DCIAppStateCR was chenged.")
					}
					if changeFlag {
						updateIndex = index
						updateIndexToPod = indexToPod
					}
					break
				}
			}
			if checkFlag {
				break
			}
		}
	}
	log.Info(cMyFuncName + " End." +
		" changeFlag: " + strconv.FormatBool(changeFlag) +
		", updateIndex: " + strconv.Itoa(updateIndex) +
		", updateIndex: " + strconv.Itoa(updateIndex))
	return changeFlag, updateIndex, updateIndexToPod
}

func (r *DCIAppStateReconciler) reconcileUpdate(
	ctx context.Context,
	log logr.Logger,
	pstCRData *examplecomv1.DCIAppState,
	myPodName string) error {
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" ConnectEndpoint: " + myPodName + ", " + pstCRData.Labels[myPodName])

	for range 1 {
		connectEndpointCRData := &examplecomv1.ConnectEndpoint{}
		err = r.Get(ctx, client.ObjectKey{
			Namespace: pstCRData.Labels[myPodName],
			Name:      myPodName}, connectEndpointCRData)
		if apierrors.IsNotFound(err) {
			err = nil
			log.Error(err, cMyFuncName+" could not find ConnectEndpointCR and failed to retrieve it."+
				" Name: "+myPodName+
				", Namespace: "+pstCRData.Labels[myPodName])
			break
		} else if err != nil {
			log.Error(err, cMyFuncName+" is unable to fetch ConnectEndpointCR."+
				" Name: "+myPodName+
				", Namespace: "+pstCRData.Labels[myPodName])
			break
		}

		var changeFlag bool
		var updateIndex int
		var updateIndexToPod int
		changeFlag, updateIndex, updateIndexToPod = statusChangeCheck(
			log,
			pstCRData,
			connectEndpointCRData)

		if !changeFlag {
			break
		}

		pstBaseConnectEndpointCRData := client.MergeFrom(connectEndpointCRData.DeepCopy())

		connectEndpointCRData.Status.To[updateIndex].ToPod[updateIndexToPod].AppStatus =
			pstCRData.Status.AppStatus

		var bFuncRetFlag bool
		bFuncRetFlag, err = createConfigFileFunc(
			log,
			connectEndpointCRData,
			examplecomv1.Create)
		if err != nil {
			break
		} else if bFuncRetFlag {
			bFuncRetFlag, err = sendSignalFunc(log)
			if err != nil {
				break
			} else if bFuncRetFlag {
				err = r.Patch(ctx, connectEndpointCRData, pstBaseConnectEndpointCRData)
				if err != nil {
					log.Error(err, cMyFuncName+" failed to update the ConnectEndpointCR."+
						" Name: "+myPodName+
						", Namespace: "+pstCRData.Labels[myPodName])
				}
			}
		}
	}
	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}
