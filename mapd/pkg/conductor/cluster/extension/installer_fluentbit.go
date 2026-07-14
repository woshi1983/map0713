// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

type FluentBit struct {
	*GenericInstaller
}

func (k *FluentBit) GetEndpoints() (endpoints []string, err error) {
	return
}

func NewFluentBit(opts *InstallerOptions) Installer {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &FluentBit{NewGenericInstaller(GenericInstallerConfig{
		Kubernetes:          opts.Kubernetes,
		Extension:           opts.Extension,
		ImageRegistry:       opts.ImageRegistry,
		LoadBalancerEnabled: opts.LoadBalancerEnabled,
		Cluster:             opts.ClusterName,
		ClusterClass:        opts.ClusterClass,
		JobNamespace:        opts.Namespace,
		Job:                 "cluster-extension-fluentbit",
		Image:               fmt.Sprintf("cluster-extension-fluentbit:%s", imageTag),
		CACerts:             opts.CACerts,
		UnmarshalEnvFunc:    unmarshalFluentBitEnv,
	}, opts.Logger)}
}

func unmarshalFluentBitEnv(ext *model.ClusterExtension, env map[string]string) error {
	if err := unmarshalEnvFromExtension(ext, env); err != nil {
		return err
	}

	var prefixes []string
	sort.Slice(ext.Related, func(i, j int) bool { return ext.Related[i].Name < ext.Related[j].Name })
	for i, rc := range ext.Related {
		prefix := "CLUSTER" + strconv.Itoa(i)
		for i := range rc.Extensions {
			each := &rc.Extensions[i]
			if each.Name == model.ClusterExtensionNameElasticSearch {
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
					if idx < 0 {
						continue
					}
					hostStr, portStr, err := net.SplitHostPort(u.Host)
					if err != nil {
						return errors.Wrapf(err, "error parsing hostport of %s", each.Name)
					}
					port, err := strconv.Atoi(portStr)
					if err != nil {
						return errors.Wrapf(err, "error parsing port number of endpoint of %s", each.Name)
					}
					if port == 9200 {
						env[prefix+"_CLUSTER_NAME"] = rc.Name
						env[prefix+"_ELASTICSEARCH_API_HOST"] = hostStr
						env[prefix+"_ELASTICSEARCH_API_PORT"] = strconv.Itoa(port)
						isTLS := false
						if u.Scheme == "https" {
							isTLS = true
						}
						env[prefix+"_ELASTICSEARCH_API_TLS"] = strconv.FormatBool(isTLS)
						prefixes = append(prefixes, prefix)
						break
					}
				}
				break
			}
		}
	}
	env["FLUENTBIT_CONTEXT_PREFIX_MAPPING"] = strings.Join(prefixes, ",")
	return nil
}

func init() {
	Registry.Register(model.ClusterExtensionNameFluentBit, NewFluentBit)
}
