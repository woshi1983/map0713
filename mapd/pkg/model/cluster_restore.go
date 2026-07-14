// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type ClusterRestore struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string    `gorm:"type:varchar(255);index"`
	Namespace      string    `gorm:"type:varchar(255);index"`
	ClusterID      NullUUID  `gorm:"type:varchar(255)"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time `gorm:"type:datetime(6)"`

	Cluster      *Cluster      `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterRestore) TableName() string {
	return "cluster_restore"
}

func ModelClusterRestore(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterRestore{}
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

type ClusterRestoreInstance struct {
	ID           uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	RestoreID    NullUUID  `gorm:"type:varchar(255)"`
	InstanceID   NullUUID  `gorm:"type:varchar(255)"`
	FederationID NullUUID  `gorm:"type:varchar(255)"`
	CreatedAt    time.Time `gorm:"type:datetime(6)"`
	UpdatedAt    time.Time `gorm:"type:datetime(6)"`

	ClusterRestore *ClusterRestore     `gorm:"foreignKey:RestoreID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Instance       *Instance           `gorm:"foreignKey:InstanceID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Federation     *InstanceFederation `gorm:"foreignKey:FederationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterRestoreInstance) TableName() string {
	return "cluster_restore_instance"
}

func ModelClusterRestoreInstance(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterRestoreInstance{}
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

const BackupAnnotationName = "map.tanzu.vmware.com/metadata"

type BackupAnnotation struct {
	Instances []BackupAnnotationInstance `json:"instances,omitempty"`
}

type BackupAnnotationInstance struct {
	Name          string `json:"name,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ReleaseID     string `json:"releaseID,omitempty"`
	FederationID  string `json:"federationID,omitempty"`
	ComponentID   string `json:"componentID,omitempty"`
	ApplicationID string `json:"applicationID,omitempty"`
}
