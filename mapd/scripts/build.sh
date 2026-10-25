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
IMAGE_REGISTRY=${1:-harbor.huaruicloud.cn}
IMAGE_REPO=${IMAGE_REPO:-map}
VERSION=$(cat "${ROOTDIR}/../VERSION")
GOPROXY=${GOPROXY:-"https://proxy.golang.org,direct"}

pushd $ROOTDIR/..
    make clean && \
    docker build --build-arg GOPROXY=${GOPROXY} -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-base:${VERSION} -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-base:latest --target base . && \
    docker build --build-arg GOPROXY=${GOPROXY} -t ${IMAGE_REPO}/map-prepare:${VERSION} -t ${IMAGE_REPO}/map-prepare:latest -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-prepare:${VERSION} -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/map-prepare:latest --target prepare . && \
    docker build --build-arg GOPROXY=${GOPROXY} -t ${IMAGE_REPO}/mapd:${VERSION} -t ${IMAGE_REPO}/mapd:latest -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/mapd:${VERSION} -t ${IMAGE_REGISTRY}/${IMAGE_REPO}/mapd:latest .
popd