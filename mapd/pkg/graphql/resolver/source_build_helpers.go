// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"sort"
	"strconv"
	"time"

	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackcorev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getBuilds(client *k8s.KpackKubernetes, namespace string, selectors ...k8s.Selector) ([]buildv1alpha1.Build, error) {
	list, err := client.GetBuilds(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertBuild(build *model.SourceBuild, obj *buildv1alpha1.Build) {
	build.Name = obj.ObjectMeta.Name
	build.Annotations = obj.ObjectMeta.Annotations
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		build.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		build.CreatedAt = time.Now()
	}

	var images []*Image
	build.Spec = &model.SourceBuildSpec{}
	build.Spec.Source = &model.SourceBuildSpecSource{}
	if obj.Spec.Source.Git != nil {
		build.Spec.Source.Git = &model.SourceBuildSpecSourceGit{}
		if obj.Spec.Source.Git.Revision != "" {
			build.Spec.Source.Git.Revision = &obj.Spec.Source.Git.Revision
		}
		if obj.Spec.Source.Git.URL != "" {
			build.Spec.Source.Git.URL = abstract.MustParseURL(obj.Spec.Source.Git.URL)
		}
	}
	if obj.Spec.Source.Blob != nil {
		build.Spec.Source.Blob = &model.SourceBuildSpecSourceBlob{}
		if obj.Spec.Source.Blob.URL != "" {
			build.Spec.Source.Blob.URL = abstract.MustParseURL(obj.Spec.Source.Blob.URL)
		}
	}
	if obj.Spec.Source.Registry != nil {
		build.Spec.Source.Registry = &model.SourceBuildSpecSourceRegistry{
			Image: &obj.Spec.Source.Registry.Image,
		}
	}
	build.Spec.Source.SubPath = &obj.Spec.Source.SubPath
	for _, each := range obj.Spec.Tags {
		image, err := ParseImage(each)
		if err != nil {
			continue
		}
		images = append(images, image)
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i].Tag < images[j].Tag || images[i].Repository < images[j].Repository || images[i].Project < images[j].Project || images[i].Registry < images[j].Registry
	})
	if len(images) > 0 {
		s := images[0].String()
		build.Spec.DistImage = &s
	}

	status := model.JobStatusPending
	build.Status = &model.SourceBuildStatus{}
	if obj.ObjectMeta.Labels != nil {
		num, ok := obj.ObjectMeta.Labels["image.kpack.io/buildNumber"]
		if ok {
			i, _ := strconv.ParseInt(num, 10, 64)
			build.Status.Num = &i
		}
	}
	for _, each := range obj.Status.Conditions {
		if each.Type == kpackcorev1alpha1.ConditionSucceeded {
			build.LastModified = each.LastTransitionTime.Inner.Time
			switch each.Status {
			case corev1.ConditionTrue:
				status = model.JobStatusSuccess
			case corev1.ConditionFalse:
				status = model.JobStatusFailure
			}
			break
		}
	}
	if status == model.JobStatusPending {
		if len(obj.Status.BuildMetadata) > 0 || len(obj.Status.StepStates) > 0 {
			status = model.JobStatusRunning
		}
	}
	build.Status.Status = &status
	if obj.Status.LatestImage != "" {
		build.Status.Image = &obj.Status.LatestImage
	}
	if len(images) > 0 {
		image := images[len(images)-1]
		if image.Tag != "" {
			s := image.String()
			build.Status.Image = &s
		}
	}

	if build.CreatedAt.IsZero() {
		build.CreatedAt = time.Unix(0, 0)
	}
	if build.LastModified.IsZero() {
		build.LastModified = build.CreatedAt
	}
}
