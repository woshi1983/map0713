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
		StorageClass:        opts.StorageClass,
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
	related := append([]model.Cluster(nil), ext.Related...)
	sort.Slice(related, func(i, j int) bool { return related[i].Name < related[j].Name })
	for _, rc := range related {
		prefix := "CLUSTER" + strconv.Itoa(len(prefixes))
		for i := range rc.Extensions {
			each := &rc.Extensions[i]
			if each.Name == model.ClusterExtensionNamePrometheus {
				if err := requireRelatedExtensionReady(rc.Name, "Prometheus", each); err != nil {
					return err
				}
				endpoints := []string{}
				if err := json.Unmarshal(each.Endpoints, &endpoints); err != nil {
					return &ErrDependency{Reason: fmt.Sprintf("invalid Prometheus endpoints on cluster %s", rc.Name)}
				}
				for _, endpoint := range endpoints {
					u, err := url.Parse(endpoint)
					if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
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
	if len(prefixes) == 0 {
		return &ErrDependency{Reason: "Grafana requires a related Prometheus cluster with a valid HTTP endpoint"}
	}
	env["GRAFANA_CONTEXT_PREFIX_MAPPING"] = strings.Join(prefixes, ",")
	return nil
}

func init() {
	Registry.Register(model.ClusterExtensionNameGrafana, NewGrafana)
}
