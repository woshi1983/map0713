// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Policy indicates a permission expression.
//
// For example:
//
// - `full:all`: full control of all resources.
// - `read:application`: readonly on `application`.
// - `write:component`: readwrite on `component`.
// - `write:application:1`: readwrite with priority 1 on `application`.
// - `modify:application:-1`: modify with priority 1 on `application`, but reverse the permission.
// - `rwxd:all`: equivalent to `full:all`.
type Policy struct {
	Perm     Perm
	Object   PolicyObject
	Priority int16
}

func (p *Policy) DeepCopy() *Policy {
	return &Policy{
		Perm:     p.Perm,
		Object:   p.Object,
		Priority: p.Priority,
	}
}

func (p *Policy) IsReversed() bool {
	return p.Priority < 0
}

func (p *Policy) AbsPriority() int16 {
	return int16(math.Abs(float64(p.Priority)))
}

func (p *Policy) String() string {
	var buf bytes.Buffer
	buf.WriteString(p.Perm.String())
	buf.WriteRune(':')
	buf.WriteString(p.Object.String())
	if p.Priority != 0 {
		buf.WriteRune(':')
		buf.WriteString(strconv.FormatInt(int64(p.Priority), 10))
	}
	return buf.String()
}

func ParsePolicy(s string) (*Policy, error) {
	ss := strings.Split(s, ":")
	if len(ss) < 2 {
		return nil, fmt.Errorf("invalid policy: %s", s)
	}
	perm, err := ParsePerm(ss[0])
	if err != nil {
		return nil, errors.Wrap(err, "invalid policy: error parsing perm")
	}
	obj, err := ParsePolicyObject(ss[1])
	if err != nil {
		return nil, errors.Wrap(err, "invalid policy: error parsing object")
	}
	var priority int64
	if len(ss) > 2 {
		priority, err = strconv.ParseInt(ss[2], 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "invalid policy: error parsing priority")
		}
	}
	return &Policy{
		Perm:     perm,
		Object:   obj,
		Priority: int16(priority),
	}, nil
}

type PolicyObject uint32

func (s PolicyObject) String() string {
	switch s {
	case PolicyObjectAll:
		return "all"
	case PolicyObjectApplication:
		return "application"
	case PolicyObjectAuditLog:
		return "audit_log"
	case PolicyObjectCluster:
		return "cluster"
	case PolicyObjectClusterProvider:
		return "cluster_provider"
	case PolicyObjectClusterResource:
		return "cluster_resource"
	case PolicyObjectComponent:
		return "component"
	case PolicyObjectImageRegistry:
		return "image_registry"
	case PolicyObjectInstance:
		return "instance"
	case PolicyObjectLoadBalancer:
		return "load_balancer"
	case PolicyObjectOrganization:
		return "organization"
	case PolicyObjectRoute:
		return "route"
	case PolicyObjectShortcut:
		return "shortcut"
	case PolicyObjectUser:
		return "user"
	}
	return "none"
}

func (s *PolicyObject) FromString(str string) error {
	switch strings.ToLower(str) {
	case "all":
		*s = PolicyObjectAll
	case "app", "application":
		*s = PolicyObjectApplication
	case "audit_log":
		*s = PolicyObjectAuditLog
	case "cluster":
		*s = PolicyObjectCluster
	case "cluster_provider":
		*s = PolicyObjectClusterProvider
	case "cluster_resource":
		*s = PolicyObjectClusterResource
	case "component":
		*s = PolicyObjectComponent
	case "image_registry":
		*s = PolicyObjectImageRegistry
	case "instance":
		*s = PolicyObjectInstance
	case "lb", "load_balancer":
		*s = PolicyObjectLoadBalancer
	case "org", "organization":
		*s = PolicyObjectOrganization
	case "route":
		*s = PolicyObjectRoute
	case "shortcut":
		*s = PolicyObjectShortcut
	case "user":
		*s = PolicyObjectUser
	default:
		*s = PolicyObjectNone
	}
	return nil
}

func (s PolicyObject) Atomic() []PolicyObject {
	var ss []PolicyObject
	for s&PolicyObjectAll != 0x00 {
		switch {
		case s&PolicyObjectAll == PolicyObjectAll:
			s ^= PolicyObjectAll
			ss = append(ss, PolicyObjectAll)
		case s&PolicyObjectApplication == PolicyObjectApplication:
			s ^= PolicyObjectApplication
			ss = append(ss, PolicyObjectApplication)
		case s&PolicyObjectAuditLog == PolicyObjectAuditLog:
			s ^= PolicyObjectAuditLog
			ss = append(ss, PolicyObjectAuditLog)
		case s&PolicyObjectCluster == PolicyObjectCluster:
			s ^= PolicyObjectCluster
			ss = append(ss, PolicyObjectCluster)
		case s&PolicyObjectClusterProvider == PolicyObjectClusterProvider:
			s ^= PolicyObjectClusterProvider
			ss = append(ss, PolicyObjectClusterProvider)
		case s&PolicyObjectClusterResource == PolicyObjectClusterResource:
			s ^= PolicyObjectClusterResource
			ss = append(ss, PolicyObjectClusterResource)
		case s&PolicyObjectComponent == PolicyObjectComponent:
			s ^= PolicyObjectComponent
			ss = append(ss, PolicyObjectComponent)
		case s&PolicyObjectImageRegistry == PolicyObjectImageRegistry:
			s ^= PolicyObjectImageRegistry
			ss = append(ss, PolicyObjectImageRegistry)
		case s&PolicyObjectInstance == PolicyObjectInstance:
			s ^= PolicyObjectInstance
			ss = append(ss, PolicyObjectInstance)
		case s&PolicyObjectLoadBalancer == PolicyObjectLoadBalancer:
			s ^= PolicyObjectLoadBalancer
			ss = append(ss, PolicyObjectLoadBalancer)
		case s&PolicyObjectOrganization == PolicyObjectOrganization:
			s ^= PolicyObjectOrganization
			ss = append(ss, PolicyObjectOrganization)
		case s&PolicyObjectRoute == PolicyObjectRoute:
			s ^= PolicyObjectRoute
			ss = append(ss, PolicyObjectRoute)
		case s&PolicyObjectShortcut == PolicyObjectShortcut:
			s ^= PolicyObjectShortcut
			ss = append(ss, PolicyObjectShortcut)
		case s&PolicyObjectUser == PolicyObjectUser:
			s ^= PolicyObjectUser
			ss = append(ss, PolicyObjectUser)
		}
	}
	return ss
}

func (s PolicyObject) Has(p PolicyObject) bool {
	return s&p == p
}

func (s *PolicyObject) Add(p PolicyObject) {
	*s |= p
}

func (s *PolicyObject) Toggle(p PolicyObject) {
	*s ^= p
}

func (s *PolicyObject) Remove(p PolicyObject) {
	*s &= *s&p ^ PolicyObjectAll
}

func ParsePolicyObject(s string) (PolicyObject, error) {
	var o PolicyObject
	if err := o.FromString(s); err != nil {
		return o, err
	}
	return o, nil
}

const (
	PolicyObjectNone        PolicyObject = 0x00
	PolicyObjectApplication PolicyObject = 0x01 << (31 - iota)
	PolicyObjectAuditLog
	PolicyObjectCluster
	PolicyObjectClusterProvider
	PolicyObjectClusterResource
	PolicyObjectComponent
	PolicyObjectEnvironment
	PolicyObjectImageRegistry
	PolicyObjectInstance
	PolicyObjectLoadBalancer
	PolicyObjectOrganization
	PolicyObjectRoute
	PolicyObjectShortcut
	PolicyObjectUser

	PolicyObjectAll PolicyObject = math.MaxUint32
)

type Perm byte

func (u Perm) String() string {
	var buf bytes.Buffer
	switch u {
	case PermNoOp:
		buf.WriteString("noop")
	case PermReadOnly:
		buf.WriteString("read")
	case PermReadWrite:
		buf.WriteString("write")
	case PermReadExec:
		buf.WriteString("exec")
	case PermModify:
		buf.WriteString("modify")
	case PermFullControl:
		buf.WriteString("full")
	default:
		if u.Has(PermReadOnly) {
			buf.WriteRune('r')
		} else {
			buf.WriteRune('-')
		}
		if u.Has(PermWriteOnly) {
			buf.WriteRune('w')
		} else {
			buf.WriteRune('-')
		}
		if u.Has(PermExecOnly) {
			buf.WriteRune('x')
		} else {
			buf.WriteRune('-')
		}
		if u.Has(PermDeleteOnly) {
			buf.WriteRune('d')
		} else {
			buf.WriteRune('-')
		}
	}
	return buf.String()
}

func (u *Perm) FromString(s string) error {
	switch strings.ToLower(s) {
	case "", "empty", "none", "noop", "null":
		*u = PermNoOp
	case "read", "readonly":
		*u = PermReadOnly
	case "write", "readwrite":
		*u = PermReadWrite
	case "exec", "readexec":
		*u = PermReadExec
	case "modify":
		*u = PermModify
	case "full", "fullcontrol":
		*u = PermFullControl
	default:
		if len(s) < 4 {
			return fmt.Errorf("invalid permission")
		}
		var v Perm
		if s[0] == 'r' {
			v.Add(PermReadOnly)
		}
		if s[1] == 'w' {
			v.Add(PermWriteOnly)
		}
		if s[2] == 'x' {
			v.Add(PermExecOnly)
		}
		if s[3] == 'd' {
			v.Add(PermDeleteOnly)
		}
		*u = v
	}
	return nil
}

func (u Perm) Has(p Perm) bool {
	return u&p == p
}

func (u *Perm) Add(p Perm) {
	*u |= p
}

func (u *Perm) Toggle(p Perm) {
	*u ^= p
}

func (u *Perm) Remove(p Perm) {
	*u &= *u&p ^ 0xFF
}

func ParsePerm(s string) (Perm, error) {
	var p Perm
	if err := p.FromString(s); err != nil {
		return p, err
	}
	return p, nil
}

const (
	PermNoOp     Perm = 0x00
	PermReadOnly Perm = 0x01 << (7 - iota)
	PermWriteOnly
	PermExecOnly
	PermDeleteOnly

	PermFullControl = PermReadOnly | PermWriteOnly | PermExecOnly | PermDeleteOnly
	PermModify      = PermReadOnly | PermWriteOnly | PermDeleteOnly
	PermReadWrite   = PermReadOnly | PermWriteOnly
	PermReadExec    = PermReadOnly | PermExecOnly
)
