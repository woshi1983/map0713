// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"time"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteBackupFromStorageLocation(client *k8s.VeleroKubernetes, namespace, name string) error {
	obj := &velerov1.DeleteBackupRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: velerov1.DeleteBackupRequestSpec{
			BackupName: name,
		},
	}
	_, err := client.CreateDeleteBackupRequest(obj)
	if err != nil && k8s.IsErrObjectAlreadyExists(err) {
		return nil
	}
	return err
}

func convertBackup(backup *model.Backup, obj *velerov1.Backup) {
	backup.Name = obj.ObjectMeta.Name
	backup.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Annotations != nil {
		var annotation model1.BackupAnnotation
		anno, ok := obj.ObjectMeta.Annotations[model1.BackupAnnotationName]
		if ok {
			json.Unmarshal([]byte(anno), &annotation)

			for _, each := range annotation.Instances {
				backup.Instances = append(backup.Instances, &model.BackupInstance{
					Name:      each.Name,
					Namespace: each.Namespace,
				})
			}
		}
	}
	if obj.ObjectMeta.Labels != nil {
		loc, ok := obj.ObjectMeta.Labels["velero.io/storage-location"]
		if ok {
			backup.Location = &model.BackupStorageLocation{
				Name: loc,
			}
		}
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		backup.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		backup.CreatedAt = time.Now()
	}

	backup.Spec = &model.BackupSpec{
		IncludedNamespaces:      obj.Spec.IncludedNamespaces,
		ExcludedNamespaces:      obj.Spec.ExcludedNamespaces,
		IncludedResources:       obj.Spec.IncludedResources,
		ExcludedResources:       obj.Spec.ExcludedResources,
		TTL:                     obj.Spec.TTL.Duration,
		IncludeClusterResources: obj.Spec.IncludeClusterResources,
	}
	if obj.Spec.DefaultVolumesToRestic != nil {
		backup.Spec.UseRestic = *obj.Spec.DefaultVolumesToRestic
	}

	switch obj.Status.Phase {
	case velerov1.BackupPhaseNew:
		backup.Status = model.BackupStatusPending
	case velerov1.BackupPhaseInProgress:
		backup.Status = model.BackupStatusRunning
	case velerov1.BackupPhaseCompleted:
		backup.Status = model.BackupStatusSuccess
	case velerov1.BackupPhaseFailedValidation, velerov1.BackupPhasePartiallyFailed, velerov1.BackupPhaseFailed:
		backup.Status = model.BackupStatusFailure
	default:
		backup.Status = model.BackupStatusUnknown
	}
	if !obj.Status.CompletionTimestamp.IsZero() {
		backup.LastModified = obj.Status.CompletionTimestamp.Time
	}

	if backup.CreatedAt.IsZero() {
		backup.CreatedAt = time.Unix(0, 0)
	}
	if backup.LastModified.IsZero() {
		backup.LastModified = backup.CreatedAt
	}
}
