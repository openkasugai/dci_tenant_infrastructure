package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/test/testutils/mock"
)

// THIS FILE IS IMPLEMENTED TO INCREASE COVERAGE FOR FAILSAFE CASES IN client PACKAGE
// ONLY FOR PARTS THAT CANNOT BE TESTED WITH envtest FOR internal/controller.

// arguments for backup of original global variables
var (
	origCapiv1AddToScheme      = capiv1.AddToScheme
	origCapdv1AddToScheme      = capdv1.AddToScheme
	origRpmv1AddToScheme       = rpmv1.AddToScheme
	origCapdockerv1AddToScheme = capdockerv1.AddToScheme
	origNewClient              = newClient
)

// TestClient_NewClientHelper is a function using testing module
// Target function : NewClientHelper()
func TestClient_NewClientHelper(t *testing.T) {
	t.Log("Start TestClient_NewClientHelper")

	defer func() {
		t.Log("Restore original global variables")
		capiv1.AddToScheme = origCapiv1AddToScheme
		capdv1.AddToScheme = origCapdv1AddToScheme
		rpmv1.AddToScheme = origRpmv1AddToScheme
		capdockerv1.AddToScheme = origCapdockerv1AddToScheme
		newClient = origNewClient
	}()

	t.Log("Generate dummy kubeconfig")
	kubeconfigPath, err := generateDummyKubeconfig()
	if err != nil {
		t.Fatalf("failed to create dummy kubeconfig: %v", err)
	}
	defer func() {
		t.Log("Remove dummy kubeconfig")
		os.Remove(kubeconfigPath)
	}()

	// test cases
	tests := []struct {
		caseName           string              // test case name
		caseDescription    string              // test case description
		inputClient        client.Client       // input client
		inputOption        *ClientHelperOption // input options
		expectedErrorIsNil bool                // whether error is nil
		expectedErrorMsg   string              // expected error messages
		prepareFunc        func()              // prepare function
		teardownFunc       func()              // teardown function
	}{
		{
			"success/invalidParameter/option/nil",
			"succeed to create ClientHelper with nil option",
			nil,
			nil,
			true,
			"",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
			},
		},
		{
			"failure/invalidEnvironment/MANAGEMENT_KUBECONFIG/dummy",
			"error happened in create new resource when DCIMachineSpec is nil",
			nil,
			nil,
			false,
			"failed to build config from kubeconfig path",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, "dummy")
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
			},
		},
		{
			"failure/capiv1.AddToScheme",
			"error happened in create new instance when capiv1.AddToScheme failed",
			nil,
			nil,
			false,
			"failed to add Cluster API scheme",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
				capiv1.AddToScheme = func(_ *runtime.Scheme) error { return errors.New("dummy error") }
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
				capiv1.AddToScheme = origCapiv1AddToScheme
			},
		},
		{
			"failure/capdv1.AddToScheme",
			"error happened in create new instance when capdv1.AddToScheme failed",
			nil,
			nil,
			false,
			"failed to add DCI provider scheme",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
				capiv1.AddToScheme = origCapiv1AddToScheme
				capdv1.AddToScheme = func(_ *runtime.Scheme) error { return errors.New("dummy error") }
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
				capdv1.AddToScheme = origCapdv1AddToScheme
			},
		},
		{
			"failure/rpmv1.AddToScheme",
			"error happened in create new instance when rpmv1.AddToScheme failed",
			nil,
			nil,
			false,
			"failed to add Resource Pool Manager scheme",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
				capiv1.AddToScheme = origCapiv1AddToScheme
				capdv1.AddToScheme = origCapdv1AddToScheme
				rpmv1.AddToScheme = func(_ *runtime.Scheme) error { return errors.New("dummy error") }
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
				rpmv1.AddToScheme = origRpmv1AddToScheme
			},
		},
		{
			"failure/capdockerv1.AddToScheme",
			"error happened in create new instance when capdockerv1.AddToScheme failed",
			nil,
			&ClientHelperOption{
				SchemeOption: &ClientHelperSchemeOption{
					EnableDocker: true,
				},
			},
			false,
			"failed to add Docker provider scheme",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
				capdockerv1.AddToScheme = func(_ *runtime.Scheme) error { return errors.New("dummy error") }
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
				capdockerv1.AddToScheme = origCapdockerv1AddToScheme
			},
		},
		{
			"failure/newClient",
			"error happened in create new instance when newClient failed",
			nil,
			&ClientHelperOption{
				SchemeOption: &ClientHelperSchemeOption{
					EnableDocker: true,
				},
			},
			false,
			"failed to create client for management cluster",
			func() {
				_ = os.Setenv(ManagementKubeconfigEnvVarName, kubeconfigPath)
				newClient = func(_ *rest.Config, _ client.Options) (client.Client, error) { return nil, errors.New("dummy error") }
			},
			func() {
				_ = os.Unsetenv(ManagementKubeconfigEnvVarName)
				newClient = origNewClient
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			if tt.prepareFunc != nil {
				tt.prepareFunc()
			}

			// call target function
			_, result := NewClientHelper(tt.inputClient, tt.inputOption)

			if tt.teardownFunc != nil {
				tt.teardownFunc()
			}

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_getManagementResource is a function using testing module
// Target function : getManagementResource()
func TestClient_getManagementResource(t *testing.T) {
	t.Log("Start TestClient_getManagementResource")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputObject        client.Object   // input object
		inputName          string          // input name
		inputNamespace     string          // input namespace
		inputKind          string          // input kind
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/obj/nil",
			"error happened in when obj is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. name",
			"D.C. namespace",
			"D.C. kind",
			false,
			"expected object to get resource is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.getManagementResource(tt.inputCtx, tt.inputObject, tt.inputName, tt.inputNamespace, tt.inputKind)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_listManagementResources is a function using testing module
// Target function : listManagementResources()
func TestClient_listManagementResources(t *testing.T) {
	t.Log("Start TestClient_listManagementResources")

	// test cases
	tests := []struct {
		caseName           string            // test case name
		caseDescription    string            // test case description
		inputCtx           context.Context   // input context
		inputObjectList    client.ObjectList // input object list
		inputNamespace     string            // input namespace
		inputKind          string            // input kind
		expectedErrorIsNil bool              // whether error is nil
		expectedErrorMsg   string            // expected error messages
	}{
		{
			"failure/invalidParameter/objList/nil",
			"error happened in when objList is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. namespace",
			"D.C. kind",
			false,
			"expected object list to list resources is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.listManagementResources(tt.inputCtx, tt.inputObjectList, tt.inputNamespace, tt.inputKind)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_createManagementResource is a function using testing module
// Target function : createManagementResource()
func TestClient_createManagementResource(t *testing.T) {
	t.Log("Start TestClient_createManagementResource")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputObject        client.Object   // input object
		inputKind          string          // input kind
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/object/nil",
			"error happened in when object is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. kind",
			false,
			"expected object to create resource is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.createManagementResource(tt.inputCtx, tt.inputObject, tt.inputKind)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_deleteManagementResource is a function using testing module
// Target function : deleteManagementResource()
func TestClient_deleteManagementResource(t *testing.T) {
	t.Log("Start TestClient_deleteManagementResource")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputObject        client.Object   // input object
		inputKind          string          // input kind
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/object/nil",
			"error happened in when object is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. kind",
			false,
			"expected object to delete resource is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.deleteManagementResource(tt.inputCtx, tt.inputObject, tt.inputKind)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_patchManagementResource is a function using testing module
// Target function : patchManagementResource()
func TestClient_patchManagementResource(t *testing.T) {
	t.Log("Start TestClient_patchManagementResource")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputObject        client.Object   // input object
		inputLastObject    client.Object   // input object
		inputKind          string          // input kind
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/object/nil",
			"error happened in when object is nil",
			context.Background(),
			nil, // invalid parameter
			nil, // invalid parameter
			"D.C. kind",
			false,
			"expected object to patch resource is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.patchManagementResource(tt.inputCtx, tt.inputObject, tt.inputLastObject, tt.inputKind)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetDCIMachineFromMachine is a function using testing module
// Target function : GetDCIMachineFromMachine()
func TestClient_GetDCIMachineFromMachine(t *testing.T) {
	t.Log("Start TestClient_GetDCIMachineFromMachine")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputMachine       *capiv1.Machine // input machine
		inputNamespace     string          // input namespace
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/Machine/nil",
			"error happened in when Machine is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. namespace",
			false,
			"expected machine is not nil, but got nil",
		},
		{
			"failure/invalidParameter/Machine/invalidKind",
			"error happened in when Machine has invalid InfrastructureRef kind",
			context.Background(),
			&capiv1.Machine{
				Spec: capiv1.MachineSpec{
					InfrastructureRef: capiv1.ContractVersionedObjectReference{
						Kind: "InvalidKind", // invalid parameter
					},
				},
			},
			"D.C. namespace",
			false,
			fmt.Sprintf("expected InfrastructureRef kind for Machine(/) is %s, but got %s", dciMachineKind, "InvalidKind"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, result := chInstance.GetDCIMachineFromMachine(tt.inputCtx, tt.inputMachine, tt.inputNamespace)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetDCIMachineTemplateFromMachineDeployment is a function using testing module
// Target function : GetDCIMachineTemplateFromMachineDeployment()
func TestClient_GetDCIMachineTemplateFromMachineDeployment(t *testing.T) {
	t.Log("Start TestClient_GetDCIMachineTemplateFromMachineDeployment")

	// test cases
	tests := []struct {
		caseName           string                    // test case name
		caseDescription    string                    // test case description
		inputCtx           context.Context           // input context
		inputMd            *capiv1.MachineDeployment // input machine deployment
		inputNamespace     string                    // input namespace
		expectedErrorIsNil bool                      // whether error is nil
		expectedErrorMsg   string                    // expected error messages
	}{
		{
			"failure/invalidParameter/MachineDeployment/nil",
			"error happened in when MachineDeployment is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. namespace",
			false,
			"expected MachineDeployment is not nil, but got nil",
		},
		{
			"failure/invalidParameter/MachineDeployment/invalidKind",
			"error happened in when MachineDeployment has invalid InfrastructureRef kind",
			context.Background(),
			&capiv1.MachineDeployment{
				Spec: capiv1.MachineDeploymentSpec{
					Template: capiv1.MachineTemplateSpec{
						Spec: capiv1.MachineSpec{
							InfrastructureRef: capiv1.ContractVersionedObjectReference{
								Kind: "InvalidKind", // invalid parameter
							},
						},
					},
				},
			},
			"D.C. namespace",
			false,
			"failed to get DCIMachineTemplate name from MachineDeployment() spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, result := chInstance.GetDCIMachineTemplateFromMachineDeployment(tt.inputCtx, tt.inputMd, tt.inputNamespace)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_CreateDCIMachineTemplateFromNodeScalingRequest is a function using testing module
// Target function : CreateDCIMachineTemplateFromNodeScalingRequest()
func TestClient_CreateDCIMachineTemplateFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestClient_CreateDCIMachineTemplateFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName           string                      // test case name
		caseDescription    string                      // test case description
		inputCtx           context.Context             // input context
		inputName          string                      // input name
		inputNamespace     string                      // input namespace
		inputNsr           *scalev1.NodeScalingRequest // input NodeScalingRequest
		expectedErrorIsNil bool                        // whether error is nil
		expectedErrorMsg   string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in when NodeScalingRequest is nil",
			context.Background(),
			"D.C. name",
			"D.C. namespace",
			nil, // invalid parameter
			false,
			"failed to create DCIMachineTemplate spec from NodeScalingRequest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.CreateDCIMachineTemplateFromNodeScalingRequest(tt.inputCtx, tt.inputName, tt.inputNamespace, tt.inputNsr)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_convertDCIMachineTemplateListToClientObject is a function using testing module
// Target function : convertDCIMachineTemplateListToClientObject()
func TestClient_convertDCIMachineTemplateListToClientObject(t *testing.T) {
	t.Log("Start TestClient_convertDCIMachineTemplateListToClientObject")

	// test cases
	tests := []struct {
		caseName        string                         // test case name
		caseDescription string                         // test case description
		inputList       *capdv1.DCIMachineTemplateList // input DCIMachineTemplateList
		expectedIsEmpty bool                           // whether result is empty
	}{
		{
			"success/DCIMachineTemplateList/nil",
			"return empty list when DCIMachineTemplateList is nil",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.convertDCIMachineTemplateListToClientObject(tt.inputList)

			if len(result) == 0 {
				if tt.expectedIsEmpty {
					// expected result
					t.Log("got empty list")
				} else {
					// unexpected result
					t.Errorf("got empty list, but expected non empty list")
				}
			} else {
				if tt.expectedIsEmpty {
					// unexpected result
					t.Errorf("got non empty list, but expected empty list")
				} else {
					// expected result
					t.Log("got non empty list")
				}
			}
		})
	}
}

// TestClient_CreateDockerMachineTemplateFromNodeScalingRequest is a function using testing module
// Target function : CreateDockerMachineTemplateFromNodeScalingRequest()
func TestClient_CreateDockerMachineTemplateFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestClient_CreateDockerMachineTemplateFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName           string                      // test case name
		caseDescription    string                      // test case description
		inputCtx           context.Context             // input context
		inputName          string                      // input name
		inputNamespace     string                      // input namespace
		inputNsr           *scalev1.NodeScalingRequest // input NodeScalingRequest
		expectedErrorIsNil bool                        // whether error is nil
		expectedErrorMsg   string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in when NodeScalingRequest is nil",
			context.Background(),
			"D.C. name",
			"D.C. namespace",
			nil, // invalid parameter
			false,
			"failed to create DockerMachineTemplate spec from NodeScalingRequest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.CreateDockerMachineTemplateFromNodeScalingRequest(tt.inputCtx, tt.inputName, tt.inputNamespace, tt.inputNsr)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetLogicalServerFromDCIMachine is a function using testing module
// Target function : GetLogicalServerFromDCIMachine()
func TestClient_GetLogicalServerFromDCIMachine(t *testing.T) {
	t.Log("Start TestClient_GetLogicalServerFromDCIMachine")

	// test cases
	tests := []struct {
		caseName           string             // test case name
		caseDescription    string             // test case description
		inputCtx           context.Context    // input context
		inputDciMachine    *capdv1.DCIMachine // input DCI machine
		expectedErrorIsNil bool               // whether error is nil
		expectedErrorMsg   string             // expected error messages
	}{
		{
			"failure/invalidParameter/DCIMachine/nil",
			"error happened in when DCIMachine is nil",
			context.Background(),
			nil, // invalid parameter
			false,
			"expected DCIMachine is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, result := chInstance.GetLogicalServerFromDCIMachine(tt.inputCtx, tt.inputDciMachine)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetMachineFromNode is a function using testing module
// Target function : GetMachineFromNode()
func TestClient_GetMachineFromNode(t *testing.T) {
	t.Log("Start TestClient_GetMachineFromNode")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputNode          *corev1.Node    // input node
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/Node/nil",
			"error happened in when Node is nil",
			context.Background(),
			nil, // invalid parameter
			false,
			"expected node is not nil, but got nil",
		},
		{
			"failure/invalidParameter/Node/noAnnotations",
			"error happened in when Node has no annotations",
			context.Background(),
			&corev1.Node{}, // invalid parameter
			false,
			" annotations is not nil, but got nil",
		},
		{
			"failure/invalidParameter/Node/invalidMachineAnnotation",
			"error happened in when Node has expected annotation(capiv1.MachineAnnotation)",
			context.Background(),
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{}, // invalid parameter
				},
			},
			false,
			"failed to get Machine name from Node",
		},
		{
			"failure/invalidParameter/Node/invalidClusterAnnotation",
			"error happened in when Node has expected cluster info annotations",
			context.Background(),
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						capiv1.MachineAnnotation: "D.C. machine name",
						// invalid parameter(missing cluster annotations)
					},
				},
			},
			false,
			"failed to get Cluster info from Node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, result := chInstance.GetMachineFromNode(tt.inputCtx, tt.inputNode)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetMachineDeploymentFromMachine is a function using testing module
// Target function : GetMachineDeploymentFromMachine()
func TestClient_GetMachineDeploymentFromMachine(t *testing.T) {
	t.Log("Start TestClient_GetMachineDeploymentFromMachine")

	// test cases
	tests := []struct {
		caseName           string          // test case name
		caseDescription    string          // test case description
		inputCtx           context.Context // input context
		inputMachine       *capiv1.Machine // input machine
		inputNamespace     string          // input namespace
		expectedErrorIsNil bool            // whether error is nil
		expectedErrorMsg   string          // expected error messages
	}{
		{
			"failure/invalidParameter/Machine/nil",
			"error happened in when Machine is nil",
			context.Background(),
			nil, // invalid parameter
			"D.C. namespace",
			false,
			"expected machine is not nil, but got nil",
		},
		{
			"failure/invalidParameter/Machine/noLabels",
			"error happened in when Machine has no labels",
			context.Background(),
			&capiv1.Machine{}, // invalid parameter
			"D.C. namespace",
			false,
			" labels is not nil, but got nil",
		},
		{
			"failure/invalidParameter/Machine/invalidLabel",
			"error happened in when Machine has expected Label(capiv1.MachineAnnotation)",
			context.Background(),
			&capiv1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{}, // invalid parameter
				},
			},
			"D.C. namespace",
			false,
			"failed to get MachineDeployment name from Machine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, result := chInstance.GetMachineDeploymentFromMachine(tt.inputCtx, tt.inputMachine, tt.inputNamespace)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetMachineDeploymentsFromTemplate is a function using testing module
// Target function : GetMachineDeploymentsFromTemplate()
func TestClient_GetMachineDeploymentsFromTemplate(t *testing.T) {
	t.Log("Start TestClient_GetMachineDeploymentsFromTemplate")

	// test cases
	tests := []struct {
		caseName               string                    // test case name
		caseDescription        string                    // test case description
		inputCtx               context.Context           // input context
		inputNamespace         string                    // input namespace
		inputMachineDeployment *capiv1.MachineDeployment // input machine deployment
		inputTemplate          client.Object             // input template
		expectedErrorIsNil     bool                      // whether error is nil
		expectedErrorMsg       string                    // expected error messages
	}{
		{
			"success/filter/namespace",
			"namespace for MachineDeployment mismatch with template namespace",
			context.Background(),
			"default", // namespace for listing MachineDeployments
			nil,
			&capdv1.DCIMachineTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "D.C. name",
					Namespace: "other-namespace", // invalid parameter
				},
			},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// create mock client
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockClient := mock.NewMockClient(mockCtrl)
			mockClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, objList client.ObjectList, _ ...client.ListOption) error {
					// set empty list
					mdList, ok := objList.(*capiv1.MachineDeploymentList)
					if !ok {
						return errors.New("failed to convert to MachineDeploymentList")
					}
					mdList.Items = []capiv1.MachineDeployment{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "machine-deployment-name",
								Namespace: tt.inputNamespace,
							},
						},
					}
					return nil
				},
			)

			// call target function
			chInstance := &clientHelperStruct{managementClusterClient: mockClient}
			_, result := chInstance.GetMachineDeploymentsFromTemplate(tt.inputCtx, tt.inputNamespace, tt.inputMachineDeployment, tt.inputTemplate)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_CreateMachineDeploymentFromNodeScalingRequest is a function using testing module
// Target function : CreateMachineDeploymentFromNodeScalingRequest()
func TestClient_CreateMachineDeploymentFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestClient_CreateMachineDeploymentFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName           string                      // test case name
		caseDescription    string                      // test case description
		inputCtx           context.Context             // input context
		inputName          string                      // input name
		inputNamespace     string                      // input namespace
		inputDmtName       string                      // input DCI machine template name
		inputNsr           *scalev1.NodeScalingRequest // input NodeScalingRequest
		inputMd            *capiv1.MachineDeployment   // input MachineDeployment
		expectedErrorIsNil bool                        // whether error is nil
		expectedErrorMsg   string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in when NodeScalingRequest is nil",
			context.Background(),
			"D.C. name",
			"D.C. namespace",
			"D.C. DCI machine template name",
			nil, // invalid parameter
			nil, // invalid parameter
			false,
			"failed to create MachineDeployment spec from NodeScalingRequest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			result := chInstance.CreateMachineDeploymentFromNodeScalingRequest(tt.inputCtx, tt.inputName, tt.inputNamespace, tt.inputDmtName, tt.inputNsr, tt.inputMd)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_IsLastMachineDeployment is a function using testing module
// Target function : IsLastMachineDeployment()
func TestClient_IsLastMachineDeployment(t *testing.T) {
	t.Log("Start TestClient_IsLastMachineDeployment")

	// test cases
	tests := []struct {
		caseName               string                    // test case name
		caseDescription        string                    // test case description
		inputCtx               context.Context           // input context
		inputMachineDeployment *capiv1.MachineDeployment // input machine deployment
		inputNamespace         string                    // input namespace
		inputClusterName       string                    // input cluster name
		expectedErrorIsNil     bool                      // whether error is nil
		expectedErrorMsg       string                    // expected error messages
	}{
		{
			"failure/filter/namespace",
			"mdsInCluster should have md, but md not found in it",
			context.Background(),
			&capiv1.MachineDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine-deployment-name",
					Namespace: "default",
				},
			},
			"default", // namespace for listing MachineDeployments
			"cluster-name",
			false,
			"failed to find the target MachineDeployment in the List",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// create mock client
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockClient := mock.NewMockClient(mockCtrl)
			mockClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			// call target function
			chInstance := &clientHelperStruct{managementClusterClient: mockClient}
			_, result := chInstance.IsLastMachineDeployment(tt.inputCtx, tt.inputMachineDeployment, tt.inputNamespace, tt.inputClusterName)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// TestClient_GetClusterNameAndNamespaceFromNode is a function using testing module
// Target function : GetClusterNameAndNamespaceFromNode()
func TestClient_GetClusterNameAndNamespaceFromNode(t *testing.T) {
	t.Log("Start TestClient_GetClusterNameAndNamespaceFromNode")

	// test cases
	tests := []struct {
		caseName           string       // test case name
		caseDescription    string       // test case description
		inputNode          *corev1.Node // input node
		expectedErrorIsNil bool         // whether error is nil
		expectedErrorMsg   string       // expected error messages
	}{
		{
			"failure/invalidParameter/object/nil",
			"error happened in when object is nil",
			nil, // invalid parameter
			false,
			"expected node is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			chInstance := &clientHelperStruct{}
			_, _, result := chInstance.GetClusterNameAndNamespaceFromNode(tt.inputNode)

			if result == nil {
				if tt.expectedErrorIsNil {
					// expected result
					t.Log("got nil")
				} else {
					// unexpected result
					t.Errorf("got nil, but expected result '%v'", tt.expectedErrorMsg)
				}
			} else {
				if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
					// unexpected result
					t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
				} else {
					// expected result
					t.Logf("got '%v'", result)
				}
			}
		})
	}
}

// generateDummyKubeconfig creates a temporary kubeconfig file with dummy data
func generateDummyKubeconfig() (string, error) {
	kubeconfig := api.NewConfig()
	contextName := "management-cluster-context"
	clusterName := "management-cluster"
	userName := "scale-controller-sa"
	kubeconfig.Clusters[clusterName] = &api.Cluster{
		Server:                   "dummy",
		CertificateAuthorityData: []byte("dummy"),
		InsecureSkipTLSVerify:    false,
	}
	kubeconfig.Contexts[contextName] = &api.Context{
		Cluster:  clusterName,
		AuthInfo: userName,
	}
	kubeconfig.AuthInfos[userName] = &api.AuthInfo{
		Token:                 "dummy",
		ClientCertificateData: []byte("dummy"),
		ClientKeyData:         []byte("dummy"),
	}
	kubeconfig.CurrentContext = contextName

	tmpFile, err := os.CreateTemp("", "temp-envtest-*.kubeconfig")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmpFile.Close()
	}()
	err = clientcmd.WriteToFile(*kubeconfig, tmpFile.Name())
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
