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

ROOTDIR=$(cd $(dirname $0); pwd -P)
REGISTRY=${REGISTRY:-$(hostname)}
REGISTRY_USERNAME=${REGISTRY_USERNAME:-""}
REGISTRY_PASSWORD=${REGISTRY_PASSWORD:-""}
CMD=$(basename $0)

usage=$(cat <<EOF
MAP chart upload script for Bitnami

Usage:
    $CMD [flags]

Flags:
    -h  Print this message.
    -r  Specify the Harbor 2.0 image registry for importing charts (default: "${REGISTRY}").
    -u  Specify the login user of image registry (default: "${REGISTRY_USERNAME}").
    -p  Specify the login password of image registry (default: "${REGISTRY_PASSWORD}").
EOF
)

function parseArgs() {
    while getopts "hr:u:p:" OPT; do
        case $OPT in
        h)
            echo "$usage"
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

    if [ ! -d "$ROOTDIR/bitnami/charts" ]; then
        echo "Error: directory $ROOTDIR/bitnami/charts does not exist yet. Please run \"download_bitnami_chart.sh\" first." 1>&2
        return 1
    fi
    if [ -z "$REGISTRY_USERNAME" ]; then
        echo "Error: a registry user is required." 1>&2
        return 1
    fi
    if [ -z "$REGISTRY_PASSWORD" ]; then
        echo "Error: a registry password is required." 1>&2
        return 1
    fi

    rm -rf $ROOTDIR/tmp
    mkdir -p $ROOTDIR/tmp

    for image in "$ROOTDIR/bitnami/images"/*.tar ; do
        file=$(basename $image)
        filename=${file%.*}

        docker run -ti --rm -v $ROOTDIR/bitnami/images:/upload -v $HOME:/root -v $ROOTDIR/tmp:/tmp \
            --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} image upload \
            --registry-verify-certs=false --registry-username "${REGISTRY_USERNAME}" \
            --registry-password "${REGISTRY_PASSWORD}" --registry-create-project \
            -i /upload/${filename}.tar -o "${REGISTRY}"
    done

    docker run -ti --rm -v $ROOTDIR/bitnami/charts:/upload -v $ROOTDIR/tmp:/tmp \
        --privileged ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} chart upload \
        -i /upload --username "$REGISTRY_USERNAME" --password "$REGISTRY_PASSWORD" \
        --insecure-skip-tls-verify https://${REGISTRY#"https://"}/chartrepo/bitnami

    rm -rf $ROOTDIR/tmp
}

main $@