// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Shortcut struct {
	ID             uuid.UUID      `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string         `gorm:"type:varchar(128);index"`
	Description    string         `gorm:"type:mediumtext"`
	Endpoint       string         `gorm:"type:mediumtext"`
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef Shortcut) TableName() string {
	return "shortcut"
}

func ModelShortcut(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Shortcut{}
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
