#! /bin/bash -eu
dir_repo_top=$(cd $(dirname $0)/..;pwd)
dir_document=${dir_repo_top}/doc
dir_document_puml=${dir_document}/puml
file_gopuml=${dir_document_puml}/scale-controller-class-diagram.puml

if [ ! -e ${dir_document_puml} ];then
    echo "Create PUML Directory: ${dir_document_puml}"
    mkdir -p ${dir_document_puml}
fi

goplantuml -recursive ${dir_repo_top} > ${file_gopuml}
