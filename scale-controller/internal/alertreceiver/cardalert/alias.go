package cardalert

import (
	"k8s.io/client-go/rest"

	client "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// InClusterConfig is an alias for rest.InClusterConfig, allowing it to be overridden in tests.
	InClusterConfig = rest.InClusterConfig

	// NewClient is an alias for client.New, allowing it to be overridden in tests.
	NewClient = client.New
)
