// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type ImageRegistry struct {
	ID             uuid.UUID         `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string            `gorm:"type:varchar(128);index"`
	Description    string            `gorm:"type:mediumtext"`
	Endpoint       string            `gorm:"type:varchar(255);index"`
	CACert         []byte            `gorm:"type:mediumblob"`
	AuthUser       string            `gorm:"type:varchar(255)"`
	AuthPassword   []byte            `gorm:"type:blob(65535)"` // AES-128 encrypted with its ID
	Type           ImageRegistryType `gorm:"type:varchar(128)"`
	OrganizationID NullUUID          `gorm:"type:varchar(255)"`
	CreatedAt      time.Time         `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time         `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt    `gorm:"type:datetime(6);index"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ImageRegistry) TableName() string {
	return "image_registry"
}

func ModelImageRegistry(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ImageRegistry{}
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

type ImageRegistryType string

func (t ImageRegistryType) String() string {
	if t == ImageRegistryTypeUnknown {
		return "UNKNOWN"
	}
	return string(t)
}

func (t *ImageRegistryType) FromString(s string) {
	switch strings.ToUpper(s) {
	case "HARBOR":
		*t = ImageRegistryTypeHarbor
	default:
		*t = ImageRegistryTypeUnknown
	}
}

const (
	ImageRegistryTypeUnknown ImageRegistryType = ""
	ImageRegistryTypeHarbor  ImageRegistryType = "HARBOR"
)
