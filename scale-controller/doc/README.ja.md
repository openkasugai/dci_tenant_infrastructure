Scale Controller Document
========================================

## 概要
本ディレクトリに格納されているスクリプトを用いることで、scale-controllerのドキュメントを作成することができる。

各ファイルの内容は以下のとおりである。

|Name|Description|
|-|-|
|scale-controller-godoc-expose.sh|godocのサーバを立ち上げる<br>終了時は`Ctrl+C`にて停止のこと|
|scale-controller-godoc-wget.sh|(非推奨)godocのサーバを立ち上げ、scale-controllerに関するページのhtmlファイルを取得する|
|scale-controller-goplantuml.sh|goplantumlを実行し、クラス図をpuml形式で取得する|

## ドキュメント作成手順
以下に各スクリプトを用いて、ドキュメントを作成する際の手順について記載する。
実行ディレクトリは本ディレクトリを想定している。

### Plantumlファイルの作成
1. puml形式のクラス図ファイルを取得する
    ```sh
    ./scale-controller-goplantuml.sh
    ```
2. `puml/`ディレクトリに格納されているpumlファイルを確認する

### godocサーバの立ち上げ
1. godocサーバを立ち上げる
    - 待ち受けポートはデフォルトで、`0.0.0.0:9999`としている
    ```sh
    ./scale-controller-godoc-expose.sh
    ```
2. Webブラウザなどから実行サーバの9999番ポートへアクセスし、scale-controllerのドキュメントを確認する
    - ブラウザから確認しても`internal/controller/`配下のファイルへのリンクは存在していないが、
      以下のアドレスへ直接アクセスすることで、コントローラのReconcile関数などのドキュメントを確認できる
      - http://Server-IP-Address:9999/pkg/scale-controller/internal/controller/
    - 外部公開していないプライベートメソッドについてはドキュメント化されない点に注意
    - 確認が終了したら、`Ctrl+C`でプロセスを停止する

### HTMLファイル(godoc)の取得(非推奨)
1. godocサーバをバックグラウンドで立ち上げ、wgetでhtmlファイルを取得し、godocサーバを終了させる
    - 待ち受けポートはデフォルトで、`127.0.0.1:55555`としている
    ```sh
    ./scale-controller-godoc-wget.sh
    ```
2. `html/`ディレクトリおよび`html/pkg`へのシンボリックリンク`pkg`が作成されるので確認する
    - [godocサーバの立ち上げ](#godocサーバの立ち上げ)同様、`internal/controller/`配下のファイルへのリンクは存在していないが、
      htmlファイル自体は取得しているため確認可能である
