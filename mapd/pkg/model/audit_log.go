// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Action         string    `gorm:"type:varchar(128);index"`
	ObjectType     string    `gorm:"type:varchar(255);index"`
	ObjectID       uuid.UUID `gorm:"type:varchar(255);index"`
	ObjectName     string    `gorm:"type:varchar(255);index"`
	UserID         uuid.UUID `gorm:"type:varchar(255);index"`
	UserName       string    `gorm:"type:varchar(255);index"`
	OrganizationID NullUUID  `gorm:"type:varchar(255);index"`
	CreatedAt      time.Time `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time `gorm:"type:datetime(6)"`
}

func (typedef AuditLog) TableName() string {
	return "audit_log"
}

func ModelAuditLog(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &AuditLog{}
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
