Scale Controller
===============================

[toc]

## Overview

Scale Controller is composed of the following controllers:
- NodeScalingRequest Controller
- CardScalingRequest Controller
- CardAlert Controller
- CardAlert AlertReceiver

The CardAlert AlertReceiver is deployed as a separate `Pod`, while the other controllers run within a single `Pod`.

### NodeScalingRequest Controller

The NodeScalingRequest Controller is deployed in a Workload Cluster created by Cluster API, and is a controller that requests the Management Cluster to scale its own cluster, that is, to add or remove Nodes. The `NodeScalingRequest` is a custom resource to realize this functionality.

To request the addition of a `Node`, set the request type to `create` and specify the specifications of the server to be added.

```yaml
kind: NodeScalingRequest
metadata:
  name: example-scaleup-request
  namespace: default
spec:
  nodeRequestType: create
  location: central-data-center1
  serverSpec:
    type: COTSServer
    gpu:
      model: H100
      num: 1
```

**Note**: The details of `spec.location` and `spec.serverSpec` are defined in DCI Manager, so please refer to the CRD of `DCIMachine` for more information.

If there is no `DCIMachineTemplate` or `MachineDeployment` in the Management Cluster that meets the requested specifications, the corresponding resources are newly created and then a request to add a `Node` is made.

To request the deletion of a `Node`, set the request type to `delete` and specify the name of the `Node` to be deleted.

```yaml
kind: NodeScalingRequest
metadata:
  name: example-scaledown-request
  namespace: default
spec:
  nodeRequestType: delete
  node: my-cluster-worker-md-0-cc2jr-h5pzg
```

If the deletion of a `Node` results in the number of replicas of the `MachineDeployment` that created the specified Node becoming zero, check whether there are other `MachineDeployment`.
  - If there are other `MachineDeployment` in the same namespace, same cluster, and same infrastructure type, and ClusterAutoScaler is disabled, delete the `MachineDeployment` used by the target `Node`.
    - When deleting a `MachineDeployment`, check whether the `DCIMachineTemplate` referenced by that `MachineDeployment` is used by other `MachineDeployments`.
    - If there are no other `MachineDeployment` using the target `DCIMachineTemplate` except itself, delete the target `DCIMachineTemplate` as well.
  - If the target `MachineDeployment` is the only `MachineDeployment`, do not delete it but set the number of replicas to zero.

When a `NodeScalingRequest` is created in the Workload Cluster, the NodeScalingRequest Controller detects it and requests the Management Cluster running Cluster API to add or remove a `Node`.
The following describes each flow. Note that as a specification of the NodeScalingRequest Controller, only one `NodeScalingRequest` can be processed at a time. Therefore, while a `NodeScalingRequest` is being processed, other `NodeScalingRequest` instances transition to a pending state and are requeued at regular intervals waiting for processing to begin.

#### Node Addition

When the NodeScalingRequest Controller detects a `NodeScalingRequest` for adding a `Node`, it operates as follows.

1. Extract the following `MachineDeployment` resources from the Management Cluster:
    - The value of `machinedeployment.spec.clusterName` matches the value of the annotation `cluster.x-k8s.io/cluster-name` of the `Node` in the Workload Cluster.
    - The value of `machinedeployment.namespace` matches the value of the annotation `cluster.x-k8s.io/cluster-namespace` of the `Node` in the Workload Cluster.
    - The value of `machinedeployment.spec.template.spec.infrastructureRef.kind` is `DCIMachineTemplate`.
    - Determine whether the `DCIMachineTemplate` indicated by `machinedeployment.spec.template.spec.infrastructureRef.name` satisfies `nodescalingrequest.spec.serverSpec`:
      - Search with AND conditions for each field set in `nodescalingrequest.spec.serverSpec`.
      - Fields not set in `nodescalingrequest.spec.serverSpec` are not considered.
      - If a field is set in `nodescalingrequest.spec.serverSpec` but is not set (nil) in `dcimachinetemplate.spec.template.spec`, treat it as not satisfied.

2. If a matching `MachineDeployment` is found in step 1, increment its `machinedeployment.spec.replicas` by 1 and set that value to `nodescalingrequest.status.desiredReplicas`.

3. If no matching `MachineDeployment` is found in step 1, create a new `MachineDeployment` and set `nodescalingrequest.status.desiredReplicas` to 1.
    - If there is no `DCIMachineTemplate` in the cluster namespace that satisfies `nodescalingrequest.spec.serverSpec`, create a new `DCIMachineTemplate` with the same contents as `nodescalingrequest.spec.serverSpec`.

4. Update `nodescalingrequest.status.phase` to `Creating`.

5. Monitor Create Events for `Node` resources in the Workload Cluster, wait until `machinedeployment.status.readyReplicas` becomes greater than or equal to `nodescalingrequest.status.desiredReplicas`, and update `nodescalingrequest.status.phase` to `Created`.

#### Node Deletion

When the NodeScalingRequest Controller detects a `NodeScalingRequest` for deleting a `Node`, it operates as follows.

1. Get the `Node` in the Workload Cluster indicated by `nodescalingrequest.spec.node`, treat the value of the annotation `cluster.x-k8s.io/machine` as the name of the `Machine` to be deleted, and treat the value of `cluster.x-k8s.io/cluster-namespace` as the namespace of the `Machine`.

2. Get the `Machine` to be deleted, treat the value of the label `cluster.x-k8s.io/deployment-name` as the name of the `MachineDeployment` to be scaled down, and use the namespace obtained in step 1 as the namespace.

3. Get the `MachineDeployment` to be deleted.

4. If `machinedeployment.spec.replicas` of the MachineDeployment to be deleted is 1 (i.e., it will become 0 after scaling down), operate as follows:
    1. Get the list of `MachineDeployment` resources and check whether there are other valid `MachineDeployment` resources.

    2. If there are no other valid `MachineDeployment` resources in step 4-1, skip the processing from step 4-3 onward and proceed to step 5.

    3. If there are other valid `MachineDeployment` resources in step 4-1, check whether the `DCIMachineTemplate` used by that `MachineDeployment` is used by other `MachineDeployment` resources.
        - Delete the target `DCIMachineTemplate` only if it is not used by other `MachineDeployment` resources.

    4. Delete the target `MachineDeployment` and set `nodescalingrequest.status.desiredReplicas` to 0.

    5. Update `nodescalingrequest.status.phase` to `Deleting`.

    6. Monitor Delete Events for `Node` resources in the Workload Cluster, and after confirming the deletion of the `Node`, update `nodescalingrequest.status.phase` to `Deleted`.

5. If `machinedeployment.spec.replicas` of the MachineDeployment to be scaled down is not 1, operate as follows:
    1. Set the value of the annotation `cluster.x-k8s.io/paused` to `true` for the target `MachineSet` in the Management Cluster.

    2. Decrement `machinedeployment.spec.replicas` by 1 and set that value to `nodescalingrequest.status.desiredReplicas`.

    3. Delete the target `Machine`.

    4. Update `nodescalingrequest.status.phase` to `Deleting`.

    5. Monitor the `Node` resources in the Workload Cluster, and when `machinedeployment.status.readyReplicas` of the target MachineDeployment becomes less than or equal to `nodescalingrequest.status.desiredReplicas`, update `nodescalingrequest.status.phase` to `Deleted`.


### CardScalingRequest Controller

The CardScalingRequest Controller is deployed in a Workload Cluster created by Cluster API, and is a controller that requests the Management Cluster to scale the number of GPUs mounted on `Nodes` within the cluster or to replace GPUs. The `CardScalingRequest` is a custom resource to realize this functionality.

**Note**: Henceforth, the values of `.spec.scalingParameter.model` and `.spec.scalingParameter.uuids` specified in `CardScalingRequest` must match the GPU model names and UUIDs described in the `LogicalServer` belonging to the Management Cluster.

To request the addition of a GPU, set the operation type to `add` and specify the model and number of GPUs to be added.
Currently, the supported card addition scenario is only the case where the number of cards of an already existing model is increased.
Therefore, adding cards of a different model is not supported.
Additionally, addition by specifying UUIDs is not supported.

```yaml
kind: CardScalingRequest
metadata:
  name: example-add-card
  namespace: default
spec:
  nodeName: node-to-be-card-added
  scalingParameter:
    op: add
    model: H100
    num: 1 # Number of cards to add
```

To request the deletion of a GPU, set the operation type to `delete` and specify either the number of cards to be deleted or their UUIDs.
Note that combining count specification and UUID specification for deletion is not supported.
Additionally, deletion resulting in zero cards is permitted.

```yaml
kind: CardScalingRequest
metadata:
  name: example-delete-card-with-num
  namespace: default
spec:
  nodeName: node-to-be-card-added
  scalingParameter:
    op: delete
    model: H100
    num: 1 # Number of cards to delete
```

```yaml
kind: CardScalingRequest
metadata:
  name: example-delete-card
  namespace: default
spec:
  nodeName: node-to-be-card-added
  scalingParameter:
    op: delete
    model: H100
    uuids:
    - GPU-01234567-89ab-cdef-0123-456789abcdef # UUID specification
```

To request the replacement of a GPU, set the operation type to `replace` and specify the UUID of the card to be replaced.

```yaml
kind: CardScalingRequest
metadata:
  name: example-replace-card
  namespace: default
spec:
  nodeName: node-to-be-card-added
  scalingParameter:
    op: replace
    model: H100
    uuids:
    - GPU-01234567-89ab-cdef-0123-456789abcdef # UUID specification
```

When a `CardScalingRequest` is created in the Workload Cluster, the CardScalingRequest Controller detects it and, as needed, creates appropriate resources in the Management Cluster running Cluster API, then requests the target `Node` to add, remove, or replace cards via `NodeScalingRequest`.
The following describes this flow. Note that as a specification of the CardScalingRequest Controller, only one `CardScalingRequest` can be processed at a time. Therefore, while a `CardScalingRequest` is being processed, other `CardScalingRequest` instances transition to a pending state and are requeued at regular intervals waiting for processing to begin.

#### Card Addition, Deletion, and Replacement Flow

When the CardScalingRequest Controller detects a `CardScalingRequest`, it operates as follows regardless of the request type.
Note that if multiple `MachineDeployment` and `DCIMachineTemplate` resources with the same specifications exist, the behavior may not be as expected.

1. Get the `Node` in the Workload Cluster indicated by `cardscalingrequest.spec.nodeName`, and retrieve the associated `Machine`, `MachineDeployment`, `DCIMachineTemplate`, `DCIMachine`, and `LogicalServer` resources in the Management Cluster.

2. Check the contents of `cardscalingrequest.spec.scalingParameter` and the `DCIMachineTemplate`, and verify whether a `DCIMachineTemplate` that satisfies the specifications after card addition or deletion exists in the same `Namespace` as the `DCIMachineTemplate` obtained in step 1. If it does not exist, create it.

3. Verify whether a `MachineDeployment` exists that references a `DCIMachineTemplate` satisfying the requested specifications from step 2 and is not targeted by ClusterAutoscaler. If it does not exist, create it.

4. Patch an annotation to the `DCIMachine` obtained in step 1 based on `cardscalingrequest.spec.scalingParameter` to request card addition, deletion, or replacement.

5. Create a `NodeScalingRequest` that deletes the `Node` obtained in step 1 and wait for its success. If the `NodeScalingRequest` fails, treat the `CardScalingRequest` as an error and terminate processing.

6. Create a `NodeScalingRequest` that adds a `Node` satisfying the requested specifications of the `DCIMachineTemplate` obtained in step 2 after card operations and wait for its success. If the `NodeScalingRequest` fails, treat the `CardScalingRequest` as an error and terminate processing.


### CardAlert Controller

The CardAlert Controller is deployed in a Workload Cluster created by Cluster API, and based on failure information of `Pod` within the cluster primarily created by the CardAlert AlertReceiver, the response policy obtained from environment variables, and the server type (`Composed`, `COTS`, etc.) of the `Node` where the failure occurred, it performs appropriate failure recovery. The `CardAlert` is a custom resource to realize this functionality.
Note that `CardAlert` is not expected to be created by users.

The failure recovery policies of the CardAlert Controller are broadly divided as follows:
- Automatic replacement of failure cards
  - Retrieved from environment variables at startup, and can be enabled or disabled
  - Default setting is enabled
- When automatic replacement is enabled: Automatic replacement policy
  - Retrieve information about the server composing the `Node` and processing changes based on its type
  - `COTS`
    - Create a `MachineAlert` resource in the Management Cluster and replace the target `Node`
  - `Composed`
    - Create a `CardScalingRequest` resource in the Workload Cluster and replace the target `Node`
  - `VM`
    - Not supported
- When automatic replacement is disabled: `Pod` failure recovery policy
  - Retrieved from environment variables at startup, selects the policy for `DeviceTaintRule`
  - Default setting is `NoExecute`
  - `NoExecute`
    - Delete `Pod` using the target failure card and attempt to create alternative `Pod`
  - `NoSchedule`
    - Do not delete `Pod` using the target failure card, but ensure that the failure card is not allocated even if the target card is requested from any process after the `Pod` is deleted

When a `CardAlert` is created in the Workload Cluster, the CardAlert Controller detects it and, according to its settings, creates appropriate resources and performs failure recovery as described above.
The following describes this flow. Note that as a specification of the CardAlert Controller, only one `CardAlert` can be processed at a time. Therefore, while a `CardAlert` is being processed, other `CardAlert` instances transition to a pending state and are requeued at regular intervals waiting for processing to begin.

#### Flow When Automatic Replacement Is Enabled

When automatic replacement is enabled, the CardAlert Controller operates as follows upon detecting a `CardAlert`:

1. Get the `Node` in the Workload Cluster indicated by `cardalert.spec.nodeName`, and retrieve the associated `Machine`, `DCIMachine`, and `LogicalServer` resources in the Management Cluster.

2. Verify that the card specified by `cardalert.spec.cardInfo` exists on the `LogicalServer` obtained in step 1. If all cards do not exist, terminate processing.

3. Retrieve the server type of the `LogicalServer` obtained in step 1 and proceed with processing as follows based on its type:
  1. If `COTS`, create a `MachineAlert` in the Management Cluster and wait for its completion. After confirming completion, delete the created `MachineAlert` and terminate `CardAlert` processing.
    - The `Namespace` in which `MachineAlert` is created is retrieved from environment variables, and must be the `Namespace` where the MachineAlert Controller is watching `MachineAlert` resources (typically `dcirpm`).
  2. If `Composed`, create a `CardScalingRequest` with OwnerReference set and wait for its completion. After confirming completion, terminate `CardAlert` processing. Note that if `CardScalingRequest` fails, treat `CardAlert` as a processing failure and terminate.
    - Deletion of `CardScalingRequest` is expected to be performed by the garbage collector upon `CardAlert` deletion, and is not performed by the CardAlert Controller.
  3. If `VM`, treat `CardAlert` as a processing failure and terminate.

#### Flow When Automatic Replacement Is Disabled

When automatic replacement is disabled, the CardAlert Controller operates as follows upon detecting a `CardAlert`:

1. Same as step 1 of the automatic replacement enabled flow

2. Same as step 2 of the automatic replacement enabled flow

3. Create a `DeviceTaintRule` for the target card and terminate `CardAlert` processing.
  - The Taint key and value are fixed values specified as `dra.example.com/unhealthy` and `Broken` respectively.
  - The Taint Effect is specified as `NoExecute` or `NoSchedule` retrieved from environment variables at startup.
  - `DeviceTaintRule` is a cluster-scoped resource, so `Namespace` is not specified.
  - Find the matching device in `ResourceSlice` based on the UUID of the GPU set as having a failure, and set the target device in `DeviceSelector`.
  - The CardAlert Controller does not delete `DeviceTaintRule` even after the failure is resolved, and expects manual deletion by the user after the failure is resolved.


### CardAlert AlertReceiver

The CardAlert AlertReceiver is deployed in a Workload Cluster created by Cluster API, and functions primarily as a server (Webhook Receiver) to receive failure information of `Pod` within the cluster from Prometheus's AlertManager, and creates `CardAlert` resources from Alert information.
This approach is based on the philosophy that Webhook Receivers are not well-suited for idempotent operations, so the receiver only performs operations to convert Alerts into custom resources, while actual failure recovery processing is delegated to a custom controller.

The specifications of the CardAlert AlertReceiver are as follows:
- Retrieve the following information from environment variables at startup:
  - `CARD_ALERT_NAMESPACE`
    - `Namespace` in which `CardAlert` is created
    - required
  - `ALERT_RECEIVER_LISTEN_ADDRESS`
    - Address for the server to listen on
    - optional
    - Defaults to `0.0.0.0` if not set
  - `ALERT_RECEIVER_LISTEN_PORT`
    - Port for the server to listen on
    - optional
    - Defaults to `8080` if not set
- Start the server process without TLS authentication and wait for Alerts from AlertManager
  - Endpoints:
    - `/healthz`
      - Endpoint used for health checks
    - `/alerts`
      - Endpoint expected as the AlertManager destination
  - When installing via Helm, a `Service` named `scale-controller-alert-receiver-service` is created, opening port `8080`. Set the following URL as the destination for AlertManager in the same cluster: (use the `Namespace` specified during `helm install`)
    - `http://scale-controller-alert-receiver-service.scale-controller-system.svc:8080/alerts`
- Create `CardAlert` based on Alerts:
  - Only Alerts with status `Firing` are targeted; `Resolved` Alerts are ignored.
  - Only Alerts named `GPU ECC Double-Bit Error Detected` are targeted; all other Alerts are ignored.
  - If labels lack any of `HostName`, `modelName`, or `UUID`, output an error and skip creating the corresponding `CardAlert`.
  - Create a `CardAlert` with a name starting with `card-alert-receiver-` using the `Namespace` retrieved from environment variables.
    - `.spec.nodeName` uses the value of the Alert label `HostName`.
    - `.spec.cards[].model` uses the value of the Alert label `modelName`.
    - `.spec.cards[].gpuUUID` uses the value of the Alert label `UUID`.
    - `.spec.cards[].migUUIDs[]` uses the value of the Alert label `dra_mig_device_uuid`. When multiple Alerts are received in a single Alert reception and among them exist Alerts with the same `HostName`, `modelName`, and `UUID`, and a `dra_mig_device_uuid` field is present, set them collectively.


## Quick Start

### Preparation

It is assumed that the source code for the Scale Controller (scale-controller), DCI Provider for Cluster API (cluster-api-provider-dci), and Resource Pool Manager (resource-pool-manager) is deployed as shown below.
This verification is intended to be performed using kind clusters.

```
|-- cluster-api-provider-dci
|-- resource-pool-manager
|-- scale-controller
```

Move to the root directory of the Scale Controller source code.

```sh
cd /path/to/scale-controller
```

### Verified Environments
- k8s: v1.36.1 (Management Cluster + Workload Cluster)
- cluster-api: v1.14.0 (v1beta2 contract)
- cert-manager: v1.19.2
- calico: n/a

### Version for quick start

Set following shell variables adjusting your environment.
Set helm_chart_version, controller_img_tag, and receiver_img_tag to the values specified in the Helm chart.

```sh
cluster_api_version=v1.14.0
cert_manager_version=v1.19.2
helm_chart_version=v0.3.9
controller_img_tag=0.3.9
receiver_img_tag=0.3.9
```

#### Setup of Management Cluster and Workload Cluster

Create the Management Cluster and Workload Cluster using Kind.
Here, the cluster names are set as "fake-management" for the Management Cluster and "fake-cluster" for the Workload Cluster.

##### Management Cluster

```sh
export M_CLUSTER_NAME="fake-management"
export CLUSTER_NAME="fake-cluster"
# If you are in a proxy environment, run the following as needed
export NO_PROXY=$NO_PROXY,172.18.0.0/16
kind_image=kindest/node:v1.36.1

# Create Management Cluster using kind
# If you do not need to specify the k8s image, the --image option is not required.
kind create cluster -n ${M_CLUSTER_NAME} --image=${kind_image}

# Deploy the Cluster API CRDs (Custom Resource Definitions) on the Management Cluster.
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/core/config/crd/bases/cluster.x-k8s.io_machinedeployments.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/core/config/crd/bases/cluster.x-k8s.io_machinesets.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/core/config/crd/bases/cluster.x-k8s.io_machines.yaml

# Deploy DCI Provider and Resource Pool Manager CRDs on the Management Cluster.
path_to_capd=../cluster-api-provider-dci
path_to_rpm=../resource-pool-manager
kubectl apply -f ${path_to_capd}/config/crd/bases/
kubectl apply -f ${path_to_rpm}/config/crd/bases/

# Create dummy Cluster API and Provider resources.
envsubst < example/cluster-api/fake/management/cluster.yaml | kubectl apply -f -
# Patch the .status field of the MachineDeployment
kubectl patch md fake-worker-md-00 --type=merge --patch '{"status": {"readyReplicas": 1}}' --subresource status
# Set OwnerReference for the MachineSet
md_uid=$(kubectl get md fake-worker-md-00 -o jsonpath='{.metadata.uid}')
patch_json=$(printf '[{
    "op": "add",
    "path": "/metadata/ownerReferences",
    "value": [{
      "apiVersion": "cluster.x-k8s.io/v1beta2",
      "kind": "MachineDeployment",
      "name": "fake-worker-md-00",
      "uid": "%s",
      "controller": true,
      "blockOwnerDeletion": true}]}]' ${md_uid})
kubectl patch ms fake-worker-ms-00 --type='json' -p="${patch_json}"
# Set OwnerReference for the DCIMachine
machine_uid=$(kubectl get machine fake-worker-machine-00 -o jsonpath='{.metadata.uid}')
patch_json=$(printf '[{
    "op": "add",
    "path": "/metadata/ownerReferences",
    "value": [{
      "apiVersion": "cluster.x-k8s.io/v1beta2",
      "kind": "Machine",
      "name": "fake-worker-machine-00",
      "uid": "%s",
      "controller": true,
      "blockOwnerDeletion": true}]}]' ${machine_uid})
kubectl patch dcimachine fake-worker-dci-machine-00 --type='json' -p="${patch_json}"

# Verify that desired replicas and ready replicas are both 1
kubectl get md fake-worker-md-00
# expected results:
# NAME                CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# fake-worker-md-00   fake-cluster               1                   1                                        2s    

# Verify that OwnerReference is set for MachineSet and others
kubectl get ms fake-worker-ms-00 -o jsonpath='{.metadata.ownerReferences}{"\n"}'
# expected results:
# [{"apiVersion":"cluster.x-k8s.io/v1beta2","blockOwnerDeletion":true,"controller":true,"kind":"MachineDeployment","name":"fake-worker-md-00","uid":"8c07d62d-6718-48e6-b3af-c975e86cfe5b"}]
kubectl get dcimachine fake-worker-dci-machine-00 -o jsonpath='{.metadata.ownerReferences}{"\n"}'
# expected results:
# [{"apiVersion":"cluster.x-k8s.io/v1beta2","blockOwnerDeletion":true,"controller":true,"kind":"Machine","name":"fake-worker-machine-00","uid":"cfec50d7-0d61-4c34-8671-c8b28f15541d"}]
```

**Note**: If the DCI Provider or Resource Pool Manager CRD files do not exist in the expected directory, generate them as described below.
```sh
make manifests -C ${path_to_capd}
make manifests -C ${path_to_rpm}
```

##### Workload Cluster

```sh
# Create Workload Cluster using kind and create dummy resources.
# If you do not need to specify the k8s image, the --image option is not required.
kind create cluster -n ${CLUSTER_NAME} --image=${kind_image}
kubectl config use-context kind-${M_CLUSTER_NAME}
# Set a kubectl alias ("kw") for accessing the Workload Cluster
alias kw="kubectl --context=kind-${CLUSTER_NAME}"
envsubst < example/cluster-api/fake/workload/cluster.yaml | kw apply -f -

# Verify Node resources
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE   VERSION
# fake-cluster-control-plane   NotReady   control-plane   8s    v1.36.1
# fake-node-00                 Unknown    <none>          0s    
# fake-node-control-plane-00   Unknown    <none>          0s    
```

#### Deploying the Scale Controller

To deploy the Scale Controller, the following preparations are required in advance.

- Deploy cert-manager
  - See “[Deploying cert-manager](#deploying-cert-manager)” below.
- Build the Scale Controller container image
  - See “[Building the Scale Controller](#building-the-scale-controller)” below.
- Create a kubeconfig for the Scale Controller to access the Management Cluster
  - An example method is shown in “[Setting up credentials](#setting-up-credentials)”.

Once these preparations are complete, install the Helm Chart with the following command.
For details, refer to [helm/README.md](helm/README.md).

Note:
- `managementKubeconfig` default value in `values.yaml` is `dummy-kubeconfig` for local rendering/tests.
- For production operation, always provide a real kubeconfig via `--set-file managementKubeconfig=...`.

##### Deploying the Scale Controller Using an Existing Helm Chart

Deploy the Scale Controller from the Helm Chart with the following command.

```sh
# Note:
#    Make sure that the current-context in your kubeconfig is set to the target Workload Cluster.
#    If not, adjust by adding the --kube-context or --kubeconfig option as needed.
cluster_context="--kube-context=kind-${CLUSTER_NAME}"
path_to_helm_chart=dist/chart/
helm_controller_image_tag=${controller_img_tag}
helm_receiver_image_tag=${receiver_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set manager.image.repository=scale-controller \
    --set manager.image.tag=${helm_controller_image_tag} \
    --set alertReceiver.container.image.repository=card-alert-receiver \
    --set alertReceiver.container.image.tag=${helm_receiver_image_tag} \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

##### Deploying the Scale Controller with Helm Chart Creation

Create the Helm Chart with the following command.

```sh
make helmchart
```

Deploy the Scale Controller from the Helm Chart with the following command.

```sh
# Note:
#    Make sure that the current-context in your kubeconfig is set to the target Workload Cluster.
#    If not, adjust by adding the --kube-context or --kubeconfig option as needed.
cluster_context="--kube-context=kind-${CLUSTER_NAME}"
path_to_helm_chart=dist/scale-controller-${helm_chart_version}.tgz
helm_controller_image_tag=${controller_img_tag}
helm_receiver_image_tag=${receiver_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set manager.image.repository=scale-controller \
    --set manager.image.tag=${helm_controller_image_tag} \
    --set alertReceiver.container.image.repository=card-alert-receiver \
    --set alertReceiver.container.image.tag=${helm_receiver_image_tag} \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

##### Deploying the Scale Controller Using Kustomize

Deploy the Scale Controller to the Workload Cluster.

```sh
export KUBECTL="kubectl --context=kind-${CLUSTER_NAME}"
make deploy-with-alert-receiver
```

Register the previously created kubeconfig as a `Secret` in the "scale-controller-system" namespace of the Workload Cluster.

```sh
kw create secret generic management-kubeconfig \
    --from-file=kubeconfig=management.kubeconfig \
    -n scale-controller-system
```

#### Deploying cert-manager

Deploy cert-manager, which issues certificates for webhooks, and make sure that its `Pod` is in the `Running` state.

```sh
kw apply -f https://github.com/cert-manager/cert-manager/releases/download/${cert_manager_version}/cert-manager.yaml
kw get pod -n cert-manager
# expected results:
# NAME                                       READY   STATUS    RESTARTS   AGE
# cert-manager-56cbc66766-5rmnk              1/1     Running   0          20s
# cert-manager-cainjector-7cf74f998f-kf66j   1/1     Running   0          20s
# cert-manager-webhook-77d44c5f9b-29rkz      1/1     Running   0          20s
```

#### Building the Scale Controller

Build the container image and load it into the Workload Cluster.

Depending on your environment, add `sudo` as needed.
Note that when passing user environment variables to the `sudo` command, you may need to add the `-E` option as appropriate.

```sh
export IMG=scale-controller:${controller_img_tag}
export ALERT_RECEIVER_IMG=card-alert-receiver:${receiver_img_tag}
make docker-build
docker images ${IMG}
# expected results:
# IMAGE                    ID             DISK USAGE   CONTENT SIZE   EXTRA
# scale-controller:0.3.9   24d68e75516c       78.7MB             0B      
docker images ${ALERT_RECEIVER_IMG}
# expected results:
# IMAGE                       ID             DISK USAGE   CONTENT SIZE   EXTRA
# card-alert-receiver:0.3.9   44cc585a4a98       40.6MB             0B        
```

```sh
kind load docker-image -n ${CLUSTER_NAME} ${IMG}
kind load docker-image -n ${CLUSTER_NAME} ${ALERT_RECEIVER_IMG}
```

**Note**
If you want to build the image with buildah instead of docker, use `buildah-build` instead of `docker-build`.

#### Setting Up Credentials

Set up authentication credentials for the Scale Controller on the Management Cluster.

```sh
kubectl apply -f example/rbac/dci/rbac-in-management-cluster.yaml
```

Create a kubeconfig from the generated credentials so that the Scale Controller can access the Management Cluster.

```sh
export SERVER_IP_ADDRESS=$(docker inspect -f '{{.NetworkSettings.Networks.kind.IPAddress}}' ${M_CLUSTER_NAME}-control-plane)
export SERVER_PORT=6443
export CA_CRT=$(kubectl get secret scale-controller-tenant-secret -o jsonpath='{.data.ca\.crt}')
export TOKEN=$(kubectl get secret scale-controller-tenant-secret -o jsonpath='{.data.token}' | base64 -d)
envsubst < example/rbac/management-cluster-kubeconfig.yaml > management.kubeconfig
```


### Operation Check (NodeScalingRequest Controller)

#### Node Addition

Create a `NodeScalingRequest` that requests the addition of a `Node` matching the existing `DCIMachineTemplate`.

```sh
kw apply -f example/nsr/dci/nsr-scaleup-existing-md.yaml
```

Verify that `nodescalingrequest.status.phase` becomes `Creating`.

```sh
kw get nsr
# expected results:
# NAME                TYPE     PHASE      AGE
# scale-up-existing   create   Creating   5s
```

Verify that` machinedeployment.spec.replicas` has increased to 2.

```sh
kubectl get md fake-worker-md-00
# expected results:
# NAME                CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# fake-worker-md-00   fake-cluster               2                   1                                        5m46s   
```

To simulate the operation of Cluster API, set `machinedeployment.status.readyReplicas` to 2 and manually add dummy `Machine` and `Node` resources.

```sh
export MD_NAME="fake-worker-md-00"
envsubst < example/cluster-api/fake/management/resources-scaleup-existing.yaml | kubectl apply -f -
machine_uid=$(kubectl get machine fake-worker-machine-01 -o jsonpath='{.metadata.uid}')
patch_json=$(printf '[{
    "op": "add",
    "path": "/metadata/ownerReferences",
    "value": [{
      "apiVersion": "cluster.x-k8s.io/v1beta2",
      "kind": "Machine",
      "name": "fake-worker-machine-01",
      "uid": "%s",
      "controller": true,
      "blockOwnerDeletion": true}]}]' ${machine_uid})
kubectl patch dcimachine fake-worker-dci-machine-01 --type='json' -p="${patch_json}"
kubectl patch md ${MD_NAME} --type=merge --patch '{"status": {"readyReplicas": 2}}' --subresource status
envsubst < example/cluster-api/fake/workload/resources-scaleup-existing.yaml | kw apply -f -
```

Verify that `nodescalingrequest.status.phase` becomes `Created`.

```sh
kw get nsr
# expected results:
# NAME                TYPE     PHASE     AGE
# scale-up-existing   create   Created   106s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has increased to 2.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               2                   2                                        6m40s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   6m38s
```

Verify that a `Node` has been added.

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE     VERSION
# fake-cluster-control-plane   Ready      control-plane   8m11s   v1.36.1
# fake-node-00                 NotReady   <none>          8m3s    
# fake-node-01                 Unknown    <none>          35s     
# fake-node-control-plane-00   NotReady   <none>          8m3s    
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-up-existing
```

#### Node Deletion

Create a `NodeScalingRequest` specifying the `Node` to be deleted.

```sh
export NODE_TO_DELETE="fake-node-01"
envsubst < example/nsr/dci/nsr-scaledown.yaml | kw apply -f -
```

Verify that `nodescalingrequest.status.phase` becomes `Deleting`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE      AGE
# scale-down   delete   Deleting   9s
```

Verify that both `machinedeployment.spec.replicas` and `machineset.spec.replicas` have decreased to 1.

```sh
kubectl get md,ms
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   2                                        9m44s   

# NAME                                            CLUSTER        DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   AGE     VERSION
# machineset.cluster.x-k8s.io/fake-worker-ms-00   fake-cluster   1                                                    9m44s   v1.36.1
```

To simulate the operation of Cluster API, manually delete the dummy `Node` and change `machinedeployment.status.readyReplicas` to 1.

```sh
kw delete node fake-node-01
kubectl patch md ${MD_NAME} --type=merge --patch '{"status": {"readyReplicas": 1}}' --subresource status
```

Verify that `nodescalingrequest.status.phase` becomes `Deleted`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   112s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has decreased to 1.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        11m   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   12m
```

Verify that a `Node` has been removed.

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE   VERSION
# fake-cluster-control-plane   Ready      control-plane   11m   v1.36.1
# fake-node-00                 NotReady   <none>          11m   
# fake-node-control-plane-00   NotReady   <none>          11m   
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-down
```

#### Node Addition with MachineDeployment Creation

Create a `NodeScalingRequest` that requests the addition of a `Node` for which no matching `DCIMachineTemplate` exists.

```sh
kw apply -f example/nsr/dci/nsr-scaleup-new-md.yaml
```

Verify that `nodescalingrequest.status.phase` becomes `Creating`.

```sh
kw get nsr
# expected results:
# NAME              TYPE     PHASE      AGE
# scale-up-new-md   create   Creating   5s
```

Confirm that a new `MachineDeployment` and `DCIMachineTemplate` are created, and that the replica count of the newly created `MachineDeployment` is 1.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        12m   
# machinedeployment.cluster.x-k8s.io/worker-md-00evb     fake-cluster               1                                                            10s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-37os8           10s
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   13m
```

Verify that the `.spec.template.spec` of the newly created `DCIMachineTemplate` matches the `spec` of the `NodeScalingRequest`.

```sh
# Function to return "same string" if two strings match
compare_string() {
  if [[ "$1" == "$2" ]];then
    echo same string
  else
    echo different string
    echo $1
    echo $2
  fi
}
export INFRA_TEMPLATE_NAME=worker-dmt-37os8 # Adjust to your environment
compare_string \
  $(kubectl get dcimachinetemplate ${INFRA_TEMPLATE_NAME} -o jsonpath='{.spec.template.spec.serverSpec}') \
  $(kw get nsr scale-up-new-md -o jsonpath='{.spec.serverSpec}')
# expected results:
# same string
compare_string \
  $(kubectl get dcimachinetemplate ${INFRA_TEMPLATE_NAME} -o jsonpath='{.spec.template.spec.location}') \
  $(kw get nsr scale-up-new-md -o jsonpath='{.spec.location}')
# expected results:
# same string
```

Also, confirm that the `.spec.template.spec.infrastructureRef` of the newly created `MachineDeployment` is set to the name of the newly created `DCIMachineTemplate`.

```sh
export MD_NAME=worker-md-00evb # Adjust to your environment
compare_string \
  $(kubectl get md ${MD_NAME} -o jsonpath='{.spec.template.spec.infrastructureRef.name}') \
  ${INFRA_TEMPLATE_NAME}
# expected results:
# same string
```

To simulate the operation of Cluster API, set `machinedeployment.status.readyReplicas` to 1 and manually add dummy `Machine` and `Node` resources.

```sh
# Create dummy resources
envsubst < example/cluster-api/fake/management/resources-scaleup-new.yaml | kubectl apply -f -
# Set OwnerReference for MachineSet
md_uid=$(kubectl get md ${MD_NAME} -o jsonpath='{.metadata.uid}')
patch_json=$(printf '[{
    "op": "add",
    "path": "/metadata/ownerReferences",
    "value": [{
      "apiVersion": "cluster.x-k8s.io/v1beta2",
      "kind": "MachineDeployment",
      "name": "%s",
      "uid": "%s",
      "controller": true,
      "blockOwnerDeletion": true}]}]' ${MD_NAME} ${md_uid})
kubectl patch ms fake-worker-ms-02 --type='json' -p="${patch_json}"
# Set OwnerReference for DCIMachine
machine_uid=$(kubectl get machine fake-worker-machine-02 -o jsonpath='{.metadata.uid}')
patch_json=$(printf '[{
    "op": "add",
    "path": "/metadata/ownerReferences",
    "value": [{
      "apiVersion": "cluster.x-k8s.io/v1beta2",
      "kind": "Machine",
      "name": "fake-worker-machine-02",
      "uid": "%s",
      "controller": true,
      "blockOwnerDeletion": true}]}]' ${machine_uid})
kubectl patch dcimachine fake-worker-dci-machine-02 --type='json' -p="${patch_json}"
# Patch the .status field of MachineDeployment
kubectl patch md ${MD_NAME} --type=merge --patch '{"status": {"readyReplicas": 1}}' --subresource status
envsubst < example/cluster-api/fake/workload/resources-scaleup-new.yaml | kw apply -f -
```

Verify that `nodescalingrequest.status.phase` becomes `Created`.

```sh
kw get nsr
# expected results:
# NAME              TYPE     PHASE     AGE
# scale-up-new-md   create   Created   2m32s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has increased to 1.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        15m     
# machinedeployment.cluster.x-k8s.io/worker-md-00evb     fake-cluster               1                   1                                        2m47s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-37os8           2m47s
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   16m
```

Verify that a `Node` has been removed.

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE   VERSION
# fake-cluster-control-plane   Ready      control-plane   15m   v1.36.1
# fake-node-00                 NotReady   <none>          15m   
# fake-node-02                 Unknown    <none>          38s   
# fake-node-control-plane-00   NotReady   <none>          15m   
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-up-new-md
```

#### Node Deletion with MachineDeployment Removal

Create a `NodeScalingRequest` specifying the `Node` to be deleted.

```sh
# set node created by 'Node Addition with MachineDeployment Creation'
export NODE_TO_DELETE="fake-node-02"
envsubst < example/nsr/dci/nsr-scaledown.yaml | kw apply -f -
```

Verify that `nodescalingrequest.status.phase` becomes `Deleting`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE      AGE
# scale-down   delete   Deleting   2s
```

Verify that the `MachineDeployment` and `MachineSet` have been deleted.

```sh
kubectl get md,ms,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        17m   

# NAME                                            CLUSTER        DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   AGE   VERSION
# machineset.cluster.x-k8s.io/fake-worker-ms-00   fake-cluster   1                                                    17m   v1.36.1

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   18m
```

To simulate the operation of Cluster API, manually delete the dummy `Machine` and `Node` resources.

```sh
kubectl delete machine fake-worker-machine-02
kw delete node fake-node-02
```

Verify that `nodescalingrequest.status.phase` becomes `Deleted`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   46s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has been removed.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        18m   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   19m
```

Verify that a Node has been removed.

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE   VERSION
# fake-cluster-control-plane   Ready      control-plane   17m   v1.36.1
# fake-node-00                 NotReady   <none>          17m   
# fake-node-control-plane-00   NotReady   <none>          17m   
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-down
```

### Cleanup

Delete the Workload Cluster and the Management Cluster.

```sh
kind delete cluster -n ${CLUSTER_NAME}
kind delete cluster -n ${M_CLUSTER_NAME}
```
