// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"fmt"
	"strconv"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
)

func convertJobToDockerfileBuild(job *model.Job) *model.DockerfileBuild {
	obj := &model.DockerfileBuild{
		Name:        job.Name,
		Annotations: job.Annotations,
		Spec:        &model.DockerfileBuildSpec{},
		Status: &model.DockerfileBuildStatus{
			Status: job.Status,
			Job:    job,
		},
		Cluster:      job.Cluster,
		CreatedAt:    job.CreatedAt,
		LastModified: job.LastModified,
	}
	if job.Spec != nil && job.Spec.Template != nil && job.Spec.Template.Spec != nil {
		for _, each := range job.Spec.Template.Spec.Containers {
			if each.Name == "main" && each.Env != nil {
				image := &Image{
					Registry:   each.Env["OUTPUT_IMAGE_REGISTRY"],
					Project:    each.Env["OUTPUT_IMAGE_REPO"],
					Repository: each.Env["OUTPUT_IMAGE_NAME"],
				}
				obj.Spec.DistImage = util.StringPtr(image.String())
				if each.Env != nil && job.Status == model.JobStatusSuccess {
					image.Tag = fmt.Sprintf("b%s.%s", each.Env["BUILD_NUM"], each.Env["OUTPUT_IMAGE_TAG"])
					path := image.String()
					obj.Status.Image = &path
				}
				if v, ok := each.Env["BUILD_NUM"]; ok {
					obj.Status.Num, _ = strconv.ParseInt(v, 10, 64)
				}
				if v, ok := each.Env["DOCKERFILE_BUILD_CONTEXT_TYPE"]; ok {
					obj.Spec.Source = util.StringPtr(v)
				}
				if v, ok := each.Env["DOCKERFILE_BUILD_GIT_BRANCH"]; ok {
					obj.Spec.Branch = util.StringPtr(v)
				}
				if v, ok := each.Env["DOCKERFILE_BUILD_GIT_SUBDIR"]; ok {
					obj.Spec.SubDir = util.StringPtr(v)
				}
				if v, ok := each.Env["DOCKERFILE_BUILD_DOCKERFILE_NAME"]; ok {
					obj.Spec.FileName = util.StringPtr(v)
				}
				if v, ok := each.Env["DOCKERFILE_BUILD_CONTEXT_URL"]; ok {
					obj.Spec.URL = abstract.MustParseURL(v)
				}
				break
			}
		}
	}
	return obj
}
