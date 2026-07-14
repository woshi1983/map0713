// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func convertInstanceFromSQLModel(instance *model1.Instance) *model.Instance {
	obj := &model.Instance{
		ID:           instance.ID,
		Name:         instance.Name,
		Namespace:    instance.Namespace,
		AutoRestart:  instance.AutoRestart,
		Status:       &model.InstanceStatus{State: model.InstanceState(instance.State.String())},
		CreatedAt:    instance.CreatedAt,
		LastModified: instance.UpdatedAt,
	}
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err := k8s.WalkManifest(instance.Manifest, func(p []byte) error {
		o := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, o)
		if err != nil {
			return err
		}
		b, err := json.Marshal(o)
		if err != nil {
			return err
		}
		obj.Manifest = append(obj.Manifest, b)
		return nil
	}); err != nil {
		panic(err)
	}
	v := []*model1.InstanceStatusTask{}
	if err := json.Unmarshal(instance.Status, &v); err != nil {
		panic(err)
	}
	for _, each := range v {
		t := &model.InstanceStatusTask{
			Name:    each.Name,
			Kind:    each.Kind,
			State:   model.InstanceState(each.State.String()),
			Message: each.Message,
		}
		if each.Namespace != "" {
			t.Namespace = &each.Namespace
		}
		obj.Status.Tasks = append(obj.Status.Tasks, t)
	}
	if instance.ClusterID.Valid {
		obj.Cluster = &model.Cluster{ID: instance.ClusterID.UUID}
	}
	if instance.ReleaseID.Valid {
		obj.Release = &model.Release{ID: instance.ReleaseID.UUID}
	}
	if instance.FederationID.Valid {
		obj.Federation = &model.InstanceFederation{ID: instance.FederationID.UUID}
	}
	if instance.ComponentID.Valid {
		obj.Component = &model.Component{ID: instance.ComponentID.UUID}
	}
	if instance.ApplicationID.Valid {
		obj.Application = &model.Application{ID: instance.ApplicationID.UUID}
	}
	if instance.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: instance.OrganizationID.UUID}
	}
	return obj
}

func finalizeInstanceQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`name` LIKE ?")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "namespace":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`namespace` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "cluster":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`cluster_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "release":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`release_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "federation":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`federation_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "component":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`component_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "application":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`application_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "organization":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`organization_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	return db
}

func getWorkloadHealthForInstance(k *k8s.Kubernetes, manifest []json.RawMessage) ([]*model.InstanceWorkloadHealth, error) {
	var workloads []*model.InstanceWorkloadHealth
LOOP:
	for _, p := range manifest {
		obj := &unstructured.Unstructured{}
		if err := json.Unmarshal(p, obj); err != nil {
			panic(errors.Wrap(err, "failed to parse object from manifest"))
		}

		workload := &model.InstanceWorkloadHealth{
			Name: obj.GetName(),
			Kind: obj.GetKind(),
		}
		if obj.GetNamespace() != "" {
			workload.Namespace = util.StringPtr(obj.GetNamespace())
		}

		switch workload.Kind {
		case "Pod":
			pod, err := k.GetPod(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			for _, cond := range pod.Status.Conditions {
				if cond.Type == "Ready" {
					if cond.Status == "True" {
						workload.ReadyReplicas = 1
					}
					break
				}
			}
			workload.Replicas = 1
			workloads = append(workloads, workload)
		case "Deployment":
			deploy, err := k.GetDeployment(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			workload.ReadyReplicas, workload.Replicas = int64(deploy.Status.ReadyReplicas), int64(deploy.Status.Replicas)
			workloads = append(workloads, workload)
		case "ReplicaSet":
			rs, err := k.GetReplicaSet(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			workload.ReadyReplicas, workload.Replicas = int64(rs.Status.ReadyReplicas), int64(rs.Status.Replicas)
			workloads = append(workloads, workload)
		case "StatefulSet":
			sts, err := k.GetStatefulSet(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			workload.ReadyReplicas, workload.Replicas = int64(sts.Status.ReadyReplicas), int64(sts.Status.Replicas)
			workloads = append(workloads, workload)
		case "DaemonSet":
			sts, err := k.GetDaemonSet(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			workload.ReadyReplicas, workload.Replicas = int64(sts.Status.NumberReady), int64(sts.Status.DesiredNumberScheduled)
			workloads = append(workloads, workload)
		case "Job":
			job, err := k.GetJob(obj.GetNamespace(), obj.GetName())
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					return nil, err
				}
				workloads = append(workloads, workload)
				continue LOOP
			}

			for _, cond := range job.Status.Conditions {
				if cond.Type == "Complete" {
					if cond.Status == "True" {
						workload.ReadyReplicas = 1
					}
					break
				}
			}
			workload.Replicas = 1
			workloads = append(workloads, workload)
		}
	}
	return workloads, nil
}
