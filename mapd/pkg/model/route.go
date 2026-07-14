// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Route struct {
	ID             uuid.UUID        `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string           `gorm:"type:varchar(128);index"`
	Description    string           `gorm:"type:mediumtext"`
	Type           LoadBalancerType `gorm:"type:varchar(128)"`
	Protocol       RouteProtocol    `gorm:"type:varchar(128)"`
	IP             string           `gorm:"type:varchar(255)"`
	Port           int64
	Spec           []byte `gorm:"type:mediumblob"`
	ProvisionPhase RouteProvision
	Provisioned    RouteProvision
	ComponentID    NullUUID       `gorm:"type:varchar(255)"`
	ApplicationID  NullUUID       `gorm:"type:varchar(255)"`
	LoadBalancerID NullUUID       `gorm:"type:varchar(255)"`
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	Component    *Component    `gorm:"foreignKey:ComponentID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Application  *Application  `gorm:"foreignKey:ApplicationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	LoadBalancer *LoadBalancer `gorm:"foreignKey:LoadBalancerID;references:ID;constraint:onDelete:SET NULL,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef Route) TableName() string {
	return "route"
}

func ModelRoute(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Route{}
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

type RouteProtocol string

func (t RouteProtocol) String() string {
	if t == RouteProtocolUnknown {
		return "UNKNOWN"
	}
	return string(t)
}

const (
	RouteProtocolUnknown RouteProtocol = ""
	RouteProtocolTCP     RouteProtocol = "TCP"
	RouteProtocolUDP     RouteProtocol = "UDP"
)

type RouteProvision int64

const (
	RouteProvisionDeleting RouteProvision = iota - 1
	RouteProvisionNone
	RouteProvisionPrepare
	RouteProvisionIP
	RouteProvisionRoute
	RouteProvisionFinish
)

type RouteAVISpec struct {
	PoolID                          string `json:"pool_id,omitempty"`
	VSVIPID                         string `json:"vs_vip_id,omitempty"`
	ApplicationProfileID            string `json:"application_profile_id,omitempty"`
	NetworkProfileID                string `json:"network_profile_id,omitempty"`
	VSID                            string `json:"vs_id,omitempty"`
	LBAlgorithm                     string `json:"lb_algorithm,omitempty"`
	LBAlgorithmConsistentHashHeader string `json:"lb_algorithm_consistent_hash_hdr,omitempty"`
	LBAlgorithmCoreNonaffinity      int64  `json:"lb_algorithm_core_nonaffinity,omitempty"`
	LBAlgorithmHash                 string `json:"lb_algorithm_hash,omitempty"`
	ManualAllocatedVSVIP            bool   `json:"manualAllocatedVsVIP,omitempty"`
	PersistentVSVIP                 bool   `json:"persistentVsVIP,omitempty"`
}

type RouteBackend struct {
	ID         uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name       string    `gorm:"type:varchar(255);index"`
	Namespace  string    `gorm:"type:varchar(255);index"`
	Type       string    `gorm:"type:varchar(128)"`
	IP         string    `gorm:"type:varchar(255)"`
	Port       int64
	Ratio      int64
	InstanceID NullUUID `gorm:"type:varchar(255)"`
	RouteID    NullUUID `gorm:"type:varchar(255)"`

	Instance *Instance `gorm:"foreignKey:InstanceID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Route    *Route    `gorm:"foreignKey:RouteID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef RouteBackend) TableName() string {
	return "route_backend"
}

func ModelRouteBackend(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &RouteBackend{}
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
