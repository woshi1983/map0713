// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	dialTCPTimeout = 5 * time.Second
)

func (c *Client) createContainerView(ctx context.Context, parentID string, viewTypes []string) (*view.ContainerView, error) {
	m := view.NewManager(c.vmomiClient.Client)

	container := &c.vmomiClient.Client.ServiceContent.RootFolder

	if parentID != "" {
		container = &types.ManagedObjectReference{}
		if !container.FromString(parentID) {
			return nil, fmt.Errorf("incorrect managed object reference format for %s", parentID)
		}
	}

	return m.CreateContainerView(ctx, *container, viewTypes, true)
}

// FindMatchingVirtualMachineTemplate finds a virtual machine template that matches the ova versions
func FindMatchingVirtualMachineTemplate(
	vms []*VSphereVirtualMachine,
	ovaVersions []string,
) (matchedTemplateVMs []*VSphereVirtualMachine, nonTemplateVMs []string) {

	for _, vm := range vms {
		for _, ovaVersion := range ovaVersions {
			if vm.OVAVersion == ovaVersion {
				if vm.IsTemplate {
					matchedTemplateVMs = append(matchedTemplateVMs, vm)
				} else {
					nonTemplateVMs = append(nonTemplateVMs, vm.Name)
				}
				break
			}
		}
	}

	return matchedTemplateVMs, nonTemplateVMs
}

// GetVCThumbprint gets the vc thumbprint
func GetVCThumbprint(host string) (string, error) {
	insecure := true

	host, port := splitHostPort(host)

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: dialTCPTimeout,
		},
		"tcp",
		fmt.Sprintf("%s:%s", host, port),
		&tls.Config{InsecureSkipVerify: insecure},
	) // #nosec
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve the vSphere thumbprint")
	}
	defer conn.Close()
	cert := conn.ConnectionState().PeerCertificates[0]

	return soap.ThumbprintSHA1(cert), nil
}

func splitHostPort(host string) (string, string) {
	ix := strings.LastIndex(host, ":")

	if ix < 0 {
		return host, VCDefaultPort
	}

	name := host[:ix]
	port := host[ix+1:]

	return name, port
}
