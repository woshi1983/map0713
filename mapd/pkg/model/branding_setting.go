package model

import (
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

// BrandingSetting contains the public display name and image for this MAP installation.
type BrandingSetting struct {
	ID        uint8     `gorm:"primaryKey"`
	NameZh    string    `gorm:"type:varchar(255);not null"`
	NameEn    string    `gorm:"type:varchar(255);not null"`
	Logo      []byte    `gorm:"type:mediumblob"`
	LogoMIME  string    `gorm:"type:varchar(32)"`
	UpdatedAt time.Time `gorm:"type:datetime(6)"`
}

func (BrandingSetting) TableName() string { return "branding_setting" }

func ModelBrandingSetting(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		if err := db.AutoMigrate(&BrandingSetting{}); err != nil {
			return err
		}
		return next(db)
	}
}
