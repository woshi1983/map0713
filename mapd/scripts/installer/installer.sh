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

CMD=$(basename $0)
DIR="$(cd "$(dirname "$0")" && pwd -P)"
source $DIR/common.sh
set -e

set +o noglob

usage=$(cat <<EOF
MAP installer utility

Please set hostname and other necessary attributes in config.yml first. DO NOT use localhost or 127.0.0.1 for hostname, because MAP needs to be accessed by external clients.

Usage:
    $CMD [flags]
    
Flags:
    --help  Print this message.
EOF
)
item=0

while [ $# -gt 0 ]; do
    case $1 in
        --help)
        note "$usage"
        exit 0;;
        *)
        note "$usage"
        exit 1;;
    esac
    shift || true
done

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $workdir

h2 "[Step $item]: checking if docker is installed ..."; item=$((item + 1))
check_docker

h2 "[Step $item]: checking docker-compose is installed ..."; item=$((item + 1))
check_dockercompose

map_tar=(map*.tar)
if [ "${#map_tar[@]}" -ne 1 ] || [ ! -f "${map_tar[0]}" ]; then
    error "Expected exactly one map*.tar core image archive in $workdir."
    exit 1
fi
h2 "[Step $item]: loading MAP images ..."; item=$((item + 1))
docker load -i "${map_tar[0]}"
if [ ! -s core-images.txt ]; then
    error "Missing core-images.txt in the offline installer."
    exit 1
fi
while IFS= read -r image; do
    [ -n "$image" ] || continue
    if ! docker image inspect "$image" >/dev/null 2>&1; then
        error "Core image was not loaded: $image"
        exit 1
    fi
done < core-images.txt
echo ""

h2 "[Step $item]: preparing environment ...";  item=$((item + 1))
if [ -n "$host" ]; then
    sed "s/^hostname: .*/hostname: $host/g" -i ./config.yml
fi

h2 "[Step $item]: preparing MAP configs ...";  item=$((item + 1))

./prepare.sh
docker-compose config --quiet
while IFS= read -r image; do
    [ -n "$image" ] || continue
    if ! docker image inspect "$image" >/dev/null 2>&1; then
        error "Compose references an image missing from the offline package: $image"
        exit 1
    fi
done < <(docker-compose config | awk '/^[[:space:]]*image:/ {print $2}')
echo ""

if [ -n "$(docker-compose ps -q)" ]; then
    note "stopping existing MAP instance ..."
    docker-compose down
fi
echo ""

h2 "[Step $item]: starting MAP ..."
docker-compose up -d --no-build

for attempt in $(seq 1 60); do
    all_healthy=true
    for container in map-log map-mariadb map-dex map-console map-mapd; do
        status=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container" 2>/dev/null || true)
        if [ "$status" != healthy ]; then
            all_healthy=false
            break
        fi
    done
    if [ "$all_healthy" = true ]; then
        success $"----MAP has been installed and started successfully.----"
        exit 0
    fi
    sleep 2
done

docker-compose ps
error "MAP did not become healthy within 120 seconds; inspect the container logs."
exit 1
