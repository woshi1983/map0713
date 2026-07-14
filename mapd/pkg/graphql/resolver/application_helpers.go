// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
)

func convertApplicationFromSQLModel(app *model1.Application) *model.Application {
	obj := &model.Application{
		ID:           app.ID,
		Name:         app.Name,
		Description:  app.Description,
		CreatedAt:    app.CreatedAt,
		LastModified: app.UpdatedAt,
	}
	if app.Avatar != "" {
		obj.Avatar = abstract.MustParseURL(app.Avatar)
	}
	if app.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: app.OrganizationID.UUID}
	}
	return obj
}

func finalizeApplicationQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
