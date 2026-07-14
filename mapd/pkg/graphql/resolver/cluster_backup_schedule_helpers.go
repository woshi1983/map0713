// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

func convertClusterBackupScheduleFromSQLModel(schedule *model1.ClusterBackupSchedule) *model.ClusterBackupSchedule {
	obj := &model.ClusterBackupSchedule{
		ID:           schedule.ID,
		Name:         schedule.Name,
		Location:     schedule.Location,
		Schedule:     schedule.Schedule,
		NextFire:     schedule.NextFire,
		CreatedAt:    schedule.CreatedAt,
		LastModified: schedule.UpdatedAt,
	}
	spec := &model1.ClusterBackupScheduleSpec{}
	if err := json.Unmarshal(schedule.Spec, spec); err != nil {
		panic(err)
	}
	obj.Spec = &model.BackupSpec{
		UseRestic:          spec.UseRestic,
		IncludedNamespaces: spec.IncludedNamespaces,
		ExcludedNamespaces: spec.ExcludedNamespaces,
		TTL:                spec.TTL,
	}
	if !schedule.LastFire.IsZero() {
		obj.LastFire = &schedule.LastFire
	}
	if schedule.ClusterID.Valid {
		obj.Cluster = &model.Cluster{ID: schedule.ClusterID.UUID}
	}
	if schedule.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: schedule.OrganizationID.UUID}
	}
	return obj
}

func finalizeClusterBackupScheduleQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`name` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "location":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`location` = ?")
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
