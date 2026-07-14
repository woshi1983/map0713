// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type ClusterProvider struct {
	ID             uuid.UUID      `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string         `gorm:"type:varchar(128);index"`
	Description    string         `gorm:"type:mediumtext"`
	Class          ClusterClass   `gorm:"type:varchar(128);index"`
	Spec           []byte         `gorm:"type:mediumblob"`
	LoadBalancerID NullUUID       `gorm:"type:varchar(255)"`
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	LoadBalancer *LoadBalancer `gorm:"foreignKey:LoadBalancerID;references:ID;constraint:onDelete:SET NULL,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterProvider) TableName() string {
	return "cluster_provider"
}

func ModelClusterProvider(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterProvider{}
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

type ClusterProviderImageRegistry struct {
	ID              uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Default         bool
	ProviderID      NullUUID `gorm:"type:varchar(255)"`
	ImageRegistryID NullUUID `gorm:"type:varchar(255)"`

	Provider      *ClusterProvider `gorm:"foreignKey:ProviderID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	ImageRegistry *ImageRegistry   `gorm:"foreignKey:ImageRegistryID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterProviderImageRegistry) TableName() string {
	return "cluster_provider_image_registry"
}

func ModelClusterProviderImageRegistry(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterProviderImageRegistry{}
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

type ClusterProviderTKGSSpec struct {
	Endpoint              string `json:"endpoint,omitempty"`
	CACert                []byte `json:"caCert,omitempty"`
	InsecureSkipTLSVerify bool   `json:"insecureSkipTLSVerify,omitempty"`
	AuthUser              string `json:"authUser,omitempty"`
	AuthPassword          []byte `json:"authPassword,omitempty"` // AES-128
	DefaultNS             string `json:"defaultNS,omitempty"`
}

type ClusterProviderTKGMSpec struct {
	KubeConfig []byte            `json:"kubeConfig,omitempty"` // AES-128
	DefaultNS  string            `json:"defaultNS,omitempty"`
	TrustedCAs []*CACert         `json:"trustedCAs,omitempty"`
	DefaultEnv map[string]string `json:"defaultEnv,omitempty"`
}

type ClusterProviderTKG2SSpec struct {
	Endpoint              string    `json:"endpoint,omitempty"`
	CACert                []byte    `json:"caCert,omitempty"`
	InsecureSkipTLSVerify bool      `json:"insecureSkipTLSVerify,omitempty"`
	AuthUser              string    `json:"authUser,omitempty"`
	AuthPassword          []byte    `json:"authPassword,omitempty"` // AES-128
	DefaultNS             string    `json:"defaultNS,omitempty"`
	TrustedCAs            []*CACert `json:"trustedCAs,omitempty"`
}

type ClusterProviderTKG2MSpec struct {
	KubeConfig            []byte    `json:"kubeConfig,omitempty"` // AES-128
	DefaultNS             string    `json:"defaultNS,omitempty"`
	TrustedCAs            []*CACert `json:"trustedCAs,omitempty"`
	VsphereNetwork        string    `json:"vsphereNetwork,omitempty"`
	VsphereResourcePool   string    `json:"vsphereResourcePool,omitempty"`
	VsphereFolder         string    `json:"vsphereFolder,omitempty"`
	VsphereDatastore      string    `json:"vsphereDatastore,omitempty"`
	AviServiceEngineGroup string    `json:"aviServiceEngineGroup,omitempty"`
	AviDataNetwork        string    `json:"aviDataNetwork,omitempty"`
	AviDataNetworkCidr    string    `json:"aviDataNetworkCidr,omitempty"`
}

type CACert struct {
	Name string `json:"name,omitempty"`
	Data []byte `json:"data,omitempty"`
}
