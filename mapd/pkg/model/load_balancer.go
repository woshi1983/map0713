// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type LoadBalancer struct {
	ID             uuid.UUID        `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string           `gorm:"type:varchar(128);index"`
	Description    string           `gorm:"type:mediumtext"`
	Type           LoadBalancerType `gorm:"type:varchar(128)"`
	Spec           []byte           `gorm:"type:mediumblob"`
	Endpoint       string           `gorm:"type:varchar(255)"`
	CACert         []byte           `gorm:"type:mediumblob"`
	InsecureTLS    bool
	AuthUser       string         `gorm:"type:varchar(255)"`
	AuthPassword   []byte         `gorm:"type:blob(65535)"` // AES-128 encrypted with its ID
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef LoadBalancer) TableName() string {
	return "load_balancer"
}

func ModelLoadBalancer(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &LoadBalancer{}
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

type LoadBalancerType string

func (t LoadBalancerType) String() string {
	if t == LoadBalancerTypeUnknown {
		return "UNKNOWN"
	}
	return string(t)
}

const (
	LoadBalancerTypeUnknown  LoadBalancerType = ""
	LoadBalancerTypeAVI      LoadBalancerType = "AVI"
	LoadBalancerTypeExternal LoadBalancerType = "EXTERNAL"
)

type LoadBalancerAVISpec struct {
	SEGroupID        string         `json:"segroup_id,omitempty"`
	LicenseTier      AVILicenseTier `json:"license_tier,omitempty"`
	PreferredNetwork string         `json:"preferred_network,omitempty"`
}

type AVILicenseTier string

const (
	AVILicenseTierUnknown    AVILicenseTier = ""
	AVILicenseTierEssentials AVILicenseTier = "ESSENTIALS"
	AVILicenseTierBasic      AVILicenseTier = "BASIC"
	AVILicenseTierEnterprise AVILicenseTier = "ENTERPRISE"
)
