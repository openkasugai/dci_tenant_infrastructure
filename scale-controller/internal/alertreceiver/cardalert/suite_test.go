package cardalert

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	scalev1 "scale-controller/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	defaultNamespace = "default"
)

var (
	ctx    context.Context
	cancel context.CancelFunc

	logger logr.Logger

	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
)

func TestAlertReceiver(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "AlertReceiver Suite")
}

var _ = BeforeSuite(func() {
	var err error

	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	log.SetLogger(logger)
	ctx, cancel = context.WithCancel(context.TODO())

	By("adding scheme for additional packages", func() {
		Expect(scalev1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	})

	// +kubebuilder:scaffold:scheme

	By("bootstrapping test environment", func() {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}

		// Retrieve the first found binary directory to allow running tests from IDEs
		if binDir := getFirstFoundEnvTestBinaryDir(); binDir != "" {
			testEnv.BinaryAssetsDirectory = binDir
		}

		// cfg is defined in this file globally.
		cfg, err = testEnv.Start()
		Expect(err).NotTo(HaveOccurred())

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	By("stopping running server", func() {
		Expect(testEnv).NotTo(BeNil())
		Expect(testEnv.Stop()).NotTo(HaveOccurred())
	})
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}
