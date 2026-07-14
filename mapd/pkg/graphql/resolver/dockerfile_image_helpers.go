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

func postDockerfileImage(client *k8s.MapKubernetes, obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.CreateDockerfileImageContext(ctx, obj)
}

func putDockerfileImage(client *k8s.MapKubernetes, obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.UpdateDockerfileImageContext(ctx, obj)
}

func getDockerfileImage(client *k8s.MapKubernetes, namespace, name string) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.GetDockerfileImageContext(ctx, namespace, name)
}

func getDockerfileImages(client *k8s.MapKubernetes, namespace string, selectors ...k8s.Selector) ([]mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetDockerfileImagesContext(ctx, namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertDockerfileImage(image *model.DockerfileImage, obj *mapv1alpha1.DockerfileImage) {
	image.Name = obj.ObjectMeta.Name
	image.Annotations = obj.ObjectMeta.Annotations
	image.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		image.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		image.CreatedAt = time.Now()
	}

	image.Spec = &model.DockerfileImageSpec{
		Source:    obj.Spec.Source,
		DistImage: obj.Spec.DistImage,
		Branch:    obj.Spec.Branch,
		SubDir:    obj.Spec.SubDir,
		FileName:  obj.Spec.FileName,
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
