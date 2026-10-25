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

set -e

MAP_PUBLIC_IMAGE_REGISTRY=${MAP_PUBLIC_IMAGE_REGISTRY:-"harbor.huaruicloud.cn"}
MAP_VERSION=${MAP_VERSION:-"v0.13.0"}

function version() {
    if [ -z "$(docker images ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} -q)" ]; then
        echo "MAP is not installed on this host." 1>&2
        exit 1
    fi
    docker run --rm ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} -v | awk '{print $3}'
}

ROOTDIR=$(cd $(dirname $0); pwd -P)
VERSION=$(version)
REGISTRY=${REGISTRY:-$(hostname)}
REGISTRY_USERNAME=${REGISTRY_USERNAME:-""}
REGISTRY_PASSWORD=${REGISTRY_PASSWORD:-""}
CMD=$(basename $0)

usage=$(cat <<EOF
MAP dependency upload script

Usage:
    $CMD [flags]

Flags:
    -h  Print this message.
    -v  Print software version.
    -r  Specify the Harbor 2.0 image registry for importing images (default: "${REGISTRY}").
    -u  Specify the login user of image registry (default: "${REGISTRY_USERNAME}").
    -p  Specify the login password of image registry (default: "${REGISTRY_PASSWORD}").
EOF
)

function parseArgs() {
    while getopts "hvr:u:p:" OPT; do
        case $OPT in
        h)
            echo "$usage"
            exit 0;;
        v)
            echo "$VERSION"
            exit 0;;
        r)
            REGISTRY=$OPTARG;;
        u)
            REGISTRY_USERNAME=$OPTARG;;
        p)
            REGISTRY_PASSWORD=$OPTARG;;
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

    rm -rf $ROOTDIR/tmp
    mkdir -p $ROOTDIR/tmp
    while read -r line; do
        docker run --rm -v $ROOTDIR:/upload -v $HOME:/root -v $ROOTDIR/tmp:/tmp --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} image upload \
            --registry-verify-certs=false --registry-username "${REGISTRY_USERNAME}" --registry-password "${REGISTRY_PASSWORD}" --registry-create-project \
            -i /upload/${line#"$ROOTDIR"} -o "${REGISTRY}"
    done < <(find $ROOTDIR -type f -name "dependencies*.tar")
    rm -rf $ROOTDIR/tmp
}

main $@
