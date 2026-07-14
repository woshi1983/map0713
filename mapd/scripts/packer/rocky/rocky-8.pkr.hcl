// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

variable "cpus" {
  type    = number
  default = 4
}

variable "datacenter" {
  type    = string
  default = "Datacenter"
}

variable "cluster" {
  type    = string
  default = ""
}

variable "datastore" {
  type    = string
  default = "Datastore"
}

variable "host" {
  type    = string
  default = ""
}

variable "resource_pool" {
  type    = string
  default = ""
}

variable "folder" {
  type    = string
  default = ""
}

variable "http_proxy" {
  type    = string
  default = ""
}

variable "https_proxy" {
  type    = string
  default = ""
}

variable "insecure_connection" {
  type    = bool
  default = true
}

variable "name" {
  type    = string
  default = "rocky-8-base"
}

variable "network" {
  type    = string
  default = "VM Network"
}

variable "no_proxy" {
  type    = bool
  default = false
}

variable "password" {
  type    = string
  default = "vmware"
}

variable "product_name" {
  type    = string
  default = "VMware Tanzu MAP"
}

variable "ram" {
  type    = number
  default = 4096
}

variable "ssh_password" {
  type    = string
  default = "changeme"
}

variable "ssh_port" {
  type    = number
  default = 22
}

variable "ssh_timeout" {
  type    = string
  default = "1h"
}

variable "ssh_username" {
  type    = string
  default = "root"
}

variable "template" {
  type    = string
  default = "photon-4_0-r1"
}

variable "username" {
  type    = string
  default = "administrator@vsphere.local"
}

variable "vcenter_server" {
  type    = string
  default = "vcenter.vsphere.local"
}

variable "docker_compose_download" {
  type    = string
}

variable "harbor_download" {
  type    = string
}

variable "cfssl_download" {
  type    = string
}

variable "cfssljson_download" {
  type    = string
}

variable "version" {
  type    = string
}

locals {
  temp_directory       = "${path.root}/../tmp"
  binary_directory     = "${path.root}/../common/sbin"
  systemd_directory    = "${path.root}/../common/systemd"
  build_directory      = "${path.root}/../../../dist"
  revision             = "${formatdate("YYYYMMDDhhmmss", timestamp())}"
}

source "vsphere-clone" "rocky-8-base" {
  boot_wait            = "5s"
  CPUs                 = var.cpus
  datacenter           = var.datacenter
  datastore            = var.datastore
  cluster              = var.cluster
  resource_pool        = var.resource_pool
  folder               = var.folder
  export {
    force              = true
    output_directory   = local.build_directory
  }
  host                 = var.host
  insecure_connection  = var.insecure_connection
  network              = var.network
  password             = var.password
  RAM                  = var.ram
  shutdown_command     = "shutdown -h now"
  ssh_pty              = true
  ssh_file_transfer_method = "sftp"
  ssh_password         = var.ssh_password
  ssh_port             = var.ssh_port
  ssh_timeout          = var.ssh_timeout
  ssh_username         = var.ssh_username
  template             = var.template
  username             = var.username
  vcenter_server       = var.vcenter_server
  vm_name              = "${replace(var.name, "-", "_")}-${replace(var.version, ".", "_")}-${local.revision}"
}

build {
  sources = ["source.vsphere-clone.rocky-8-base"]

  provisioner "shell-local" {
    inline             = [
      "rm -rf ${local.temp_directory}",
      "mkdir -p ${local.temp_directory}",
      "wget -O ${local.temp_directory}/harbor-offline-installer.tgz ${var.harbor_download}",
      "mkdir -p ${local.temp_directory}/bin",
      "wget -O ${local.temp_directory}/bin/docker-compose ${var.docker_compose_download}",
      "wget -O ${local.temp_directory}/bin/cfssl ${var.cfssl_download}",
      "wget -O ${local.temp_directory}/bin/cfssljson ${var.cfssljson_download}",
      "cp ${local.build_directory}/map-offline-installer-${replace(var.version, ".", "_")}.tgz ${local.temp_directory}/map-offline-installer.tgz",
      "mkdir -p ${local.binary_directory}",
      "mkdir -p ${local.systemd_directory}"
    ]
  }
  provisioner "shell" {
    environment_vars   = [
      "http_proxy=${var.http_proxy}",
      "https_proxy=${var.https_proxy}",
      "no_proxy=${var.no_proxy}"
    ]
    execute_command    = "{{ .Vars }} sh -eux '{{ .Path }}'"
    expect_disconnect  = true
    scripts            = [
      "${path.root}/scripts/prepare.sh"
    ]
  }

  provisioner "file" {
    source             = "${local.binary_directory}/"
    destination        = "/usr/sbin"
    generated          = true
  }

  provisioner "file" {
    source             = "${local.temp_directory}/bin/"
    destination        = "/usr/bin/"
    generated          = true
  }

  provisioner "file" {
    source             = "${local.systemd_directory}/"
    destination        = "/usr/lib/systemd/system/"
    generated          = true
  }

  provisioner "file" {
    source             = "${local.temp_directory}/harbor-offline-installer.tgz"
    destination        = "/installer/harbor-offline-installer.tgz"
    generated          = true
  }

  provisioner "file" {
    source             = "${local.temp_directory}/map-offline-installer.tgz"
    destination        = "/installer/map-offline-installer.tgz"
    generated          = true
  }

  provisioner "shell" {
    inline             = [
      "chmod +x /usr/bin/docker-compose",
      "chmod +x /usr/bin/cfssl",
      "chmod +x /usr/bin/cfssljson",
      "chmod +x /usr/sbin/vmbootstrap",
      "chmod +x /usr/sbin/vmsupport"
    ]
    execute_command    = "{{ .Vars }} sh -eux '{{ .Path }}'"
  }

  provisioner "shell" {
    environment_vars   = [
      "http_proxy=${var.http_proxy}",
      "https_proxy=${var.https_proxy}",
      "no_proxy=${var.no_proxy}",
      "MAPD_VERSION=${var.version}"
    ]
    execute_command    = "{{ .Vars }} sh -eux '{{ .Path }}'"
    scripts            = [
      "${path.root}/scripts/bootstrap.sh"
    ]
  }

  post-processor "shell-local" {
    environment_vars   = [
      "MAPD_VERSION=${var.version}",
      "OVA_PRODUCT_NAME=${var.product_name}"
    ]
    inline             = [
      "${path.root}/post_scripts/generate_ovf.sh '${replace(var.name, "-", "_")}-${replace(var.version, ".", "_")}-${local.revision}' ${local.build_directory}",
      "ovftool '${local.build_directory}/${replace(var.name, "-", "_")}-${replace(var.version, ".", "_")}-${local.revision}'.ovf '${local.build_directory}/${replace(var.name, "-", "_")}.ova'",
      "mv '${local.build_directory}/${replace(var.name, "-", "_")}.ova' '${local.build_directory}/${replace(var.name, "-", "_")}-${replace(var.version, ".", "_")}-${local.revision}'.ova"
    ]
    execute_command    = ["bash", "-c", "{{ .Vars }} sh -eux '{{ .Script }}'"]
  }
}
