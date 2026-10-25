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

ROOTDIR=$(cd $(dirname $0); pwd -P)
DOWNLOAD_CHART_IMAGE=${DOWNLOAD_CHART_IMAGE:-true}
SKIP_EXISTING_IMAGE=${SKIP_EXISTING_IMAGE:-false}
CMD=$(basename $0)

usage=$(cat <<EOF
MAP chart download script for Bitnami

Usage:
    $CMD [flags]

Flags:
    -h  Print this message.
    -N  Do not download images for Helm chart.
    -I  Do not download existing images.
EOF
)

function parseArgs() {
    while getopts "hN" OPT; do
        case $OPT in
        h)
            echo "$usage"
            exit 0;;
        N)
            DOWNLOAD_CHART_IMAGE=false;;
        I)
            SKIP_EXISTING_IMAGE=true;;
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
    mkdir -p $ROOTDIR/tmp && \
    docker run -ti --rm -v $ROOTDIR:/download -v $ROOTDIR/tmp:/tmp \
        --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} chart download https://charts.bitnami.com/bitnami \
        -o /download/bitnami/charts

    if [ ! -d "$ROOTDIR/bitnami/charts" ]; then
        rm -rf $ROOTDIR/tmp
        return 0
    fi

    if [ q"$DOWNLOAD_CHART_IMAGE" == q"true" ]; then
        mkdir -p $ROOTDIR/bitnami/images
        for chart in "$ROOTDIR/bitnami/charts"/*.tgz ; do
            file=$(basename $chart)
            filename=${file%.*}
            
            if [ q"$SKIP_EXISTING_IMAGE" == q"true" ] && [ -f "$ROOTDIR/bitnami/images/${filename}.tar" ]; then
                continue
            fi

            docker run -ti --rm -v "$ROOTDIR/bitnami":/download -v $HOME:/root -v $ROOTDIR/tmp:/tmp \
                --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} image download -c /download/charts/$file \
                -o "/download/images/${filename}.tar"
        done
    fi

    rm -rf $ROOTDIR/tmp
}

main $@
