// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

type TBS struct {
	*GenericInstaller
}

func (k *TBS) GetEndpoints() (endpoints []string, err error) {
	return
}

func NewTBS(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &TBS{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-tbs",
		Image:               fmt.Sprintf("cluster-extension-tbs:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalEnvFromExtension,
	}, opts.Logger)}
}

func init() {
	Registry.Register(model.ClusterExtensionNameTBS, NewTBS)
}
