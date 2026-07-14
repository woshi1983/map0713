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

MAP_PUBLIC_IMAGE_REGISTRY=${MAP_PUBLIC_IMAGE_REGISTRY:-"registry-stg.eng.vmware.com"}
MAP_VERSION=${MAP_VERSION:-"latest"}

function version() {
    if [ -z "$(docker images ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} -q)" ]; then
        echo "MAP is not installed on this host." 1>&2
        exit 1
    fi
    docker run --rm ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} -v | awk '{print $3}'
}

ROOTDIR=$(cd $(dirname $0); pwd -P)
VERSION=$(version)
CMD=$(basename $0)
PT_PART=${PT_PART:-"0"}

usage=$(cat <<EOF
MAP dependency download script

Usage:
    $CMD [flags]

Flags:
    -h  Print this message.
    -v  Print software version.
    -1  Only download part 1.
    -2  Only download part 2.
EOF
)

function parseArgs() {
    while getopts "hv12" OPT; do
        case $OPT in
        h)
            echo "$usage"
            exit 0;;
        v)
            echo "$VERSION"
            exit 0;;
        1)
            PT_PART=1;;
        2)
            PT_PART=2;;
        *)
            echo "$usage"
            exit 1;;
        esac
    done
}

function downloadPart1() {
    rm -rf $ROOTDIR/tmp
    mkdir -p $ROOTDIR/tmp && \
    docker run -ti --rm -v $ROOTDIR:/download -v $HOME:/root -v $ROOTDIR/tmp:/tmp \
        --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} image download \
        --registry-verify-certs=false -I ${MAP_PUBLIC_IMAGE_REGISTRY}/map/dependency:${VERSION} \
        -o /download/dependencies-part1-${VERSION//./_}.tar
    rm -rf $ROOTDIR/tmp
}

function downloadPart2() {
    rm -rf $ROOTDIR/tmp
    mkdir -p $ROOTDIR/tmp && \
    docker run -ti --rm -v $ROOTDIR:/download -v $HOME:/root -v $ROOTDIR/tmp:/tmp \
        --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} image download \
        --registry-verify-certs=false -I ${MAP_PUBLIC_IMAGE_REGISTRY}/map/dependency-tbs:${VERSION} \
        -o /download/dependencies-part2-${VERSION//./_}.tar
    rm -rf $ROOTDIR/tmp
}

function main() {
    parseArgs $@

    if [ $PT_PART -eq 1 ]; then
        downloadPart1
    elif [ $PT_PART -eq 2 ]; then
        downloadPart2
    else
        downloadPart1 && downloadPart2
    fi
}

main $@