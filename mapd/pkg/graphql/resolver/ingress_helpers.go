// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"net"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	networkingv1 "k8s.io/api/networking/v1"
)

func getIngresses(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]networkingv1.Ingress, error) {
	list, err := client.GetIngresses(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertIngress(ingress *model.Ingress, obj *networkingv1.Ingress) {
	ingress.Name = obj.ObjectMeta.Name
	ingress.Namespace = obj.ObjectMeta.Namespace
	if obj.ObjectMeta.Annotations != nil {
		ingress.Annotations = obj.ObjectMeta.Annotations
	}
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				ingress.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				ingress.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				ingress.Instance = &model.Instance{ID: id}
			}
		}
	}
	ingress.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		ingress.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		ingress.CreatedAt = time.Now()
	}

	ingress.Spec = &model.IngressSpec{IngressClass: obj.Spec.IngressClassName}
	for _, each := range obj.Spec.TLS {
		tlsCfg := &model.IngressTLS{Hosts: each.Hosts}
		if each.SecretName != "" {
			tlsCfg.Secret = util.StringPtr(each.SecretName)
		}
		ingress.Spec.TLS = append(ingress.Spec.TLS, tlsCfg)
	}
	for _, each := range obj.Spec.Rules {
		rule := &model.IngressRule{}
		if each.Host != "" {
			rule.Host = util.StringPtr(each.Host)
		}
		if each.HTTP != nil {
			rule.HTTP = &model.HTTPIngressRuleValue{}
			for _, path := range each.HTTP.Paths {
				p := &model.HTTPIngressPath{}
				if path.Path != "" {
					p.Path = util.StringPtr(path.Path)
				}
				if path.PathType != nil {
					p.PathType = util.StringPtr(string(*path.PathType))
				}
				p.Backend = &model.IngressBackend{}
				if path.Backend.Service != nil {
					p.Backend.Service = &model.IngressServiceBackend{
						Name: path.Backend.Service.Name,
						Port: &model.ServiceBackendPort{},
					}
					if path.Backend.Service.Port.Name != "" {
						p.Backend.Service.Port.Name = util.StringPtr(path.Backend.Service.Port.Name)
					}
					if path.Backend.Service.Port.Number != 0 {
						p.Backend.Service.Port.Number = util.Int64Ptr(int64(path.Backend.Service.Port.Number))
					}
				}
				rule.HTTP.Paths = append(rule.HTTP.Paths, p)
			}
		}
		ingress.Spec.Rules = append(ingress.Spec.Rules, rule)
	}

	ingress.Status = &model.IngressStatus{}
	for _, each := range obj.Status.LoadBalancer.Ingress {
		v := &model.IngressStatusIngress{
			IP: net.ParseIP(each.IP),
		}
		if each.Hostname != "" {
			v.Hostname = &each.Hostname
		}
		ingress.Status.Ingress = append(ingress.Status.Ingress, v)
	}
	if ingress.CreatedAt.IsZero() {
		ingress.CreatedAt = time.Unix(0, 0)
	}
	if ingress.LastModified.IsZero() {
		ingress.LastModified = ingress.CreatedAt
	}
}
