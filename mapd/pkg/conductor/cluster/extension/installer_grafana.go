// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

type Grafana struct {
	*GenericInstaller
}

func (k *Grafana) GetEndpoints() (endpoints []string, err error) {
	return k.GenericInstaller.GetEndpoints("tcp", map[int32]string{
		80:  "http",
		443: "https",
	})
}

func NewGrafana(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &Grafana{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-grafana",
		ServiceNamespace:    "tanzu-system-dashboards",
		Services:            []string{"grafana"},
		Image:               fmt.Sprintf("cluster-extension-grafana:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalGrafanaEnv,
	}, opts.Logger)}
}

func unmarshalGrafanaEnv(ext *model.ClusterExtension, env map[string]string) error {
	if err := unmarshalEnvFromExtension(ext, env); err != nil {
		return err
	}

	var prefixes []string
	sort.Slice(ext.Related, func(i, j int) bool { return ext.Related[i].Name < ext.Related[j].Name })
	for i, rc := range ext.Related {
		prefix := "CLUSTER" + strconv.Itoa(i)
		for i := range rc.Extensions {
			each := &rc.Extensions[i]
			if each.Name == model.ClusterExtensionNamePrometheus {
				endpoints := []string{}
				if err := json.Unmarshal(each.Endpoints, &endpoints); err != nil {
					return err
				}
				for _, endpoint := range endpoints {
					u, err := url.Parse(endpoint)
					if err != nil {
						return errors.Wrapf(err, "error parsing endpoint of %s", each.Name)
					}
					idx := strings.Index(u.Host, ":")
					if idx > 0 {
						continue
					}

					env[prefix+"_CLUSTER_NAME"] = rc.Name
					env[prefix+"_PROMETHEUS_APISERVER"] = endpoint
					prefixes = append(prefixes, prefix)
					break
				}
				break
			}
		}
	}
	env["GRAFANA_CONTEXT_PREFIX_MAPPING"] = strings.Join(prefixes, ",")
	return nil
}

func init() {
	Registry.Register(model.ClusterExtensionNameGrafana, NewGrafana)
}
