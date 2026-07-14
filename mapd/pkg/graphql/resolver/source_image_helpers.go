// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackcorev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getImages(client *k8s.KpackKubernetes, namespace string, selectors ...k8s.Selector) ([]buildv1alpha1.Image, error) {
	list, err := client.GetImages(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertImage(image *model.SourceImage, obj *buildv1alpha1.Image) {
	image.Name = obj.ObjectMeta.Name
	image.Annotations = obj.ObjectMeta.Annotations
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		image.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		image.CreatedAt = time.Now()
	}

	image.Spec = &model.SourceImageSpec{}
	image.Spec.Source = &model.SourceImageSpecSource{}
	if obj.Spec.Source.Git != nil {
		image.Spec.Source.Git = &model.SourceImageSpecSourceGit{
			Revision: util.StringPtr(obj.Spec.Source.Git.Revision),
			URL:      abstract.MustParseURL(obj.Spec.Source.Git.URL),
		}
	}
	if obj.Spec.Source.Blob != nil {
		image.Spec.Source.Blob = &model.SourceImageSpecSourceBlob{
			URL: abstract.MustParseURL(obj.Spec.Source.Blob.URL),
		}
	}
	if obj.Spec.Source.Registry != nil {
		image.Spec.Source.Registry = &model.SourceImageSpecSourceRegistry{
			Image: &obj.Spec.Source.Registry.Image,
		}
	}
	image.Spec.Source.SubPath = util.StringPtr(obj.Spec.Source.SubPath)
	if obj.Spec.Build != nil {
		image.Spec.Env = make(map[string]string)
		for _, each := range obj.Spec.Build.Env {
			image.Spec.Env[each.Name] = each.Value
		}
	}
	if obj.Spec.Tag != "" {
		image.Spec.DistImage = util.StringPtr(obj.Spec.Tag)
	}

	image.Status = &model.SourceImageStatus{}
	ready := false
	if obj.Status.Conditions != nil {
		for _, each := range obj.Status.Conditions {
			if each.Type == kpackcorev1alpha1.ConditionReady {
				image.LastModified = each.LastTransitionTime.Inner.Time
				ready = each.Status == corev1.ConditionTrue
				break
			}
		}
	}
	image.Status.Ready = util.BoolPtr(ready)

	if image.CreatedAt.IsZero() {
		image.CreatedAt = time.Unix(0, 0)
	}
	if image.LastModified.IsZero() {
		image.LastModified = image.CreatedAt
	}
}
