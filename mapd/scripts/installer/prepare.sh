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

# If compiling source code this dir is MAP's make dir.
# If installing MAP via package, this dir is MAP's root dir.
if [[ -n "$MAP_BUNDLE_DIR" ]]; then
    map_prepare_path=$MAP_BUNDLE_DIR
else
    map_prepare_path="$( cd "$(dirname "$0")" ; pwd -P )"
fi
echo "prepare base dir is set to ${map_prepare_path}"

if [ ! -f $map_prepare_path/config.yml ]; then
    echo "no config file: $map_prepare_path/config.yml." 1>&2
    exit 1
fi

# Run prepare script
docker run --rm -v $map_prepare_path:/config \
    --privileged \
    ${MAP_PUBLIC_IMAGE_REGISTRY}/map/map-prepare:${MAP_VERSION} config setup -f /config/config.yml -o /config $@
