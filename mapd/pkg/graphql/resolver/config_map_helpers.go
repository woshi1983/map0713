// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getConfigMaps(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]corev1.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetConfigMapsContext(ctx, namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertConfigMap(cm *model.ConfigMap, obj *corev1.ConfigMap) {
	cm.Name = obj.ObjectMeta.Name
	cm.Namespace = obj.ObjectMeta.Namespace
	cm.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				cm.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				cm.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				cm.Instance = &model.Instance{ID: id}
			}
		}
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		cm.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		cm.CreatedAt = time.Now()
	}

	cm.Data = make(map[string]string)
	if obj.Data != nil {
		for k, v := range obj.Data {
			cm.Data[k] = v
		}
	}
	if obj.BinaryData != nil {
		for k, v := range obj.BinaryData {
			cm.Data[k] = string(v)
		}
	}
	if cm.CreatedAt.IsZero() {
		cm.CreatedAt = time.Unix(0, 0)
	}
	if cm.LastModified.IsZero() {
		cm.LastModified = cm.CreatedAt
	}
}
