// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"net/url"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	mapv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/mapd/v1alpha1"
)

func getBinaryImages(client *k8s.MapKubernetes, namespace string, selectors ...k8s.Selector) ([]mapv1alpha1.BinaryImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetBinaryImagesContext(ctx, namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertBinaryImage(image *model.BinaryImage, obj *mapv1alpha1.BinaryImage) {
	image.Name = obj.ObjectMeta.Name
	image.Annotations = obj.ObjectMeta.Annotations
	image.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		image.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		image.CreatedAt = time.Now()
	}

	image.Spec = &model.BinaryImageSpec{
		BaseImage:  obj.Spec.BaseImage,
		DistImage:  obj.Spec.DistImage,
		BinaryDir:  obj.Spec.BinaryDir,
		ProcessUID: obj.Spec.ProcessUID,
	}
	v, err := url.Parse(obj.Spec.URL)
	if err == nil {
		image.Spec.URL = v
	}

	if image.CreatedAt.IsZero() {
		image.CreatedAt = time.Unix(0, 0)
	}
	if image.LastModified.IsZero() {
		image.LastModified = image.CreatedAt
	}
}
