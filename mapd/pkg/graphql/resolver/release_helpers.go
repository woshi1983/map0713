// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func convertReleaseFromSQLModel(release *model1.Release) *model.Release {
	obj := &model.Release{
		ID:           release.ID,
		Name:         release.Name,
		Num:          release.Num,
		CreatedAt:    release.CreatedAt,
		LastModified: release.UpdatedAt,
	}
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err := k8s.WalkManifest(release.Manifest, func(p []byte) error {
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
	if release.ComponentID.Valid {
		obj.Component = &model.Component{ID: release.ComponentID.UUID}
	}
	if release.ApplicationID.Valid {
		obj.Application = &model.Application{ID: release.ApplicationID.UUID}
	}
	if release.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: release.OrganizationID.UUID}
	}
	return obj
}

func finalizeReleaseQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
