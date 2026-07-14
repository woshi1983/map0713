// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type SystemInfo struct {
	Name      SystemInfoEntry `gorm:"type:varchar(255);primaryKey"`
	Value     string          `gorm:"type:varchar(255)"`
	UpdatedAt time.Time       `gorm:"type:datetime(6)"`
}

func (typedef SystemInfo) TableName() string {
	return "system_info"
}

func ModelSystemInfo(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &SystemInfo{}
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

type SystemInfoEntry string

const (
	SystemInfoVersion SystemInfoEntry = "version"
)
