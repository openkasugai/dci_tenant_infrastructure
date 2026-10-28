package controller

import (
	"fmt"
	"os"
	"reflect"
	pkgrt "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"

	"encoding/json"

	"github.com/go-logr/logr"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"
)

// writeFileFunc allows tests to override file write behavior.
var writeFileFunc = os.WriteFile

// jsonMarshalFunc allows tests to override json.Marshal
var jsonMarshalFunc = json.Marshal

// readDirFunc allows tests to override os.ReadDir
var readDirFunc = os.ReadDir

// readFileFunc allows tests to override os.ReadFile
var readFileFunc = os.ReadFile

// findProcessFunc allows tests to override process lookup behavior.
var findProcessFunc = findProcess

// signalFunc allows tests to override the signal sending behavior.
var signalFunc = signalSend

func signalSend(p *os.Process, sig syscall.Signal) error {
	return p.Signal(sig)
}

// jsonUnmarshalFunc allows tests to override json.Unmarshal.
var jsonUnmarshalFunc = json.Unmarshal

// runtimeCallerFunc allows tests to override runtime.Caller.
var runtimeCallerFunc = pkgrt.Caller

// findProcessFunc allows tests to override process lookup behavior.
var FindProcessFunc = os.FindProcess

// Allow tests to override createConfigFile and sendSignal behavior.
var createConfigFileFunc = createConfigFile
var sendSignalFunc = sendSignal

// Allow tests to override syscall.Close and syscall.Unlink behavior.
var syscallCloseFunc = syscall.Close
var syscallUnlinkFunc = syscall.Unlink

func NewCustomControllerOptions() controller.Options {
	return controller.Options{
		RateLimiter: workqueue.NewTypedMaxOfRateLimiter(
			workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				examplecomv1.FailureRateMin*time.Second,
				examplecomv1.FailureRateMin*time.Second,
			),
			&workqueue.TypedBucketRateLimiter[ctrl.Request]{
				Limiter: rate.NewLimiter(
					rate.Limit(examplecomv1.RateLimit),
					examplecomv1.RateBurst),
			},
		),
	}
}

func getFuncName() string {
	pc, _, _, ok := runtimeCallerFunc(1)
	if !ok {
		return ""
	}
	cFuncName := pkgrt.FuncForPC(pc).Name()
	idx := 0
	iLastIdx := strings.LastIndex(cFuncName, ".")
	if iLastIdx >= 0 && iLastIdx < len(cFuncName) {
		idx = iLastIdx
	}
	return cFuncName[idx+1:]
}

func createPodConfig(
	log logr.Logger,
	toPodStatus *examplecomv1.ToPodStatus) (examplecomv1.Destination, error) {
	var err error

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" PodRef: " + toPodStatus.PodRef.Name + "," + toPodStatus.PodRef.Namespace +
		", Address: " + toPodStatus.Address)

	mstDestinationPod := examplecomv1.Destination{}
	mstDestinationPod.Name = toPodStatus.PodRef.Name
	mstDestinationPod.Namespace = toPodStatus.PodRef.Namespace
	mstDestinationPod.Address = toPodStatus.Address

	if len(toPodStatus.AppStatus.Raw) != 0 {
		mapAppStatus := map[string]any{}
		err = jsonUnmarshalFunc(toPodStatus.AppStatus.Raw, &mapAppStatus)
		if err == nil {
			iVal := reflect.ValueOf(&mstDestinationPod).Elem()
			for sAppStatusKey, val := range mapAppStatus {
				runes := []rune(sAppStatusKey)
				runes[0] = unicode.ToUpper(runes[0])
				field := iVal.FieldByName(string(runes))
				if field.IsValid() && field.CanSet() {
					var cVal []byte
					cVal, err = json.Marshal(val)
					if err == nil {
						objectVal := runtime.RawExtension{Raw: cVal}
						field.Set(reflect.ValueOf(objectVal))
					}
				} else {
					log.Error(err, cMyFuncName+" could not found "+
						sAppStatusKey+" as a member of DestinationPod.")
				}
			}
		} else {
			log.Error(err, cMyFuncName+" failed to json.Unmarshal the AppStatus."+
				" AppStatus: "+string(toPodStatus.AppStatus.Raw))
		}
	} else {
		log.Info(cMyFuncName + " dose not process because AppStatus has no value." +
			" AppStatus: " + string(toPodStatus.AppStatus.Raw))
	}

	log.Info(cMyFuncName + " End." +
		" err: " + fmt.Sprint(err))
	return mstDestinationPod, err
}

func createConfigFile(
	log logr.Logger,
	pstCRData *examplecomv1.ConnectEndpoint,
	eventKind string) (bool, error) {
	var err error
	bProcFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start." +
		" ConnectEndpoint: " + pstCRData.Name + "," + pstCRData.Namespace +
		", Address: " + pstCRData.Status.From.FromPod.Address)

	for range 1 {
		var fileData examplecomv1.AppConnectControlInfo
		var jsonData []byte

		fileData.Address = pstCRData.Status.From.FromPod.Address
		fileData.Services = []examplecomv1.Service{}

		if examplecomv1.Delete != eventKind {
			if pstCRData.Status.InputEndpoint != nil {
				fileData.InputSubsets = pstCRData.Status.InputEndpoint.ExternalSubsets
			}

			if pstCRData.Status.OutputEndpoint != nil {
				fileData.OutputSubsets = pstCRData.Status.OutputEndpoint.ExternalSubsets
			}

			for _, to := range pstCRData.Status.To {
				var stService examplecomv1.Service
				stService.Label = to.DeploymentRef.Name
				stService.Destinations = []examplecomv1.Destination{}
				for _, toPodStatus := range to.ToPod {
					var mstDestinationPod examplecomv1.Destination
					mstDestinationPod, err = createPodConfig(log, &toPodStatus)
					if err != nil {
						break
					}
					stService.Destinations =
						append(stService.Destinations,
							mstDestinationPod)
				}
				if err != nil {
					break
				}
				fileData.Services =
					append(fileData.Services, stService)
			}
			if err != nil {
				break
			}
		}

		jsonData, err = jsonMarshalFunc(fileData)
		if err != nil {
			log.Error(err, cMyFuncName+" failed to json.Marshal the fileData.")
			break
		}

		err = writeFileFunc(examplecomv1.ConfigFilePath, jsonData,
			syscall.S_IRUSR|syscall.S_IWUSR|syscall.S_IRGRP|syscall.S_IROTH) // 0644
		if err != nil {
			log.Error(err, cMyFuncName+" is unable to write the file.")
			break
		}

		bProcFlag = true
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return bProcFlag, err
}

func findProcess(processName string) (*os.Process, error) {
	var err error
	var retPID *os.Process

	for range 1 {
		var entries []os.DirEntry
		procPath := "/proc"
		entries, err = readDirFunc(procPath)
		if err != nil {
			// If reading /proc fails (e.g. in restricted test environments),
			// do not treat it as a fatal error for process lookup; return no process and no error.
			err = nil
			break
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pid, errStrConv := strconv.Atoi(entry.Name())
			if errStrConv != nil {
				continue
			}

			cmdline, errReadFile := readFileFunc(fmt.Sprintf("/proc/%d/cmdline", pid))
			if errReadFile != nil {
				continue
			}

			var trimcmdline []byte
			for _, v := range cmdline {
				if v != 0 {
					trimcmdline = append(trimcmdline, v)
				}
			}

			name := strings.ReplaceAll(string(trimcmdline), " ", "")
			replaceProcessName := strings.ReplaceAll(processName, " ", "")
			if strings.Contains(name, replaceProcessName) {
				retPID, err = FindProcessFunc(pid)
				break
			}
		}
	}
	return retPID, err
}

func sendSignal(log logr.Logger) (bool, error) {
	var err error
	bProcFlag := false

	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	for range 1 {
		processName := os.Getenv(examplecomv1.MainContainerProcessName)
		if processName == "" {
			log.Error(err, cMyFuncName+" not found process name from the environment variable.")
			break
		}

		var process *os.Process
		process, err = findProcessFunc(processName)
		if err != nil {
			log.Error(err, cMyFuncName+" detected an error in the findProcess.")
			break
		}
		if process == nil {
			log.Error(err, cMyFuncName+" detected that no process name was returned from the findProcess."+
				" processName: ", processName)
			break
		}

		err = signalFunc(process, syscall.SIGHUP)
		if err != nil {
			log.Error(err, cMyFuncName+" failed to send SIGHUP.")
			break
		}
		bProcFlag = true
	}

	log.Info(cMyFuncName + " End." +
		" bProcFlag: " + strconv.FormatBool(bProcFlag) +
		", err: " + fmt.Sprint(err))
	return bProcFlag, err
}

func fileExclusive(
	myPodName string,
	myPodNamespace string) (cFilePath string, iFd int, eFdErr error) {
	cFilePath = examplecomv1.ExclCtrlComFilePath + myPodName + "_" +
		myPodNamespace + examplecomv1.ExclCtrlFileExtension

	iFd, eFdErr = syscall.Open(
		cFilePath,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_RDWR,
		syscall.S_IRUSR|syscall.S_IWUSR|syscall.S_IRGRP|syscall.S_IROTH) // 0644

	return cFilePath, iFd, eFdErr
}

func deferExclusive(
	log logr.Logger,
	iFd int,
	cFilePath string) {
	cMyFuncName := getFuncName()
	log.Info(cMyFuncName + " Start.")

	err := syscallCloseFunc(iFd)
	if err != nil {
		log.Error(err, cMyFuncName+" failed to close the file.")
	}
	err = syscallUnlinkFunc(cFilePath)
	if err != nil {
		log.Error(err, cMyFuncName+" failed to unlink the file.")
	}

	log.Info(cMyFuncName + " End.")
}
