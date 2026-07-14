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

yum install -y cloud-init
cat > /etc/cloud/cloud.cfg <<EOF
users:
 - default

disable_root: 1
ssh_pwauth:   1

mount_default_fields: [~, ~, 'auto', 'defaults,nofail,x-systemd.requires=cloud-init.service', '0', '2']
resize_rootfs_tmp: /dev
ssh_deletekeys:   1
ssh_genkeytypes:  ['rsa', 'ecdsa', 'ed25519']
syslog_fix_perms: ~
disable_vmware_customization: false

cloud_init_modules:
 - disk_setup
 - migrator
 - bootcmd
 - write-files
 - growpart
 - resizefs
 - set_hostname
 - update_hostname
 - update_etc_hosts
 - rsyslog
 - users-groups
 - ssh

cloud_config_modules:
 - mounts
 - locale
 - set-passwords
 - rh_subscription
 - yum-add-repo
 - timezone
 - disable-ec2-metadata
 - runcmd

cloud_final_modules:
 - package-update-upgrade-install
 - puppet
 - chef
 - salt-minion
 - mcollective
 - rightscale_userdata
 - scripts-per-once
 - scripts-per-boot
 - scripts-per-instance
 - scripts-user
 - ssh-authkey-fingerprints
 - keys-to-console
 - phone-home
 - final-message
 - power-state-change

system_info:
  default_user:
    name: cloud-user
    lock_passwd: true
    gecos: Cloud User
    groups: [adm, systemd-journal]
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    shell: /bin/bash
  distro: rhel
  paths:
    cloud_dir: /var/lib/cloud
    templates_dir: /etc/cloud/templates
  ssh_svcname: sshd

# vim:syntax=yaml
EOF
systemctl enable cloud-init-local.service
systemctl enable cloud-init.service
systemctl enable cloud-config.service
systemctl enable cloud-final.service
firewall-cmd --permanent --new-service=vmware-system-map
firewall-cmd --permanent --service=vmware-system-map --add-port=8080/tcp --add-port=8443/tcp
firewall-cmd --permanent --add-service=vmware-system-map
firewall-cmd --permanent --new-service=vmware-harbor
firewall-cmd --permanent --service=vmware-harbor --add-port=80/tcp --add-port=443/tcp
firewall-cmd --permanent --add-service=vmware-harbor
firewall-cmd --reload
systemctl enable vmbootstrap-base.service
systemctl enable vmbootstrap-network.service
systemctl enable vmbootstrap-harbor.service
systemctl enable vmbootstrap-map.service
yum install -y yum-utils
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum install -y docker-ce docker-ce-cli containerd.io && systemctl enable docker
yum clean all
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "64m",
    "max-file": "2"
  }
}
EOF
new_passwd=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 64 ; echo '')
echo $new_passwd | passwd --stdin root
echo > /var/log/secure
echo > /var/log/dnf.log
echo > /var/log/dnf.librepo.log
echo > /var/log/dnf.rpm.log
echo > /var/log/lastlog
echo > /var/log/messages
echo > /root/.bash_history
rm -f /var/log/vmware-*.log
history -c