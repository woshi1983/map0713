package model

import (
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

// VKSSupervisorLicense is shared by all VKS providers pointing to one Supervisor.
type VKSSupervisorLicense struct {
	SupervisorUID string `gorm:"type:varchar(128);primaryKey"`
	Token         string `gorm:"type:text;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func ModelVKSSupervisorLicense(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		if err := db.AutoMigrate(&VKSSupervisorLicense{}); err != nil {
			return err
		}
		return next(db)
	}
}
