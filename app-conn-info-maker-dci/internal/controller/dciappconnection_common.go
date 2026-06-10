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
	pkgrt "runtime"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"

	examplecomv1 "app-conn-info-maker-dci/api/v1"

	hostdevicenetwork "github.com/Mellanox/network-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"encoding/json"
	"strconv"
	"strings"
)

var ctrlSetControllerReferenceFunc = ctrl.SetControllerReference
var jsonUnmarshalFunc = json.Unmarshal

func getFinalizerName(
	rClient client.Client,
	pstCRData *examplecomv1.DCIAppConnection) string {
	// Value to set in the finalizer
	gvks, _, _ := rClient.Scheme().ObjectKinds(pstCRData)
	return strings.ToLower(gvks[0].Kind) + ".finalizers." +
		strings.ToLower(gvks[0].Group+"/"+gvks[0].Version)
}

func getFromConnectionInformation(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	cDCIAppConnectionName string,
	cDCIAppConnectionNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	iConnIdx int,
	iIOMode int) (error, bool) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", DCIAppConnectionRef: " + cDCIAppConnectionName + "," + cDCIAppConnectionNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag) +
		", iConnIdx: " + strconv.Itoa(iConnIdx) +
		", iIOMode: " + strconv.Itoa(iIOMode))
	var err error
	bProcFlag := false

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// List Get Pod Resource
		var pstCRDeployment *appsv1.Deployment
		var pstCRPods *corev1.PodList
		var bFuncRet bool
		err, bFuncRet, pstCRDeployment, pstCRPods = getDeploymentList(
			ctx,
			rClient,
			log,
			cDeploymentName,
			cDeploymentNamespace,
			bAdditionalProcFlag,
			bPodControllerFlag)

		if err != nil {
			break
		}
		if !bFuncRet {
			break
		}

		if pstCRDeployment == nil {
			if bPodControllerFlag {
				log.Info(cFuncName + " detected that Deployment resource does not exist.")
				bProcFlag = true
			} else if !bPodControllerFlag && !bAdditionalProcFlag {
				log.Info(cFuncName + " detected that Deployment resource does not exist.")
				bProcFlag = true
			}
			break
		}

		bProcFlag = true
		cLabel, bExist := pstCRDeployment.Labels["dciappconnection"]

		if !bExist {
			if bPodControllerFlag {
				log.Info(cFuncName + " could not find dciappconnection label in Deployment resource.")
				break
			} else if !bPodControllerFlag && !bAdditionalProcFlag {
				log.Info(cFuncName + " could not find dciappconnection label in Deployment resource.")
				break
			}
		} else {
			if len(cLabel) <= 0 {
				if bPodControllerFlag {
					log.Info(cFuncName + " detected that the value of the dciappconnection label in Deployment resource is empty.")
					break
				} else if !bPodControllerFlag && !bAdditionalProcFlag {
					log.Info(cFuncName + " detected that the value of the dciappconnection label in Deployment resource is empty.")
					break
				}
			} else {
				cNameNamespace := strings.Split(cLabel, "_")
				if len(cNameNamespace) == 1 {
					cNameNamespace = append(cNameNamespace, "default")
				}
				if cNameNamespace[0] != cDCIAppConnectionName || cNameNamespace[1] != cDCIAppConnectionNamespace {
					if bPodControllerFlag {
						log.Info(cFuncName + " detected that the dciappconnection label does not match DCIAppConnectionCR within Name/Namespace.")
						break
					} else if !bPodControllerFlag && !bAdditionalProcFlag {
						log.Info(cFuncName + " detected that the dciappconnection label does not match DCIAppConnectionCR within Name/Namespace.")
						break
					} else {
						log.Error(err, cFuncName+" detected that the dciappconnection label does not match DCIAppConnectionCR withinName/Namespace. It is not a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
				}
			}
		}

		if !pstCRDeployment.DeletionTimestamp.IsZero() {
			if !bPodControllerFlag && bAdditionalProcFlag {
				log.Info(cFuncName + " detected that Deployment resource has DeletionTimestamp.")
				bProcFlag = false
				break
			}
		}

		err, bFuncRet = getFromConnectionInformationPod(
			ctx,
			rClient,
			log,
			cDeploymentName,
			cDeploymentNamespace,
			cDCIAppConnectionName,
			cDCIAppConnectionNamespace,
			bAdditionalProcFlag,
			bPodControllerFlag,
			pstFromPodInfos,
			iConnIdx,
			iIOMode,
			pstCRDeployment,
			pstCRPods)

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

func getFromConnectionInformationPod(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	cDCIAppConnectionName string,
	cDCIAppConnectionNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool,
	pstFromPodInfos *examplecomv1.FromPodInfos,
	iConnIdx int,
	iIOMode int,
	pstCRDeployment *appsv1.Deployment,
	pstCRPods *corev1.PodList) (error, bool) {
	bProcFlag := true

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", DCIAppConnectionRef: " + cDCIAppConnectionName + "," + cDCIAppConnectionNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag) +
		", iConnIdx: " + strconv.Itoa(iConnIdx) +
		", iIOMode: " + strconv.Itoa(iIOMode))
	var err error
	var bFuncRet bool
	iPodCnt := 0

	for iPodIdx := 0; len(pstCRPods.Items) > iPodIdx; iPodIdx++ {

		log.Info(cFuncName + " set FromPod Info. Pod Index: " + strconv.Itoa(len(pstCRPods.Items)))

		stCRPod := pstCRPods.Items[iPodIdx]

		// Check Running Phase
		if corev1.PodRunning != stCRPod.Status.Phase {
			if !bPodControllerFlag && bAdditionalProcFlag {
				log.Info(cFuncName + " detected that Pod resource Phase is not Running." +
					" Name: " + stCRPod.Name +
					", Namespace: " + stCRPod.Namespace)
				bProcFlag = false
				break
			}
		}

		// Get 2nd Nic Address
		var c2ndNicAddress string
		err, bFuncRet = get2ndNicAddress(
			ctx,
			rClient,
			log,
			&stCRPod,
			&c2ndNicAddress)
		if err != nil {
			bProcFlag = false
			break
		}
		if !bFuncRet {
			if bAdditionalProcFlag {
				bProcFlag = false
				break
			}
		}

		bHoldFlag := true
		cLabel, bExist := stCRPod.Labels["dciappconnection"]
		if !bExist {
			if !bAdditionalProcFlag {
				bHoldFlag = false
				log.Info("debug: " + cFuncName + " could not find dciappconnection in the Pod's labels.")
			}
		} else {
			if len(cLabel) <= 0 {
				if !bAdditionalProcFlag {
					bHoldFlag = false
					log.Info("debug: " + cFuncName + " detected that the value of the dciappconnection label in the Pod is empty.")
				}
			} else {
				cNameNamespace := strings.Split(cLabel, "_")
				if len(cNameNamespace) == 1 {
					cNameNamespace = append(cNameNamespace, "default")
				}
				if cNameNamespace[0] != cDCIAppConnectionName ||
					cNameNamespace[1] != cDCIAppConnectionNamespace {
					if bPodControllerFlag {
						bHoldFlag = false
						log.Info("debug: " + cFuncName + " detected that it does not match the value of the dciappconnection label in the Pod.")
					} else if !bAdditionalProcFlag {
						bHoldFlag = false
						log.Info("debug: " + cFuncName + " detected that it does not match the value of the dciappconnection label in the Pod.")
					} else {
						log.Error(err, cFuncName+" detected that it does not match the value of the dciappconnection label in the Pod. It is not a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
				}
			}
		}
		if bHoldFlag {
			var stFromPod examplecomv1.FromPodInfo
			stFromPod.DeploymentRef.Name = cDeploymentName
			stFromPod.DeploymentRef.Namespace = cDeploymentNamespace
			stFromPod.DeploymentR = pstCRDeployment
			stFromPod.PodR = &stCRPod
			stFromPod.Address = c2ndNicAddress
			stFromPod.IOMode = iIOMode

			stConnectEndpoint := &examplecomv1.ConnectEndpoint{}
			err = rClient.Get(ctx,
				client.ObjectKey{
					Name:      stCRPod.Name,
					Namespace: cDCIAppConnectionNamespace,
				}, stConnectEndpoint)

			if err != nil {
				if !errors.IsNotFound(err) {
					log.Error(err, cFuncName+" failed to get ConnectEndpointCR.")
					bProcFlag = false
					break
				}
				err = nil
			} else {
				if !stConnectEndpoint.DeletionTimestamp.IsZero() {
					if !bPodControllerFlag && bAdditionalProcFlag {
						log.Info(cFuncName + " detected that ConnectEndpointCR has a DeletionTimestamp. It is not a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
					if bPodControllerFlag && bAdditionalProcFlag {
						log.Info(cFuncName + " detected that ConnectEndpointCR has a DeletionTimestamp. It is a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
				}
				stFromPod.ConnectEndpointCR = stConnectEndpoint
			}
			if iPodCnt == 0 {
				stFromPod.MstPodFlag = true
			}
			if !stCRPod.DeletionTimestamp.IsZero() {
				stFromPod.DeleteFlag = true
				log.Info(cFuncName + " detected that Pod resource has DeletionTimestamp." + " Name:" + stFromPod.PodR.Name)
			}
			pstFromPodInfos.FromPods = append(
				pstFromPodInfos.FromPods,
				stFromPod)

			iPodCnt++
		}
	}
	if bProcFlag && iPodCnt == 0 {
		var stFromPod examplecomv1.FromPodInfo
		stFromPod.DeploymentRef.Name = cDeploymentName
		stFromPod.DeploymentRef.Namespace = cDeploymentNamespace
		stFromPod.DeploymentR = pstCRDeployment
		stFromPod.IOMode = iIOMode
		stFromPod.MstPodFlag = true
		pstFromPodInfos.FromPods = append(
			pstFromPodInfos.FromPods,
			stFromPod)
	}
	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func getToConnectionInformation(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	cDCIAppConnectionName string,
	cDCIAppConnectionNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool,
	pstToPodInfos *examplecomv1.ToPodInfos,
	iConnIdx int) (error, bool) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", DCIAppConnectionRef: " + cDCIAppConnectionName + "," + cDCIAppConnectionNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag) +
		", iConnIdx: " + strconv.Itoa(iConnIdx))
	var err error
	bProcFlag := false

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// List Get Pod Resource
		var pstCRDeployment *appsv1.Deployment
		var pstCRPods *corev1.PodList
		var bFuncRet bool
		err, bFuncRet, pstCRDeployment, pstCRPods = getDeploymentList(
			ctx,
			rClient,
			log,
			cDeploymentName,
			cDeploymentNamespace,
			bAdditionalProcFlag,
			bPodControllerFlag)
		if nil != err {
			break
		}
		if !bFuncRet {
			break
		}

		stToPods := pstToPodInfos.ToPodInfos[iConnIdx]
		stToPods.DeploymentRef.Name = cDeploymentName
		stToPods.DeploymentRef.Namespace = cDeploymentNamespace
		pstToPodInfos.ToPodInfos[iConnIdx] = stToPods

		if nil == pstCRDeployment {
			if bPodControllerFlag {
				log.Info(cFuncName + " detected that Deployment resource does not exist.")
				bProcFlag = true
			} else if !bPodControllerFlag && !bAdditionalProcFlag {
				log.Info(cFuncName + " detected that Deployment resource does not exist.")
				bProcFlag = true
			}
			break
		}

		bProcFlag = true
		cLabel, bExist := pstCRDeployment.Labels["dciappconnection"]
		if bExist {
			if len(cLabel) <= 0 {
				if bPodControllerFlag {
					log.Info(cFuncName + " detected that the value of the dciappconnection label in Deployment resource is empty.")
					break
				} else if !bPodControllerFlag && !bAdditionalProcFlag {
					log.Info(cFuncName + " detected that the value of the dciappconnection label in Deployment resource is empty.")
					break
				}
			} else {
				cNameNamespace := strings.Split(cLabel, "_")
				if len(cNameNamespace) == 1 {
					cNameNamespace = append(cNameNamespace, "default")
				}
				if cNameNamespace[0] != cDCIAppConnectionName || cNameNamespace[1] != cDCIAppConnectionNamespace {
					if bPodControllerFlag {
						log.Info(cFuncName + " detected that the dciappconnection label does not match DCIAppConnectionCR within Name/Namespace.")
						break
					} else if !bPodControllerFlag && !bAdditionalProcFlag {
						log.Info(cFuncName + " detected that the dciappconnection label does not match DCIAppConnectionCR within Name/Namespace.")
						break
					} else {
						log.Error(err, cFuncName+" detected that the dciappconnection label does not match DCIAppConnectionCR within Name/Namespace.. It is not a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
				}
			}
		} else {
			if bPodControllerFlag {
				log.Info(cFuncName + " could not find dciappconnection label in Deployment resource.")
				break
			} else if !bPodControllerFlag && !bAdditionalProcFlag {
				log.Info(cFuncName + " could not find dciappconnection label in Deployment resource.")
				break
			}
		}

		// Handle Deployment DeletionTimestamp semantics
		if !pstCRDeployment.DeletionTimestamp.IsZero() {
			if !bPodControllerFlag && bAdditionalProcFlag {
				log.Info(cFuncName + " detected that Deployment resource has DeletionTimestamp.")
				bProcFlag = false
				break
			}
		}

		stToPods.DeploymentR = pstCRDeployment
		pstToPodInfos.ToPodInfos[iConnIdx] = stToPods

		var pstServiceR *corev1.Service
		var pstEndpointSliceR *discoveryv1.EndpointSlice

		err, bFuncRet, pstServiceR, pstEndpointSliceR = getToConnectionInformationService(
			ctx,
			rClient,
			log,
			cDeploymentName,
			cDeploymentNamespace,
			cDCIAppConnectionName,
			cDCIAppConnectionNamespace,
			bAdditionalProcFlag,
			bPodControllerFlag)

		if err != nil {
			bProcFlag = false
			break
		}

		if !bFuncRet {
			bProcFlag = false
			break
		}
		stToPods.ServiceR = pstServiceR
		stToPods.EndpointSliceR = pstEndpointSliceR
		pstToPodInfos.ToPodInfos[iConnIdx] = stToPods

		var arToPod []examplecomv1.ToPodInfo
		err, bFuncRet, arToPod = getToConnectionInformationPod(
			ctx,
			rClient,
			log,
			cDeploymentName,
			cDeploymentNamespace,
			cDCIAppConnectionName,
			cDCIAppConnectionNamespace,
			bAdditionalProcFlag,
			bPodControllerFlag,
			pstCRPods)
		if err != nil {
			bProcFlag = false
			break
		}
		if !bFuncRet {
			bProcFlag = false
			break
		}
		stToPods.ToPods = arToPod
		pstToPodInfos.ToPodInfos[iConnIdx] = stToPods
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func getToConnectionInformationService(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	cDCIAppConnectionName string,
	cDCIAppConnectionNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool) (error, bool, *corev1.Service, *discoveryv1.EndpointSlice) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", DCIAppConnectionRef: " + cDCIAppConnectionName + "," + cDCIAppConnectionNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag))

	var err error
	bProcFlag := false
	var pstServiceR *corev1.Service
	var pstEndpointSliceR *discoveryv1.EndpointSlice

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Get Service Resource
		stServiceR := &corev1.Service{}
		err = rClient.Get(ctx,
			client.ObjectKey{
				Name:      cDeploymentName,
				Namespace: cDCIAppConnectionNamespace,
			}, stServiceR)
		if nil != err {
			if !errors.IsNotFound(err) {
				log.Error(err, cFuncName+" failed to get Service resource.")
				break
			}
			// Service not found
			err = nil
		} else {
			if !stServiceR.DeletionTimestamp.IsZero() {
				if !bPodControllerFlag && bAdditionalProcFlag {
					log.Info(cFuncName + " detected that Service resource has a DeletionTimestamp. It is not a call from the Pod controller, but a request for additional processing.")
					break
				}
			}
			pstServiceR = stServiceR
		}

		// Get Endpoint Resource only if we have a Service resource
		stEndpointSliceR := &discoveryv1.EndpointSlice{}
		err = rClient.Get(ctx,
			client.ObjectKey{
				Name:      cDeploymentName,
				Namespace: cDCIAppConnectionNamespace,
			}, stEndpointSliceR)
		if nil != err {
			if !errors.IsNotFound(err) {
				log.Error(err, cFuncName+" failed to get EndpointSlice resource.")
				break
			}
			// EndpointSlice not found
			err = nil
		} else {
			if !stEndpointSliceR.DeletionTimestamp.IsZero() {
				if bAdditionalProcFlag {
					log.Info(cFuncName + " detected that EndpointSlice resource has a DeletionTimestamp. It is not a call from the Pod controller, but a request for additional processing.")
					break
				}
			}
			pstEndpointSliceR = stEndpointSliceR
		}

		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag, pstServiceR, pstEndpointSliceR
}

func getToConnectionInformationPod(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	cDCIAppConnectionName string,
	cDCIAppConnectionNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool,
	pstCRPods *corev1.PodList) (error, bool, []examplecomv1.ToPodInfo) {
	bProcFlag := true
	var stToPodInfos []examplecomv1.ToPodInfo

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", DCIAppConnectionRef: " + cDCIAppConnectionName + "," + cDCIAppConnectionNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag))
	var err error
	var bFuncRet bool
	iPodCnt := 0

	for iPodIdx := 0; len(pstCRPods.Items) > iPodIdx; iPodIdx++ {

		stCRPod := pstCRPods.Items[iPodIdx]

		// Check Running Phase
		if corev1.PodRunning != stCRPod.Status.Phase {
			if !bPodControllerFlag && bAdditionalProcFlag {
				log.Error(err, cFuncName+" detected that Pod resource Phase is not Running.Not a call from the Pod controller. And is a request for additional processing."+
					" Name: "+stCRPod.Name+
					", Namespace: "+stCRPod.Namespace)
				bProcFlag = false
				break
			} else {
				log.Info(cFuncName + " detected that Pod resource Phase is not Running." +
					" Name: " + stCRPod.Name +
					", Namespace: " + stCRPod.Namespace)
			}
		}

		// Get 2nd Nic Address
		var c2ndNicAddress string
		err, bFuncRet = get2ndNicAddress(
			ctx,
			rClient,
			log,
			&stCRPod,
			&c2ndNicAddress)
		if err != nil {
			bProcFlag = false
			break
		}
		if !bFuncRet {
			if !bPodControllerFlag && bAdditionalProcFlag {
				bProcFlag = false
				break
			}
		}

		bHoldFlag := true
		cLabel, bExist := stCRPod.Labels["dciappconnection"]
		if !bExist {
			if !bAdditionalProcFlag {
				bHoldFlag = false
			}
		} else {
			if len(cLabel) <= 0 {
				if !bAdditionalProcFlag {
					bHoldFlag = false
				}
			} else {
				cNameNamespace := strings.Split(cLabel, "_")
				if len(cNameNamespace) == 1 {
					cNameNamespace = append(cNameNamespace, "default")
				}
				if cNameNamespace[0] != cDCIAppConnectionName ||
					cNameNamespace[1] != cDCIAppConnectionNamespace {
					if bPodControllerFlag {
						bHoldFlag = false
					} else if !bAdditionalProcFlag {
						bHoldFlag = false
					} else {
						log.Info(cFuncName + " detected that it does not match the value of the dciappconnection label in the Pod. It is not a call from the Pod controller, but a request for additional processing.")
						bProcFlag = false
						break
					}
				}
			}
		}

		if bHoldFlag {
			// Make ToPod Information
			var stToPod examplecomv1.ToPodInfo
			stToPod.PodR = &stCRPod
			stToPod.Address = c2ndNicAddress
			if !stCRPod.DeletionTimestamp.IsZero() {
				stToPod.DeleteFlag = true
				log.Info(cFuncName + " detected that Pod resource has DeletionTimestamp." + " Name:" + stToPod.PodR.Name)
			}
			iPodCnt++
			stToPodInfos = append(stToPodInfos, stToPod)
		}
	}
	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag, stToPodInfos
}

func getDeploymentList(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cDeploymentName string,
	cDeploymentNamespace string,
	bAdditionalProcFlag bool,
	bPodControllerFlag bool) (error, bool, *appsv1.Deployment, *corev1.PodList) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + cDeploymentName + "," + cDeploymentNamespace +
		", bAdditionalProcFlag: " + strconv.FormatBool(bAdditionalProcFlag) +
		", bPodControllerFlag: " + strconv.FormatBool(bPodControllerFlag))
	var err error
	bProcFlag := false
	var pstCRDeployment *appsv1.Deployment
	pstCRPods := &corev1.PodList{}

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		pstGetCRDeployment := &appsv1.Deployment{}
		// Get Deployment Resource
		err = rClient.Get(ctx,
			client.ObjectKey{
				Name:      cDeploymentName,
				Namespace: cDeploymentNamespace,
			},
			pstGetCRDeployment)
		if nil != err {
			if errors.IsNotFound(err) {
				err = nil
				log.Info(cFuncName + " could not find Deployment resource.")
				if !bPodControllerFlag && bAdditionalProcFlag {
					// do nothing
				} else {
					bProcFlag = true
				}
			} else {
				log.Error(err, cFuncName+" failed to get Deployment resource.")
			}
			break
		}
		cName, bExist := pstGetCRDeployment.Labels["app.kubernetes.io/name"]
		if !bExist {
			log.Info(cFuncName + " could not find Deployment resource label app.kubernetes.io/name.")
			if !bPodControllerFlag && bAdditionalProcFlag {
				// do nothing
			} else {
				bProcFlag = true
			}
			break
		}
		pstCRDeployment = pstGetCRDeployment

		// List Get ReplicaSet Resource
		pstCRReplicaSets := &appsv1.ReplicaSetList{}
		err = rClient.List(ctx, pstCRReplicaSets,
			&client.ListOptions{
				Namespace: cDeploymentNamespace,
				LabelSelector: labels.SelectorFromSet(
					map[string]string{
						"app.kubernetes.io/name": cName,
					},
				),
			})
		if nil != err {
			log.Error(err, cFuncName+" failed to list ReplicaSet resource.")
			break
		}

		if len(pstCRReplicaSets.Items) == 0 {
			if !bPodControllerFlag && bAdditionalProcFlag {
				log.Info(cFuncName + " detected that ReplicaSet resource list is empty." +
					" Name: " + cDeploymentName +
					", Namespace: " + cDeploymentNamespace)
				break
			}
		}

		bFuncRet := true

		for iReplicaIdx := 0; len(pstCRReplicaSets.Items) > iReplicaIdx; iReplicaIdx++ {

			pstCRReplicaSet := pstCRReplicaSets.Items[iReplicaIdx]
			bReplicaSetDeletedFlag := false

			if !pstCRReplicaSet.DeletionTimestamp.IsZero() {
				if !bPodControllerFlag && bAdditionalProcFlag {
					log.Info(cFuncName + " detested that ReplicaSet resource has DeletionTimestamp.")
					bFuncRet = false
					break
				}
				if bPodControllerFlag && bAdditionalProcFlag {
					bReplicaSetDeletedFlag = true
				}
			}

			if !bReplicaSetDeletedFlag {

				cPodHash, exist := pstCRReplicaSet.Labels["pod-template-hash"]
				if !exist {
					if !bPodControllerFlag && bAdditionalProcFlag {
						log.Info(cFuncName + "getDeploymentList could not find pod-template-hash label in ReplicaSet resource.")
						bFuncRet = false
						break
					}
				}

				pstCRPodList := &corev1.PodList{}
				// List Get Pod Resource
				err = rClient.List(ctx,
					pstCRPodList,
					&client.ListOptions{
						Namespace: pstCRReplicaSet.Namespace,
						LabelSelector: labels.SelectorFromSet(
							labels.Set{
								"app.kubernetes.io/name": cName,
								"pod-template-hash":      cPodHash,
							},
						),
					})
				if nil != err {
					log.Error(err, cFuncName+" failed to list Pod resource.")
					break
				}
				if len(pstCRPodList.Items) == 0 {
					log.Info(cFuncName + " detected that Pod resource list is empty." +
						" Name: " + pstCRReplicaSet.Name +
						", Namespace: " + pstCRReplicaSet.Namespace)
				} else {
					log.Info(cFuncName + "Pod resource list num: " + strconv.Itoa(len(pstCRPodList.Items)) +
						" Name: " + pstCRReplicaSet.Name +
						", Namespace: " + pstCRReplicaSet.Namespace)
					for iPodIdx := 0; len(pstCRPodList.Items) > iPodIdx; iPodIdx++ {
						pstCRPods.Items = append(pstCRPods.Items,
							pstCRPodList.Items[iPodIdx])
					}
				}
			}
		}
		if bFuncRet {
			if pstGetCRDeployment.Spec.Replicas != nil {
				if *pstGetCRDeployment.Spec.Replicas != int32(len(pstCRPods.Items)) {
					if !bPodControllerFlag && bAdditionalProcFlag {
						log.Error(err, cFuncName+" detected  a mismatch in the number of Pod resource. It is not a call from the Pod controller, but a request for additional processing.")
						bFuncRet = false
					} else {
						log.Info(cFuncName + " Pod resource count: expected: " + strconv.Itoa(int(*pstGetCRDeployment.Spec.Replicas)) + ", actual: " + strconv.Itoa(len(pstCRPods.Items)))
					}
				}
			} else {
				if len(pstCRPods.Items) != 1 {
					if !bPodControllerFlag && bAdditionalProcFlag {
						log.Error(err, cFuncName+" detected  a mismatch in the number of Pod resource. It is not a call from the Pod controller, but a request for additional processing.")
						bFuncRet = false
					} else {
						log.Info(cFuncName + " Pod resource count: expected: 1 , actual: " + strconv.Itoa(len(pstCRPods.Items)))
					}
				}
			}
		}
		if bFuncRet {
			bProcFlag = true
		}
	}

	log.Info(cFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag, pstCRDeployment, pstCRPods
}

const (
	Interface2ndNicDef = "net1"
)

func get2ndNicAddress(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	pstCRPod *corev1.Pod,
	pc2ndNicAddress *string) (error, bool) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" PodRef: " + pstCRPod.Name + "," + pstCRPod.Namespace)
	bProcFlag := false
	var err error
	*pc2ndNicAddress = ""

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Parse network-status annotation directly to obtain 2nd NIC address.
		bHostDeviceNetworkFlag := false
		cHostDeviceNWName, exists := pstCRPod.Annotations["k8s.v1.cni.cncf.io/networks"]
		// If a HostDeviceNetwork name is configured, attempt to fetch it. Tests may leave
		// cHostDeviceNWName empty which skips this lookup.
		if exists {
			if len(cHostDeviceNWName) != 0 {
				stHostDeviceNW := &hostdevicenetwork.HostDeviceNetwork{}
				err = rClient.Get(ctx,
					client.ObjectKey{
						Name:      cHostDeviceNWName,
						Namespace: pstCRPod.Namespace,
					},
					stHostDeviceNW)

				if errors.IsNotFound(err) {
					log.Error(err, cFuncName+" could not find HostDeviceNetwork."+
						" Name: "+cHostDeviceNWName+
						", Namespace: "+pstCRPod.Namespace)
					err = nil
				} else if err != nil {
					log.Error(err, cFuncName+" failed to fetch HostDeviceNetwork."+
						" Name: "+cHostDeviceNWName+
						", Namespace: "+pstCRPod.Namespace)
					break
				} else {
					bHostDeviceNetworkFlag = true
				}
			}
		} else {
			log.Info(cFuncName + " could not find Pod resource annotations[k8s.v1.cni.cncf.io/networks]." +
				" Name: " + pstCRPod.Name +
				", Namespace: " + pstCRPod.Namespace)
		}

		if !bHostDeviceNetworkFlag {
			bProcFlag = true
			break
		}

		// Parse network-status annotation directly to obtain 2nd NIC address.
		cAnnotationJSON, exists := pstCRPod.Annotations["k8s.v1.cni.cncf.io/network-status"]
		if !exists {
			log.Info(cFuncName + " could not find Pod resource annotations[k8s.v1.cni.cncf.io/network-status]." +
				" Name: " + pstCRPod.Name +
				", Namespace: " + pstCRPod.Namespace)
			break
		}

		if len(cAnnotationJSON) == 0 {
			log.Info(cFuncName + " detected that Pod resource annotations[k8s.v1.cni.cncf.io/network-status] length is Zero." +
				" Name: " + pstCRPod.Name +
				", Namespace: " + pstCRPod.Namespace)
			break
		}

		log.Info("Pod 2nd NIC Info." +
			" Name: " + pstCRPod.Name +
			", Namespace: " + pstCRPod.Namespace +
			", Annotation[k8s.v1.cni.cncf.io/network-status]: " +
			cAnnotationJSON + ".")

		bytes := []byte(cAnnotationJSON)

		var stInterfaces []examplecomv1.NetworkStatus
		err = jsonUnmarshalFunc(bytes, &stInterfaces)
		if nil != err {
			log.Error(err, cFuncName+" failed to json.unmarshal annotation."+
				" Name: "+pstCRPod.Name+
				", Namespace: "+pstCRPod.Namespace)
			err = nil
			break
		}

		b2ndNicFlag := false
		for _, stInterface := range stInterfaces {
			if Interface2ndNicDef != stInterface.Interface {
				continue
			}
			if len(stInterface.Ips) == 0 {
				continue
			}
			*pc2ndNicAddress = stInterface.Ips[0]
			b2ndNicFlag = true
			break
		}
		if !b2ndNicFlag {
			log.Info(cFuncName + " detected that there is no 2nd NIC." +
				" Name: " + pstCRPod.Name +
				", Namespace: " + pstCRPod.Namespace)
			break
		}
		bProcFlag = true
	}

	log.Info(cFuncName + " End." +
		" *pc2ndNicAddress: " + *pc2ndNicAddress +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return err, bProcFlag
}

func makeConnectEndpoint(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	rScheme *runtime.Scheme,
	pstCRConnection *examplecomv1.DCIAppConnection,
	pstFromPodInfo *examplecomv1.FromPodInfo,
	pstToPods *examplecomv1.ToPodInfos) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop
		pstCRConnectEndpoint := &examplecomv1.ConnectEndpoint{}

		bAddFlag := false
		if !pstFromPodInfo.DeleteFlag {

			if nil == pstFromPodInfo.ConnectEndpointCR {
				bAddFlag = true
			}

			var pstBaseCRConnectEndpoint client.Patch

			if bAddFlag {
				pstCRConnectEndpoint = &examplecomv1.ConnectEndpoint{}
				pstCRConnectEndpoint.Name = pstFromPodInfo.PodR.Name
				pstCRConnectEndpoint.Namespace = pstCRConnection.Namespace
				err = ctrlSetControllerReferenceFunc(
					pstCRConnection,
					pstCRConnectEndpoint,
					rScheme)
				if nil != err {
					log.Error(err,
						cFuncName+" failed to set ConnectEndpointCR OwnerReference."+
							" Name: "+pstCRConnectEndpoint.Name+
							", Namespace: "+pstCRConnectEndpoint.Namespace)
					break
				}
			} else {
				pstBaseCRConnectEndpoint = client.MergeFrom(
					pstFromPodInfo.ConnectEndpointCR.DeepCopy())
				pstCRConnectEndpoint = pstFromPodInfo.ConnectEndpointCR
			}

			// Set From Deployment Connection Information (Sync Spec/Status)
			pstCRConnectEndpoint.Spec.From.DeploymentRef.Name = pstFromPodInfo.DeploymentRef.Name
			pstCRConnectEndpoint.Spec.From.DeploymentRef.Namespace = pstFromPodInfo.DeploymentRef.Namespace
			pstCRConnectEndpoint.Status.From.DeploymentRef.Name = pstFromPodInfo.DeploymentRef.Name
			pstCRConnectEndpoint.Status.From.DeploymentRef.Namespace = pstFromPodInfo.DeploymentRef.Namespace

			// Set From Pod Connection Information (Sync Spec/Status)
			pstCRConnectEndpoint.Spec.From.FromPod.PodRef.Name = pstFromPodInfo.PodR.Name
			pstCRConnectEndpoint.Spec.From.FromPod.PodRef.Namespace = pstFromPodInfo.PodR.Namespace
			pstCRConnectEndpoint.Spec.From.FromPod.Address = pstFromPodInfo.Address
			pstCRConnectEndpoint.Status.From.FromPod.PodRef.Name = pstFromPodInfo.PodR.Name
			pstCRConnectEndpoint.Status.From.FromPod.PodRef.Namespace = pstFromPodInfo.PodR.Namespace
			pstCRConnectEndpoint.Status.From.FromPod.Address = pstFromPodInfo.Address

			pstInputEndpoint := examplecomv1.ExternalEndpoint{}
			// Set Input-Output Endpoint Information
			if examplecomv1.IOModeIN == pstFromPodInfo.IOMode ||
				examplecomv1.IOModeINOUT == pstFromPodInfo.IOMode {
				pstInputEndpoint.ExternalSubsets =
					pstCRConnection.Spec.InputEndpoint.ExternalSubsets
				pstCRConnectEndpoint.Spec.InputEndpoint =
					&pstInputEndpoint
			}
			pstOutputEndpoint := examplecomv1.ExternalEndpoint{}
			if examplecomv1.IOModeOUT == pstFromPodInfo.IOMode ||
				examplecomv1.IOModeINOUT == pstFromPodInfo.IOMode {
				pstOutputEndpoint.ExternalSubsets =
					pstCRConnection.Spec.OutputEndpoint.ExternalSubsets
				pstCRConnectEndpoint.Spec.OutputEndpoint =
					&pstOutputEndpoint
			}

			stToDeplotments := []examplecomv1.ToDeployment{}
			for iFromDeployIdx := 0; len(pstFromPodInfo.DeploymentIndexs) > iFromDeployIdx; iFromDeployIdx++ {
				iToDeploymentIndex := pstFromPodInfo.DeploymentIndexs[iFromDeployIdx]

				stToPods := pstToPods.ToPodInfos[iToDeploymentIndex]

				log.Info(cFuncName + " refers to the From array and sets ConnectEndpointCR." +
					" FromIndex: " + strconv.Itoa(iFromDeployIdx) +
					", FromName: " + pstFromPodInfo.PodR.Name +
					", ToIndex: " + strconv.Itoa(int(iToDeploymentIndex)) +
					", Tolen: " + strconv.Itoa(len(stToPods.ToPods)))

				stEndpoints := []examplecomv1.ToPod{}

				for iToPodIdx := 0; len(stToPods.ToPods) > iToPodIdx; iToPodIdx++ {
					stToPod := stToPods.ToPods[iToPodIdx]
					if stToPod.DeleteFlag {
						continue
					}
					stEndpoint := examplecomv1.ToPod{}
					log.Info(cFuncName + " refers to the To array and sets ConnectEndpointCR." +
						" Name: " + stToPod.PodR.Name)
					stEndpoint.PodRef.Name = stToPod.PodR.Name
					stEndpoint.PodRef.Namespace = stToPod.PodR.Namespace
					stEndpoint.Address = stToPod.Address
					stEndpoints = append(stEndpoints, stEndpoint)
				}
				stToDeplotment := examplecomv1.ToDeployment{}
				stToDeplotment.DeploymentRef.Name = stToPods.DeploymentRef.Name
				stToDeplotment.DeploymentRef.Namespace = stToPods.DeploymentRef.Namespace
				stToDeplotment.ToPod = stEndpoints
				stToDeplotments = append(stToDeplotments, stToDeplotment)
			}
			pstCRConnectEndpoint.Spec.To = stToDeplotments

			if bAddFlag {
				err = rClient.Create(ctx, pstCRConnectEndpoint)
				if nil != err {
					log.Error(err, cFuncName+" failed to create ConnectEndpointCR."+
						" Name: "+pstCRConnectEndpoint.Name+
						", Namespace: "+pstCRConnectEndpoint.Namespace)
					break
				}
				log.Info(cFuncName + " created ConnectEndpointCR successfully." +
					" Name: " + pstCRConnectEndpoint.Name +
					", Namespace: " + pstCRConnectEndpoint.Namespace)
			} else {
				err = rClient.Patch(ctx,
					pstCRConnectEndpoint,
					pstBaseCRConnectEndpoint)
				if nil != err {
					log.Error(err, cFuncName+" failed to patch ConnectEndpointCR."+
						" Name: "+pstCRConnectEndpoint.Name+
						", Namespace: "+pstCRConnectEndpoint.Namespace)
					break
				}
				log.Info(cFuncName + " patched ConnectEndpointCR successfully." +
					" Name: " + pstCRConnectEndpoint.Name +
					", Namespace: " + pstCRConnectEndpoint.Namespace)
			}
			pstFromPodInfo.ConnectEndpointCR = pstCRConnectEndpoint
		}
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func deleteConnectEndpoint(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	pstCRConnectEndpoint *examplecomv1.ConnectEndpoint) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start.")

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		if pstCRConnectEndpoint != nil {
			// Delete ConnectEndpoint Resource
			err = rClient.Delete(ctx, pstCRConnectEndpoint)
			if nil != err {
				log.Error(err, cFuncName+" failed to delete ConnectEndpointCR."+
					" Name: "+pstCRConnectEndpoint.Name+
					", Namespace: "+pstCRConnectEndpoint.Namespace)
				break
			}
			log.Info(cFuncName + " deleted ConnectEndpointCR successfully." +
				" Name: " + pstCRConnectEndpoint.Name +
				", Namespace: " + pstCRConnectEndpoint.Namespace)
		}
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

var unsetDciPodConnectionRelationshipFunc = unsetDciPodConnectionRelationship

func unsetDciPodConnectionRelationship(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cLabelKey string,
	cFinalizerName string,
	pstCRPod *corev1.Pod) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" PodRef: " + pstCRPod.Name + "," + pstCRPod.Namespace)

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Update Pod Resource
		pstBaseCRPod := client.MergeFrom(pstCRPod.DeepCopy())

		// Delete labels
		delete(pstCRPod.Labels, cLabelKey)
		// Delete Finalizer
		controllerutil.RemoveFinalizer(pstCRPod, cFinalizerName)
		log.Info(cFuncName + " removed the finalizer.")

		err = rClient.Patch(ctx, pstCRPod, pstBaseCRPod)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch Pod resource."+
				" Name: "+pstCRPod.Name+
				", Namespace: "+pstCRPod.Namespace)
			break
		}
		log.Info(cFuncName + " patched Pod resource successfully." +
			" Name: " + pstCRPod.Name +
			", Namespace: " + pstCRPod.Namespace)
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func unsetDciDeploymentConnectionRelationship(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cLabelKey string,
	cFinalizerName string,
	pstDeploymentR *appsv1.Deployment) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" DeploymentRef: " + pstDeploymentR.Name + "," + pstDeploymentR.Namespace)

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Update Deployment Resource
		pstBaseCRDeployment := client.MergeFrom(pstDeploymentR.DeepCopy())

		// Delete labels
		delete(pstDeploymentR.Labels, cLabelKey)
		// Delete Finalizer
		controllerutil.RemoveFinalizer(pstDeploymentR, cFinalizerName)
		log.Info(cFuncName + " removed the finalizer.")

		err = rClient.Patch(ctx, pstDeploymentR, pstBaseCRDeployment)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch Deployment resource."+
				" Name: "+pstDeploymentR.Name+
				", Namespace: "+pstDeploymentR.Namespace)
			break
		}
		log.Info(cFuncName + " patched Deployment resource successfully." +
			" Name: " + pstDeploymentR.Name +
			", Namespace: " + pstDeploymentR.Namespace)
	}

	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func setDciPodConnectionRelationship(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cLabelKey string,
	cLabelVal string,
	cFinalizerName string,
	pstCRPod *corev1.Pod) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" LabelKey: " + cLabelKey +
		", LabelVal: " + cLabelVal +
		", FinalizerName: " + cFinalizerName +
		", PodRef: " + pstCRPod.Name + "," + pstCRPod.Namespace)

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Update Pod Resource
		pstBaseCRPod := client.MergeFrom(pstCRPod.DeepCopy())

		// Add Labels
		pstCRPod.Labels[cLabelKey] = cLabelVal
		// Add Finalizer
		controllerutil.AddFinalizer(pstCRPod, cFinalizerName)
		log.Info(cFuncName + " added a finalizer to Pod resource.")

		err = rClient.Patch(ctx, pstCRPod, pstBaseCRPod)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch Pod resource."+
				" Name: "+pstCRPod.Name+
				", Namespace: "+pstCRPod.Namespace)
			break
		}
		log.Info(cFuncName + " patched Pod resource successfully." +
			" Name: " + pstCRPod.Name +
			", Namespace: " + pstCRPod.Namespace)
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func setDciDeploymentConnectionRelationship(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	cLabelKey string,
	cLabelVal string,
	cFinalizerName string,
	cDeploymentR *appsv1.Deployment) error {

	var err error

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" LabelKey: " + cLabelKey +
		", LabelVal: " + cLabelVal +
		", FinalizerName: " + cFinalizerName +
		", DeploymentRef: " + cDeploymentR.Name + "," + cDeploymentR.Namespace)

	for doWhile := 0; doWhile < 1; doWhile++ { //nolint:staticcheck // SA4008: This loop is intentionally only a one-time loop

		// Update Deployment Resource
		pstBaseCRDeployment := client.MergeFrom(cDeploymentR.DeepCopy())

		// Add labels
		cDeploymentR.Labels[cLabelKey] = cLabelVal
		// Add Finalizer
		controllerutil.AddFinalizer(cDeploymentR, cFinalizerName)
		log.Info(cFuncName + " added a finalizer to Deployment resource.")

		err = rClient.Patch(ctx, cDeploymentR, pstBaseCRDeployment)
		if nil != err {
			log.Error(err, cFuncName+" failed to patch Deployment resource."+
				" Name: "+cDeploymentR.Name+
				", Namespace: "+cDeploymentR.Namespace)
			break
		}
		log.Info(cFuncName + " patched Deployment resource successfully." +
			" Name: " + cDeploymentR.Name +
			", Namespace: " + cDeploymentR.Namespace)
	}
	log.Info(cFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return err
}

func getAppendEndpointSlice(
	ctx context.Context,
	rClient client.Client,
	log logr.Logger,
	pstCRPod *corev1.Pod,
	pstCREndpointSlice *discoveryv1.EndpointSlice,
	c2ndNicAddress string) {

	cFuncName := getFuncName()
	log.Info(cFuncName + " Start." +
		" PodRef: " + pstCRPod.Name + "," + pstCRPod.Namespace +
		", EndpointSliceRef: " + pstCREndpointSlice.Name + "," + pstCREndpointSlice.Namespace)
	bIpNewFlag := false

	gvks, _, _ := rClient.Scheme().ObjectKinds(pstCRPod)
	cKind := strings.ToLower(gvks[0].Kind)

	stEndpoints := []discoveryv1.Endpoint{}
	for iEndpointIdx := 0; len(pstCREndpointSlice.Endpoints) > iEndpointIdx; iEndpointIdx++ {
		stEndpoint := pstCREndpointSlice.Endpoints[iEndpointIdx]

		if cKind == strings.ToLower(stEndpoint.TargetRef.Kind) &&
			pstCRPod.Name == stEndpoint.TargetRef.Name &&
			pstCRPod.Namespace == stEndpoint.TargetRef.Namespace {
			var cAddresses []string
			cAddresses = append(cAddresses, c2ndNicAddress)
			stEndpoint.Addresses = cAddresses
			log.Info(cFuncName + " set 2nd NIC address on EndpointSlice resource." +
				" Address: " + c2ndNicAddress +
				", Name: " + pstCREndpointSlice.Name +
				", Namespace: " + pstCREndpointSlice.Namespace)
			bIpNewFlag = true
		}
		stEndpoints = append(stEndpoints, stEndpoint)
	}

	if !bIpNewFlag {
		var stEndpoint discoveryv1.Endpoint
		stEndpoint.Addresses = append(
			stEndpoint.Addresses,
			c2ndNicAddress)
		stEndpoint.TargetRef = &corev1.ObjectReference{}
		stEndpoint.TargetRef.Kind = cKind
		stEndpoint.TargetRef.Name = pstCRPod.Name
		stEndpoint.TargetRef.Namespace = pstCRPod.Namespace

		log.Info(cFuncName + " added 2nd NIC address to EndpointSlice resource." +
			" Address: " + c2ndNicAddress +
			", Name: " + pstCREndpointSlice.Name +
			", Namespace: " + pstCREndpointSlice.Namespace)
		stEndpoints = append(stEndpoints, stEndpoint)
	}

	pstCREndpointSlice.Endpoints = stEndpoints

	log.Info(cFuncName + " End.")
	return
}

func getFuncName() string {
	pc, _, _, _ := pkgrt.Caller(1)
	cFuncName := pkgrt.FuncForPC(pc).Name()
	idx := 0
	iLastIdx := strings.LastIndex(cFuncName, ".")
	if iLastIdx >= 0 && iLastIdx < len(cFuncName) {
		idx = iLastIdx
	}
	return cFuncName[idx+1:]
}
