#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
GOLANGCI_LINT_PKG="github.com/golangci/golangci-lint/cmd/golangci-lint"
GOLANGCI_LINT_VER="v1.59.1"

cd "${REPO_ROOT}"
source "scripts/util.sh"

command golangci-lint &>/dev/null || util::install_tools ${GOLANGCI_LINT_PKG} ${GOLANGCI_LINT_VER}

golangci-lint --version

export GOFLAGS=-mod=vendor
golangci-lint cache clean

if (golangci-lint run -v); then
    echo 'Congratulations!  All Go source files have passed staticcheck.'
else
    echo # print one empty line, separate from warning messages.
    echo 'Please review the above warnings.'
    echo 'If the above warnings do not make sense, feel free to file an issue.'
    exit 1
fi
