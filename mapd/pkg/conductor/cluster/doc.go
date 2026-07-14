// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

const (
	defNamespace                       = "vmware-system-map"
	defRegistrySecret                  = "vmware-system-map-registry"
	defKubeConfigSecret                = "vmware-system-map-kubeconfig"
	defServiceAccount                  = "vmware-system-map"
	defClusterAdminClusterRoleBinding  = "vmware-system-map-cluster-admin"
	defPspPrivilegedClusterRoleBinding = "vmware-system-map-psp-priviledged"
	defSSHPrivateKeySecret             = "vmware-system-map-ssh"

	defTkgServiceConfiguration = "tkg-service-configuration"

	defLabelManaged  = "map.tanzu.vmware.com/managed"
	defLabelInternal = "map.tanzu.vmware.com/internal"
	defLabelIssuer   = "map.tanzu.vmware.com/issuer"
)

type ErrCriticalFailure struct {
	Err error
}

func (e *ErrCriticalFailure) Error() string {
	return e.Err.Error()
}

func IsCriticalError(err error) bool {
	_, ok := err.(*ErrCriticalFailure)
	return ok
}
