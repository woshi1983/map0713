// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

const (
	DefaultStandardPackageRepositoryName  = "harbor-standard-v210"
	DefaultStandardPackageRepositoryImage = "harbor.huaruicloud.cn/tkg/packages/standard/repo:v2.1.0"
)

type StandardPackageRepository struct {
	ID          uuid.UUID      `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name        string         `gorm:"type:varchar(128);index"`
	Description string         `gorm:"type:mediumtext"`
	Image       string         `gorm:"type:varchar(512);index"`
	Default     bool           `gorm:"type:bool;index"`
	CreatedAt   time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt   time.Time      `gorm:"type:datetime(6)"`
	DeletedAt   gorm.DeletedAt `gorm:"type:datetime(6);index"`
}

func (typedef StandardPackageRepository) TableName() string {
	return "standard_package_repository"
}

func ModelStandardPackageRepository(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &StandardPackageRepository{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		var count int64
		if err := db.Model(typedef).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return db.Create(&StandardPackageRepository{
			ID:      uuid.New(),
			Name:    DefaultStandardPackageRepositoryName,
			Image:   DefaultStandardPackageRepositoryImage,
			Default: true,
		}).Error
	}
}
