// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package model contains model definitions of SQL database.
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Upgrade exports all defined model to db.
func Upgrade(tx *gorm.DB, log *zap.SugaredLogger) error {
	defer log.Sync()

	var currentVersion *semver.Version
	info := &SystemInfo{Name: SystemInfoVersion}
	if migrator := tx.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator(); migrator.HasTable(info) {
		if err := mapd.RegisterModel(tx, ModelSystemInfo); err != nil {
			return errors.Wrap(err, "error initializing database models")
		}
		err := tx.Take(info).Error
		if err != nil {
			return err
		}
		currentVersion, err = semver.NewVersion(info.Value)
		if err != nil {
			return errors.Wrap(err, "error reading previous version")
		}
		log.Infof("found previous version info: v%v.", currentVersion)
	} else {
		log.Infof("previous version info is not found.")
	}
	err := Upgrader.Upgrade(tx, currentVersion,
		ModelApplication,
		ModelAuditLog,
		ModelClusterBackupScheduleComponent,
		ModelClusterBackupSchedule,
		ModelClusterExtensionRelatedCluster,
		ModelClusterExtension,
		ModelClusterImageRegistry,
		ModelClusterProviderImageRegistry,
		ModelClusterProvider,
		ModelClusterQuotaUsage,
		ModelClusterRestoreInstance,
		ModelClusterRestore,
		ModelCluster,
		ModelInstanceFederation,
		ModelComponent,
		ModelImageRegistry,
		ModelInstance,
		ModelLoadBalancer,
		ModelOrganizationQuotaUsage,
		ModelOrganizationQuota,
		ModelOrganization,
		ModelRelease,
		ModelRoleBinding,
		ModelRouteBackend,
		ModelRoute,
		ModelShortcut,
		ModelStandardPackageRepository,
		ModelSystemInfo,
		ModelUserBinding,
	)
	if err != nil {
		return err
	}
	if err := tx.Save(&SystemInfo{
		Name:  SystemInfoVersion,
		Value: mapd.Version,
	}).Error; err != nil {
		return err
	}
	log.Infof("upgraded database for version %s.", mapd.Version)
	return nil
}

// NullUUID can be used with the standard sql package to represent a
// UUID value that can be NULL in the database
type NullUUID struct {
	UUID  uuid.UUID
	Valid bool
}

func (u NullUUID) String() string {
	if !u.Valid {
		return "<nil>"
	}
	return u.UUID.String()
}

// Value implements the driver.Valuer interface.
func (u NullUUID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	// Delegate to UUID Value function
	return u.UUID.Value()
}

// Scan implements the sql.Scanner interface.
func (u *NullUUID) Scan(src interface{}) error {
	if src == nil {
		u.UUID, u.Valid = uuid.UUID{}, false
		return nil
	}

	// Delegate to UUID Scan function
	u.Valid = true
	return u.UUID.Scan(src)
}

// NewNullUUID creates a new NullUUID with given id.
func NewNullUUID(id uuid.UUID) NullUUID {
	return NullUUID{UUID: id, Valid: true}
}

func IsDeadlockError(err error) bool {
	switch e := err.(type) {
	case *mysql.MySQLError:
		return e.Number == 1205 || e.Number == 1213
	}
	return false
}

type DatabaseUpgrader func(db *gorm.DB, ver *semver.Version) (*semver.Version, error)

type UpgraderRegistry struct {
	mux   sync.Mutex
	init  DatabaseUpgrader
	funcs map[string]DatabaseUpgrader
}

func (r *UpgraderRegistry) Register(ver *semver.Version, upgrader DatabaseUpgrader) *UpgraderRegistry {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.funcs[ver.String()] = upgrader
	return r
}

func (r *UpgraderRegistry) Upgrade(db *gorm.DB, currentVersion *semver.Version, registers ...mapd.RegisterMiddleware) error {
	ver, err := r.init(db, currentVersion)
	if err != nil {
		return err
	}

	if err := mapd.RegisterModel(db, registers...); err != nil {
		return errors.Wrap(err, "error initializing database models")
	}

	latestVersion := semver.MustParse(mapd.Version)
	for ver.LessThan(latestVersion) {
		upgrader, ok := r.funcs[ver.String()]
		if !ok {
			return fmt.Errorf("version %v is not upgradable", ver)
		}
		ver, err = upgrader(db, ver)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewUpgraderRegistry(init DatabaseUpgrader) *UpgraderRegistry {
	return &UpgraderRegistry{
		init:  init,
		funcs: make(map[string]DatabaseUpgrader),
	}
}

const minVersionRequirement = "v0.11.0"

func checkMinimumVersionRequirement(db *gorm.DB, ver *semver.Version) (*semver.Version, error) {
	if ver == nil {
		// Extract just the version part (e.g., "v0.12.0" from "v0.12.0 (revision@branch)")
		fullVer := strings.TrimSpace(mapd.Version)
		if idx := strings.Index(fullVer, " "); idx > 0 {
			fullVer = fullVer[:idx]
		}
		return semver.MustParse(fullVer), nil
	}

	minVer := semver.MustParse(minVersionRequirement)
	if ver.LessThan(minVer) {
		return nil, fmt.Errorf("current version does not meet the minimum version requirement. Please upgrade to version %s first", minVersionRequirement)
	}
	return ver, nil
}

var Upgrader = NewUpgraderRegistry(checkMinimumVersionRequirement)

func init() {
	Upgrader.Register(semver.MustParse("v0.11.0"), upgrader0_11_0To0_11_1)
	Upgrader.Register(semver.MustParse("v0.11.1"), upgrader0_11_1To0_12_0)
}

// upgrader0_11_0To0_11_1 upgrades v0.11.0 to v0.11.1
func upgrader0_11_0To0_11_1(db *gorm.DB, ver *semver.Version) (*semver.Version, error) {
	if err := db.Model(&Cluster{}).Where("`deleted_at` IS NULL").Updates(map[string]interface{}{"insecure_tls": true}).Error; err != nil {
		return nil, err
	}
	var providersToUpgrade []ClusterProvider
	if err := db.Where("`class` = ? AND `deleted_at` IS NULL", ClusterClassTKGS).Find(&providersToUpgrade).Error; err != nil {
		return nil, err
	}
	for _, provider := range providersToUpgrade {
		spec := &ClusterProviderTKGSSpec{}
		if err := json.Unmarshal(provider.Spec, spec); err != nil {
			return nil, errors.Wrap(err, "unable to read provider spec")
		}
		spec.InsecureSkipTLSVerify = true

		b, err := json.Marshal(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to save provider spec")
		}
		if err := db.Model(&ClusterProvider{ID: provider.ID}).Updates(map[string]interface{}{"spec": string(b)}).Error; err != nil {
			return nil, err
		}
	}
	if err := db.Model(&LoadBalancer{}).Where("`deleted_at` IS NULL").Updates(map[string]interface{}{"insecure_tls": true}).Error; err != nil {
		return nil, err
	}
	return semver.MustParse("v0.11.1"), nil
}

// upgrader0_11_1To0_12_0 upgrades v0.11.1 to v0.12.0
func upgrader0_11_1To0_12_0(db *gorm.DB, ver *semver.Version) (*semver.Version, error) {
	return semver.MustParse("v0.12.0"), nil
}
