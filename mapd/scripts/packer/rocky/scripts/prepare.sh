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

mkdir -p /installer
sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/selinux/config
sed -i 's/ rhgb quiet/ net.ifnames=0 biosdevname=0 rhgb quiet/' /etc/default/grub
if [ -f /boot/grub2/grub.cfg ]; then
    grub2-mkconfig -o /boot/grub2/grub.cfg
elif [ -f /boot/efi/EFI/rocky/grub.cfg ]; then
    grub2-mkconfig -o /boot/efi/EFI/rocky/grub.cfg
else
    echo "Error preparing OS: grub2 configuration file was not found!" 1>&2
    exit 1
fi
mv /etc/sysconfig/network-scripts/ifcfg-ens192 /etc/sysconfig/network-scripts/ifcfg-eth0
sed -i 's/^DEVICE=.*$/DEVICE=eth0/g' /etc/sysconfig/network-scripts/ifcfg-eth0
yum install -y chrony net-tools python3 && systemctl enable chronyd && systemctl start chronyd
timedatectl set-timezone UTC
yum update -y && reboot