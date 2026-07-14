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
set -o noglob

ROOTDIR=$(cd $(dirname $0); pwd)
IMAGE_REGISTRY=${1:-registry-stg.eng.vmware.com}
IMAGE_REPO=${IMAGE_REPO:-map}
VERSION=$(cat "${ROOTDIR}/../VERSION")


workdir=${ROOTDIR}/../dist
mkdir -p ${workdir}
pushd ${workdir}
    rm -rf map-offline-installer
    cp -r ${ROOTDIR}/installer map-offline-installer
    cmd="docker save -o map-offline-installer/map-${VERSION//./_}.tar"
    while read line; do
        evalline=$(eval "echo ${line}")
        if [ -n ${evalline} ]; then
            type=$(echo ${evalline} | awk -F ',' '{print $1}')
            name=$(echo ${evalline} | awk -F ',' '{print $2}')
            case $type in
            image)
                docker pull $name

                cmd="$cmd $name"
                if [ q"$(echo ${line} | awk -F ':' '{print $2}')" == q"\${VERSION}" ]; then
                    latest_name="${name%$VERSION}latest"
                    docker tag $name $latest_name

                    cmd="$cmd $latest_name"
                fi
                ;;
            bundle)
                echo "WARN: auto-skipped ${name} due to: bundle images shall not be included in core images!" 1>&2
                ;;
            esac
        fi
    done < ${ROOTDIR}/core.csv
    eval "$cmd" && \
    tar -czf map-offline-installer-${VERSION//./_}.tgz map-offline-installer

    rm -rf map-online-installer
    cp -r ${ROOTDIR}/installer map-online-installer && \
    tar -czf map-online-installer-${VERSION//./_}.tgz map-online-installer
popd