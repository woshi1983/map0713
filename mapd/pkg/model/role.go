// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Role string

func (r Role) String() string {
	normalized := r.Normalize()
	if normalized == RoleUnknown {
		return "UNKNOWN"
	}
	return string(normalized)
}

func (r Role) Normalize() Role {
	switch strings.ToUpper(strings.TrimSpace(string(r))) {
	case string(RoleGuest):
		return RoleGuest
	case string(RoleUser):
		return RoleUser
	case string(RoleAdmin):
		return RoleAdmin
	default:
		return RoleUnknown
	}
}

const (
	RoleUnknown Role = ""
	RoleGuest   Role = "GUEST"
	RoleUser    Role = "USER"
	RoleAdmin   Role = "ADMIN"
)

type RoleBinding struct {
	ID             uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Role           Role      `gorm:"type:varchar(128)"`
	UserBindingID  NullUUID  `gorm:"type:varchar(255)"`
	OrganizationID NullUUID  `gorm:"type:varchar(255)"`

	UserBinding  *UserBinding  `gorm:"foreignKey:UserBindingID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef RoleBinding) TableName() string {
	return "role_binding"
}

func ModelRoleBinding(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &RoleBinding{}
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

var (
	RoleBindingPolicyAdminRawBase64 string
	RoleBindingPolicyUserRawBase64  string
	RoleBindingPolicyGuestRawBase64 string
	SuperAdminPolicyRaw             = "full:all"

	roleBindingPolicyMap = make(map[Role][]*Policy)
	superAdminPolicySet  []*Policy
)

func init() {
	registerBase64Policies(roleBindingPolicyMap, RoleAdmin, RoleBindingPolicyAdminRawBase64)
	registerBase64Policies(roleBindingPolicyMap, RoleUser, RoleBindingPolicyUserRawBase64)
	registerBase64Policies(roleBindingPolicyMap, RoleGuest, RoleBindingPolicyGuestRawBase64)

	p, err := ParsePolicy(SuperAdminPolicyRaw)
	if err != nil {
		panic(err)
	}
	superAdminPolicySet = []*Policy{p}
}

func registerBase64Policies(policyMap map[Role][]*Policy, role Role, rawStr string) {
	raw, err := base64.StdEncoding.DecodeString(rawStr)
	if err != nil {
		panic(err)
	}
	policies, err := readPolicyFromReader(bytes.NewReader([]byte(raw)))
	if err != nil {
		panic(err)
	}
	policyMap[role] = policies
}

func readPolicyFromReader(r io.Reader) ([]*Policy, error) {
	reader := bufio.NewReader(r)
	var (
		p, b     []byte
		isPrefix bool
		policies []*Policy
		err      error
	)
	for {
		b, isPrefix, err = reader.ReadLine()
		p = append(p, b...)
		if isPrefix {
			continue
		}
		p = bytes.TrimSpace(p)
		if len(p) > 0 {
			policy, e := ParsePolicy(string(p))
			if e != nil {
				return policies, e
			}
			policies = append(policies, policy)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return policies, err
		}
		p = p[:0]
	}
	return policies, nil
}

func GetRoleBindingPolicies(role Role, isSuperAdmin bool) []*Policy {
	var policies []*Policy
	if isSuperAdmin {
		for _, p := range superAdminPolicySet {
			policies = append(policies, p.DeepCopy())
		}
		return policies
	}
	role = role.Normalize()
	ps, ok := roleBindingPolicyMap[role]
	if !ok {
		return policies
	}
	for _, p := range ps {
		policies = append(policies, p.DeepCopy())
	}
	return policies
}
