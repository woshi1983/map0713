// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getSecrets(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]corev1.Secret, error) {
	list, err := client.GetSecrets(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertSecret(secret *model.Secret, obj *corev1.Secret) {
	secret.Name = obj.ObjectMeta.Name
	secret.Namespace = obj.ObjectMeta.Namespace
	secret.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				secret.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				secret.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				secret.Instance = &model.Instance{ID: id}
			}
		}
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		secret.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		secret.CreatedAt = time.Now()
	}

	if obj.Data != nil {
		secret.Data = make(map[string]string)
		for k, v := range obj.Data {
			secret.Data[k] = string(v)
		}
	}
	typ := string(obj.Type)
	secret.Type = &typ

	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = time.Unix(0, 0)
	}
	if secret.LastModified.IsZero() {
		secret.LastModified = secret.CreatedAt
	}
}
