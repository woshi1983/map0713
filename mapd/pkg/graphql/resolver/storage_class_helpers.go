// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	storagev1 "k8s.io/api/storage/v1"
)

func getStorageClasses(client *k8s.Kubernetes, selectors ...k8s.Selector) ([]storagev1.StorageClass, error) {
	list, err := client.GetStorageClasses(selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertStorageClass(sc *model.StorageClass, obj *storagev1.StorageClass) {
	sc.Name = obj.ObjectMeta.Name
	sc.Annotations = obj.ObjectMeta.Annotations
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				sc.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				sc.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				sc.Instance = &model.Instance{ID: id}
			}
		}
	}
	sc.Labels = obj.ObjectMeta.Labels
	sc.Provisioner = obj.Provisioner
	sc.Parameters = obj.Parameters
	if obj.ReclaimPolicy != nil {
		sc.ReclaimPolicy = util.StringPtr(string(*obj.ReclaimPolicy))
	}
	sc.MountOptions = obj.MountOptions
	sc.AllowVolumeExpansion = obj.AllowVolumeExpansion
	if obj.VolumeBindingMode != nil {
		sc.VolumeBindingMode = util.StringPtr(string(*obj.VolumeBindingMode))
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		sc.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		sc.CreatedAt = time.Now()
	}

	if sc.CreatedAt.IsZero() {
		sc.CreatedAt = time.Unix(0, 0)
	}
	if sc.LastModified.IsZero() {
		sc.LastModified = sc.CreatedAt
	}
}
