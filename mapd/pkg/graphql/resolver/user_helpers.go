// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
)

// Users have no organization_id column. Their role bindings determine which
// organizations may expose them to a non-superadmin caller.
func filterVisibleUsers(db *gorm.DB, sess *auth.SessionInfo) *gorm.DB {
	if sess.Admin {
		return db
	}
	roles := db.Session(&gorm.Session{NewDB: true}).Model(&model1.RoleBinding{}).Select("user_binding_id")
	roles = sess.FilterQueryByOrganization(roles, model1.PermReadOnly, model1.PolicyObjectUser)
	return db.Where("`user_binding`.`id` IN (?)", roles)
}

func convertUserFromSQLModel(user *model1.UserBinding) *model.User {
	obj := &model.User{
		ID:           user.ID,
		Name:         user.Name,
		DisplayName:  user.FullName,
		Email:        user.Email,
		Remark:       user.Remark,
		Admin:        user.Admin,
		Managed:      user.Managed,
		CreatedAt:    user.CreatedAt,
		LastModified: user.UpdatedAt,
	}
	return obj
}

func finalizeUserQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "displayName":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`full_name` LIKE ?")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "email":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`email` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	return db
}
