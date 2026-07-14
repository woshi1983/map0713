// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Component struct {
	ID             uuid.UUID      `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string         `gorm:"type:varchar(128);index"`
	Description    string         `gorm:"type:mediumtext"`
	Type           ComponentType  `gorm:"type:varchar(128);index"`
	Avatar         string         `gorm:"type:mediumtext"`
	Template       []byte         `gorm:"type:mediumblob"`
	ApplicationID  NullUUID       `gorm:"type:varchar(255)"`
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	Application  *Application  `gorm:"foreignKey:ApplicationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef Component) TableName() string {
	return "component"
}

func ModelComponent(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Component{}
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

type ComponentType string

func (s ComponentType) String() string {
	if s == ComponentTypeUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ComponentTypeUnknown ComponentType = ""
	ComponentTypeBasic   ComponentType = "BASIC"
	ComponentTypeHelm    ComponentType = "HELM"
	ComponentTypeRaw     ComponentType = "RAW"
)

type ComponentHelmTemplate struct {
	URL       string            `json:"url,omitempty"`
	Values    json.RawMessage   `json:"values,omitempty"`
	ExtraInfo map[string]string `json:"extraInfo,omitempty"`
}
