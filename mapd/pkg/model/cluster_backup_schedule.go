// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type ClusterBackupSchedule struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string    `gorm:"type:varchar(255);index"`
	Namespace      string    `gorm:"type:varchar(255);index"`
	Spec           []byte    `gorm:"type:mediumblob"`
	Location       string    `gorm:"type:varchar(255);index"`
	Schedule       string    `gorm:"type:varchar(128)"`
	Fired          int64
	LastFire       time.Time `gorm:"type:datetime(6)"`
	NextFire       time.Time `gorm:"type:datetime(6)"`
	ClusterID      NullUUID  `gorm:"type:varchar(255)"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time `gorm:"type:datetime(6)"`

	Cluster      *Cluster      `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterBackupSchedule) TableName() string {
	return "cluster_backup_schedule"
}

func ModelClusterBackupSchedule(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterBackupSchedule{}
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

type ClusterBackupScheduleComponent struct {
	ID          uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	ComponentID NullUUID  `gorm:"type:varchar(255)"`
	ScheduleID  NullUUID  `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"type:datetime(6)"`
	UpdatedAt   time.Time `gorm:"type:datetime(6)"`

	Component *Component             `gorm:"foreignKey:ComponentID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Schedule  *ClusterBackupSchedule `gorm:"foreignKey:ScheduleID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterBackupScheduleComponent) TableName() string {
	return "cluster_backup_schedule_component"
}

func ModelClusterBackupScheduleComponent(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterBackupScheduleComponent{}
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

type ClusterBackupScheduleSpec struct {
	UseRestic          bool     `json:"useRestic"`
	IncludedNamespaces []string `json:"includedNamespaces"`
	// the following namespaces will be automatically excluded by default:
	//
	//   - `avi-system`
	//
	//   - `build-service`
	//
	//   - `cert-manager`
	//
	//   - `default`
	//
	//   - `kpack`
	//
	//   - `kube-node-lease`
	//
	//   - `kube-public`
	//
	//   - `kube-system`
	//
	//   - `stacks-operator-system`
	//
	//   - `tanzu-package-repo-global`
	//
	//   - `tanzu-system-dashboards`
	//
	//   - `tanzu-system-ingress`
	//
	//   - `tanzu-system-logging`
	//
	//   - `tanzu-system-monitoring`
	//
	//   - `tkg-system`
	//
	//   - `vmware-system-auth`
	//
	//   - `vmware-system-cloud-provider`
	//
	//   - `vmware-system-csi`
	//
	//   - `vmware-system-elasic`
	//
	//   - `vmware-system-map`
	//
	//   - `vmware-system-octant`
	ExcludedNamespaces []string      `json:"excludedNamespaces"`
	TTL                time.Duration `json:"ttl"`
}
