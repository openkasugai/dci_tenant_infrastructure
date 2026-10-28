Node Scaling Request Controller Test Programs
=============================================

[toc]

## Overview

|Name|Commands|Description|
|-|-|-|
|[Envtest](#envtest)|`make test`|Execute Envtest|
|[e2e](#e2e)|`make test-e2e`|Execute e2e test|

### Preparation

- It is assumed that the Scale Controller (scale-controller), Cluster API Provider for DCI (cluster-api-provider-dci) and Resource Pool Manager (resource-pool-manager) source code are deployed.

  ```
  |-- cluster-api-provider-dci/
  |-- resource-pool-manager/
  |-- scale-controller/
      |-test/ # this directory
  ```

- It is assumed that you have already installed packages of go language for cluster-api, and Docker provider to get CustomResourceDefinition.

- All environment variables for path in this file should be set as an absolute path.

- The following environment variables are assumed to be set:
  ```sh
  export PATH_TO_SCALE_CONTROLLER=/path/to/scale-controller
  ```
  or
  ```sh
  # when current directory is this directory(scale-controller/test/)
  export PATH_TO_SCALE_CONTROLLER=$(cd ../;pwd)
  ```

___

### Envtest

This test needs following environment variables for path to CRD(CustomResourceDefinition),
and you can set these parameters as Makefile variables too:
|Name|Description|Example(confirm your environment)|
|-|-|-|
|ENV_DOCKER_PROVIDER_ENABLE|flag for Docker-provider|set `true` if you want to enable test for docker provider, otherwise disabled|

Envtest is executed through `go test` by following commands:

- Execute Envtest for controller&webhook (for DCI)
  ```sh
  # change directory to the TOP directory
  cd ${PATH_TO_SCALE_CONTROLLER}

  # Set valid path(and version) for your environment to each crd directories if needed
  export ENV_DOCKER_PROVIDER_ENABLE=false

  # Execute test
  make test
  # expected results:
  # :
  # ok      scale-controller/cmd/alertreceiver      1.470s  coverage: 1.5% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/cmd/controller 2.615s  coverage: 7.3% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/alertreceiver/cardalert       9.964s  coverage: 4.4% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/controller    139.699s        coverage: 75.0% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/helper/client 2.338s  coverage: 5.3% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/helper/utils  2.338s  coverage: 1.7% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/webhook/v1alpha1      20.760s coverage: 10.0% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler      2.410s  coverage: 1.0% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/alerts       2.440s  coverage: 0.6% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/alerts/examples      2.409s  coverage: 0.8% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/healthz      2.505s  coverage: 0.4% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/receiver     2.473s  coverage: 2.2% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # go tool cover -html=cover.out -o coverage.html # coverage file
  ```

- Execute Envtest for controller&webhook (for DCI and Docker)
  ```sh
  # change directory to the TOP directory
  cd ${PATH_TO_SCALE_CONTROLLER}

  # Set valid path(and version) for your environment to each crd directories if needed
  export ENV_DOCKER_PROVIDER_ENABLE=true

  # Execute test
  make test
  # expected results:
  # :
  # ok      scale-controller/cmd/alertreceiver      0.422s  coverage: 1.5% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/cmd/controller 0.694s  coverage: 7.3% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/alertreceiver/cardalert       8.502s  coverage: 4.4% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/controller    144.297s        coverage: 77.2% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/helper/client 0.361s  coverage: 5.3% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/helper/utils  0.327s  coverage: 1.7% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/internal/webhook/v1alpha1      17.659s coverage: 10.0% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler      0.361s  coverage: 1.0% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/alerts       0.382s  coverage: 0.6% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/alerts/examples      0.375s  coverage: 0.8% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/handler/healthz      0.380s  coverage: 0.4% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # ok      scale-controller/pkg/alertreceiver/receiver     0.394s  coverage: 2.2% of statements in scale-controller/api/v1alpha1, scale-controller/cmd/alertreceiver, scale-controller/cmd/controller, scale-controller/internal/alertreceiver/cardalert, scale-controller/internal/controller, scale-controller/internal/helper/client, scale-controller/internal/helper/utils, scale-controller/internal/webhook/v1alpha1, scale-controller/pkg/alertreceiver, scale-controller/pkg/alertreceiver/common, scale-controller/pkg/alertreceiver/handler, scale-controller/pkg/alertreceiver/handler/alerts, scale-controller/pkg/alertreceiver/handler/alerts/examples, scale-controller/pkg/alertreceiver/handler/healthz, scale-controller/pkg/alertreceiver/receiver
  # go tool cover -html=cover.out -o coverage.html
  ```

___

### e2e

You should install following tools at first:
- kind: v0.32.0+
- clusterctl: v1.14.0+
- helm
- kubebuilder: v4.15.0+

You can execute e2e test with kind clusters by following commands:

- Test for scale-controller for DCI provider installed by `helm install`
  ```sh
  cd ${PATH_TO_SCALE_CONTROLLER}
  make test-e2e
  ```

- Test for scale-controller for DCI provider installed by `make deploy`
  ```sh
  cd ${PATH_TO_SCALE_CONTROLLER}
  make test-e2e E2E_EXTRA_OPTS="-i kustomize"
  ```

- Test for scale-controller for Docker provider installed by `helm install`
  ```sh
  cd ${PATH_TO_SCALE_CONTROLLER}
  make test-e2e E2E_EXTRA_OPTS="-p docker"
  ```

- Test for scale-controller for Docker provider installed by `make deploy`
  - This test is not supported yet.
  ```sh
  cd ${PATH_TO_SCALE_CONTROLLER}
  make test-e2e E2E_EXTRA_OPTS="-i kustomize -p docker"
  ```

After the test completes, the `${PATH_TO_SCALE_CONTROLLER}/test/e2e/e2e-logs` directory contains the results of the execution log.
