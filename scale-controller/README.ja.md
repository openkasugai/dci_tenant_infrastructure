Scale Controller
===============================

[toc]

## 概要

Scale Controller は、以下のコントローラ群によって構成されている。
なお、 CardAlert AlertReceiver は別 `Pod` としてデプロイされ、その他のコントローラは一つの `Pod` 内にて動作する。
- NodeScalingRequest Controller
- CardScalingRequest Controller
- CardAlert Controller
- CardAlert AlertReceiver

### NodeScalingRequest Controller

NodeScalingRequest Controller は、 Cluster API で作成した Workload Cluster にデプロイされ、デプロイされた自 Cluster のスケーリングつまり、 `Node` の追加または削除を Management Cluster に対して要求するコントローラであり、 `NodeScalingRequest` は本機能を実現するためのカスタムリソースである。

`Node` の追加を要求する場合、要求種別として `create` を設定し、追加したいサーバのスペックを指定する。

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

**Note**: `spec.location` および `spec.serverSpec`の詳細については、 DCI Manager にて定義されているため、詳細については `DCIMachine` の CRD などを参照のこと。

要求スペックを満たす `DCIMachineTemplate` または `MachineDeployment` が Management Cluster に存在しない場合は、新たなに該当リソース作成した上で `Node` 追加要求を行う。

`Node` の削除を要求する場合、要求種別として `delete` を設定し、削除したい `Node` の名前を指定する。

```yaml
kind: NodeScalingRequest
metadata:
  name: example-scaledown-request
  namespace: default
spec:
  nodeRequestType: delete
  node: my-cluster-worker-md-0-cc2jr-h5pzg
```

`Node` 削除の結果、指定した `Node` を作成していた `MachineDeployment` のレプリカ数が0となるような場合には、他に `MachineDeployment` が存在するか確認する。
  - 同 Namespace 、同 Cluster　、同インフラ種別かつ ClusterAutoScaler 無効な `MachineDeployment` が他に存在する場合は、対象の `Node` で使用していた `MachineDeployment` を削除する
    - `MachineDeployment` を削除する場合は、その `MachineDeployment` が参照していた `DCIMachineTemplate` が他の `MachineDeployment` で使用されているかを確認する
    - 対象の `DCIMachineTemplate` を使用している `MachineDeployment` が自身の他に存在しない場合は、対象の `DCIMachineTemplate` も併せて削除する。
  - 対象の `MachineDeployment` が唯一の `MachineDeployment` である場合は、削除せずレプリカ数を0とする

`NodeScalingRequest` を Workload Cluster に作成すると、 NodeScalingRequest Controller がこれを検知し、 Cluster API が動作する Management Cluster に対して `Node` の追加、または削除を要求する。
以下にそれぞれのフローを示す。なお、 NodeScalingRequest Controller の仕様として、同時に最大一つの `NodeScalingRequest` しか処理しないため、ある `NodeScalingRequest` が処理進行中は、他の `NodeScalingRequest` は保留状態に遷移して、一定時間ごとにリキューして処理の開始を待つ。

#### Node 追加のフロー

`Node` 追加の `NodeScalingRequest` を検知した NodeScalingRequest Controller は次のように動作する。

1. Management Cluster の `MachineDeployment` から次にあてはまるものを抽出する
    - `machinedeployment.spec.clusterName` の値が、 Workload Cluster の `Node` のアノテーション `cluster.x-k8s.io/cluster-name` の値と一致する
    - `machinedeployment.namespace` の値が、 Workload Cluster の `Node` のアノテーション `cluster.x-k8s.io/cluster-namespace` の値と一致する
    - `machinedeployment.spec.template.spec.infrastructureRef.kind` の値が `DCIMachineTemplate` を満たす
    - `machinedeployment.spec.template.spec.infrastructureRef.name` で示される `DCIMachineTemplate` が、 `nodescalingrequest.spec.serverSpec` を満たすか判定する
      - `nodescalingrequest.spec.serverSpec` に設定された各フィールドについて AND 条件で検索
      - `nodescalingrequest.spec.serverSpec` に設定されていないフィールドはDon't care
      - `nodescalingrequest.spec.serverSpec` に設定されており、 `dcimachinetemplate.spec.template.spec` が未設定（nil）の場合は満たさないものとして扱う

2. 1.で一致する `MachineDeployment` を抽出できた場合、その `machinedeployment.spec.replicas` を +1 し、その値を `nodescalingrequest.status.desiredReplicas` に設定する

3. 1.で一致する `MachineDeployment` を抽出できなかった場合、 `MachineDeployment` を新たに作成し、 `nodescalingrequest.status.desiredReplicas` に 1 を設定する
    - `nodescalingrequest.spec.serverSpec` を満たす `DCIMachineTemplate` が同 Namespace 内に存在しない場合、 `nodescalingrequest.spec.serverSpec` と同じ内容で `DCIMachineTemplate` を新たに作成する

4. `nodescalingrequest.status.phase` を `Creating` に更新する

5. Workload Cluster の `Node` に対する Create Event を監視し、 `machinedeployment.status.readyReplicas` が`nodescalingrequest.status.desiredReplicas` 以上になるまで待ち受け、 `nodescalingrequest.status.phase` を `Created` に更新する

#### Node 削除のフロー

`Node` の削除の `NodeScalingRequest` を検知した NodeScalingRequest Controller は次のように動作する。

1. `nodescalingrequest.spec.node` で示される Workload Cluster の `Node` を取得し、アノテーション `cluster.x-k8s.io/machine` の値を削除対象の `Machine` の名前とし、 `cluster.x-k8s.io/cluster-namespace` の値を `Machine` の namespace とする

2. 削除対象の `Machine` を取得し、ラベル `cluster.x-k8s.io/deployment-name` の値をスケールダウン対象の `MachineDeployment` の名前とし、1.で取得した namespace を namespace とする

3. 削除対象の `MachineDeployment` を取得する

4. 削除対象の `machinedeployment.spec.replicas` が 1 の場合(つまりスケールダウンにより0になる場合)は次のように動作する
    1. `MachineDeployment` のリストを取得し、他に妥当な `MachineDeployment` リソースが存在するか確認する

    2. 4-1.で他に妥当な `MachineDeployment` が存在しない場合、4-3.以降の処理を実行せずに5.を実行する

    3. 4-1.で他に妥当な `MachineDeployment` が存在する場合、その `MachineDeployment` が使用している `DCIMachineTemplate` が他の`MachineDeployment` にて使用されているか確認する
        - 他の `MachineDeployment` にて使用されていない場合のみ、対象の `DCIMachineTemplate` を削除する

    4. 削除対象の `MachineDeployment` を削除し、 `nodescalingrequest.status.desiredReplicas` に 0 を設定する

    5. `nodescalingrequest.status.phase` を `Deleting` に更新する

    6. Workload Cluster の `Node` の Delete Event を監視し、`Node` 削除確認後に `nodescalingrequest.status.phase` を `Deleted` に更新する

5. スケールダウン対象の `machinedeployment.spec.replicas` が 1 以外の場合は次のように動作する
    1. Management Cluster における削除対象の `MachineSet` に、アノテーション `cluster.x-k8s.io/paused` の値を `true` に設定する

    2. 削除対象の `machinedeployment.spec.replicas` を -1 し、その値を `nodescalingrequest.status.desiredReplicas` に設定する

    3. 削除対象の `Machine` を削除する

    4. `nodescalingrequest.status.phase` を `Deleting` に更新する

    5. Workload Cluster の `Node` を監視し、削除対象の `machinedeployment.status.readyReplicas` が `nodescalingrequest.status.desiredReplicas` 以下になったら`nodescalingrequest.status.phase` を `Deleted` に更新する


### CardScalingRequest Controller

CardScalingRequest Controller は、 Cluster API で作成した Workload Cluster にデプロイされ、デプロイされた自 Cluster 内に存在している `Node` に搭載されている GPU の枚数スケーリング及び交換を Management Cluster に対して要求するコントローラであり、 `CardScalingRequest` は本機能を実現するためのカスタムリソースである。

**Note**: 以降 `CardScalingRequest` にて記載される `.spec.scalingParameter.model` 、 `.spec.scalingParameter.uuids` の値は、 Management Cluster に属している `LogicalServer` に記載されている GPU のモデル名、 UUID と一致している必要がある。

GPU の追加を要求する場合、要求種別として `add` を設定し、追加したい GPU のモデルと追加したい枚数を指定する。
なお、現在サポートしているカード追加は、すでに存在しているモデルのカード枚数を増やすケースのみであり、
そのため、別のモデルのカードを追加しようとするようなカードの枚数追加は実施できない。
また、 UUID を指定して追加することもできない。

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
    num: 1 # 枚数指定
```

GPU の削除を要求する場合、要求種別として `delete` を設定し、削除したいカードの枚数か、 その UUID のいずれかを指定する。
なお、枚数指定と UUID 指定を組み合わせて削除することは出来ない。
また、0枚になるようなカード削除は許容される。

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
    num: 1 # 枚数指定
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
    - GPU-01234567-89ab-cdef-0123-456789abcdef # UUID指定
```

GPU の交換を要求する場合、要求種別として `replace` を設定し、交換したいカードの UUID を指定する。

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
    - GPU-01234567-89ab-cdef-0123-456789abcdef # UUID指定
```

`CardScalingRequest` を Workload Cluster に作成すると、 CardScalingRequest Controller がこれを検知し、必要に応じて Cluster API が動作する Management Cluster に対して適切なリソースの作成を行った後に、 `NodeScalingRequest` を通じて対象の `Node` に対してカード増減・交換を要求する。
以下にそのフローを示す。なお、 CardScalingRequest Controller の仕様として、同時に最大一つの `CardScalingRequest` しか処理しないため、ある `CardScalingRequest` が処理進行中は、他の `CardScalingRequest` は保留状態に遷移して、一定時間ごとにリキューして処理の開始を待つ。

#### カード増減・交換のフロー

`CardScalingRequest` を検知した CardScalingRequest Controller は、その要求種別によらず、次のように動作する。
なお、同スペックの `MachineDeployment`, `DCIMachineTemplate` が複数存在する場合期待した動作とならない場合がある点に注意すること。

1. `cardscalingrequest.spec.nodeName` で示される Workload Cluster の `Node` を取得し、該当リソースに紐づいている Management Cluster 上の `Machine`, `MachineDeployment`, `DCIMachineTemplate`, `DCIMachine`, `LogicalServer` を取得する。

2. `cardscalingrequest.spec.scalingParameter` の内容と `DCIMachineTemplate` を確認し、カード追加・削除後のスペックを満たすような `DCIMachineTemplate` が 1. で取得した `DCIMachineTemplate` と同 `Namespace` 上に存在しているかを確認し、存在していなければ作成する。

3. 2.で取得した要求スペックを満たすような `DCIMachineTemplate` を参照しており、 `ClusterAutoscaler` の対象となっていない `MachineDeployment` が存在しているかを確認し、存在していなければ作成する。

4. 1.で取得した `DCIMachine` に対して、 `cardscalingrequest.spec.scalingParameter` をもとにしてカード増減・交換要求をアノテーションを付与する。

5. 1.で取得した `Node` を削除するような `NodeScalingRequest` を作成し、その成功を待つ。 `NodeScalingRequest` が失敗した場合は、 `CardScalingRequest` もエラーとして処理を終了する。

5. 2.で取得したカード増減後の `DCIMachineTemplate` の要求スペックを満たすような `Node` を追加するための `NodeScalingRequest` を作成し、その成功を待つ。 `NodeScalingRequest` が失敗した場合は、 `CardScalingRequest` もエラーとして処理を終了する。


### CardAlert Controller

CardAlert Controller は、 Cluster API で作成した Workload Cluster にデプロイされ、主に CardAlert AlertReceiver によって作成された自 Cluster 内に存在している `Pod` の障害情報を元に作成された `CardAlert` リソース、環境変数から取得した対応方針及び障害が発生した `Node` の種別 (`Composed`,`COTS`,...) をもとに適切な障害対応を行うコントローラであり、 `CardAlert` は本機能を実現するためのカスタムリソースである。
なお、 `CardAlert` はユーザの手によって作成されることを期待していない。

CardAlert Controller の障害対応時の対応方針は大別すると以下のとおりである。
- 障害カードの自動交換
  - 起動時に環境変数から取得され、有効化するか無効化するかを選択可能
  - デフォルト設定は有効
- 自動交換有効化時: 自動交換方針
  - `Node` を構成するサーバの情報を取得し、その種別により下記の通り処理が変わる
  - `COTS`
    - Management Cluster に対して `MachineAlert` リソースを作成し、対象の `Node` を交換する
  - `Composed`
    - Workload Cluster に対して `CardScalingRequest` リソースを作成し、対象の `Node` を交換する
  - `VM`
    - サポート対象外
- 自動交換無効化時: `Pod` 障害対応方針
  - 起動時に環境変数から取得され、障害発生時に障害発生しているカードに対する `DeviceTaintRule` の処理方針を選択可能
  - デフォルト設定は `NoExecute`
  - `NoExecute`
    - 対象の障害カードを使用している `Pod` を削除し別の `Pod` を作成しようとする
  - `NoSchedule`
    - 対象の障害カードを使用している `Pod` を削除しないが、仮に該当 `Pod` 削除後に同種カードを要求されても払い出さないようにする

> **Note**
自動交換無効時に作成する `DeviceTaintRule` は現時点ではアルファ版の機能であり、 k8s 起動時に feature-gate の設定などを別途実施する必要がある。
また、 K8s `v1.34.2` の `DeviceTaintRule` のリソース定義が `v1.35.0` 以降では変更され、期待通り動作しない可能性がある。

`CardAlert` を Workload Cluster に作成すると、 CardAlert Controller がこれを検知し、上記の通りその設定に応じて適切なリソースを作成し障害対応を実施する。
以下にそのフローを示す。なお、 CardAlert Controller の仕様として、同時に最大一つの `CardAlert` しか処理しないため、ある `CardAlert` が処理進行中は、他の `CardAlert` は保留状態に遷移して、一定時間ごとにリキューして処理の開始を待つ。

#### 自動交換有効時フロー

自動交換有効時の CardAlert Controller は `CardAlert` を検知したとき、次のように動作する。

1. `cardalert.spec.nodeName` で示される Workload Cluster の `Node` を取得し、該当リソースに紐づいている Management Cluster 上の `Machine`, `DCIMachine`, `LogicalServer` を取得する。

2. 1.で取得した `LogicalServer` 上に `cardalert.spec.cardInfo` で指定されているカードが存在しているか確認し、すべてのカードが存在していなければ処理を終了する。

3. 1.で取得した `LogicalServer` のサーバ種別を取得し、その種別に応じて以下の通り処理を進める。
  1. `COTS` の場合は、 Management Cluster 上に `MachineAlert` を作成しその終了を待つ。終了を確認後、作成した `MachineAlert` を削除し、 `CardAlert` も処理を終了する。
    - `MachineAlert` を作成する `Namespace` は環境変数から取得するが、 MachineAlert Controller が `MachineAlert` を Watch している `Namespace` (通常 `dcirpm`) である必要がある点に注意。
  2. `Composed` の場合は、 OwnerReference を設定しつつ `CardScalingRequest` を作成し、その終了を待つ。終了を確認後、 `CardAlert` も処理を終了する。なお、 `CardScalingRequest` が失敗した場合は、 `CardAlert` も処理失敗として処理を終了する。
    - `CardScalingRequest` の削除は `CardAlert` 削除に伴うガーベッジコレクタによって実施されることを期待しており CardAlert Controller による削除は実施しない。
  3. `VM` の場合は、 `CardAlert` を処理失敗として処理を終了する。

#### 自動交換無効時フロー

自動交換無効時の CardAlert Controller は `CardAlert` を検知したとき、次のように動作する。

1. 自動交換有効時の1.と同じ

2. 自動交換有効時の2.と同じ

3. 対象のカードに対する `DeviceTaintRule` を作成し、 `CardAlert` は処理を終了する。
  - Taint の Key, Value はそれぞれ固定値として、 `dra.example.com/unhealthy`, `Broken` を指定する。
  - Taint の Effect は起動時に環境変数から取得した `NoExecute` か `NoSchedule` を指定する。
  - `DeviceTaintRule` は Cluster-scoped リソースであるため、 `Namespace` は 指定しない。
  - 障害があると設定する DeviceSelector では、 GPU の UUID をもとに CEL 構文によるデバイス一致確認を行う。
    - MIG の UUID は有無チェックのみ行い、その値は使わない。
    - **NOTE**: CEL 構文によるデバイス一致確認は、k8s `v1.35.0` 以降サポートしていない可能性がある。
  - CardAlert Controller は `DeviceTaintRule` の削除は行わず、障害解決などした際にユーザによる手動削除を期待している。


### CardAlert AlertReceiver

CardAlert AlertReceiver は、 Cluster API で作成した Workload Cluster にデプロイされ、主に Prometheus の AlertManager から送信される自 Cluster の `Pod` の障害情報を受け取るためのサーバ (Webhook Receiver) として動作し、 Alert 情報を `CardAlert` として作成する。
これは、 Webhook Receiver が冪等的な操作を得意でないため、 Alert をカスタムリソースへ成形する操作のみを実施し、実際の障害対応の処理についてはカスタムコントローラに実施させる思想である。

CardAlert AlertReceiver の仕様を以下に示す。
- 起動時に環境変数から以下の情報を取得する。
  - `CARD_ALERT_NAMESPACE`
    - `CardAlert` を作成する `Namespace`
    - required
  - `ALERT_RECEIVER_LISTEN_ADDRESS`
    - サーバとして Listen するアドレス
    - optional
    - 未設定時は `0.0.0.0`
  - `ALERT_RECEIVER_LISTEN_PORT`
    - サーバとして Listen するポート
    - optional
    - 未設定時は `8080`
- TLS 認証無でサーバプロセスを起動し、 AlertManager からの Alert を待ち受ける
  - Endpoint
    - `/healthz`
      - health check に用いるエンドポイント
    - `/alerts`
      - AlertManager の送信先として期待しているエンドポイント
  - helm によるインストール時は、 `scale-controller-alert-receiver-service` という名前の `Service` を `8080` 番ポートを開いて作成するため、以下の URL を同クラスタの AlertManager の送信先として設定すればよい。 (`Namespace` は `helm install` 時に指定したものを使用のこと)
    - `http://scale-controller-alert-receiver-service.scale-controller-system.svc:8080/alerts`
- Alert をもとに `CardAlert` を作成する。
  - ステータスが `Firing` の Alert のみを対象とし、 `Resolved` の Alert は無視する。
  - 名前が `GPU ECC Double-Bit Error Detected` の Alert のみを対象とし、それ以外の Alert は無視する。
  - ラベルに `HostName`, `modelName`, `UUID` のいずれかが欠けている場合はエラー出力をし、対象 Alert 相当の `CardAlert` 作成は Skip する。
  - 環境変数から取得した `Namespace` を使用して `card-alert-receiver-` から始まる名前の `CardAlert` を作成する。
    - `.spec.nodeName` には Alert のラベル `HostName` の値を使用する。
    - `.spec.cards[].model` には Alert のラベル `modelName` の値を使用する。
    - `.spec.cards[].gpuUUID` には Alert のラベル `UUID` の値を使用する。
    - `.spec.cards[].migUUIDs[]` には Alert のラベル `dra_mig_device_uuid` の値を使用する。一度の Alert 受信で複数 Alert を受信し、その中で同一の `HostName`, `modelName`, `UUID` を持つ Alert が存在し、かつ `dra_mig_device_uuid` というフィールドがあればそれをまとめて設定する。


## 動作確認手順

### 事前準備

Scale Controller (scale-controller) 、 DCI 独自プロバイダ (cluster-api-provider-dci) およびリソースプール管理 (resource-pool-manager) のソースコードが以下の通り展開されていることを前提とする。
また、本動作確認は kind cluster にて実施することを想定している。

```
├── cluster-api-provider-dci
├── resource-pool-manager
├── scale-controller
```

Scale Controller のソースコードのルートディレクトリに移動する。

```sh
cd /path/to/scale-controller
```

### 確認済環境
- k8s: v1.34.2 (Management Cluster + Workload Cluster)
- cluster-api: v1.12.0 (v1beta2 contract)
- cert-manager: v1.19.2
- calico: n/a

### バージョン指定

以下のシェル変数を適宜設定する。

```sh
cluster_api_version=v1.12.0
cert_manager_version=v1.19.2
helm_chart_version=v0.3.0
controller_img_tag=0.3.0
receiver_img_tag=0.3.0
```

#### Management Cluster および Workload Cluster のセットアップ

Kind を使用して、 Management Cluster および Workload Cluster を作成する。
ここではクラスター名をそれぞれ "fake-management" 、"fake-cluster" とする。

##### Management Cluster

```sh
export M_CLUSTER_NAME="fake-management"
export CLUSTER_NAME="fake-cluster"
# If you are in a proxy environment, run the following as needed
export NO_PROXY=$NO_PROXY,172.18.0.0/16
kind_image=kindest/node:v1.34.2

# Create Management Cluster using kind
# If you do not need to specify the k8s image, the --image option is not required.
kind create cluster -n ${M_CLUSTER_NAME} --image=${kind_image}

# Deploy the Cluster API CRDs (Custom Resource Definitions) on the Management Cluster.
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/config/crd/bases/cluster.x-k8s.io_machinedeployments.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/config/crd/bases/cluster.x-k8s.io_machinesets.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api/refs/tags/${cluster_api_version}/config/crd/bases/cluster.x-k8s.io_machines.yaml

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
# fake-worker-md-00   fake-cluster               1                   1                                        3s    

# Verify that OwnerReference is set for MachineSet and others
kubectl get ms fake-worker-ms-00 -o jsonpath='{.metadata.ownerReferences}{"\n"}'
# expected results:
# [{"apiVersion":"cluster.x-k8s.io/v1beta2","blockOwnerDeletion":true,"controller":true,"kind":"MachineDeployment","name":"fake-worker-md-00","uid":"6973cc52-0ab5-41d1-add2-0e8dbf0c53ed"}]
kubectl get dcimachine fake-worker-dci-machine-00 -o jsonpath='{.metadata.ownerReferences}{"\n"}'
# expected results:
# [{"apiVersion":"cluster.x-k8s.io/v1beta2","blockOwnerDeletion":true,"controller":true,"kind":"Machine","name":"fake-worker-machine-00","uid":"251ac098-c939-41c1-8998-20fe9d9c5431"}]
```

**Note**: 期待したディレクトリに DCI 独自プロバイダ、リソースプール管理の CRD ファイルが存在しない場合は以下の通り生成する。
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
# fake-cluster-control-plane   NotReady   control-plane   7s    v1.34.2
# fake-node-00                 Unknown    <none>          0s    
# fake-node-control-plane-00   Unknown    <none>          0s    
```

#### Scale Controller のデプロイ

Scale Controller をデプロイするには、事前に次の用意が必要。

- cert-manager のデプロイ
  - 下の「[cert-manager のデプロイ](#cert-manager-のデプロイ)」を参照のこと
- Scale Controller のコンテナイメージ作成
  - 下の「[Scale Controller のビルド](#scale-controller-のビルド)」を参照のこと
- Scale Controller が Management Cluster にアクセスするための kubeconfig 作成
  - 「[認証情報の設定](#認証情報の設定)」に作成方法の例を示した

以上の準備ができたら、次のコマンドで Helm Chart をインストールする。
詳細は、[helm/README.md](helm/README.md) を参照のこと。

##### 作成済みの Helm Chart を使用した Scale Controller のデプロイ

次のコマンドで Helm Chart から Scale Controller をデプロイする。

```sh
# 注意：
#    kubeconfig の current-context が対象の Workload Cluster 向けであることを確認のこと。
#    そうでない場合は --kube-context 、 --kubeconfig オプションを追加するなど調整のこと。
cluster_context="--kube-context=kind-${CLUSTER_NAME}"
path_to_helm_chart=dist/chart/
controller_image_tag=${controller_img_tag}
receiver_image_tag=${receiver_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set controllerManager.container.image.repository=scale-controller \
    --set controllerManager.container.image.tag=${controller_image_tag} \
    --set alertReceiver.container.image.repository=card-alert-receiver \
    --set alertReceiver.container.image.tag=${receiver_image_tag} \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

##### Helm Chart の作成を伴う Scale Controller のデプロイ

次のコマンドで Helm Chart を作成する。

```sh
make helmchart
```

次のコマンドで Helm Chart から Scale Controller をデプロイする。

```sh
# 注意：
#    kubeconfig の current-context が対象の Workload Cluster 向けであることを確認のこと。
#    そうでない場合は --kube-context 、 --kubeconfig オプションを追加するなど調整のこと。
cluster_context="--kube-context=kind-${CLUSTER_NAME}"
path_to_helm_chart=dist/scale-controller-${helm_chart_version}.tgz
controller_image_tag=${controller_img_tag}
receiver_image_tag=${receiver_img_tag}
path_to_management_kubeconfig=management.kubeconfig
helm install scale-controller ${path_to_helm_chart} \
    ${cluster_context} \
    --create-namespace \
    --namespace scale-controller-system \
    --set controllerManager.container.image.repository=scale-controller \
    --set controllerManager.container.image.tag=${controller_image_tag} \
    --set alertReceiver.container.image.repository=card-alert-receiver \
    --set alertReceiver.container.image.tag=${receiver_image_tag} \
    --set-file managementKubeconfig=${path_to_management_kubeconfig}
```

##### Kustomize を使用した Scale Controller のデプロイ

Workload Cluster に Scale Controller をデプロイする。

```sh
export KUBECTL="kubectl --context=kind-${CLUSTER_NAME}"
make deploy-with-alert-receiver
```

事前に作成した kubeconfig を Workload Cluster の Namespace "scale-controller-system" 内に `Secret` として登録する。

```sh
kw create secret generic management-kubeconfig \
    --from-file=kubeconfig=management.kubeconfig \
    -n scale-controller-system
```

#### cert-manager のデプロイ

Webhook 用の証明書を発行する cert-manager をデプロイし、その `Pod` が `Running` になることを確認する。

```sh
kw apply -f https://github.com/cert-manager/cert-manager/releases/download/${cert_manager_version}/cert-manager.yaml
kw get pod -n cert-manager
# expected results:
# NAME                                      READY   STATUS    RESTARTS   AGE
# cert-manager-79559475b4-sv574             1/1     Running   0          19s
# cert-manager-cainjector-966fc8fbc-5nxh6   1/1     Running   0          19s
# cert-manager-webhook-854cf5f458-4h9ss     1/1     Running   0          19s
```

#### Scale Controller のビルド

コンテナイメージをビルドし、 Workload Cluster にロードする。

環境に応じて `sudo` を付与するなど適宜対応すること。
なお、 `sudo` コマンドに一般ユーザの環境変数を渡す場合、適宜 `-E` オプションを付与する必要がある点に注意。

```sh
export IMG=scale-controller:${controller_img_tag}
export ALERT_RECEIVER_IMG=card-alert-receiver:${receiver_img_tag}
make docker-build
docker images ${IMG}
# expected results:
# IMAGE                    ID             DISK USAGE   CONTENT SIZE   EXTRA
# scale-controller:0.3.0   a69fe6e286de       82.4MB             0B        
docker images ${ALERT_RECEIVER_IMG}
# expected results:
# IMAGE                       ID             DISK USAGE   CONTENT SIZE   EXTRA
# card-alert-receiver:0.3.0   7a46ede9b1d8       43.6MB             0B
```

```sh
kind load docker-image -n ${CLUSTER_NAME} ${IMG}
kind load docker-image -n ${CLUSTER_NAME} ${ALERT_RECEIVER_IMG}
```

**Note**
docker の代わりに buildah でイメージをビルドしたい場合は、 `docker-build` の代わりに、 `buildah-build` を用いること。

#### 認証情報の設定

Management Cluster に Scale Controller 用の認証情報を設定する。

```sh
kubectl apply -f example/rbac/rbac-in-management-cluster.yaml
```

作成した認証情報から Scale Controller が Management Cluster にアクセスするための kubeconfig を作成する。

```sh
export SERVER_IP_ADDRESS=$(docker inspect -f '{{.NetworkSettings.Networks.kind.IPAddress}}' ${M_CLUSTER_NAME}-control-plane)
export SERVER_PORT=6443
export CA_CRT=$(kubectl get secret scale-controller-tenant-secret -o jsonpath='{.data.ca\.crt}')
export TOKEN=$(kubectl get secret scale-controller-tenant-secret -o jsonpath='{.data.token}' | base64 -d)
envsubst < example/rbac/management-cluster-kubeconfig.yaml > management.kubeconfig
```


### 動作確認 (NodeScalingRequest Controller)

#### Node 追加

既存の `DCIMachineTemplate` を満たす `Node` 追加を要求する `NodeScalingRequest` を作成する。

```sh
kw apply -f example/nsr/dci/nsr-scaleup-existing-md.yaml
```

`nodescalingrequest.status.phase` が `Creating` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME                TYPE     PHASE      AGE
# scale-up-existing   create   Creating   5s
```

`machinedeployment.spec.replicas` が 2 に増加していることを確認する。

```sh
kubectl get md fake-worker-md-00
# expected results:
# NAME                CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# fake-worker-md-00   fake-cluster               2                   1                                        5m46s   
```

Cluster API の動作をシミュレーションするため、 `machinedeployment.status.readyReplicas` を 2 に設定し、ダミーの `Machine` および `Node` を手動で追加する。

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

`nodescalingrequest.status.phase` が `Created` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME                TYPE     PHASE     AGE
# scale-up-existing   create   Created   106s
```

`machinedeployment.spec.replicas` 及び `machinedeployment.status.readyReplicas` が 2 に増加したことを確認する。

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               2                   2                                        6m40s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   6m38s
```

`Node` が追加されたことを確認する。

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE     VERSION
# fake-cluster-control-plane   Ready      control-plane   6m      v1.34.2
# fake-node-00                 NotReady   <none>          5m53s   
# fake-node-01                 Unknown    <none>          22s     
# fake-node-control-plane-00   NotReady   <none>          5m53s   
```

`NodeScalingRequest` を削除する。

```sh
kw delete nsr scale-up-existing
```

#### Node 削除

削除する `Node` を指定した `NodeScalingRequest` を作成する。

```sh
export NODE_TO_DELETE="fake-node-01"
envsubst < example/nsr/dci/nsr-scaledown.yaml | kw apply -f -
```

`nodescalingrequest.status.phase` が `Deleting` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE      AGE
# scale-down   delete   Deleting   9s
```

`machinedeployment.spec.replicas` および `machineset.spec.replicas` が 1 に減少していることを確認する。

```sh
kubectl get md,ms
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   2                                        7m54s   

# NAME                                            CLUSTER        DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   AGE     VERSION
# machineset.cluster.x-k8s.io/fake-worker-ms-00   fake-cluster   1                                                    7m52s   v1.34.2
```

Cluster API の動作をシミュレーションするため、ダミーの `Node` を手動で削除し、`machinedeployment.status.readyReplicas` を 1 に変更する。

```sh
kw delete node fake-node-01
kubectl patch md ${MD_NAME} --type=merge --patch '{"status": {"readyReplicas": 1}}' --subresource status
```

`nodescalingrequest.status.phase` が `Deleted` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   61s
```

`machinedeployment.spec.replicas` および `machinedeployment.status.readyReplicas` が 1 に減少したことを確認する。

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        9m18s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   9m16s
```

`Node` が削除されたことを確認する。

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE     VERSION
# fake-cluster-control-plane   Ready      control-plane   7m3s    v1.34.2
# fake-node-00                 NotReady   <none>          6m56s   
# fake-node-control-plane-00   NotReady   <none>          6m56s   
```

`NodeScalingRequest` を削除する。

```sh
kw delete nsr scale-down
```

#### MachineDeployment の作成を伴う Node 追加

スペックを満たす `DCIMachineTemplate` が存在しない `Node` の追加を要求する `NodeScalingRequest` を作成する。

```sh
kw apply -f example/nsr/dci/nsr-scaleup-new-md.yaml
```

`nodescalingrequest.status.phase` が `Creating` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME              TYPE     PHASE      AGE
# scale-up-new-md   create   Creating   5s
```

`MachineDeployment` および `DCIMachineTemplate` が新たに作成され、新規に作成された `MachineDeployment` のレプリカ数が 1 であることを確認する。

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        10m   
# machinedeployment.cluster.x-k8s.io/worker-md-uw350     fake-cluster               1                                                            11s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   10m
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-mk55f           12s
```

新たに作成された `DCIMachineTemplate` の `.spec.template.spec` が、 `NodeScalingRequest` の `spec` と一致することを確認する。

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
export INFRA_TEMPLATE_NAME=worker-dmt-mk55f # 環境に合わせる
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

また、新たに作成された `MachineDeployment` の `.spec.template.spec.infrastructureRef` に、新たに作成された `DCIMachineTemplate` の名前が設定されていることを確認する。

```sh
export MD_NAME=worker-md-uw350　# 環境に合わせる
compare_string \
  $(kubectl get md ${MD_NAME} -o jsonpath='{.spec.template.spec.infrastructureRef.name}') \
  ${INFRA_TEMPLATE_NAME}
# expected results:
# same string
```

Cluster API の動作をシミュレーションするため、 `machinedeployment.status.readyReplicas` を 1 に設定し、ダミーの `Machine` および `Node` を手動で追加する。

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

`nodescalingrequest.status.phase` が `Created` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME              TYPE     PHASE     AGE
# scale-up-new-md   create   Created   5m20s
```

`machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` が 1 であることを確認する.

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE     VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        14m     
# machinedeployment.cluster.x-k8s.io/worker-md-uw350     fake-cluster               1                   1                                        3m49s   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   14m
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-mk55f           3m50s
```

`Node` が追加されたことを確認する。

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE     VERSION
# fake-cluster-control-plane   Ready      control-plane   9m45s   v1.34.2
# fake-node-00                 NotReady   <none>          9m38s   
# fake-node-02                 NotReady   <none>          66s     
# fake-node-control-plane-00   NotReady   <none>          9m38s   
```

`NodeScalingRequest` を削除する。

```sh
kw delete nsr scale-up-new-md
```

#### MachineDeployment の削除を伴う Node 削除

削除する `Node` を指定した `NodeScalingRequest` を作成する。

```sh
# set node created by 'Node Addition with MachineDeployment Creation'
export NODE_TO_DELETE="fake-node-02"
envsubst < example/nsr/dci/nsr-scaledown.yaml | kw apply -f -
```

`nodescalingrequest.status.phase` が `Deleting` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE      AGE
# scale-down   delete   Deleting   2s
```

`MachineDeployment` および `MachineSet` が削除されていることを確認する。

```sh
kubectl get md,ms,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        15m   

# NAME                                            CLUSTER        DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   AGE   VERSION
# machineset.cluster.x-k8s.io/fake-worker-ms-00   fake-cluster   1                                                    15m   v1.34.2

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   15m
```

Cluster API の動作をシミュレーションするため、ダミーの `Machine` および `Node` を手動で削除する。

```sh
kubectl delete machine fake-worker-machine-02
kw delete node fake-node-02
```

`nodescalingrequest.status.phase` が `Deleted` になることを確認する。

```sh
kw get nsr
# expected results:
# NAME         TYPE     PHASE     AGE
# scale-down   delete   Deleted   98s
```

`machinedeployment.spec.replicas` and `machinedeployment.status.readyReplicas` が削除されたことを確認する。

```sh
kubectl get md,dcimachinetemplate
# expected results:
# NAME                                                   CLUSTER        AVAILABLE   DESIRED   CURRENT   READY   AVAILABLE   UP-TO-DATE   PHASE   AGE   VERSION
# machinedeployment.cluster.x-k8s.io/fake-worker-md-00   fake-cluster               1                   1                                        16m   

# NAME                                                                          AGE
# dcimachinetemplate.infrastructure.cluster.x-k8s.io/worker-dmt-dc1-cots-g1n1   16m
```

`Node` が削除されたことを確認する。

```sh
kw get node
# expected results:
# NAME                         STATUS     ROLES           AGE   VERSION
# fake-cluster-control-plane   Ready      control-plane   10m   v1.34.2
# fake-node-00                 NotReady   <none>          10m   
# fake-node-control-plane-00   NotReady   <none>          10m   
```

`NodeScalingRequest` を削除する。

```sh
kw delete nsr scale-down
```

### 後片付け

Workload Cluster および Management Cluster を削除する。

```sh
kind delete cluster -n ${CLUSTER_NAME}
kind delete cluster -n ${M_CLUSTER_NAME}
```
