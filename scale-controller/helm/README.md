Helm chart for the Scale Controller
===========================

## Overview
This document describes how to create, install and uninstall a helm chart for the Scale Controller.

## Prerequisites
- kubebuilder v4.10.0+ (recommended: v4.10.1+)
- helm v3.17.0+

## Create a helm chart

You can create a helm chart by running
```sh
make helmchart
```
under the directory containing the `Makefile`.
The helm chart will be created under the `dist` directory with the name `scale-controller-<version>.tgz`.

Generation flow (helm/v2-alpha):
- Input: `config` (kustomize)
- Step 1: generate `dist/install.yaml`
- Step 2: generate `dist/chart` from `dist/install-scale-controller.yaml`
- Step 3: complement uncovered parts with `helm/patches/chart`
- Values policy: keep `helm/patches/chart/values.yaml` as user-editable source and sync it into `dist/chart/values.yaml`

## Install

The image repositories of Scale Controller must be accessible from the Workload cluster.
You can install Scale Controller by running the following command.

```sh
 # Set appropriate values
path_to_helm_chart=dist/scale-controller-v0.3.0.tgz
controller_image_tag=0.3.0
receiver_image_tag=0.3.0
path_to_management_kubeconfig=/path/to/management.kubeconfig
```

Note:
- `managementKubeconfig` default value in `values.yaml` is `dummy-kubeconfig` for local rendering/tests.
- For production operation, always provide a real kubeconfig via `--set-file managementKubeconfig=...`.

### Install with Alert Receiver
```sh
helm install scale-controller ${path_to_helm_chart} \
    --create-namespace \
    --namespace scale-controller-system \
    --set manager.image.repository=scale-controller \
    --set manager.image.tag=${controller_image_tag} \
    --set alertReceiver.container.image.repository=card-alert-receiver \
    --set alertReceiver.container.image.tag=${receiver_image_tag} \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

If you see the following message, the operation was successful.

```sh
# Expected results:
# NAME: scale-controller
# LAST DEPLOYED: Mon Apr  6 14:43:02 2026
# NAMESPACE: scale-controller-system
# STATUS: deployed
# REVISION: 1
# TEST SUITE: None
```

### Install without Alert Receiver
```sh
helm install scale-controller ${path_to_helm_chart} \
    --create-namespace \
    --namespace scale-controller-system \
    --set manager.image.repository=scale-controller \
    --set manager.image.tag=${controller_image_tag} \
    --set alertReceiver.enable=false \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

If you see the following message, the operation was successful.

```sh
# Expected results:
# NAME: scale-controller
# LAST DEPLOYED: Mon Apr  6 14:36:50 2026
# NAMESPACE: scale-controller-system
# STATUS: deployed
# REVISION: 1
# TEST SUITE: None
```

## Uninstall

You can uninstall Node Scaling Request Controller by running

```sh
helm uninstall scale-controller -n scale-controller-system
# Expected results:
# These resources were kept due to the resource policy:
# [CustomResourceDefinition] cardalerts.cluster.my.domain
# [CustomResourceDefinition] cardscalingrequests.cluster.my.domain
# [CustomResourceDefinition] nodescalingrequests.cluster.my.domain
# [Certificate] serving-cert
# [Certificate] metrics-certs
# 
# release "scale-controller" uninstalled
```

If you want to delete the Namespace or CustomResourceDefinition, delete it manually.

```sh
kubectl delete namespace scale-controller-system
# Expected results:
# namespace "scale-controller-system" deleted
kubectl delete crd nodescalingrequests.cluster.my.domain cardscalingrequests.cluster.my.domain cardalerts.cluster.my.domain
# Expected results:
# customresourcedefinition.apiextensions.k8s.io "nodescalingrequests.cluster.my.domain" deleted
# customresourcedefinition.apiextensions.k8s.io "cardscalingrequests.cluster.my.domain" deleted
# customresourcedefinition.apiextensions.k8s.io "cardalerts.cluster.my.domain" deleted
```
