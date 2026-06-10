Scale Controller Document
========================================

## Overview
Scripts stored in this directory can be used to create documents for the scale-controller.

The contents of each script are as follows.

|Name|Description|
|-|-|
|scale-controller-godoc-expose.sh|Set up a godoc server<br>When finished, stop by `Ctrl+C`|
|scale-controller-godoc-wget.sh|(Deprecated) Start the godoc server and get the html file for the page about the scale-controller|
|scale-controller-goplantuml.sh|Run goplantuml to get the class diagram in puml format|

## How to create Documents
The procedure for creating a document using each script is described below.
This directory is assumed to be the execution directory.

### Creating a Plantuml File
1. Get class diagram file in puml format
    ```sh
    ./scale-controller-goplantuml.sh
    ```
2. Check the puml file stored in the `puml/`directory

### Starting the godoc Server
1. Start the godoc server
    - Listen port defaults to `0.0.0.0:9999`
    ```sh
    ./scale-controller-godoc-expose.sh
    ```
2. Access port 9999 of the active server from a web browser, and check the scale-controller documentation.
    - From the browser, there is no link to the file under `internal/controller/`.
      You can view documentation such as the controller's Reconcile() function by going directly to the following address:
      - http://Server-IP-Address:9999/pkg/scale-controller/internal/controller/
    - Note that private methods that are not exposed are not documented
    - When you finish checking, press `Ctrl+C` to stop the process.

### (Deprecated)Obtaining HTML Files(godoc)
1. Start the godoc server in the background, get the html file with wget, and terminate the godoc server.
    - Listen port defaults to `127.0.0.1:55555`
    ```sh
    ./scale-controller-godoc-wget.sh
    ```
2. Check that the `html/`directory and the symbolic link `pkg` to `html/pkg` are created.
    - Like [Starting the godoc Server](#starting-the-godoc-server), there is no link to the file under `internal/controller/`, but the html file is obtained and can be seen
