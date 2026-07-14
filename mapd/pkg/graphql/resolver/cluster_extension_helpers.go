// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"log"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

func convertClusterExtensionFromSQLModel(ext *model1.ClusterExtension) *model.ClusterExtension {
	obj := &model.ClusterExtension{
		ID:           ext.ID,
		Name:         model.ClusterExtensionName(ext.Name.String()),
		Spec:         make(map[string]string),
		State:        model.ClusterExtensionState(ext.State.String()),
		Status:       model.ClusterExtensionStatus(ext.Status.String()),
		CreatedAt:    ext.CreatedAt,
		LastModified: ext.UpdatedAt,
	}
	endpoints := ext.Endpoints
	if len(endpoints) == 0 {
		endpoints = []byte("[]")
	}
	var ss []string
	if err := json.Unmarshal(endpoints, &ss); err != nil {
		log.Printf("WARN: failed to unmarshal endpoints for extension %s: %v, endpoints=%q", ext.ID, err, string(endpoints))
		ss = []string{}
	}
	for _, s := range ss {
		obj.Endpoints = append(obj.Endpoints, abstract.MustParseURL(s))
	}
	spec := ext.Spec
	if len(spec) == 0 {
		spec = []byte("{}")
	}
	if err := json.Unmarshal(spec, &obj.Spec); err != nil {
		log.Printf("WARN: failed to unmarshal spec for extension %s: %v, spec=%q", ext.ID, err, string(spec))
		obj.Spec = make(map[string]string)
	}
	if ext.ClusterID.Valid {
		obj.Cluster = &model.Cluster{ID: ext.ClusterID.UUID}
	}
	if ext.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: ext.OrganizationID.UUID}
	}
	return obj
}

func finalizeClusterExtensionQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "status":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`status` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "related":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` IN (SELECT `extension_id` FROM `cluster_extension_related_cluster` WHERE `guest_id` = ?)")
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
