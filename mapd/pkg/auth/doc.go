// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

// JWTClaims defines the claims to be used for auth.
type JWTClaims struct {
	jwt.StandardClaims

	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	AccessTokenHash   string `json:"at_hash,omitempty"`
}

func appendAuditLog(db *gorm.DB, action, objType string, objID uuid.UUID, objName string, userID uuid.UUID, userName string) error {
	l := &model.AuditLog{
		ID:         uuid.New(),
		Action:     action,
		ObjectType: objType,
		ObjectID:   objID,
		ObjectName: objName,
		UserID:     userID,
		UserName:   userName,
	}
	return db.Create(l).Error
}
