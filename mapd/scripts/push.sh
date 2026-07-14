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
    docker push ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-prepare:${VERSION} && \
    docker push ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-prepare:latest && \
    docker push ${IMAGE_REGISTRY}/${IMAGE_REPO}/mapd:${VERSION} && \
    docker push ${IMAGE_REGISTRY}/${IMAGE_REPO}/mapd:latest
    
    cmd="docker run --rm -v $HOME:/root --privileged ${IMAGE_REGISTRY}/map/map-prepare:latest image index --registry-verify-certs=false -o ${IMAGE_REGISTRY}/${IMAGE_REPO}/dependency:${VERSION}"
    while read line; do
        line=$(eval "echo ${line}")
        if [ -n ${line} ]; then
            type=$(echo ${line} | awk -F ',' '{print $1}')
            name=$(echo ${line} | awk -F ',' '{print $2}')
            case $type in
            image)
                cmd="${cmd} -i ${name}";;
            bundle)
                cmd="${cmd} -b ${name}";;
            esac
        fi
    done < ${ROOTDIR}/dependencies_part1.csv
    eval $cmd || echo "You can run the following command manually: $cmd"

    cmd="docker run --rm -v $HOME:/root --privileged ${IMAGE_REGISTRY}/map/map-prepare:latest image index --registry-verify-certs=false -o ${IMAGE_REGISTRY}/${IMAGE_REPO}/dependency-tbs:${VERSION}"
    while read line; do
        line=$(eval "echo ${line}")
        if [ -n ${line} ]; then
            type=$(echo ${line} | awk -F ',' '{print $1}')
            name=$(echo ${line} | awk -F ',' '{print $2}')
            case $type in
            image)
                cmd="${cmd} -i ${name}";;
            bundle)
                cmd="${cmd} -b ${name}";;
            esac
        fi
    done < ${ROOTDIR}/dependencies_part2.csv
    eval $cmd || echo "You can run the following command manually: $cmd"

    cmd="docker run --rm -v $HOME:/root --privileged ${IMAGE_REGISTRY}/map/map-prepare:latest image index --registry-verify-certs=false -o ${IMAGE_REGISTRY}/${IMAGE_REPO}/flings/dependency:${VERSION}"
    while read line; do
        line=$(eval "echo ${line}")
        if [ -n ${line} ]; then
            type=$(echo ${line} | awk -F ',' '{print $1}')
            name=$(echo ${line} | awk -F ',' '{print $2}')
            case $type in
            image)
                cmd="${cmd} -i ${name}";;
            bundle)
                cmd="${cmd} -b ${name}";;
            esac
        fi
    done < ${ROOTDIR}/flings/dependencies.csv
    eval $cmd || echo "You can run the following command manually: $cmd"
popd