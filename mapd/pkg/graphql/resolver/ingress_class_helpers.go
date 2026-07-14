// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	networkingv1 "k8s.io/api/networking/v1"
)

func getIngressClasses(client *k8s.Kubernetes, selectors ...k8s.Selector) ([]networkingv1.IngressClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetIngressClassesContext(ctx, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertIngressClass(cls *model.IngressClass, obj *networkingv1.IngressClass) {
	cls.Name = obj.ObjectMeta.Name
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				cls.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				cls.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				cls.Instance = &model.Instance{ID: id}
			}
		}
	}
	cls.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		cls.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		cls.CreatedAt = time.Now()
	}

	if cls.CreatedAt.IsZero() {
		cls.CreatedAt = time.Unix(0, 0)
	}
	if cls.LastModified.IsZero() {
		cls.LastModified = cls.CreatedAt
	}
}
