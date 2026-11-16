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
ROOTDIR=$(cd "$(dirname "$0")" && pwd)
VERSION=$(cat "${ROOTDIR}/../VERSION")
workdir="${ROOTDIR}/../dist"
mkdir -p "$workdir"

images=()
while IFS=, read -r type name; do
    [ "$type" = image ] || continue
    name="${name/\$\{VERSION\}/$VERSION}"
    case "$name" in
        harbor.huaruicloud.cn/*)
            echo "Public Harbor image is forbidden in the offline core package: $name" >&2
            exit 1;;
    esac
    docker image inspect "$name" >/dev/null || {
        echo "Missing local core image: $name" >&2
        exit 1
    }
    images+=("$name")
done < "${ROOTDIR}/core.csv"

if [ "${#images[@]}" -eq 0 ]; then
    echo "No core images listed in ${ROOTDIR}/core.csv" >&2
    exit 1
fi

package_dir="${workdir}/map-offline-installer"
rm -rf "$package_dir"
mkdir -p "$package_dir"
cp "${ROOTDIR}/installer/common.sh" \
   "${ROOTDIR}/installer/config.yml.tmpl" \
   "${ROOTDIR}/installer/installer.sh" \
   "${ROOTDIR}/installer/prepare.sh" \
   "${ROOTDIR}/installer/upload_dependency.sh" \
   "$package_dir/"
printf '%s\n' "${images[@]}" > "${package_dir}/core-images.txt"
docker save -o "${package_dir}/map-${VERSION//./_}.tar" "${images[@]}"
tar -czf "${workdir}/map-offline-installer-${VERSION//./_}.tgz" -C "$workdir" map-offline-installer
