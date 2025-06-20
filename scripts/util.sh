#!/usr/bin/env bash

set -ex
set -u
set -o pipefail

# This script holds featuregate bash variables and utility functions.

# This function installs a Go tools by 'go get' command.
# Parameters:
#  - $1: package name, such as "sigs.k8s.io/controller-tools/cmd/controller-gen"
#  - $2: package version, such as "v0.4.1"
# Note:
#   Since 'go get' command will resolve and add dependencies to current module, that may update 'go.mod' and 'go.sum' file.
#   So we use a temporary directory to install the tools.
function util::install_tools() {
    local package="$1"
    local version="$2"

    temp_path=$(mktemp -d)
    pushd "${temp_path}" >/dev/null
    GO111MODULE=on go install "${package}"@"${version}"
    GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
    export PATH=$PATH:$GOPATH/bin
    popd >/dev/null
    rm -rf "${temp_path}"
}

function util::kubectl_with_retry() {
    local ret=0
    local count=0
    for i in {1..10}; do
        kubectl "$@"
        ret=$?
        if [[ ${ret} -ne 0 ]]; then
            echo "kubectl $@ failed, retrying(${i} times)"
            sleep 1
            continue
        else
            ((count++))
            # sometimes pod status is from running to error to running
            # so we need check it more times
            if [[ ${count} -ge 3 ]]; then
                return 0
            fi
            sleep 1
            continue
        fi
    done

    echo "kubectl $@ failed"
    kubectl "$@"
    return ${ret}
}

function util::wait_pod_ready() {
    local pod_label=$1
    local pod_namespace=$2
    local timeout=${3:-30m}
    local kubeconfig=$4
    local pod_label_key=${5:-app}
    local context=${6:-kind-gsc}

    echo "wait the $pod_label ready..."
    set +e
    util::kubectl_with_retry --kubeconfig "${kubeconfig}" --context "${context}" wait --for=condition=Ready --timeout="${timeout}" pods -l "${pod_label_key}"="${pod_label}" -n "${pod_namespace}"
    ret=$?
    set -e
    if [ $ret -ne 0 ]; then
        echo "kubectl describe info: $(kubectl --kubeconfig "${kubeconfig}" --context "${context}" describe pod -l "${pod_label_key}"="${pod_label}" -n "${pod_namespace}")"
    fi
    return ${ret}
}

function util::get_local_ip() {
    node_ip=$(ip -o route get to 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')
    echo "$node_ip"
    return $?
}

function util::get_container_port() {
    container_name=${1}
    target_port=${2}
    local str
    local IFS
    local result
    str=$(docker inspect --format='{{json .NetworkSettings.Ports}}' "${container_name}" | jq -r 'to_entries[] | "\(.key) -> \(.value[0].HostIp):\(.value[0].HostPort)"' | grep "${target_port}")
    IFS=':' read -r -a array <<<"$str"
    result=${array[-1]}
    echo "$result"
}
