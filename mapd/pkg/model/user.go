// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type UserBinding struct {
	ID        uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name      string    `gorm:"type:varchar(255);index"`      // generated
	FullName  string    `gorm:"type:varchar(255);index"`      // at:name
	Email     string    `gorm:"type:varchar(255);index"`      // at:email
	Remark    string    `gorm:"type:varchar(255);index"`      // at:preferred_username
	Admin     bool
	Managed   bool
	CreatedAt time.Time `gorm:"type:datetime(6)"`
	UpdatedAt time.Time `gorm:"type:datetime(6)"`
}

func (typedef UserBinding) TableName() string {
	return "user_binding"
}

func ModelUserBinding(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &UserBinding{}
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
