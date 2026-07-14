// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

func convertAuditLogFromSQLModel(al *model1.AuditLog) *model.AuditLog {
	obj := &model.AuditLog{
		ID:           al.ID,
		Action:       al.Action,
		ObjectType:   al.ObjectType,
		ObjectID:     al.ObjectID,
		ObjectName:   al.ObjectName,
		UserName:     al.UserName,
		User:         &model.User{ID: al.UserID},
		CreatedAt:    al.CreatedAt,
		LastModified: al.UpdatedAt,
	}
	if al.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: al.OrganizationID.UUID}
	}
	return obj
}

func finalizeAuditLogQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "action":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`action` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "objectType":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`object_type` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "objectName":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`object_name` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "object":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`object_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "userName":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`user_name` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "user":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`user_id` = ?")
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
				db = db.Where(strings.Join(exprs, " OR ")+" OR `organization_id` IS NULL", args...)
			}
		case "from_time":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`updated_at` > ?")

				t, err := time.Parse(time.RFC3339Nano, val)
				if err != nil {
					continue
				}
				args = append(args, t.Local().Format("2006-01-02 15:04:05"))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "until_time":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`updated_at` < ?")

				t, err := time.Parse(time.RFC3339Nano, val)
				if err != nil {
					continue
				}
				args = append(args, t.Local().Format("2006-01-02 15:04:05"))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	return db
}

func appendAuditLog(db *gorm.DB, action, objType string, objID uuid.UUID, objName string, userID uuid.UUID, userName string, orgID uuid.UUID) error {
	l := &model1.AuditLog{
		ID:         uuid.New(),
		Action:     action,
		ObjectType: objType,
		ObjectID:   objID,
		ObjectName: objName,
		UserID:     userID,
		UserName:   userName,
	}
	if orgID != uuid.Nil {
		l.OrganizationID = model1.NewNullUUID(orgID)
	}
	return db.Create(l).Error
}
