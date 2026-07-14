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

func convertJobToBinaryBuild(job *model.Job) *model.BinaryBuild {
	obj := &model.BinaryBuild{
		Name:        job.Name,
		Annotations: job.Annotations,
		Spec:        &model.BinaryBuildSpec{},
		Status: &model.BinaryBuildStatus{
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
				if v, ok := each.Env["BINARY_DEPLOY_DIR"]; ok {
					obj.Spec.BinaryDir = util.StringPtr(v)
				}
				if v, ok := each.Env["BINARY_BUILD_BASE_IMAGE_NAME"]; ok {
					obj.Spec.BaseImage = util.StringPtr(v)
				}
				if v, ok := each.Env["BINARY_FILE_URL"]; ok {
					obj.Spec.URL = abstract.MustParseURL(v)
				}
				if v, ok := each.Env["BINARY_PROCESS_RUN_UID"]; ok {
					i, _ := strconv.ParseInt(v, 10, 64)
					obj.Spec.ProcessUID = &i
				}
				break
			}
		}
	}
	return obj
}
