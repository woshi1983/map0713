// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Organization struct {
	ID          uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name        string    `gorm:"type:varchar(128);index"`
	Description string    `gorm:"type:mediumtext"`
	Avatar      string    `gorm:"type:mediumtext"`
	Link        string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"type:datetime(6)"`
	UpdatedAt   time.Time `gorm:"type:datetime(6)"`
}

func (typedef Organization) TableName() string {
	return "organization"
}

func ModelOrganization(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Organization{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}

type OrganizationQuota struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	CPUs           int64     `gorm:"column:cpus"`
	MemoryMi       int64     `gorm:"column:memory_mi"`
	StorageGi      int64     `gorm:"column:storage_gi"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef OrganizationQuota) TableName() string {
	return "organization_quota"
}

func ModelOrganizationQuota(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &OrganizationQuota{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}

type OrganizationQuotaUsage struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	CPUs           int64     `gorm:"column:cpus"`
	MemoryMi       int64     `gorm:"column:memory_mi"`
	StorageGi      int64     `gorm:"column:storage_gi"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef OrganizationQuotaUsage) TableName() string {
	return "organization_quota_usage"
}

func ModelOrganizationQuotaUsage(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &OrganizationQuotaUsage{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}
