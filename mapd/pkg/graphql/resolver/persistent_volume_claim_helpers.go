// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getPersistentVolumeClaims(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]corev1.PersistentVolumeClaim, error) {
	list, err := client.GetPersistentVolumeClaims(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertPersistentVolumeClaim(pvc *model.PersistentVolumeClaim, obj *corev1.PersistentVolumeClaim) {
	pvc.Name = obj.ObjectMeta.Name
	pvc.Namespace = obj.ObjectMeta.Namespace
	pvc.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				pvc.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				pvc.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				pvc.Instance = &model.Instance{ID: id}
			}
		}
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		pvc.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		pvc.CreatedAt = time.Now()
	}

	pvc.Spec = &model.PersistentVolumeClaimSpec{}
	pvc.Spec.AccessModes = pvc.Spec.AccessModes[:0]
	for _, each := range obj.Spec.AccessModes {
		pvc.Spec.AccessModes = append(pvc.Spec.AccessModes, string(each))
	}
	pvc.Spec.Resources = &model.PersistentVolumeClaimSpecResources{}
	if obj.Spec.Resources.Limits != nil {
		pvc.Spec.Resources.Limits = &model.PersistentVolumeClaimSpecResourcesLimits{}
		if v := obj.Spec.Resources.Limits.Storage(); v != nil && !v.IsZero() {
			pvc.Spec.Resources.Limits.Storage = util.StringPtr(v.String())
		}
	}
	if obj.Spec.Resources.Requests != nil {
		pvc.Spec.Resources.Requests = &model.PersistentVolumeClaimSpecResourcesRequests{}
		if v := obj.Spec.Resources.Requests.Storage(); v != nil && !v.IsZero() {
			pvc.Spec.Resources.Requests.Storage = util.StringPtr(v.String())
		}
	}
	if obj.Spec.StorageClassName != nil {
		pvc.Spec.StorageClass = obj.Spec.StorageClassName
	}
	if obj.Spec.VolumeMode != nil {
		pvc.Spec.VolumeMode = util.StringPtr(string(*obj.Spec.VolumeMode))
	}
	pvc.Status = util.StringPtr(string(obj.Status.Phase))

	if pvc.CreatedAt.IsZero() {
		pvc.CreatedAt = time.Unix(0, 0)
	}
	if pvc.LastModified.IsZero() {
		pvc.LastModified = pvc.CreatedAt
	}
}
