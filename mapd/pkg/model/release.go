// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Release struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string    `gorm:"type:varchar(128);index"`
	Num            int64     `gorm:"index"`
	CRDs           []byte    `gorm:"type:mediumblob;column:crds"`
	Manifest       []byte    `gorm:"type:mediumblob"`
	ComponentID    NullUUID  `gorm:"type:varchar(255)"`
	ApplicationID  NullUUID  `gorm:"type:varchar(255)"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time `gorm:"type:datetime(6)"`

	Component    *Component    `gorm:"foreignKey:ComponentID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Application  *Application  `gorm:"foreignKey:ApplicationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef Release) TableName() string {
	return "release"
}

func ModelRelease(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Release{}
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
