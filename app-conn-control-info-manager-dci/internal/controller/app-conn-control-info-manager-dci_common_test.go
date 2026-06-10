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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplecomv1 "app-conn-control-info-manager-dci/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	ProcessCmd    string = "../.././test_process.sh"
	ProcessCmdEnv string = "/bin/bash -x " + ProcessCmd
)

const (
	RawExtensionName string = "RawExtension"
	RawExtensionPath string = "k8s.io/apimachinery/pkg/runtime"
)

func getMgr(mgr ctrl.Manager) (ctrl.Manager, error) {
	if mgr == nil {
		return ctrl.NewManager(cfg, ctrl.Options{
			Scheme: testScheme,
		})
	}
	return mgr, nil
}

func processStart() {
	cmd := exec.Command(ProcessCmd)
	_ = cmd.Start()

	go func() {
		_ = cmd.Wait()
	}()
}

func createShellfile() {
	content := []byte(
		"#!/bin/bash -x\n" +
			"while true; do sleep 2; done",
	)

	_ = os.WriteFile(ProcessCmd, content,
		(syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IXUSR |
			syscall.S_IRGRP | syscall.S_IXGRP |
			syscall.S_IROTH | syscall.S_IXOTH))
}

func getTestProcess() (pgrepst string) {
	pgrepCmd := exec.Command("pgrep", "-o", "test_process.sh")
	pgrep, _ := pgrepCmd.Output()
	pgrepst = strings.TrimSpace(string(pgrep))
	fmt.Println("get process : ", pgrepst)
	return pgrepst
}

func checkTestProcess(pid string) bool {
	checkFlag := false

	pgrepCmd := exec.Command("pgrep", "-o", "test_process.sh")
	pgrep, _ := pgrepCmd.Output()
	pgrepst := strings.TrimSpace(string(pgrep))
	pgrepSlice := strings.Split(pgrepst, "\n")
	fmt.Println("get process : ", pgrepSlice)
	for _, v := range pgrepSlice {
		if pid == v {
			checkFlag = true
			break
		}
	}
	return checkFlag
}

func deleteProcess() {
	pgrepCmd := exec.Command("pgrep", "-o", "test_process.sh")
	pgrep, _ := pgrepCmd.Output()
	pgrepst := strings.TrimSpace(string(pgrep))
	pgrepSlice := strings.Split(pgrepst, "\n")
	for _, v := range pgrepSlice {
		_, _ = exec.Command("kill", v).Output()
	}
}

func deleteShellfile() {
	err := os.Remove(ProcessCmd)
	fmt.Println("deleteShellfile result : ", err)
}

func createConnectEndpoint(ctx context.Context, connectEndpointCR *examplecomv1.ConnectEndpoint) error {
	tmp := &examplecomv1.ConnectEndpoint{}
	*tmp = *connectEndpointCR
	tmp.TypeMeta = connectEndpointCR.TypeMeta
	err := k8sClient.Create(ctx, tmp)
	if err != nil {
		return err
	}
	return nil
}

func createDCIAppState(ctx context.Context, dciAppStateCR *examplecomv1.DCIAppState) error {
	tmp := &examplecomv1.DCIAppState{}
	*tmp = *dciAppStateCR
	tmp.TypeMeta = dciAppStateCR.TypeMeta
	err := k8sClient.Create(ctx, tmp)
	if err != nil {
		return err
	}
	return nil
}

func showappConnectControlInfo() error {
	var err error
	var configFileBytes []byte
	var configFileData examplecomv1.AppConnectControlInfo

	configFileBytes, err = os.ReadFile(examplecomv1.ConfigFilePath)
	if err != nil {
		fmt.Println("unable to readfile. FileName = " + examplecomv1.ConfigFilePath)
		return err
	}
	if err = json.Unmarshal(configFileBytes, &configFileData); err != nil {
		fmt.Println("unable to unmarshal. FileName = " + examplecomv1.ConfigFilePath)
		return err
	}

	fmt.Println("====================================================")
	fmt.Println("AppConnectControlInfo Data")
	fmt.Println("Address : ", configFileData.Address)

	for _, stServiceData := range configFileData.Services {
		fmt.Println("Services.Label : ", stServiceData.Label)

		for podIndex := range stServiceData.Destinations {
			destinationPodData := &stServiceData.Destinations[podIndex]
			val := reflect.ValueOf(*destinationPodData)
			typ := reflect.TypeOf(*destinationPodData)
			for idx := range val.NumField() {
				fieldType := typ.Field(idx).Type
				fieldName := typ.Field(idx).Name
				fieldValue := val.Field(idx).Interface()
				if fieldType.Name() == "RawExtension" &&
					fieldType.PkgPath() == "k8s.io/apimachinery/pkg/runtime" {
					if raw, ok := fieldValue.(runtime.RawExtension); ok {
						if raw.Raw != nil {
							var m map[string]interface{}
							err := json.Unmarshal(raw.Raw, &m)
							if err == nil {
								fmt.Printf("field: %s, map: %+v\n", fieldName, m)
							} else {
								fmt.Println("unmarshal error:", err)
							}
						}
					}
				} else {
					fmt.Printf("field: %s, value: %v\n", fieldName, fieldValue)
				}
			}
		}
	}
	for _, inputSubsetData := range configFileData.InputSubsets {
		fmt.Println("InputSubsets : ", inputSubsetData)
	}
	for _, outputSubsetData := range configFileData.OutputSubsets {
		fmt.Println("OutputSubsets : ", outputSubsetData)
	}
	fmt.Println("====================================================")

	return nil
}

func showDCIAppState(ctx context.Context, name, namespace string) error {
	var dciAppState examplecomv1.DCIAppState
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &dciAppState)
	if err != nil {
		fmt.Println("NG")
		return err
	}

	fmt.Println("====================================================")
	fmt.Println("DCIAppState CR Data : ", dciAppState.Name, " ", dciAppState.Namespace)
	fmt.Println("ObjectMeta.Labels : ", dciAppState.Labels)
	fmt.Println("ObjectMeta.Finalizers : ", dciAppState.Finalizers)
	fmt.Println("Spec.Type : ", dciAppState.Spec.Type)

	var ifAppStateAppStatus map[string]interface{}
	_ = json.Unmarshal(dciAppState.Status.AppStatus.Raw,
		&ifAppStateAppStatus)
	fmt.Println("Status.AppStatus : ", ifAppStateAppStatus)

	fmt.Println("====================================================")

	return nil
}

func showConnectEndpoint(ctx context.Context, namespace string) error {
	var connectEndpointData examplecomv1.ConnectEndpoint
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      "dciappstate-101",
		Namespace: namespace,
	}, &connectEndpointData)
	if err != nil {
		return err
	}

	fmt.Println("====================================================")
	fmt.Println("ConnectEndpoint CR Data")
	fmt.Println("Spec.From.DeploymentRef : ", connectEndpointData.Spec.From.DeploymentRef)
	fmt.Println("Spec.From.FromPod.PodRef : ", connectEndpointData.Spec.From.FromPod.PodRef)
	fmt.Println("Spec.From.FromPod.Address : ", connectEndpointData.Spec.From.FromPod.Address)
	for toIndex, toData := range connectEndpointData.Spec.To {
		fmt.Printf("Spec.To[%d]", toIndex)
		fmt.Println(".DeploymentRef : ", toData.DeploymentRef)
		for toPodIndex, toPodData := range toData.ToPod {
			fmt.Printf("Spec.To[%d].ToPod[%d]", toIndex, toPodIndex)
			fmt.Println(".PodRef : ", toPodData.PodRef)
			fmt.Printf("Spec.To[%d].ToPod[%d]", toIndex, toPodIndex)
			fmt.Println(".Address : ", toPodData.Address)
		}
	}
	if connectEndpointData.Spec.InputEndpoint != nil {
		for _, inputEndpointExternalSubsetsData := range connectEndpointData.Spec.InputEndpoint.ExternalSubsets {
			fmt.Println("Spec.InputEndpoint.ExternalSubsets : ",
				inputEndpointExternalSubsetsData)
		}
	}
	if connectEndpointData.Spec.OutputEndpoint != nil {
		for _, outputEndpointExternalSubsetsData := range connectEndpointData.Spec.OutputEndpoint.ExternalSubsets {
			fmt.Println("Spec.OutputEndpoint.ExternalSubsets : ",
				outputEndpointExternalSubsetsData)
		}
	}

	fmt.Println("Status.Condition : ", connectEndpointData.Status.Condition)
	fmt.Println("Status.From.DeploymentRef : ", connectEndpointData.Status.From.DeploymentRef)
	fmt.Println("Status.From.FromPod.PodRef : ", connectEndpointData.Status.From.FromPod.PodRef)
	fmt.Println("Status.From.FromPod.Address : ", connectEndpointData.Status.From.FromPod.Address)
	for toIndex, toData := range connectEndpointData.Status.To {
		fmt.Printf("Status.To[%d]", toIndex)
		fmt.Println(".DeploymentRef : ", toData.DeploymentRef)
		for toPodIndex, toPodData := range toData.ToPod {
			fmt.Printf("Status.To[%d].ToPod[%d]", toIndex, toPodIndex)
			fmt.Println("Status.To.PodRef : ", toPodData.PodRef)
			fmt.Printf("Status.To[%d].ToPod[%d]", toIndex, toPodIndex)
			fmt.Println("Status.To.Address : ", toPodData.Address)
			var mapAppStatus map[string]interface{}
			_ = json.Unmarshal(toPodData.AppStatus.Raw, &mapAppStatus)
			fmt.Printf("Status.To[%d].ToPod[%d]", toIndex, toPodIndex)
			fmt.Println(".AppStatus : ", mapAppStatus)
		}
	}
	if connectEndpointData.Status.InputEndpoint != nil {
		for _, inputEndpointExternalSubsetsData := range connectEndpointData.Status.InputEndpoint.ExternalSubsets {
			fmt.Println("Status.InputEndpoint.ExternalSubsets : ",
				inputEndpointExternalSubsetsData)
		}
	}
	if connectEndpointData.Status.OutputEndpoint != nil {
		for _, outputEndpointExternalSubsetsData := range connectEndpointData.Status.OutputEndpoint.ExternalSubsets {
			fmt.Println("Status.OutputEndpoint.ExternalSubsets : ",
				outputEndpointExternalSubsetsData)
		}
	}
	fmt.Println("====================================================")
	return nil
}
