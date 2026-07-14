// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
)

type Octant struct {
	*GenericInstaller
}

func (k *Octant) GetEndpoints() (endpoints []string, err error) {
	return k.GenericInstaller.GetEndpoints("http", nil)
}

func NewOctant(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &Octant{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-octant",
		ServiceNamespace:    "vmware-system-octant",
		Services:            []string{"octant"},
		Image:               fmt.Sprintf("cluster-extension-octant:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalOctantEnv,
	}, opts.Logger)}
}

func unmarshalOctantEnv(ext *model.ClusterExtension, env map[string]string) error {
	if err := unmarshalEnvFromExtension(ext, env); err != nil {
		return err
	}

	var prefixes []string
	sort.Slice(ext.Related, func(i, j int) bool { return ext.Related[i].Name < ext.Related[j].Name })
	for i, each := range ext.Related {
		c, err := cryptoutil.NewAESGCMCipher(each.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}

		prefix := "CLUSTER" + strconv.Itoa(i)
		env[prefix+"_CLUSTER_NAME"] = each.Name
		env[prefix+"_CLUSTER_APISERVER"] = each.Endpoint
		env[prefix+"_CLUSTER_TOKEN"] = string(c.Decrypt(each.AuthToken))
		prefixes = append(prefixes, prefix)
	}
	env["OCTANT_CONTEXT_PREFIX_MAPPING"] = strings.Join(prefixes, ",")
	return nil
}

func init() {
	Registry.Register(model.ClusterExtensionNameOctant, NewOctant)
}
