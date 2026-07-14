// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type ClusterExtension struct {
	ID             uuid.UUID              `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           ClusterExtensionName   `gorm:"type:varchar(128);index"`
	Endpoints      []byte                 `gorm:"type:blob"`
	Spec           []byte                 `gorm:"type:mediumblob"`
	State          ClusterExtensionState  `gorm:"type:varchar(128);index"`
	Status         ClusterExtensionStatus `gorm:"type:varchar(128);index"`
	ClusterID      NullUUID               `gorm:"type:varchar(255)"`
	OrganizationID NullUUID               `gorm:"type:varchar(255)"`
	CreatedAt      time.Time              `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time              `gorm:"type:datetime(6)"`

	Cluster      *Cluster      `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Related      []Cluster     `gorm:"-"`
}

func (typedef ClusterExtension) TableName() string {
	return "cluster_extension"
}

func ModelClusterExtension(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterExtension{}
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

type ClusterExtensionName string

func (s ClusterExtensionName) String() string {
	if s == ClusterExtensionNameUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ClusterExtensionNameUnknown       ClusterExtensionName = ""
	ClusterExtensionNameTBS           ClusterExtensionName = "TBS"
	ClusterExtensionNameOctant        ClusterExtensionName = "OCTANT"
	ClusterExtensionNameContour       ClusterExtensionName = "CONTOUR"
	ClusterExtensionNameElasticSearch ClusterExtensionName = "ELASTICSEARCH"
	ClusterExtensionNameFluentBit     ClusterExtensionName = "FLUENTBIT"
	ClusterExtensionNameGrafana       ClusterExtensionName = "GRAFANA"
	ClusterExtensionNamePrometheus    ClusterExtensionName = "PROMETHEUS"
	ClusterExtensionNameWebTerminal   ClusterExtensionName = "WEB_TERMINAL"
	ClusterExtensionNameNFSDriverNFS  ClusterExtensionName = "CSI_DRIVER_NFS"
)

type ClusterExtensionState string

func (s ClusterExtensionState) String() string {
	if s == ClusterExtensionStateUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ClusterExtensionStateUnknown   ClusterExtensionState = ""
	ClusterExtensionStateEnabling  ClusterExtensionState = "ENABLING"
	ClusterExtensionStateEnabled   ClusterExtensionState = "ENABLED"
	ClusterExtensionStateDisabling ClusterExtensionState = "DISABLING"
)

type ClusterExtensionStatus string

func (s ClusterExtensionStatus) String() string {
	if s == ClusterExtensionStatusPending {
		return "PENDING"
	}
	return string(s)
}

const (
	ClusterExtensionStatusPending      ClusterExtensionStatus = ""
	ClusterExtensionStatusProvisioning ClusterExtensionStatus = "PROVISIONING"
	ClusterExtensionStatusProvisioned  ClusterExtensionStatus = "PROVISIONED"
	ClusterExtensionStatusFailed       ClusterExtensionStatus = "FAILED"
)

type ClusterExtensionRelatedCluster struct {
	ID          uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	ExtensionID NullUUID  `gorm:"type:varchar(255)"`
	ClusterID   NullUUID  `gorm:"type:varchar(255)"`
	GuestID     NullUUID  `gorm:"type:varchar(255)"`

	ClusterExtension *ClusterExtension `gorm:"foreignKey:ExtensionID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Cluster          *Cluster          `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Guest            *Cluster          `gorm:"foreignKey:GuestID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterExtensionRelatedCluster) TableName() string {
	return "cluster_extension_related_cluster"
}

func ModelClusterExtensionRelatedCluster(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterExtensionRelatedCluster{}
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
