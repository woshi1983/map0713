// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Instance struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string    `gorm:"type:varchar(128);index"`
	Namespace      string    `gorm:"type:varchar(128);index"`
	CRDs           []byte    `gorm:"type:mediumblob;column:crds"`
	AutoRestart    bool
	PrevManifest   []byte        `gorm:"type:mediumblob"`
	Manifest       []byte        `gorm:"type:mediumblob"`
	State          InstanceState `gorm:"type:varchar(128);index"`
	Status         []byte        `gorm:"type:mediumblob"`
	ClusterID      NullUUID      `gorm:"type:varchar(255)"`
	ReleaseID      NullUUID      `gorm:"type:varchar(255)"`
	FederationID   NullUUID      `gorm:"type:varchar(255)"`
	ComponentID    NullUUID      `gorm:"type:varchar(255)"`
	ApplicationID  NullUUID      `gorm:"type:varchar(255)"`
	OrganizationID NullUUID      `gorm:"type:varchar(255)"`
	CreatedAt      time.Time     `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time     `gorm:"type:datetime(6)"`

	Cluster      *Cluster            `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Release      *Release            `gorm:"foreignKey:ReleaseID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Federation   *InstanceFederation `gorm:"foreignKey:FederationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Component    *Component          `gorm:"foreignKey:ComponentID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Application  *Application        `gorm:"foreignKey:ApplicationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization       `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef Instance) TableName() string {
	return "instance"
}

func ModelInstance(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Instance{}
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

type InstanceStatusTask struct {
	Name      string        `json:"name,omitempty"`
	Namespace string        `json:"namespace,omitempty"`
	Kind      string        `json:"kind,omitempty"`
	State     InstanceState `json:"state,omitempty"`
	Retry     int64         `json:"retry,omitempty"`
	Message   string        `json:"message,omitempty"`
}

type InstanceState string

func (s InstanceState) String() string {
	if s == InstanceStatePending {
		return "PENDING"
	}
	return string(s)
}

const (
	InstanceStatePending      InstanceState = ""
	InstanceStateProvisioning InstanceState = "PROVISIONING"
	InstanceStateSkipped      InstanceState = "SKIPPED" // when it is used on instance itself, it means that instance is pending for a cluster restore.
	InstanceStateSuccess      InstanceState = "SUCCESS"
	InstanceStateFailure      InstanceState = "FAILURE"
	InstanceStateDeleting     InstanceState = "DELETING"
)
