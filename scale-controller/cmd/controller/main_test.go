package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/ptr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	cfg "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	scalev1 "scale-controller/api/v1alpha1"
	"scale-controller/internal/helper/utils"
	webhookv1 "scale-controller/internal/webhook/v1alpha1"
	"scale-controller/test/testutils/mock"
)

// arguments for backup of original global variables
var (
	origOsArgs                                = slices.Clone(os.Args)
	origNewManager                            = ctrl.NewManager
	origGetConfigOrDie                        = ctrl.GetConfigOrDie
	origNewControllerManagedBy                = ctrl.NewControllerManagedBy
	origNewNodeScalingRequestWebhookManagedBy = webhookv1.NewNodeScalingRequestWebhookManagedBy
	origNewCardScalingRequestWebhookManagedBy = webhookv1.NewCardScalingRequestWebhookManagedBy
	origSetupSignalHandler                    = ctrl.SetupSignalHandler
)

// TestScaleController_run is a white-box-test function using testing module
// Target function : run()
func TestScaleController_run(t *testing.T) {
	t.Log("Start TestScaleController_run")
	// backup original values
	var (
		c      = gomock.NewController(t)
		zl     = zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = zerologr.New(&zl)
	)
	defer func() {
		t.Log("Restore original global variables")
		os.Args = origOsArgs
		ctrl.GetConfigOrDie = origGetConfigOrDie
		ctrl.NewManager = origNewManager
		ctrl.NewControllerManagedBy = origNewControllerManagedBy
		webhookv1.NewNodeScalingRequestWebhookManagedBy = origNewNodeScalingRequestWebhookManagedBy
		webhookv1.NewCardScalingRequestWebhookManagedBy = origNewCardScalingRequestWebhookManagedBy
		ctrl.SetupSignalHandler = origSetupSignalHandler

		t.Log("Finish gomock controller")
		c.Finish()
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

	t.Log("Generate dummy cert/key pair")
	pemDir, err := generateDummyCertKeyPair()
	if err != nil {
		t.Fatalf("failed to generate dummy cert/key pair: %v", err)
	}
	defer func() {
		t.Log("Remove dummy cert/key pair")
		os.RemoveAll(pemDir)
	}()

	// test cases
	tests := []struct {
		caseName        string // test case name
		caseDescription string // test case description
		expectedCode    int    // expected exit code from run()
		prepareFunc     func() // function to prepare for test case
	}{
		{
			"failure/invalidCertPath/certManager",
			"error happened in initializing webhook certificate watcher",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--webhook-cert-path=dummy_path",
				}
				// replace global variables
				os.Args = args
			},
		},
		{
			"failure/invalidCertPath/metrics",
			"error happened in initializing metrics certificate watcher",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--metrics-cert-path=dummy_path",
				}
				// replace global variables
				os.Args = args
			},
		},
		{
			"failure/ctrl.NewManager",
			"error happened in creating manager instance",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(nil)
			},
		},
		{
			"failure/setupWithManager/getClientHelperOption/GetProviderTypeFromEnv/empty",
			"error happened in utils.GetProviderTypeFromEnv",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				// setup environment variables
				_ = os.Setenv(utils.ProviderTypeEnvVarName, "")
			},
		},
		{
			"failure/setupWithManager/NewClientHelper",
			"error happened in helperclient.NewClientHelper",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("MANAGEMENT_KUBECONFIG")
			},
		},
		{
			"failure/setupWithManagerNodeScalingRequest/controller",
			"error happened in NodeScalingRequestReconciler.SetupWithManager",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Setenv("NSR_TTL_SECONDS_AFTER_FINISHED", "")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"failure/setupWithManagerNodeScalingRequest/webhook",
			"error happened in NodeScalingRequestValidator.SetupWithManager",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy([]int{1})
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"failure/setupWithManagerCardScalingRequest/controller",
			"error happened in CardScalingRequestReconciler.SetupWithManager",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("CSR_TTL_SECONDS_AFTER_FINISHED", "")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"failure/setupWithManagerCardScalingRequest/webhook",
			"error happened in CardScalingRequestValidator.SetupWithManager",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy([]int{1})
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"failure/setupWithManagerCardAlert/controller",
			"error happened in CardAlertReconciler.SetupWithManager",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"failure/setupWithManager/mgr.Add/metrics",
			"error happened in mgr.Add for metrics",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--metrics-cert-path=" + pemDir,
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(errors.New("dummy-error"))
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"failure/setupWithManager/mgr.Add/webhook",
			"error happened in mgr.Add for metrics",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--webhook-cert-path=" + pemDir,
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(errors.New("dummy-error"))
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"failure/setupWithManager/mgr.AddHealthzCheck",
			"error happened in mgr.AddHealthzCheck",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("dummy-error"))
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"failure/setupWithManager/mgr.AddReadyzCheck",
			"error happened in mgr.AddReadyzCheck",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("dummy-error"))
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"failure/mgr.Start",
			"error happened in mgr.Start",
			1,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().Start(gomock.Any()).AnyTimes().Return(errors.New("dummy-error"))
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				ctrl.SetupSignalHandler = newDummySetupSignalHandler()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"success/disableWebhooks",
			"succeed to run with disabling webhooks",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().Start(gomock.Any()).AnyTimes().Return(nil)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				ctrl.SetupSignalHandler = newDummySetupSignalHandler()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "false")
			},
		},
		{
			"success/enableWebhooks/dci",
			"succeed to run with enabling webhooks for DCI provider",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--webhook-cert-path=" + pemDir,
					"--metrics-cert-path=" + pemDir,
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().Start(gomock.Any()).AnyTimes().Return(nil)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				ctrl.SetupSignalHandler = newDummySetupSignalHandler()
				// setup environment variables
				_ = os.Unsetenv(utils.ProviderTypeEnvVarName)
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"success/enableWebhooks/docker",
			"succeed to run with enabling webhooks for Docker provider",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--webhook-cert-path=" + pemDir,
					"--metrics-cert-path=" + pemDir,
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().Start(gomock.Any()).AnyTimes().Return(nil)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				ctrl.SetupSignalHandler = newDummySetupSignalHandler()
				// setup environment variables
				_ = os.Setenv(utils.ProviderTypeEnvVarName, "docker")
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("ENABLE_CARD_ALERT", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
			},
		},
		{
			"success/disableCardScalingRequest",
			"succeed to run with disabling CardScalingRequest",
			0,
			func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				args := []string{
					os.Args[0],
					"--webhook-cert-path=" + pemDir,
					"--metrics-cert-path=" + pemDir,
				}
				// create mock for manager
				mgr := mock.NewMockManager(c)
				mgr.EXPECT().GetClient().AnyTimes().Return(nil)
				mgr.EXPECT().GetScheme().AnyTimes().Return(nil)
				mgr.EXPECT().GetEventRecorder(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().GetLogger().AnyTimes().Return(logger)
				mgr.EXPECT().Add(gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddHealthzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().AddReadyzCheck(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				mgr.EXPECT().Start(gomock.Any()).AnyTimes().Return(nil)
				// replace global variables
				os.Args = args
				ctrl.GetConfigOrDie = newDummyGetConfigOrDie()
				ctrl.NewManager = newDummyNewManager(mgr)
				ctrl.NewControllerManagedBy = newDummyNewControllerManagedBy()
				webhookv1.NewNodeScalingRequestWebhookManagedBy = newDummyNewNodeScalingRequestWebhookManagedBy(nil)
				webhookv1.NewCardScalingRequestWebhookManagedBy = newDummyNewCardScalingRequestWebhookManagedBy(nil)
				ctrl.SetupSignalHandler = newDummySetupSignalHandler()
				// setup environment variables
				_ = os.Setenv(utils.ProviderTypeEnvVarName, "docker")
				_ = os.Unsetenv("NSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CSR_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Unsetenv("CARD_ALERT_TTL_SECONDS_AFTER_FINISHED")
				_ = os.Setenv("MACHINE_ALERT_NAMESPACE", "default")
				_ = os.Setenv("ENABLE_CARD_ALERT", "true")
				_ = os.Setenv("ENABLE_CARD_SCALING_REQUEST", "false")
				_ = os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
				_ = os.Setenv("ENABLE_WEBHOOKS", "true")
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
			exitCode := run()

			if exitCode != tt.expectedCode {
				// unexpected exit code
				t.Errorf("got %d, but expected exit code %d", exitCode, tt.expectedCode)
			} else {
				// expected exit code
				t.Logf("got expected exit code %d", exitCode)
			}
		})
	}
}

func TestScaleController_main(t *testing.T) {
	t.Log("Start TestScaleController_run")

	t.Log("Generate dummy kubeconfig")
	kubeconfigPath, err := generateDummyKubeconfig()
	if err != nil {
		t.Fatalf("failed to create dummy kubeconfig: %v", err)
	}
	defer func() {
		t.Log("Remove dummy kubeconfig")
		os.Remove(kubeconfigPath)
	}()

	t.Log("Generate dummy cert/key pair")
	pemDir, err := generateDummyCertKeyPair()
	if err != nil {
		t.Fatalf("failed to generate dummy cert/key pair: %v", err)
	}
	defer func() {
		t.Log("Remove dummy cert/key pair")
		os.RemoveAll(pemDir)
	}()

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
				cmd := exec.Command(os.Args[0], "-test.run=TestScaleController_main_1_helper")
				cmd.Env = append(os.Environ(),
					"ENV_IS_ENABLED_HELPER=1",
					"ENV_KUBECONFIG="+kubeconfigPath,
					"ENV_PEM_DIR="+pemDir,
				)
				return cmd
			},
			1,
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
			}

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
		})
	}
}

func TestScaleController_main_1_helper(t *testing.T) {
	if os.Getenv("ENV_IS_ENABLED_HELPER") != "1" {
		t.Skip("skip helper in parent process")
	}

	kubeconfigPath := os.Getenv("ENV_KUBECONFIG")
	pemDir := os.Getenv("ENV_PEM_DIR")
	os.Args = append(os.Args,
		"--kubeconfig="+kubeconfigPath,
		"--metrics-cert-path="+pemDir,
	)
	os.Setenv("MANAGEMENT_KUBECONFIG", kubeconfigPath)
	os.Setenv("ENABLE_CARD_ALERT", "false")
	os.Setenv("ENABLE_CARD_SCALING_REQUEST", "false")

	main()

	// main will exit with code 1, so if we reach here, it means the exit code is not as expected
	os.Exit(0)
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

// generateDummyPemFile creates a temporary PEM file with given content
func generateDummyPemFile(fileName string, certPEM []byte) error {
	// create temporary file
	tmpFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	// write certificate content to the file
	if _, err := tmpFile.Write(certPEM); err != nil {
		return err
	}
	return nil
}

// generateDummyCertKeyPair generates a dummy self-signed certificate and private key
// into temporary files. It returns the directory path containing the files.
func generateDummyCertKeyPair() (string, error) {
	// generate a dummy self-signed certificate and private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	// create a template for the certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "dummy",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	// create the self-signed certificate PEM data
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", err
	}
	certPEMData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	// create the private key PEM data
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	keyPEMData := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	// create temporary directory to store cert and key files
	pemDir, err := os.MkdirTemp("/tmp", "dummycert-*")
	if err != nil {
		return "", err
	}
	certPEMPath := filepath.Join(pemDir, "tls.crt")
	keyPEMPath := filepath.Join(pemDir, "tls.key")

	// generate cert file
	err = generateDummyPemFile(certPEMPath, certPEMData)
	if err != nil {
		return "", err
	}

	// generate key file
	err = generateDummyPemFile(keyPEMPath, keyPEMData)
	if err != nil {
		return "", err
	}

	return pemDir, nil
}

// newDummyNewManager returns a dummy function to replace ctrl.NewManager
func newDummyNewManager(mgr manager.Manager) func(*rest.Config, manager.Options) (manager.Manager, error) {
	if mgr == nil {
		return func(_ *rest.Config, _ manager.Options) (manager.Manager, error) {
			return nil, errors.New("mocked error in ctrl.NewManager")
		}
	}
	return func(_ *rest.Config, _ manager.Options) (manager.Manager, error) {
		return mgr, nil
	}
}

// newDummyGetConfigOrDie returns a dummy function to replace ctrl.GetConfigOrDie
func newDummyGetConfigOrDie() func() *rest.Config {
	return func() *rest.Config {
		return &rest.Config{}
	}
}

// newDummyNewControllerManagedBy returns a dummy function to replace ctrl.NewControllerManagedBy
func newDummyNewControllerManagedBy() func(manager.Manager) *ctrl.Builder {
	s := runtime.NewScheme()
	if err := scalev1.AddToScheme(s); err != nil {
		panic(err)
	}
	return func(_ manager.Manager) *ctrl.Builder {
		m, _ := origNewManager(&rest.Config{}, ctrl.Options{
			Scheme: s,
			Controller: cfg.Controller{
				SkipNameValidation: ptr.To(true),
			},
		})
		return origNewControllerManagedBy(m).
			WithOptions(controller.Options{SkipNameValidation: ptr.To(true)})
	}
}

// newDummyNewWebhookManagedBy returns a dummy function to replace ctrl.NewWebhookManagedBy
func newDummyNewWebhookManagedBy[T runtime.Object](
	errorCalls []int, defaultFunc func(manager.Manager) *builder.WebhookBuilder[T],
) func(manager.Manager) *builder.WebhookBuilder[T] {
	countCall := 0 // closure variable to count calls
	s := runtime.NewScheme()
	if err := scalev1.AddToScheme(s); err != nil {
		panic(err)
	}
	errorScheme := runtime.NewScheme()
	return func(_ manager.Manager) *builder.WebhookBuilder[T] {
		usingScheme := s
		countCall++
		if slices.Contains(errorCalls, countCall) {
			usingScheme = errorScheme
		}
		m, _ := origNewManager(&rest.Config{}, ctrl.Options{
			Scheme: usingScheme,
			Controller: cfg.Controller{
				SkipNameValidation: ptr.To(true),
			},
		})
		return defaultFunc(m)
	}
}

// newDummyNewWebhookManagedBy returns a dummy function to replace ctrl.NewWebhookManagedBy
func newDummyNewNodeScalingRequestWebhookManagedBy(errorCalls []int) func(manager.Manager) *builder.WebhookBuilder[*scalev1.NodeScalingRequest] {
	return newDummyNewWebhookManagedBy[*scalev1.NodeScalingRequest](errorCalls, origNewNodeScalingRequestWebhookManagedBy)
}

// newDummyNewWebhookManagedBy returns a dummy function to replace ctrl.NewWebhookManagedBy
func newDummyNewCardScalingRequestWebhookManagedBy(errorCalls []int) func(manager.Manager) *builder.WebhookBuilder[*scalev1.CardScalingRequest] {
	return newDummyNewWebhookManagedBy[*scalev1.CardScalingRequest](errorCalls, origNewCardScalingRequestWebhookManagedBy)
}

// newDummySetupSignalHandler returns a dummy function to replace ctrl.SetupSignalHandler
func newDummySetupSignalHandler() func() context.Context {
	return context.Background
}
