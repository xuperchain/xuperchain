#!/bin/bash

LOG_INFO(){
    content=${1}
    echo -e "\033[32m[INFO] ${content}\033[0m"
}

dirs=($(ls -l ${SHELL_FOLDER} | awk '/^d/ {print $NF}'))
for dir in ${dirs[*]}
do
    echo ${dir}/conf/server.yaml
    if [[ -f "${dir}/conf/server.yaml" && -f "${dir}/control.sh" ]];then
        LOG_INFO "try to start ${dir}"
        cd ${dir}
        bash control.sh start
        cd ..
    fi    
done
wait