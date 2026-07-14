// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type InstanceFederation struct {
	ID             uuid.UUID        `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string           `gorm:"type:varchar(128);index"`
	Description    string           `gorm:"type:mediumtext"`
	Status         FederationStatus `gorm:"type:varchar(128)"`
	ComponentID    NullUUID         `gorm:"type:varchar(255)"`
	ApplicationID  NullUUID         `gorm:"type:varchar(255)"`
	OrganizationID NullUUID         `gorm:"type:varchar(255)"`
	CreatedAt      time.Time        `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time        `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt   `gorm:"type:datetime(6);index"`

	Component    *Component    `gorm:"foreignKey:ComponentID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Application  *Application  `gorm:"foreignKey:ApplicationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef InstanceFederation) TableName() string {
	return "instance_federation"
}

func ModelInstanceFederation(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &InstanceFederation{}
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

type FederationStatus string

func (s FederationStatus) String() string {
	if s == FederationStatusPending {
		return "PENDING"
	}
	return string(s)
}

const (
	FederationStatusPending      FederationStatus = ""
	FederationStatusProvisioning FederationStatus = "PROVISIONING"
	FederationStatusProvisioned  FederationStatus = "PROVISIONED"
	FederationStatusDeleting     FederationStatus = "DELETING"
)
