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

h2 "[Step $item]: checking if docker is installed ..."; let item+=1
check_docker

h2 "[Step $item]: checking docker-compose is installed ..."; let item+=1
check_dockercompose

if [ -f map*.tar ]; then
    h2 "[Step $item]: loading MAP images ..."; let item+=1
    docker load -i ./map*.tar
fi
echo ""

h2 "[Step $item]: preparing environment ...";  let item+=1
if [ -n "$host" ]; then
    sed "s/^hostname: .*/hostname: $host/g" -i ./config.yml
fi

h2 "[Step $item]: preparing MAP configs ...";  let item+=1

./prepare.sh
echo ""

if [ -n "$(docker-compose ps -q)" ]; then
    note "stopping existing MAP instance ..."
    docker-compose down -v
fi
echo ""

h2 "[Step $item]: starting MAP ..."
docker-compose up -d

success $"----MAP has been installed and started successfully.----"