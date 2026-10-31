// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"

	"github.com/dexidp/dex/api/v2"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Dex does not expose the secret of an existing client through this API version.
// Keep the generated secret in MAP's database so all mapd processes use the
// same value across restarts and overlapping starts.
type dexClientSecret struct {
	ID          string `gorm:"type:varchar(255);primaryKey"`
	Secret      string `gorm:"type:varchar(255);not null"`
	Initialized bool   `gorm:"not null"`
}

func (dexClientSecret) TableName() string { return "dex_client_secret" }

func initPersistentDexClient(dex api.DexClient, cfg *oauth2.Config) error {
	// The INSERT contains credential material; never let GORM's SQL logger print it.
	db := mapd.Database().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	if err := db.AutoMigrate(&dexClientSecret{}); err != nil {
		return fmt.Errorf("prepare OIDC client secret storage: %w", err)
	}

	configured := mapd.GlobalConfig.Auth.IDPClientSecret
	candidate := configured
	if candidate == "" {
		candidate = randutil.StringN(64)
	}
	// A unique ID makes concurrent starts choose a single secret. No process
	// uses its candidate until it has read the winning row.
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dexClientSecret{
		ID: cfg.ClientID, Secret: candidate,
	}).Error; err != nil {
		return fmt.Errorf("store OIDC client secret: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var stored dexClientSecret
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", cfg.ClientID).Take(&stored).Error; err != nil {
			return fmt.Errorf("read OIDC client secret: %w", err)
		}
		if configured != "" && configured != stored.Secret {
			return fmt.Errorf("MAPD_IDP_CLIENT_SECRET differs from the stored OIDC client secret")
		}
		cfg.ClientSecret = stored.Secret
		if err := initDexClient(dex, cfg, !stored.Initialized, configured != ""); err != nil {
			return err
		}
		if !stored.Initialized {
			if err := tx.Model(&dexClientSecret{}).Where("id = ?", cfg.ClientID).Update("initialized", true).Error; err != nil {
				return fmt.Errorf("mark OIDC client secret initialized: %w", err)
			}
		}
		return nil
	})
}
