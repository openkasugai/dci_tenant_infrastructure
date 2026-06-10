#! /bin/bash -eu

# print usage
usage() {
    echo "usage: $0 OPTIONS"
    echo "OPTIONS:"
    echo " -h,--help             : print usage"
    echo " -c,--create-cluster   : create clusters for test"
    echo "                       : this option requires options: -m,-w"
    echo " -d,--delete-cluster   : delete clusters for test"
    echo "                       : this option requires options: -m,-w"
    echo " -x,--execute-test     : execute tests"
    echo "                       : this option requires options: -m,-w"
    echo " -m,--management NAME  : name for management cluster"
    echo " -w,--workload NAME    : name for workload cluster"
    echo " -p,--provider NAME    : specify supporting provider"
    echo "                       : available providers(DCI,Docker)"
    echo "                       : alert receiver will NOT be launched when provider is Docker"
    echo " -i,--installer TYPE   : specify install method for scale-controller"
    echo "                       : available methods(kustomize,helm)"
    echo " -j,--config-json PATH : specify path to config.json"
    echo "                       : default is config.json"
    echo "config.json:"
    echo " You can set parameters used in this script by config.json file too."
    echo " The command line options overwrite the settings in config.json if you specify them."
    echo " For more details about config.json's parameters, see 'shrc/parameters.shrc'."
}

# main process
main() {
    echo "################################################################################"
    echo "Starting scale-controller e2e test script..."
    if [[ $# == 0 ]];then
        usage $0
        exit 0
    fi

    setup

    local _cluster_name_management=""
    local _cluster_name_workload=""

    local _provider_name="DCI"
    local _installer_type="helm"
    local _is_enabled_debug="false"
    local _is_enabled_create_cluster=0
    local _is_enabled_delete_cluster=0
    local _is_enabled_execute_test=0
    local _config_json_path="${dir_execute}/config.json"

    # parse arguments
    while (( $# > 0 ))
    do
        case $1 in
            -h | --help)
                usage $0
                exit 0
                ;;
            -c | --create-cluster)
                _is_enabled_create_cluster=1
                ;;
            -d | --delete-cluster)
                _is_enabled_delete_cluster=1
                ;;
            -m | -m=* | --management | --management=*)
                if [[ "$1" =~ ^--management= ]]; then
                    _cluster_name_management=${1#--management=}
                elif [[ "$1" =~ ^-m= ]]; then
                    _cluster_name_management=${1#-m=}
                elif [[ $# == 1 ]] || [[ "$2" =~ ^-+ ]]; then
                    echo "'management' requires an argument." 1>&2
                    usage $0
                    exit 1
                else
                    _cluster_name_management="$2"
                    shift
                fi
                ;;
            -w | -w=* | --workload | --workload=*)
                if [[ "$1" =~ ^--workload= ]]; then
                    _cluster_name_workload=${1#--workload=}
                elif [[ "$1" =~ ^-w= ]]; then
                    _cluster_name_workload=${1#-w=}
                elif [[ $# == 1 ]] || [[ "$2" =~ ^-+ ]]; then
                    echo "'workload' requires an argument." 1>&2
                    usage $0
                    exit 1
                else
                    _cluster_name_workload="$2"
                    shift
                fi
                ;;
            -x | --execute-test)
                _is_enabled_execute_test=1
                ;;
            -p | -p=* | --provider | --provider=*)
                if [[ "$1" =~ ^--provider= ]]; then
                    _provider_name=${1#--provider=}
                elif [[ "$1" =~ ^-p= ]]; then
                    _provider_name=${1#-w=}
                elif [[ $# == 1 ]] || [[ "$2" =~ ^-+ ]]; then
                    traced_echo "'provider' requires an argument." 1>&2
                    exit 1
                else
                    _provider_name="$2"
                    shift
                fi
                ;;
            -i | -i=* | --installer | --installer=*)
                if [[ "$1" =~ ^--installer= ]]; then
                    _installer_type=${1#--installer=}
                elif [[ "$1" =~ ^-i= ]]; then
                    _installer_type=${1#-w=}
                elif [[ $# == 1 ]] || [[ "$2" =~ ^-+ ]]; then
                    traced_echo "'installer' requires an argument." 1>&2
                    exit 1
                else
                    _installer_type="$2"
                    shift
                fi
                ;;
            -j | -j=* | --config-json | --config-json=*)
                if [[ "$1" =~ ^--config-json= ]]; then
                    _config_json_path=${1#--config-json=}
                elif [[ "$1" =~ ^-j= ]]; then
                    _config_json_path=${1#-j=}
                elif [[ $# == 1 ]] || [[ "$2" =~ ^-+ ]]; then
                    traced_echo "'config-json' requires an argument." 1>&2
                    exit 1
                else
                    _config_json_path="$2"
                    shift
                fi
                ;;
            --enable-debug)
                _is_enabled_debug="true"
                ;;
            -*)
                echo "invalid option : $1"
                usage $0
                exit 1
                ;;
            *)
                echo "unexpected parameter : $1"
                usage $0
                exit 1
                ;;
        esac
        shift
    done

    # check config.json
    check_config_json "${_config_json_path}"

    echo "Checking parameters started..."
    if [[ ${_cluster_name_management} == "" ]] || [[ ${_cluster_name_workload} == "" ]];then
        echo "** Invalid parameters:"
        echo "   - cluster_name_management : ${_cluster_name_management}"
        echo "   - cluster_name_workload : ${_cluster_name_workload}"
        echo "** Please use -m and -w option"
        echo ""
        usage $0
        exit 1
    fi
    set_cluster_name_m "${_cluster_name_management}"
    set_cluster_name_w "${_cluster_name_workload}"
    echo "Checking parameters finished successfully."

    # create or delete clusters
    if [[ ${_is_enabled_create_cluster} == 1 ]];then
        echo "Creating clusters started..."
        setup_clusters "${_provider_name}" "${_installer_type}" "${_is_enabled_debug}"
        echo "Creating clusters finished successfully."
    fi

    if [[ ${_is_enabled_execute_test} == 1 ]];then
        echo "Executing tests started..."
        execute_test "${_provider_name}" "${dir_logs}"
        echo "Executing tests finished successfully."
    fi

    if [[ ${_is_enabled_delete_cluster} == 1 ]];then
        echo "Deleting clusters started..."
        cleanup_all "${_provider_name}" "${_installer_type}" "${_is_enabled_debug}"
        echo "Deleting clusters finished successfully."
    fi

    collect_logs_after_all "${dir_logs}"

    echo "Scale-controller e2e test script finished successfully."
    echo "################################################################################"
}

setup() {
    echo "Setting up environment for scale-controller e2e test script started..."
    # setup
    local _dir_script=$(cd $(dirname $0);pwd)
    local _dir_repo_top=$(cd ${_dir_script}/../..;pwd)
    local _dir_execute="$(pwd)"

    # Source shrc files
    for f in $(ls ${_dir_script}/shrc/*.shrc); do
        echo "Sourcing ${f}"
        . ${f}
    done

    # Set global variables based on paths for this script and execution
    set_dir_script "${_dir_script}"
    set_dir_repo_top "${_dir_repo_top}"
    set_dir_execute "${_dir_execute}"

    echo "Setting up environment for scale-controller e2e test script finished successfully."
}

##########################
# Start execution
##########################
time_execute="$(TZ='Asia/Tokyo' date +%F-%H-%M-%S)"
dir_logs=logs/${time_execute}

# Generate logfile path
logfile_execute="$(basename ${0%.*})-${time_execute}.log"

mkdir -p ${dir_logs}
main $* | tee ${logfile_execute}
mv ${logfile_execute} ${dir_logs}
