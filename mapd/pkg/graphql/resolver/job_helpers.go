// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	batchv1 "k8s.io/api/batch/v1"
)

func getJobs(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]batchv1.Job, error) {
	list, err := client.GetJobs(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertJob(job *model.Job, obj *batchv1.Job) {
	job.Name = obj.ObjectMeta.Name
	job.Namespace = obj.ObjectMeta.Namespace
	job.Annotations = obj.ObjectMeta.Annotations
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				job.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				job.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				job.Instance = &model.Instance{ID: id}
			}
		}
	}
	job.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		job.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		job.CreatedAt = time.Now()
	}

	job.Spec = &model.JobSpec{
		ActiveDeadlineSeconds: obj.Spec.ActiveDeadlineSeconds,
		BackoffLimit:          util.Int32PtrToInt64Ptr(obj.Spec.BackoffLimit),
		Completions:           util.Int32PtrToInt64Ptr(obj.Spec.Completions),
		Parallelism:           util.Int32PtrToInt64Ptr(obj.Spec.Parallelism),
		Template:              &model.PodTemplate{},
	}
	convertPodTemplate(job.Spec.Template, obj.Spec.Template, job.Namespace, job.Cluster)

	if obj.Status.Succeeded > 0 {
		job.Status = model.JobStatusSuccess
	} else if obj.Status.Failed > 0 || (obj.Spec.ActiveDeadlineSeconds != nil && time.Since(obj.GetCreationTimestamp().Time) > time.Duration(*obj.Spec.ActiveDeadlineSeconds)*time.Second) {
		job.Status = model.JobStatusFailure
	} else if total := obj.Status.Active + obj.Status.Failed + obj.Status.Succeeded; obj.Status.Active > 0 || (obj.Spec.BackoffLimit != nil && total < *obj.Spec.BackoffLimit) {
		job.Status = model.JobStatusRunning
	} else {
		job.Status = model.JobStatusUnknown
	}
	if !obj.Status.CompletionTime.IsZero() {
		job.LastModified = obj.Status.CompletionTime.Time
	}

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Unix(0, 0)
	}
	if job.LastModified.IsZero() {
		job.LastModified = job.CreatedAt
	}
}
