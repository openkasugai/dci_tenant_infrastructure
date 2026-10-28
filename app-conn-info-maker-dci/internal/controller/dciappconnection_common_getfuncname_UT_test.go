package controller

import (
	"fmt"
	"testing"
)

func TestUTGetFuncNameCallerNg(t *testing.T) {
	// ------------------------------
	//  TestUTGetFuncNameCallerNg: runtimeCallerFunc returns ok=false
	//  Expected behavior: Returns empty string
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_common_getfuncname_UT_test.go")
	fmt.Println(">>>>> Start TestUTGetFuncNameCallerNg")
	fmt.Println(">>>>>")

	orgRuntimeCallerFunc := runtimeCallerFunc
	runtimeCallerFunc = func(skip int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	if got := getFuncName(); got != "" {
		t.Fatalf("getFuncName() = %q, want empty string", got)
	} else {
		fmt.Println(">>>>> Result:", got)
	}

	runtimeCallerFunc = orgRuntimeCallerFunc

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End TestUTGetFuncNameCallerNg")
	fmt.Println("<<<<<")
}
