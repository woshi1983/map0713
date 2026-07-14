#!/usr/bin/env bash

# Copyright (c) 2023 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set +e
set -o noglob

ROOTDIR=$(cd $(dirname $0); pwd)
ENTRYPOINT=${ENTRYPOINT:-"${ROOTDIR}/../../packer/rocky/rocky-8.pkr.hcl"}
VALUES=${VALUES:-"${ROOTDIR}/../../packer/values.json"}
VERSION=$(cat "${ROOTDIR}/../../../VERSION")
CMD=$(basename $0)

usage=$(cat <<EOF
MAP OVA build script

Usage:
    $CMD [flags]

Flags:
    -I  Specify Packer script\'s entrypoint file.
    -f  Specify the values file for Packer script.
EOF
)

function parseArgs() {
    while getopts "hI:f:" OPT; do
        case $OPT in
        h)
            echo "$usage"
            exit 0;;
        I)
            ENTRYPOINT=$OPTARG;;
        f)
            VALUES=$OPTARG;;
        :)
            echo "Error: -${OPTARG} requires an argument."
            exit 1;;
        *)
            echo "$usage"
            exit 1;;
        esac
    done
}

function main() {
    parseArgs $@

    packer init -upgrade "$ENTRYPOINT" && \
        packer build -var="version=${VERSION}" -var="product_name=VMware Tanzu MAP (Flings Edition)" -var-file="$VALUES" -on-error=ask "$ENTRYPOINT"
}

main $@