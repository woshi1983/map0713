// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"net"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getServices(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]corev1.Service, error) {
	list, err := client.GetServices(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertService(service *model.Service, obj *corev1.Service) {
	service.Name = obj.ObjectMeta.Name
	service.Namespace = obj.ObjectMeta.Namespace
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				service.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				service.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				service.Instance = &model.Instance{ID: id}
			}
		}
	}
	service.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		service.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		service.CreatedAt = time.Now()
	}

	service.Spec = &model.ServiceSpec{
		Selector: obj.Spec.Selector,
	}
	if obj.Spec.ClusterIP != "" {
		service.Spec.ClusterIP = &obj.Spec.ClusterIP
	}
	if obj.Spec.ExternalTrafficPolicy != "" {
		service.Spec.ExternalTrafficPolicy = util.StringPtr(string(obj.Spec.ExternalTrafficPolicy))
	}
	for _, each := range obj.Spec.Ports {
		port := &model.ServiceSpecPort{}
		if each.Name != "" {
			port.Name = &each.Name
		}
		if each.NodePort != 0 {
			port.NodePort = util.Int64Ptr(int64(each.NodePort))
		}
		if each.Port != 0 {
			port.Port = util.Int64Ptr(int64(each.Port))
		}
		if each.TargetPort.IntVal != 0 && each.TargetPort.StrVal != "" {
			port.TargetPort = &each.TargetPort
		}
		if each.Protocol != "" {
			port.Protocol = util.StringPtr(string(each.Protocol))
		}
		service.Spec.Ports = append(service.Spec.Ports, port)
	}
	if obj.Spec.SessionAffinity != "" {
		service.Spec.SessionAffinity = util.StringPtr(string(obj.Spec.SessionAffinity))
	}
	if obj.Spec.Type != "" {
		service.Spec.Type = util.StringPtr(string(obj.Spec.Type))
	}

	service.Status = &model.ServiceStatus{}
	for _, each := range obj.Status.LoadBalancer.Ingress {
		service.Status.Ingress = append(service.Status.Ingress, &model.ServiceStatusIngress{
			IP: net.ParseIP(each.IP),
		})
	}
	if service.CreatedAt.IsZero() {
		service.CreatedAt = time.Unix(0, 0)
	}
	if service.LastModified.IsZero() {
		service.LastModified = service.CreatedAt
	}
}

func convertServiceSpecFromInput(input *model.CreateServiceSpec) *model.ServiceSpec {
	spec := &model.ServiceSpec{
		Type: input.Type,
	}
	for _, each := range input.Ports {
		port := &model.ServiceSpecPort{
			Name:       each.Name,
			Port:       &each.Port,
			Protocol:   each.Protocol,
			TargetPort: &each.TargetPort,
		}
		spec.Ports = append(spec.Ports, port)
	}
	return spec
}
