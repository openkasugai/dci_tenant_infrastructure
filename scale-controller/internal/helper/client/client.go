package client

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/controller-runtime/pkg/client"

	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	capdockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"

	capdv1 "cluster-api-provider-dci/api/v1beta2"
	rpmv1 "resource-pool-manager/api/v1"
)

// ClientHelper defines helper functions that use clients.
type clientHelperStruct struct {
	// client for workload cluster
	client.Client

	// client for management cluster
	managementClusterClient client.Client
}

// ClientHelperOption defines all options for NewClientHelper.
type ClientHelperOption struct {
	// Client for Management Cluster
	// If nil, a new client will be created.
	ManagementClusterClient client.Client

	// Scheme options for the client helper
	// only used when ManagementClusterClient is nil.
	SchemeOption *ClientHelperSchemeOption
}

// ClientHelperSchemeOption defines options for scheme used in ClientHelper.
type ClientHelperSchemeOption struct {
	// Scheme used for the management cluster client, if nil, a default scheme will be created.
	Scheme *runtime.Scheme

	// EnableClusterAPI indicates whether to enable Cluster API scheme.
	// if Scheme is nil, default is true.
	EnableClusterAPI bool

	// EnableDCI indicates whether to enable DCI provider and Resource Pool Manager scheme.
	// if Scheme is nil, default is true.
	EnableDCI bool

	// EnableDocker indicates whether to enable Docker provider scheme.
	// if Scheme is nil, default is false.
	EnableDocker bool
}

// Ensure clientHelperStruct implements the ClientHelper interface.
var _ ClientHelper = &clientHelperStruct{}

// newClient is a variable to hold the function for creating a new client, which can be overridden in tests.
var newClient = client.New

const (
	ManagementKubeconfigEnvVarName = "MANAGEMENT_KUBECONFIG"
)

// NewClientHelper returns a new ClientHelper.
func NewClientHelper(
	workloadClusterClient client.Client, option *ClientHelperOption,
) (ClientHelper, error) {
	if option == nil {
		option = &ClientHelperOption{}
	}
	managementClusterClient, err := getManagementClusterClient(option.ManagementClusterClient, option.SchemeOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get management cluster client: %w", err)
	}
	return &clientHelperStruct{
		Client:                  workloadClusterClient,
		managementClusterClient: managementClusterClient,
	}, nil
}

// getManagementClusterClient is a function to get the ClusterAPI resource of a ManagementCluster.
func getManagementClusterClient(managementClusterClient client.Client, option *ClientHelperSchemeOption) (client.Client, error) {
	// Set default options if nil
	if option == nil {
		option = &ClientHelperSchemeOption{
			EnableClusterAPI: true,
			EnableDCI:        true,
			EnableDocker:     false,
		}
	}

	// Return provided management cluster client if exists
	if managementClusterClient != nil {
		return managementClusterClient, nil
	}

	// Get kubeconfig path from environment variable
	kubeconfigPath := os.Getenv(ManagementKubeconfigEnvVarName)
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("environment variable %s is not set", ManagementKubeconfigEnvVarName)
	}

	// Build kubeconfig from the kubeconfig path
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig path: %s: %w", kubeconfigPath, err)
	}

	// Use provided scheme or create a new one
	scheme := option.Scheme

	if option.Scheme == nil {
		// Create scheme for the management cluster client
		scheme = runtime.NewScheme()
		if option.EnableClusterAPI {
			if err = capiv1.AddToScheme(scheme); err != nil {
				return nil, fmt.Errorf("failed to add Cluster API scheme: %w", err)
			}
		}
		if option.EnableDCI {
			if err = capdv1.AddToScheme(scheme); err != nil {
				return nil, fmt.Errorf("failed to add DCI provider scheme: %w", err)
			}
			if err = rpmv1.AddToScheme(scheme); err != nil {
				return nil, fmt.Errorf("failed to add Resource Pool Manager scheme: %w", err)
			}
		}
		if option.EnableDocker {
			if err = capdockerv1.AddToScheme(scheme); err != nil {
				return nil, fmt.Errorf("failed to add Docker provider scheme: %w", err)
			}
		}
	}

	// Create management cluster client
	c, err := newClient(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		err = fmt.Errorf("failed to create client for management cluster: %w", err)
	}

	return c, err
}
