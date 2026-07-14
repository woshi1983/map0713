// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package vc

import (
	"encoding/json"
	"net/url"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

type VSphereVirtualMachine struct {
	Name          string
	Moid          string
	OVAVersion    string
	DistroName    string
	DistroVersion string
	DistroArch    string
	IsTemplate    bool
}

type VSphereDatacenter struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere datacenter
func (m *VSphereDatacenter) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereDatacenter) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereDatacenter) UnmarshalBinary(b []byte) error {
	var res VSphereDatacenter
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

type VSphereDatastore struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere datastore
func (m *VSphereDatastore) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereDatastore) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereDatastore) UnmarshalBinary(b []byte) error {
	var res VSphereDatastore
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

type VSphereFolder struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere folder
func (m *VSphereFolder) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereFolder) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereFolder) UnmarshalBinary(b []byte) error {
	var res VSphereFolder
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

type VSphereNetwork struct {

	// display name
	DisplayName string `json:"displayName,omitempty"`

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere network
func (m *VSphereNetwork) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereNetwork) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereNetwork) UnmarshalBinary(b []byte) error {
	var res VSphereNetwork
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

type VSphereResourcePool struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere resource pool
func (m *VSphereResourcePool) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereResourcePool) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereResourcePool) UnmarshalBinary(b []byte) error {
	var res VSphereResourcePool
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// vCenter Managed Object Type Names
const (
	TypeCluster         = "ClusterComputeResource"
	TypeComputeResource = "ComputeResource"
	TypeHostGroup       = "HostGroup"
	TypeResourcePool    = "ResourcePool"
	TypeDatacenter      = "Datacenter"
	TypeFolder          = "Folder"
	TypeDatastore       = "Datastore"
	TypeHost            = "HostSystem"
	TypeNetwork         = "Network"
	TypeDvpg            = "DistributedVirtualPortgroup"
	TypeDvs             = "VmwareDistributedVirtualSwitch"
	TypeOpaqueNetwork   = "OpaqueNetwork"
	TypeVirtualMachine  = "VirtualMachine"
)

//go:generate counterfeiter -o ../fakes/vcclientfactory.go --fake-name VcClientFactory . VcClientFactory

// VcClientFactory a factory for creating VC clients
type VcClientFactory interface {
	NewClient(vcURL *url.URL, thumbprint string, insecure bool) (*Client, error)
}

type vcClientFactory struct{}

// NewClient creates new clusterclient
func (c *vcClientFactory) NewClient(vcURL *url.URL, thumbprint string, insecure bool) (*Client, error) { //nolint:gocritic
	return NewClient(vcURL, thumbprint, insecure)
}

// NewVcClientFactory creates new vcclient factory
func NewVcClientFactory() VcClientFactory {
	return &vcClientFactory{}
}

// VSphereManagementObject v sphere management object
// swagger:model VSphereManagementObject
type VSphereManagementObject struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// parent moid
	ParentMoid string `json:"parentMoid,omitempty"`

	// path
	Path string `json:"path,omitempty"`

	// resource type
	// Enum: [datacenter cluster hostgroup folder respool vm datastore host network]
	ResourceType string `json:"resourceType,omitempty"`
}

// Validate validates this v sphere management object
func (m *VSphereManagementObject) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateResourceType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var vSphereManagementObjectTypeResourceTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["datacenter","cluster","hostgroup","folder","respool","vm","datastore","host","network"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		vSphereManagementObjectTypeResourceTypePropEnum = append(vSphereManagementObjectTypeResourceTypePropEnum, v)
	}
}

const (

	// VSphereManagementObjectResourceTypeDatacenter captures enum value "datacenter"
	VSphereManagementObjectResourceTypeDatacenter string = "datacenter"

	// VSphereManagementObjectResourceTypeCluster captures enum value "cluster"
	VSphereManagementObjectResourceTypeCluster string = "cluster"

	// VSphereManagementObjectResourceTypeHostgroup captures enum value "hostgroup"
	VSphereManagementObjectResourceTypeHostgroup string = "hostgroup"

	// VSphereManagementObjectResourceTypeFolder captures enum value "folder"
	VSphereManagementObjectResourceTypeFolder string = "folder"

	// VSphereManagementObjectResourceTypeRespool captures enum value "respool"
	VSphereManagementObjectResourceTypeRespool string = "respool"

	// VSphereManagementObjectResourceTypeVM captures enum value "vm"
	VSphereManagementObjectResourceTypeVM string = "vm"

	// VSphereManagementObjectResourceTypeDatastore captures enum value "datastore"
	VSphereManagementObjectResourceTypeDatastore string = "datastore"

	// VSphereManagementObjectResourceTypeHost captures enum value "host"
	VSphereManagementObjectResourceTypeHost string = "host"

	// VSphereManagementObjectResourceTypeNetwork captures enum value "network"
	VSphereManagementObjectResourceTypeNetwork string = "network"
)

// prop value enum
func (m *VSphereManagementObject) validateResourceTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, vSphereManagementObjectTypeResourceTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *VSphereManagementObject) validateResourceType(formats strfmt.Registry) error {

	if swag.IsZero(m.ResourceType) { // not required
		return nil
	}

	// value enum
	if err := m.validateResourceTypeEnum("resourceType", "body", m.ResourceType); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *VSphereManagementObject) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereManagementObject) UnmarshalBinary(b []byte) error {
	var res VSphereManagementObject
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
