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

package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/exec"
	"testing"

	"k8s.io/client-go/rest"

	"scale-controller/internal/alertreceiver/cardalert"
	receiver "scale-controller/pkg/alertreceiver"
)

// arguments for backup of original global variables
var (
	origNewAlertReceiver = receiver.NewAlertReceiver
	origInClusterConfig  = cardalert.InClusterConfig
)

// TestAlertReceiver_main is a white-box-test function using testing module
// Target function : run()
func TestAlertReceiver_run(t *testing.T) {
	t.Log("Start TestAlertReceiver_run")
	defer func() {
		t.Log("Restore original global variables")
		receiver.NewAlertReceiver = origNewAlertReceiver
		cardalert.InClusterConfig = origInClusterConfig
	}()

	// test cases
	tests := []struct {
		caseName        string // test case name
		caseDescription string // test case description
		expectedCode    int    // test result code
		prepareFunc     func() // function to prepare for test case
		cleanupFunc     func() // function to cleanup after test case
	}{
		{
			"failure/cardalert/NewAlertReceiver",
			"error happened in cardalert.NewAlertReceiver()",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				os.Setenv(cardAlertNamespaceEnvVarName, "default")
				// Set mock function
				cardalert.InClusterConfig = newMockInClusterConfig()
				receiver.NewAlertReceiver = newMockNewAlertReceiver(errors.New("mock error in NewAlertReceiver"), nil)
			},
			func() {
				os.Unsetenv(cardAlertNamespaceEnvVarName)
			},
		},
		{
			"failure/alertReceiver/Serve",
			"error happened in alertReceiver.Serve()",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				os.Setenv(cardAlertNamespaceEnvVarName, "default")
				// Set mock function
				cardalert.InClusterConfig = newMockInClusterConfig()
				receiver.NewAlertReceiver = newMockNewAlertReceiver(nil, errors.New("mock error in Serve"))
			},
			func() {
				os.Unsetenv(cardAlertNamespaceEnvVarName)
				cardalert.InClusterConfig = origInClusterConfig
				receiver.NewAlertReceiver = origNewAlertReceiver
			},
		},
		{
			"success",
			"succeed to run alert receiver",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				os.Setenv(cardAlertNamespaceEnvVarName, "default")
				// Set mock function
				cardalert.InClusterConfig = newMockInClusterConfig()
				receiver.NewAlertReceiver = newMockNewAlertReceiver(nil, nil)
			},
			func() {
				os.Unsetenv(cardAlertNamespaceEnvVarName)
				cardalert.InClusterConfig = origInClusterConfig
				receiver.NewAlertReceiver = origNewAlertReceiver
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)
			// execute prepare function if exists
			if tt.prepareFunc != nil {
				tt.prepareFunc()
			}

			// call target function
			code := run()

			// execute cleanup function if exists
			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}

			if code != tt.expectedCode {
				// unexpected result
				t.Errorf("got %v, but expected error state %v", code, tt.expectedCode)
			} else {
				// expected result
				t.Logf("got expected error state %v", code)
			}
		})
	}
}

// TestAlertReceiver_main is a black-box-test function using testing module
// Target function : main()
func TestAlertReceiver_main(t *testing.T) {
	t.Log("Start TestAlertReceiver_main")

	// test cases
	tests := []struct {
		caseName        string // test case name
		caseDescription string // test case description
		cmd             func() *exec.Cmd
		expectedCode    int    // expected exit code from run()
		prepareFunc     func() // function to prepare for test case
	}{
		{
			"failure/exitCode/1",
			"exit with code 1 when run() failed",
			func() *exec.Cmd {
				cmd := exec.Command(os.Args[0], "-test.run=TestAlertReceiver_main_1_helper")
				cmd.Env = append(os.Environ(), "ENV_IS_ENABLED_HELPER=1")
				return cmd
			},
			1,
			func() {},
		},
		{
			"success/exitCode/0",
			"exit with code 0 when run() succeeded",
			func() *exec.Cmd {
				cmd := exec.Command(os.Args[0], "-test.run=TestAlertReceiver_main_0_helper")
				cmd.Env = append(os.Environ(), "ENV_IS_ENABLED_HELPER=1")
				return cmd
			},
			0,
			func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			t.Log(tt.caseDescription)
			// execute prepare function if exists
			if tt.prepareFunc != nil {
				tt.prepareFunc()
			}

			// call target function
			err := tt.cmd().Run()

			if err == nil {
				exitCode := 0
				if exitCode != tt.expectedCode {
					t.Fatalf("got nil error, but expected exit code %d", tt.expectedCode)
				} else {
					t.Logf("got expected exit code %d", exitCode)
				}
			} else {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("got %v, but expected ExitError", err)
				}

				if exitCode := exitErr.ExitCode(); exitCode != tt.expectedCode {
					// unexpected exit code
					t.Errorf("got %d, but expected exit code %d", exitCode, tt.expectedCode)
				} else {
					// expected exit code
					t.Logf("got expected exit code %d", exitCode)
				}
			}
		})
	}
}

// TestAlertReceiver_main_1_helper is a helper function for TestAlertReceiver_main
// to test the case when run() returns error code and main() exits with code 1.
func TestAlertReceiver_main_1_helper(t *testing.T) {
	if os.Getenv("ENV_IS_ENABLED_HELPER") != "1" {
		t.Skip("skip helper in parent process")
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Unsetenv(cardAlertNamespaceEnvVarName) // Unset required environment variable to cause error

	main()
}

// TestAlertReceiver_main_0_helper is a helper function for TestAlertReceiver_main
// to test the case when run() returns error code and main() exits with code 0.
func TestAlertReceiver_main_0_helper(t *testing.T) {
	if os.Getenv("ENV_IS_ENABLED_HELPER") != "1" {
		t.Skip("skip helper in parent process")
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Setenv(cardAlertNamespaceEnvVarName, "default")
	// Set mock function
	cardalert.InClusterConfig = newMockInClusterConfig()
	receiver.NewAlertReceiver = newMockNewAlertReceiver(nil, nil)

	main()

	// main will not exit when it succeeds, so we call os.Exit(0) explicitly to indicate success after main() returns
	os.Exit(0)
}

type mockAlertReceiver struct {
	serveError error
}

func (m *mockAlertReceiver) Serve(_ context.Context) error {
	return m.serveError
}

func newMockNewAlertReceiver(newError, serveError error) receiver.AlertReceiverNewFunc {
	return func(_ *receiver.AlertReceiverNewOptions) (receiver.AlertReceiver, error) {
		return &mockAlertReceiver{serveError: serveError}, newError
	}
}

func newMockInClusterConfig() func() (*rest.Config, error) {
	return func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}
}
