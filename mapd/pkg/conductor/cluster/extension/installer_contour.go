// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

type Contour struct {
	*GenericInstaller
}

func (k *Contour) GetEndpoints() (endpoints []string, err error) {
	return k.GenericInstaller.GetEndpoints("tcp", map[int32]string{
		80:  "http",
		443: "https",
	})
}

func NewContour(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &Contour{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-contour",
		ServiceNamespace:    "tanzu-system-ingress",
		Services:            []string{"envoy"},
		Image:               fmt.Sprintf("cluster-extension-contour:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalEnvFromExtension,
	}, opts.Logger)}
}

func init() {
	Registry.Register(model.ClusterExtensionNameContour, NewContour)
}
