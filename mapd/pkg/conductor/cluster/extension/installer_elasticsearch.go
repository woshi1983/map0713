// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

type ElasticSearch struct {
	*GenericInstaller
}

func (k *ElasticSearch) GetEndpoints() (endpoints []string, err error) {
	return k.GenericInstaller.GetEndpoints("tcp", map[int32]string{
		80:   "http",
		443:  "https",
		9200: "http",
	})
}

func NewElasticSearch(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &ElasticSearch{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-elasticsearch",
		ServiceNamespace:    "vmware-system-elastic",
		Services:            []string{"elasticsearch-kibana", "elasticsearch-coordinating-only"},
		Image:               fmt.Sprintf("cluster-extension-elasticsearch:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalEnvFromExtension,
	}, opts.Logger)}
}

func init() {
	Registry.Register(model.ClusterExtensionNameElasticSearch, NewElasticSearch)
}
