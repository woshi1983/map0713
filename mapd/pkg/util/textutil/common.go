// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil

import (
	"net"

	"golang.org/x/net/idna"
)

// IsDomainName returns true if str is a valid IDNA string.
func IsDomainName(str string) bool {
	_, err := idna.Lookup.ToASCII(str)
	return err == nil && net.ParseIP(str) == nil
}

// IsIPAddressOrDomainName returns true if str is an IP address or IDNA string.
func IsIPAddressOrDomainName(str string) bool {
	_, err := idna.Lookup.ToASCII(str)
	return net.ParseIP(str) != nil || err == nil
}
