// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/kvstore"
	"gorm.io/gorm"
)

type UserInfo struct {
	ID                 string                  `json:"id,omitempty"`
	Name               string                  `json:"name,omitempty"`
	FullName           string                  `json:"fullName,omitempty"`
	Email              string                  `json:"email,omitempty"`
	Remark             string                  `json:"remark,omitempty"`
	Admin              bool                    `json:"admin,omitempty"`
	Managed            bool                    `json:"managed,omitempty"`
	DefaultOrgnization *Organization           `json:"def_org,omitempty"`
	Organizations      []*Organization         `json:"orgs,omitempty"`
	Role               map[string]model.Role   `json:"role,omitempty"` // org_id -> role
	Perm               map[string]*SessionPerm `json:"perm,omitempty"` // org_id -> perm
}

type Organization struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type SessionInfo struct {
	UserID   uuid.UUID
	UserName string
	RemoteIP string
	Admin    bool
	Perm     map[uuid.UUID]*SessionPerm
}

type SessionPerm struct {
	R model.PolicyObject `json:"r"`
	W model.PolicyObject `json:"w"`
	X model.PolicyObject `json:"x"`
	D model.PolicyObject `json:"d"`
}

func (s SessionInfo) Can(p model.Perm, obj model.PolicyObject, organization uuid.UUID) bool {
	if s.Admin {
		return true
	}
	perm := s.Perm[organization]
	if perm == nil {
		return false
	}
	if p.Has(model.PermReadOnly) && !perm.R.Has(obj) {
		return false
	}
	if p.Has(model.PermWriteOnly) && !perm.W.Has(obj) {
		return false
	}
	if p.Has(model.PermExecOnly) && !perm.X.Has(obj) {
		return false
	}
	if p.Has(model.PermDeleteOnly) && !perm.D.Has(obj) {
		return false
	}
	return true
}

// FilterQueryByOrganization returns a filtered DB session by its associated organization. Be aware of that the
// session may not be correctly cloned when using this. Make sure the session is clear before calling this if
// you want to reuse the session.
func (s SessionInfo) FilterQueryByOrganization(db *gorm.DB, perm model.Perm, obj model.PolicyObject) *gorm.DB {
	return s.FilterQueryByOrganizationWithField(db, perm, obj, "`organization_id`")
}

func (s SessionInfo) FilterQueryByOrganizationWithField(db *gorm.DB, perm model.Perm, obj model.PolicyObject, fieldName string) *gorm.DB {
	if s.Admin {
		return db
	}

	if len(s.Perm) == 0 {
		return db.Where("FALSE")
	}

	var (
		expr   []string
		params []interface{}
	)
	for org := range s.Perm {
		if !s.Can(perm, obj, org) {
			continue
		}
		expr = append(expr, "?")
		params = append(params, org)
	}
	if len(expr) == 0 {
		return db.Where("FALSE")
	}
	return db.Where(fieldName+" IN ("+strings.Join(expr, ", ")+")", params...)
}

type SessionCache struct {
	ActiveUsers kvstore.KVStore

	store kvstore.KVStore
}

func (c *SessionCache) OnMiddleware(uid uuid.UUID, username string) {
	c.ActiveUsers.SetTTL(uid.String(), []byte(username), time.Minute*10)
}

func (s *SessionCache) OnLogin(uid uuid.UUID, username string) error {
	s.ActiveUsers.SetTTL(uid.String(), []byte(username), time.Minute*10)
	return appendAuditLog(mapd.Database(), "login", "User", uid, username, uid, username)
}

func (c *SessionCache) Set(token *oidc.IDToken, userInfo *UserInfo) error {
	if token.AccessTokenHash == "" {
		return nil
	}

	p, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	c.store.SetTTL(token.AccessTokenHash, p, time.Until(token.Expiry))
	return nil
}

func (c *SessionCache) Get(token *oidc.IDToken) (*UserInfo, error) {
	if token.AccessTokenHash == "" {
		return nil, nil
	}

	p := c.store.Get(token.AccessTokenHash)
	if p == nil {
		return nil, nil
	}

	info := &UserInfo{}
	if err := json.Unmarshal(p, info); err != nil {
		return nil, err
	}
	return info, nil
}

func newSessionCache() *SessionCache {
	return &SessionCache{
		ActiveUsers: kvstore.NewMemoryKVStore(),
		store:       kvstore.NewMemoryKVStore(),
	}
}

var (
	RuntimeSessionCache = newSessionCache()
)
