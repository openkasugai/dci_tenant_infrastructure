package utils

import (
	"strings"
	"testing"

	resourcev1 "k8s.io/api/resource/v1beta2"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"

	scalev1 "scale-controller/api/v1alpha1"
)

// THIS FILE IS IMPLEMENTED TO INCREASE COVERAGE FOR FAILSAFE CASES IN utils PACKAGE
// ONLY FOR PARTS THAT CANNOT BE TESTED WITH envtest FOR internal/controller.

// TestUtils_createRandomString is a function using testing module
// Target function : createRandomString()
func TestUtils_createRandomString(t *testing.T) {
	t.Log("Start TestUtils_createRandomString")

	// test cases
	tests := []struct {
		caseName        string // test case name
		caseDescription string // test case description
		inputLength     int    // input length of generated string
		expectedLength  int    // expected length of generated string
	}{
		{
			"invalidLength/zero",
			"if input length is zero, default length is used",
			0,
			defaultSuffixLength,
		},
		{
			"invalidLength/negative",
			"if input length is negative, default length is used",
			-1,
			defaultSuffixLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := createRandomString(tt.inputLength)

			if len(result) != tt.expectedLength {
				// unexpected length
				t.Errorf("got %v, but expected result %v(length)", result, tt.expectedLength)
			} else {
				// expected length
				t.Logf("got %v", result)
			}
		})
	}
}

// TestUtils_isSameGPUSpec is a function using testing module
// Target function : isSameGPUSpec()
func TestUtils_isSameGPUSpec(t *testing.T) {
	t.Log("Start TestUtils_isSameGPUSpec")

	// test cases
	tests := []struct {
		caseName        string         // test case name
		caseDescription string         // test case description
		inputRequest    *rpmv1.GPUSpec // input request object
		inputTemplate   *rpmv1.GPUSpec // input template object
		expectedResult  bool           // expected result
	}{
		{
			"request/nil,template/nil",
			"should return true when both of request and template are nil",
			nil,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := isSameGPUSpec(tt.inputRequest, tt.inputTemplate)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_IsAllocatableServerSpec is a function using testing module
// Target function : IsAllocatableServerSpec()
func TestUtils_IsAllocatableServerSpec(t *testing.T) {
	t.Log("Start TestUtils_IsAllocatableServerSpec")

	// test cases
	tests := []struct {
		caseName        string            // test case name
		caseDescription string            // test case description
		inputRequest    *rpmv1.ServerSpec // input request object
		inputTemplate   *rpmv1.ServerSpec // input template object
		expectedResult  bool              // expected result
	}{
		{
			"request/nil,template/d.c.",
			"should return true when request is nil",
			nil,
			nil,
			true,
		},
		{
			"request/non-nil,template/nil",
			"should return false when request is non-nil and template is nil",
			&rpmv1.ServerSpec{},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := IsAllocatableServerSpec(tt.inputRequest, tt.inputTemplate)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_IsSameServerSpec is a function using testing module
// Target function : IsSameServerSpec()
func TestUtils_IsSameServerSpec(t *testing.T) {
	t.Log("Start TestUtils_IsSameServerSpec")

	// test cases
	tests := []struct {
		caseName        string            // test case name
		caseDescription string            // test case description
		inputRequest    *rpmv1.ServerSpec // input request object
		inputTemplate   *rpmv1.ServerSpec // input template object
		expectedResult  bool              // expected result
	}{
		{
			"request/nil,template/nil",
			"should return true when both of request and template are nil",
			nil,
			nil,
			true,
		},
		{
			"request/nil,template/non-nil",
			"should return false when request is nil and template is non-nil",
			nil,
			&rpmv1.ServerSpec{},
			false,
		},
		{
			"request/non-nil,template/nil",
			"should return false when request is non-nil and template is nil",
			&rpmv1.ServerSpec{},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := IsSameServerSpec(tt.inputRequest, tt.inputTemplate)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_HasAllUUIDsInLogicalServer is a function using testing module
// Target function : HasAllUUIDsInLogicalServer()
func TestUtils_HasAllUUIDsInLogicalServer(t *testing.T) {
	t.Log("Start TestUtils_HasAllUUIDsInLogicalServer")

	// test cases
	tests := []struct {
		caseName        string         // test case name
		caseDescription string         // test case description
		inputUUIDs      []string       // input UUIDs
		inputGPUSpec    *rpmv1.GPUSpec // input GPUSpec object
		expectedResult  bool           // expected result
	}{
		{
			"gpuSpec/nil,subset/nil",
			"should return true when UUIDs slice is nil",
			nil,
			nil,
			true,
		},
		{
			"gpuSpec/non-nil,subset/nil",
			"should return false when UUIDs slice is non-nil and GPUSpec object is nil",
			[]string{"dummy-uuid"},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := HasAllUUIDsInLogicalServer(tt.inputUUIDs, tt.inputGPUSpec)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_HasAnyUUIDsInLogicalServer is a function using testing module
// Target function : HasAnyUUIDsInLogicalServer()
func TestUtils_HasAnyUUIDsInLogicalServer(t *testing.T) {
	t.Log("Start TestUtils_HasAnyUUIDsInLogicalServer")

	// test cases
	tests := []struct {
		caseName        string         // test case name
		caseDescription string         // test case description
		inputUUIDs      []string       // input UUIDs
		inputGPUSpec    *rpmv1.GPUSpec // input GPUSpec object
		expectedResult  bool           // expected result
	}{
		{
			"gpuSpec/nil,subset/nil",
			"should return true when UUIDs slice is nil",
			nil,
			nil,
			false,
		},
		{
			"gpuSpec/non-nil,subset/nil",
			"should return false when UUIDs slice is non-nil and GPUSpec object is nil",
			[]string{"dummy-uuid"},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := HasAnyUUIDsInLogicalServer(tt.inputUUIDs, tt.inputGPUSpec)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_IsSameDeviceTaintSelector is a function using testing module
// Target function : IsSameDeviceTaintSelector()
func TestUtils_IsSameDeviceTaintSelector(t *testing.T) {
	t.Log("Start TestUtils_IsSameDeviceTaintSelector")

	// test cases
	tests := []struct {
		caseName        string                          // test case name
		caseDescription string                          // test case description
		inputSelector1  *resourcev1.DeviceTaintSelector // input DeviceTaintSelector object 1
		inputSelector2  *resourcev1.DeviceTaintSelector // input DeviceTaintSelector object 2
		expectedResult  bool                            // expected result
	}{
		{
			"selector1/nil,selector2/nil",
			"should return true when both DeviceTaintSelector objects are nil",
			nil,
			nil,
			true,
		},
		{
			"selector1/non-nil,selector2/nil",
			"should return false when the first DeviceTaintSelector object is non-nil and the second is nil",
			&resourcev1.DeviceTaintSelector{},
			nil,
			false,
		},
		{
			"selector1/nil,selector2/non-nil",
			"should return false when the first DeviceTaintSelector object is nil and the second is non-nil",
			nil,
			&resourcev1.DeviceTaintSelector{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			result := IsSameDeviceTaintSelector(tt.inputSelector1, tt.inputSelector2)

			if result != tt.expectedResult {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedResult)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_NewDCIMachineTemplateFromCardScalingRequest is a function using testing module
// Target function : NewDCIMachineTemplateFromCardScalingRequest()
func TestUtils_NewDCIMachineTemplateFromCardScalingRequest(t *testing.T) {
	t.Log("Start TestUtils_NewDCIMachineTemplateFromCardScalingRequest")

	// test cases
	tests := []struct {
		caseName            string                      // test case name
		caseDescription     string                      // test case description
		inputCsr            *scalev1.CardScalingRequest // input csr object
		inputDCIMachineSpec *capdv1.DCIMachineSpec      // input DCIMachineSpec object
		expectedErrorMsg    string                      // expected error messages
	}{
		{
			"failure/invalidParameter/CardScalingRequest/nil",
			"error happened in create new resource when CardScalingRequest is nil",
			nil,
			&capdv1.DCIMachineSpec{},
			"expected CardScalingRequest and DCIMachineSpec are not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := NewDCIMachineTemplateFromCardScalingRequest(tt.inputCsr, tt.inputDCIMachineSpec)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_NewDCIMachineTemplateFromNodeScalingRequest is a function using testing module
// Target function : NewDCIMachineTemplateFromNodeScalingRequest()
func TestUtils_NewDCIMachineTemplateFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestUtils_NewDCIMachineTemplateFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName         string                      // test case name
		caseDescription  string                      // test case description
		inputName        string                      // input name
		inputNamespace   string                      // input namespace
		inputNsr         *scalev1.NodeScalingRequest // input nsr object
		expectedErrorMsg string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in create new resource when NodeScalingRequest is nil",
			"D.C. name",
			"D.C. namespace",
			nil,
			"expected NodeScalingRequest and ServerSpec is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := NewDCIMachineTemplateFromNodeScalingRequest(tt.inputName, tt.inputNamespace, tt.inputNsr)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_NewDockerMachineTemplateFromNodeScalingRequest is a function using testing module
// Target function : NewDockerMachineTemplateFromNodeScalingRequest()
func TestUtils_NewDockerMachineTemplateFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestUtils_NewDockerMachineTemplateFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName         string                      // test case name
		caseDescription  string                      // test case description
		inputName        string                      // input name
		inputNamespace   string                      // input namespace
		inputNsr         *scalev1.NodeScalingRequest // input nsr object
		expectedErrorMsg string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in create new resource when NodeScalingRequest is nil",
			"D.C. name",
			"D.C. namespace",
			nil,
			"expected NodeScalingRequest and Location is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := NewDockerMachineTemplateFromNodeScalingRequest(tt.inputName, tt.inputNamespace, tt.inputNsr)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_NewMachineDeploymentFromCardScalingRequest is a function using testing module
// Target function : NewMachineDeploymentFromCardScalingRequest()
func TestUtils_NewMachineDeploymentFromCardScalingRequest(t *testing.T) {
	t.Log("Start TestUtils_NewMachineDeploymentFromCardScalingRequest")

	// test cases
	tests := []struct {
		caseName               string                      // test case name
		caseDescription        string                      // test case description
		inputCsr               *scalev1.CardScalingRequest // input csr object
		inputMachineDeployment *capiv1.MachineDeployment   // input MachineDeployment object
		expectedErrorMsg       string                      // expected error messages
	}{
		{
			"failure/invalidParameter/CardScalingRequest/nil",
			"error happened in create new resource when CardScalingRequest is nil",
			nil,
			nil,
			"expected CardScalingRequest and its NewDCIMachineInfo are not nil, but got nil",
		},
		{
			"failure/invalidParameter/CardScalingRequest/Status/NewDCIMachineInfo/nil",
			"error happened in create new resource when CardScalingRequest's NewDCIMachineInfo is nil",
			&scalev1.CardScalingRequest{},
			nil,
			"expected CardScalingRequest and its NewDCIMachineInfo are not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := NewMachineDeploymentFromCardScalingRequest(tt.inputCsr, tt.inputMachineDeployment)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_NewMachineDeploymentFromNodeScalingRequest is a function using testing module
// Target function : NewMachineDeploymentFromNodeScalingRequest()
func TestUtils_NewMachineDeploymentFromNodeScalingRequest(t *testing.T) {
	t.Log("Start TestUtils_NewMachineDeploymentFromNodeScalingRequest")

	// test cases
	tests := []struct {
		caseName                 string                      // test case name
		caseDescription          string                      // test case description
		inputName                string                      // input name
		inputNamespace           string                      // input namespace
		inputMachineTemplateName string                      // input machine template name
		inputNsr                 *scalev1.NodeScalingRequest // input nsr object
		inputMachineDeployment   *capiv1.MachineDeployment   // input MachineDeployment object
		expectedErrorMsg         string                      // expected error messages
	}{
		{
			"failure/invalidParameter/NodeScalingRequest/nil",
			"error happened in create new resource when NodeScalingRequest is nil",
			"D.C. name",
			"D.C. namespace",
			"D.C. machine template name",
			nil,
			nil,
			"expected NodeScalingRequest is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := NewMachineDeploymentFromNodeScalingRequest(tt.inputName, tt.inputNamespace, tt.inputMachineTemplateName, tt.inputNsr, tt.inputMachineDeployment)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_GenerateServerSpecAfterScaling is a function using testing module
// Target function : GenerateServerSpecAfterScaling()
func TestUtils_GenerateServerSpecAfterScaling(t *testing.T) {
	t.Log("Start TestUtils_GenerateServerSpecAfterScaling")

	// test cases
	tests := []struct {
		caseName              string                    // test case name
		caseDescription       string                    // test case description
		inputServerSpec       *rpmv1.ServerSpec         // input server spec object
		inputScalingParameter *scalev1.ScalingParameter // input scaling parameter object
		expectedErrorMsg      string                    // expected error messages
	}{
		{
			"failure/invalidParameter/ServerSpec/nil",
			"error happened in create new resource when ServerSpec is nil",
			nil,
			nil,
			"expected ServerSpec and ScalingParameter are not nil, but got nil",
		},
		{
			"failure/invalidParameter/ScalingParameter/NumAndUUIDsNotSpecified",
			"error happened in create new resource when both of Num and UUIDs for ScalingParameter are not specified",
			&rpmv1.ServerSpec{GPU: &rpmv1.GPUSpec{}},
			&scalev1.ScalingParameter{},
			"expected either Num or UUIDs is specified in scaling parameter, but both are not specified",
		},
		{
			"failure/invalidParameter/ScalingParameter/InvalidOp",
			"error happened in create new resource when ScalingParameter has invalid operation",
			&rpmv1.ServerSpec{GPU: &rpmv1.GPUSpec{}},
			&scalev1.ScalingParameter{
				Num: 2,
				Op:  capdv1.DCIMachineCardScalingOp("InvalidOp"),
			},
			"unsupported scaling parameter operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := GenerateServerSpecAfterScaling(tt.inputServerSpec, tt.inputScalingParameter)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}

// TestUtils_GenerateDCIMachineSpecAfterScaling is a function using testing module
// Target function : GenerateDCIMachineSpecAfterScaling()
func TestUtils_GenerateDCIMachineSpecAfterScaling(t *testing.T) {
	t.Log("Start TestUtils_GenerateDCIMachineSpecAfterScaling")

	// test cases
	tests := []struct {
		caseName            string                      // test case name
		caseDescription     string                      // test case description
		inputDCIMachineSpec *capdv1.DCIMachineSpec      // input DCI machine spec object
		inputCsr            *scalev1.CardScalingRequest // input csr object
		expectedErrorMsg    string                      // expected error messages
	}{
		{
			"failure/invalidParameter/DCIMachineSpec/nil",
			"error happened in create new resource when DCIMachineSpec is nil",
			nil,
			nil,
			"expected DCIMachineSpec is not nil, but got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)

			// call target function
			_, result := GenerateDCIMachineSpecAfterScaling(tt.inputDCIMachineSpec, tt.inputCsr)

			if !strings.Contains(result.Error(), tt.expectedErrorMsg) {
				// unexpected result
				t.Errorf("got '%v', but expected result '%v'", result, tt.expectedErrorMsg)
			} else {
				// expected result
				t.Logf("got '%v'", result)
			}
		})
	}
}
