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

func convertImageRegistryFromSQLModel(registry *model1.ImageRegistry) *model.ImageRegistry {
	obj := &model.ImageRegistry{
		ID:           registry.ID,
		Name:         registry.Name,
		Description:  registry.Description,
		Endpoint:     abstract.MustParseURL(registry.Endpoint),
		AuthUser:     registry.AuthUser,
		CaCert:       string(registry.CACert),
		Type:         model.ImageRegistryType(registry.Type.String()),
		CreatedAt:    registry.CreatedAt,
		LastModified: registry.UpdatedAt,
	}
	if registry.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: registry.OrganizationID.UUID}
	}
	return obj
}

func finalizeImageRegistryQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "pushable":
			db = db.Where("`auth_user` != ''")
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
