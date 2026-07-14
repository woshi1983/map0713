// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
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

func convertStandardPackageRepositoryFromSQLModel(repo *model1.StandardPackageRepository) *model.StandardPackageRepository {
	return &model.StandardPackageRepository{
		ID:           repo.ID,
		Name:         repo.Name,
		Description:  repo.Description,
		Image:        repo.Image,
		Default:      repo.Default,
		CreatedAt:    repo.CreatedAt,
		LastModified: repo.UpdatedAt,
	}
}

func finalizeStandardPackageRepositoryQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "default":
			for _, val := range each.Value {
				if strings.EqualFold(val, "true") {
					db = db.Where("`default` = ?", true)
					break
				}
				if strings.EqualFold(val, "false") {
					db = db.Where("`default` = ?", false)
					break
				}
			}
		}
	}
	return db
}

func canStandardPackageRepository(sess *auth.SessionInfo, perm model1.Perm) bool {
	if sess.Admin {
		return true
	}
	for organization := range sess.Perm {
		if sess.Can(perm, model1.PolicyObjectClusterProvider, organization) {
			return true
		}
	}
	return false
}
