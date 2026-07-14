// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
)

func convertOrganizationFromSQLModel(organization *model1.Organization) *model.Organization {
	obj := &model.Organization{
		ID:           organization.ID,
		Name:         organization.Name,
		Description:  organization.Description,
		Quota:        &model.OrganizationQuota{ID: organization.ID},
		CreatedAt:    organization.CreatedAt,
		LastModified: organization.UpdatedAt,
	}
	if organization.Avatar != "" {
		obj.Avatar = abstract.MustParseURL(organization.Avatar)
	}
	if organization.Link != "" {
		obj.Link = abstract.MustParseURL(organization.Link)
	}
	return obj
}

func finalizeOrganizationQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		}
	}
	return db
}

func getOrganizationQuotaUsage(db *gorm.DB, id uuid.UUID) (cpus, memMi, storageGi int64) {
	var count int64
	if err := db.Model(&model1.OrganizationQuotaUsage{}).Where("`organization_id` = ?", id).Count(&count).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	n := count / 100
	m := count % 100
	if m > 0 {
		n++
	}
	for i := int64(0); i < n; i++ {
		us := []model1.OrganizationQuotaUsage{}
		if err := db.Where("`organization_id` = ?", id).Offset(int(i) * 100).Limit(100).Find(&us).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		for _, u := range us {
			cpus += u.CPUs
			memMi += u.MemoryMi
			storageGi += u.StorageGi
		}
	}
	return
}

func getOrganizationQuota(db *gorm.DB, id uuid.UUID) (cpus, memMi, storageGi int64) {
	quota := &model1.OrganizationQuota{}
	if err := db.Where("`organization_id` = ?", id).Take(quota).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(&ErrDBConnection{err})
	}
	return quota.CPUs, quota.MemoryMi, quota.StorageGi
}

func validateOrganizationQuotaUsage(db *gorm.DB, id uuid.UUID) bool {
	usedCPUs, usedMemMi, usedStorageGi := getOrganizationQuotaUsage(db, id)
	totalCPUs, totalMemMi, totalStorageGi := getOrganizationQuota(db, id)
	return (totalCPUs < 1 || usedCPUs <= totalCPUs) &&
		(totalMemMi < 1 || usedMemMi <= totalMemMi) &&
		(totalStorageGi < 1 || usedStorageGi <= totalStorageGi)
}
