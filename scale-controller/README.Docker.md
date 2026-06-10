Scale Controller for Docker Provider
===================================================

[toc]

## Note

Only differences from [README.md](README.md) are described in this file.

Processing for DCI Provider cannot be executed when DockerProvider Mode is enabled.


## Overview

Scale Controller is consist of below controllers:
- NodeScalingRequest Controller
- CardScalingRequest Controller

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
  location: docker-machine-template-name-you-need
  serverSpec:
    type: COTSServer # dummy
```

**Note**: The `spec.location` filed dedicates the name of DockerMachineTemplate you need.
And the `spec.serverSpec` field is ignored by NodeScalingRequest Controller when DockerProvider Mode is enabled, but `spec.serverSpec.type` should be set to avoid failure of  `kubectl apply` for NodeScalingRequest.

If there is no `DockerMachineTemplate` or `MachineDeployment` in the Management Cluster that meets the requested specifications, the corresponding resources are newly created and then a request to add a `Node` is made.


To request deletion of a node, see [Overview](README.md#overview).

When a `NodeScalingRequest` is created in the Workload Cluster, the NodeScalingRequest Controller detects it and requests the Management Cluster running Cluster API to add or remove a `Node`.
The following describes each flow.

#### Node Addition

When the NodeScalingRequest Controller detects a `NodeScalingRequest` for adding a `Node`, it operates as follows.

1. Extract the following `MachineDeployment` resources from the Management Cluster:
    - The value of `machinedeployment.spec.clusterName` matches the value of the annotation `cluster.x-k8s.io/cluster-name` of the `Node` in the Workload Cluster.
    - The value of `machinedeployment.namespace` matches the value of the annotation `cluster.x-k8s.io/cluster-namespace` of the `Node` in the Workload Cluster.
    - The value of `machinedeployment.spec.template.spec.infrastructureRef.kind` is `DockerMachineTemplate`.
    - The value of `machinedeployment.spec.template.spec.infrastructureRef.name` matches the value of `nodescalingrequest.spec.location`

2. See [Node Addition:2](README.md#node-addition).

3. If no matching `MachineDeployment` is found in step 1, create a new `MachineDeployment` and set `nodescalingrequest.status.desiredReplicas` to 1.
    - If there is no `DockerMachineTemplate` in the cluster namespace that satisfies `nodescalingrequest.spec.location`, create a new `DockerMachineTemplate` with the same name as `nodescalingrequest.spec.location`.

4. See [Node Addition:4](README.md#node-addition).

5. See [Node Addition:5](README.md#node-addition).


#### Node Deletion

See [Node Deletion](README.md#node-deletion).

### Card Scaling Request Controller

The CardScalingRequest Controller is only supported when using the DCI Provider, so do not use it when the Docker Provider is enabled.


## Quick Start

### Preparation

You don't have to prepare source code for DCI Provider (cluster-api-provider-dci) and Resource Pool Manager (resource-pool-manager).

Move to the root directory of the Scale Controller source code.

```sh
cd /path/to/scale-controller
```

### Verified Environments
- k8s: v1.34.2 (Management Cluster + Workload Cluster)
- cluster-api: v1.12.0 (v1beta2 contract)
- cert-manager: v1.19.2
- calico: v3.31.2

### Version for quick start

Set following shell variables adjusting your environment.

```sh
calico_version=v3.31.2
cluster_api_version=v1.12.0
cert_manager_version=v1.19.2
helm_chart_version=v0.3.0
controller_img_tag=0.3.0
```

#### Setup of Management Cluster and Workload Cluster

Create the Management Cluster using `kind` and the Workload Cluster using `kubectl apply` with `clusterctl`.

##### Management Cluster

```sh
export M_CLUSTER_NAME="my-management"
export CLUSTER_NAME="my-simple-cluster"
# If you are in a proxy environment, run the following as needed
export NO_PROXY=$NO_PROXY,172.18.0.0/16

# Create Management Cluster using kind
kind create cluster -n ${M_CLUSTER_NAME} --config example/kind/kind-cluster-with-extramounts.yaml

# Deploy the Cluster API and Docker Provider on the Management Cluster.
clusterctl init --infrastructure docker

# Wait until all pods are running.
kubectl get pod -A
# expected results:
# NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
# capd-system                         capd-controller-manager-5697dcf77d-cqx4v                         1/1     Running   0          2m6s
# capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-7976f6bf86-5qh7r       1/1     Running   0          2m8s
# capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-784ff8d7fd-hwdld   1/1     Running   0          2m7s
# capi-system                         capi-controller-manager-84546dfc89-tr27t                         1/1     Running   0          2m8s
# cert-manager                        cert-manager-69fd4bc5fc-d4mzp                                    1/1     Running   0          2m27s
# cert-manager                        cert-manager-cainjector-85b6d7fc67-9wztl                         1/1     Running   0          2m27s
# cert-manager                        cert-manager-webhook-cfbc49fc8-4l7p5                             1/1     Running   0          2m27s
# kube-system                         coredns-66bc5c9577-b2f4w                                         1/1     Running   0          3m10s
# kube-system                         coredns-66bc5c9577-fj8v9                                         1/1     Running   0          3m10s
# kube-system                         etcd-my-management-control-plane                                 1/1     Running   0          3m17s
# kube-system                         kindnet-gpl44                                                    1/1     Running   0          3m11s
# kube-system                         kube-apiserver-my-management-control-plane                       1/1     Running   0          3m17s
# kube-system                         kube-controller-manager-my-management-control-plane              1/1     Running   0          3m17s
# kube-system                         kube-proxy-tgq79                                                 1/1     Running   0          3m11s
# kube-system                         kube-scheduler-my-management-control-plane                       1/1     Running   0          3m17s
# local-path-storage                  local-path-provisioner-5c4cdb564f-qtfnl                          1/1     Running   0          3m10s
```

##### Workload Cluster

```sh
# Pull calico images written in simple-cluster.yaml if needed
docker pull docker.io/calico/kube-controllers:${calico_version}
docker pull docker.io/calico/cni:${calico_version}
docker pull docker.io/calico/node:${calico_version}
```

```sh
# Create Workload Cluster using yaml file for Cluster API
envsubst < example/cluster-api/cluster-class-disabled/simple-cluster.yaml | kubectl apply -f -

# After a while, check "NODE NAME" is set for Machine
# FYI: Cluster will be available after installing calico.
kubectl get cluster,machine
# expected results:
# NAME                                         CLUSTERCLASS   AVAILABLE   CP DESIRED   CP AVAILABLE   CP UP-TO-DATE   W DESIRED   W AVAILABLE   W UP-TO-DATE   PHASE         AGE    VERSION
# cluster.cluster.x-k8s.io/my-simple-cluster                  False       1            0              1               1           0             1              Provisioned   4m1s   

# NAME                                               CLUSTER             NODE NAME                                   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE     VERSION
# machine.cluster.x-k8s.io/controlplane-4klb7        my-simple-cluster   my-simple-cluster-controlplane-4klb7        False   False       True         Running   3m59s   v1.34.2
# machine.cluster.x-k8s.io/worker-md-0-jjzlz-sts28   my-simple-cluster   my-simple-cluster-worker-md-0-jjzlz-sts28   False   False       True         Running   3m45s   v1.34.2

# Create kubeconfig file for workload cluster and set alias
clusterctl get kubeconfig ${CLUSTER_NAME} > ${PWD}/${CLUSTER_NAME}.kubeconfig
alias kw="kubectl --kubeconfig ${PWD}/${CLUSTER_NAME}.kubeconfig"

# Install Calico.
kw apply -f https://raw.githubusercontent.com/projectcalico/calico/${calico_version}/manifests/calico.yaml
# Confirm all pods are running.
kw get pod -A
# expected results:
# NAMESPACE     NAME                                                           READY   STATUS    RESTARTS   AGE
# kube-system   calico-kube-controllers-5c676f698c-rxhc5                       1/1     Running   0          111s
# kube-system   calico-node-8d7sv                                              1/1     Running   0          111s
# kube-system   calico-node-dfzbb                                              1/1     Running   0          111s
# kube-system   coredns-66bc5c9577-f75g2                                       1/1     Running   0          5m53s
# kube-system   coredns-66bc5c9577-vfpxv                                       1/1     Running   0          5m53s
# kube-system   etcd-my-simple-cluster-controlplane-q8s5j                      1/1     Running   0          5m57s
# kube-system   kube-apiserver-my-simple-cluster-controlplane-q8s5j            1/1     Running   0          5m59s
# kube-system   kube-controller-manager-my-simple-cluster-controlplane-q8s5j   1/1     Running   0          5m55s
# kube-system   kube-proxy-7kqdg                                               1/1     Running   0          4m26s
# kube-system   kube-proxy-r5zft                                               1/1     Running   0          5m53s
# kube-system   kube-scheduler-my-simple-cluster-controlplane-q8s5j            1/1     Running   0          5m55s

# Verify that desired replicas and ready replicas are both 1
kubectl get md,dockermachinetemplate
# expected results:
# NAME                                             CLUSTER             AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE    VERSION
# machinedeployment.cluster.x-k8s.io/worker-md-0   my-simple-cluster   True        1         1         1       1           1            Running   8m1s   v1.34.2

# NAME                                                                 AGE
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/controlplane   8m1s
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/worker         8m1s

# Verify Node resources
kw get node
# expected results:
# NAME                                        STATUS   ROLES           AGE     VERSION
# my-simple-cluster-controlplane-q8s5j        Ready    control-plane   6m56s   v1.34.2
# my-simple-cluster-worker-md-0-65g6b-92qqz   Ready    <none>          5m19s   v1.34.2
```

#### Deploying the Scale Controller

See [Setup of Management Cluster and Workload Cluster](README.md#setup-of-management-cluster-and-workload-cluster) about description.

##### Deploying the Scale Controller Using an Existing Helm Chart

You need to add --set option for a environment variable(`PROVIDER_NAME`) to switch provider mode.

```sh
# Note:
#    Make sure that the current-context in your kubeconfig is set to the target Workload Cluster.
#    If not, adjust by adding the --kube-context or --kubeconfig option as needed.
cluster_context="--kubeconfig=${CLUSTER_NAME}.kubeconfig"
path_to_helm_chart=dist/chart/
helm_controller_image_tag=${controller_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set controllerManager.container.env.PROVIDER_NAME=Docker \
    --set controllerManager.container.image.repository=scale-controller \
    --set controllerManager.container.image.tag=${helm_controller_image_tag} \
    --set alertReceiver.enable=false \
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
cluster_context="--kubeconfig=${CLUSTER_NAME}.kubeconfig"
path_to_helm_chart=dist/scale-controller-${helm_chart_version}.tgz
helm_controller_image_tag=${controller_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set controllerManager.container.env.PROVIDER_NAME=Docker \
    --set controllerManager.container.image.repository=scale-controller \
    --set controllerManager.container.image.tag=${helm_controller_image_tag} \
    --set alertReceiver.enable=false \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

##### Deploying the Scale Controller Using Kustomize

**Note:**
When you want to deploy Node Scaling Request Controller by `make deploy` without using helm, you should change value of the environment variable under `config/manager/manager.yaml` that determines whether or not you are in DockerProvider mode.

```yaml
        # This field determines supporting provider: DCI, Docker
        - name: PROVIDER_NAME
          value: Docker
```

```sh
export KUBECTL="kubectl --kubeconfig ${PWD}/${CLUSTER_NAME}.kubeconfig"
make deploy
```

Register the previously created kubeconfig as a `Secret` in the "scale-controller-system" namespace of the Workload Cluster.

```sh
kw create secret generic management-kubeconfig \
    --from-file=kubeconfig=management.kubeconfig \
    -n scale-controller-system
```

#### Deploying cert-manager

See [Deploying cert-manager](README.md#deploying-cert-manager).

#### Building the Scale Controller

Build the container image and load it into the Workload Cluster.

Depending on your environment, add `sudo` as needed.
Note that when passing user environment variables to the `sudo` command, you may need to add the `-E` option as appropriate.

```sh
export IMG=scale-controller:${controller_img_tag}
make docker-build
docker images ${IMG}
# expected results:
# IMAGE                    ID             DISK USAGE   CONTENT SIZE   EXTRA
# scale-controller:0.3.0   a69fe6e286de       82.4MB             0B        
```

```sh
kind load docker-image -n ${CLUSTER_NAME} ${IMG}
```

#### Setting Up Credentials

Set up authentication credentials for the Scale Controller on the Management Cluster.

```sh
kubectl apply -f example/rbac/rbac-in-management-cluster-docker.yaml
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

Create a `NodeScalingRequest` that requests the addition of a `Node` matching the existing `DockerMachineTemplate`.

```sh
kw apply -f example/nsr/docker/nsr-scaleup-existing-md.yaml 
```

Wait a while and verify that `nodescalingrequest.status.phase` becomes `Created`.

```sh
kw get nsr
# expected results:
NAME                TYPE     PHASE     AGE
scale-up-existing   create   Created   4m26s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has increased to 2.

```sh
kubectl get md,dockermachinetemplate
# expected results:
# NAME                                             CLUSTER             AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE   VERSION
# machinedeployment.cluster.x-k8s.io/worker-md-0   my-simple-cluster   True        2         2         2       2           2            Running   17m   v1.34.2

# NAME                                                                 AGE
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/controlplane   17m
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/worker         17m
```

Verify that a Node has been added.

```sh
kw get node
# expected results:
# NAME                                        STATUS   ROLES           AGE     VERSION
# my-simple-cluster-controlplane-q8s5j        Ready    control-plane   16m     v1.34.2
# my-simple-cluster-worker-md-0-65g6b-92qqz   Ready    <none>          14m     v1.34.2
# my-simple-cluster-worker-md-0-65g6b-944xx   Ready    <none>          4m30s   v1.34.2
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-up-existing
```

#### Node Deletion

Create a `NodeScalingRequest` specifying the `Node` to be deleted.

```sh
export NODE_TO_DELETE="my-simple-cluster-worker-md-0-65g6b-944xx" # Adjust to your environment
envsubst < example/nsr/docker/nsr-scaledown.yaml | kw apply -f -
```

Wait a while and verify that `nodescalingrequest.status.phase` becomes `Deleted`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   9s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has decreased to 1.

```sh
kubectl get md,dockermachinetemplate
# expected results:
# NAME                                             CLUSTER             AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE   VERSION
# machinedeployment.cluster.x-k8s.io/worker-md-0   my-simple-cluster   True        1         1         1       1           1            Running   20m   v1.34.2

# NAME                                                                 AGE
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/controlplane   20m
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/worker         20m
```

Verify that a Node has been removed.

```sh
kw get node
# expected results:
# NAME                                        STATUS   ROLES           AGE   VERSION
# my-simple-cluster-controlplane-q8s5j        Ready    control-plane   19m   v1.34.2
# my-simple-cluster-worker-md-0-65g6b-92qqz   Ready    <none>          17m   v1.34.2
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-down
```

#### Node Addition with MachineDeployment Creation

Create a `NodeScalingRequest` that requests the addition of a `Node` for which no matching `DockerMachineTemplate` exists.

```sh
kw apply -f example/nsr/docker/nsr-scaleup-new-md.yaml
```

Verify that `nodescalingrequest.status.phase` becomes `Created`.

```sh
kw get nsr
# expected results:
# NAME              TYPE     PHASE     AGE
# scale-up-new-md   create   Created   36s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has increased to 1.

```sh
kubectl get md,dockermachinetemplate
# expected results:
# NAME                                                 CLUSTER             AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE   VERSION
# machinedeployment.cluster.x-k8s.io/worker-md-0       my-simple-cluster   True        1         1         1       1           1            Running   22m   v1.34.2
# machinedeployment.cluster.x-k8s.io/worker-md-8ggbp   my-simple-cluster   True        1         1         1       1           1            Running   63s   v1.34.2

# NAME                                                                            AGE
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/controlplane              22m
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/new-docker-machine-type   63s
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/worker                    22m
```

Verify that a Node has been removed.

```sh
kw get node
# expected results:
# NAME                                            STATUS   ROLES           AGE   VERSION
# my-simple-cluster-controlplane-q8s5j            Ready    control-plane   20m   v1.34.2
# my-simple-cluster-worker-md-0-65g6b-92qqz       Ready    <none>          19m   v1.34.2
# my-simple-cluster-worker-md-8ggbp-rlvd6-rvmt4   Ready    <none>          75s   v1.34.2
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-up-new-md
```

#### Node Deletion with MachineDeployment Removal

Create a `NodeScalingRequest` specifying the `Node` to be deleted.

```sh
# set node created by 'Node Addition with MachineDeployment Creation'
export NODE_TO_DELETE="my-simple-cluster-worker-md-8ggbp-rlvd6-rvmt4"
envsubst < example/nsr/docker/nsr-scaledown.yaml | kw apply -f -
```

Verify that `nodescalingrequest.status.phase` becomes `Deleted`.

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   5s
```

Verify that `machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` has increased to 1.

```sh
kubectl get md,dockermachinetemplate
# expected results:
# NAME                                             CLUSTER             AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE   VERSION
# machinedeployment.cluster.x-k8s.io/worker-md-0   my-simple-cluster   True        1         1         1       1           1            Running   23m   v1.34.2

# NAME                                                                 AGE
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/controlplane   23m
# dockermachinetemplate.infrastructure.cluster.x-k8s.io/worker         23m
```

Verify that a Node has been removed.

```sh
kw get node
# expected results:
# NAME                                        STATUS   ROLES           AGE   VERSION
# my-simple-cluster-controlplane-q8s5j        Ready    control-plane   22m   v1.34.2
# my-simple-cluster-worker-md-0-65g6b-92qqz   Ready    <none>          20m   v1.34.2
```

Delete the `NodeScalingRequest`.

```sh
kw delete nsr scale-down
```

### Cleanup

Delete the Workload Cluster.

```sh
kubectl delete cluster ${CLUSTER_NAME}
```

Delete the Management Cluster.

```sh
kind delete cluster -n ${M_CLUSTER_NAME}
```
