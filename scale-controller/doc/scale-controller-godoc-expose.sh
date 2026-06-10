#! /bin/bash -eu
godoc_server_ipaddr=0.0.0.0
godoc_server_port=9999
godoc_server=${godoc_server_ipaddr}:${godoc_server_port}

pushd $(dirname $0)

godoc -http=${godoc_server}

popd
