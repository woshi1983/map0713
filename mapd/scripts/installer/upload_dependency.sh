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

MAP_VERSION=${MAP_VERSION:-"v0.13.0"}
MAP_PREPARE_IMAGE="map/map-prepare:${MAP_VERSION}"

function version() {
    if ! docker image inspect "$MAP_PREPARE_IMAGE" >/dev/null 2>&1; then
        echo "MAP is not installed on this host." 1>&2
        exit 1
    fi
    docker run --rm "$MAP_PREPARE_IMAGE" -v | awk '{print $3}'
}

ROOTDIR=$(cd $(dirname $0); pwd -P)
REGISTRY=${REGISTRY-}
REGISTRY_USERNAME=${REGISTRY_USERNAME:-""}
REGISTRY_PASSWORD=${REGISTRY_PASSWORD:-""}
INPUT_TAR=""
CMD=$(basename $0)

usage=$(cat <<EOF
MAP dependency upload script

Usage:
    $CMD [flags]

Flags:
    -h  Print this message.
    -v  Print software version.
    -r  Specify the internal image registry for importing images (required).
    -i  Use a dependency tar file outside this directory.
    -u  Specify the login user of image registry (default: "${REGISTRY_USERNAME}").
    -p  Specify the login password of image registry (defaults to REGISTRY_PASSWORD environment variable).
EOF
)

function parseArgs() {
    while getopts "hvr:u:p:i:" OPT; do
        case $OPT in
        h)
            echo "$usage"
            exit 0;;
        v)
            version
            exit 0;;
        r)
            REGISTRY=$OPTARG;;
        u)
            REGISTRY_USERNAME=$OPTARG;;
        p)
            REGISTRY_PASSWORD=$OPTARG;;
        i)
            INPUT_TAR=$OPTARG;;
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
    parseArgs "$@"

    if [ -z "$REGISTRY" ]; then
        echo "Specify the internal registry with -r; no public registry is used by default." >&2
        exit 1
    fi
    registry_host=${REGISTRY#https://}
    registry_host=${registry_host#http://}
    case "$registry_host" in
        harbor.huaruicloud.cn|harbor.huaruicloud.cn/*|harbor.huaruicloud.cn:*)
            echo "The public Harbor registry is not an allowed upload destination." >&2
            exit 1;;
    esac

    dependency_files=()
    if [ -n "$INPUT_TAR" ]; then
        if [ ! -f "$INPUT_TAR" ]; then
            echo "Dependency archive not found: $INPUT_TAR" >&2
            exit 1
        fi
        dependency_files+=("$(realpath "$INPUT_TAR")")
    else
        while IFS= read -r line; do
            dependency_files+=("$line")
        done < <(find "$ROOTDIR" -type f -name 'dependencies*.tar')
    fi
    if [ "${#dependency_files[@]}" -eq 0 ]; then
        echo "No dependency archive found; use -i to select the part1 tar file." >&2
        exit 1
    fi

    mkdir -p "$ROOTDIR/tmp"
    for archive in "${dependency_files[@]}"; do
        archive_dir=$(dirname "$archive")
        archive_name=$(basename "$archive")
        docker run --rm -v "$archive_dir:/upload" -v "$HOME:/root" -v "$ROOTDIR/tmp:/tmp" --privileged "$MAP_PREPARE_IMAGE" image upload \
            --registry-verify-certs=false --registry-username "${REGISTRY_USERNAME}" --registry-password "${REGISTRY_PASSWORD}" --registry-create-project \
            -i "/upload/$archive_name" -o "${REGISTRY}"
    done
    rm -rf "$ROOTDIR/tmp"
}

main "$@"
