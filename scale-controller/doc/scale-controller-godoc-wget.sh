#! /bin/bash -eu
dir_repo_top=$(cd $(dirname $0)/..;pwd)
dir_document=${dir_repo_top}/doc
dir_document_html=${dir_document}/html
dir_document_godoc=${dir_document_html}/scale-controller-godoc
godoc_server_ipaddr=127.0.0.1
godoc_server_port=55555
godoc_server=${godoc_server_ipaddr}:${godoc_server_port}

if [[ $(printenv NO_PROXY | wc -l) == 0 ]];then
    export NO_PROXY=127.0.0.1
elif [[ $(printenv NO_PROXY | grep -e '127.0.0.1' | wc -l) == 0 ]];then
    export NO_PROXY=$NO_PROXY,127.0.0.1
fi
if [[ $(printenv no_proxy | wc -l) == 0 ]];then
    export no_proxy=127.0.0.1
elif [[ $(printenv no_proxy | grep -e '127.0.0.1' | wc -l) == 0 ]];then
    export no_proxy=$no_proxy,127.0.0.1
fi

pushd ${dir_document}

if [ -e pkg ];then
    rm pkg
fi
if [ -e ${dir_document_godoc} ];then
    rm -r ${dir_document_godoc}
fi
if [ ! -e ${dir_document_html} ];then
    echo "Create HTML Directory: ${dir_document_html}"
    mkdir -p ${dir_document_html}
fi

godoc -http=${godoc_server} >/dev/null 2>&1 &
pid_godoc_server=$!
echo "Wait for establishing godoc server 3 seconds"
sleep 3

wget -r -np "http://${godoc_server}/pkg/scale-controller/" 2> /dev/null
wget -r -np "http://${godoc_server}/pkg/scale-controller/internal/" 2> /dev/null
wget -r -np "http://${godoc_server}/pkg/scale-controller/internal/controller/" 2> /dev/null

kill ${pid_godoc_server}

mv "${godoc_server}" ${dir_document_godoc}
ln -sf ${dir_document_godoc}/pkg
echo "Create Symbolic link: ${dir_document}/pkg"

popd
