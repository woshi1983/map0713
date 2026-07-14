// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
)

func TestIsValidDomainName(t *testing.T) {
	assert.True(t, textutil.IsDomainName("com"))
	assert.True(t, textutil.IsDomainName("vmware.com"))
	assert.True(t, textutil.IsDomainName("www.vmware.com"))
	assert.False(t, textutil.IsDomainName("www.vmware,.com"))
	assert.False(t, textutil.IsDomainName("-vmware.com"))
	assert.True(t, textutil.IsDomainName("my-vmware.com"))
	assert.False(t, textutil.IsDomainName("1.1.1.1"))
	assert.False(t, textutil.IsDomainName("vmware-.com"))
	assert.True(t, textutil.IsDomainName("a.b.c.d.vmware.com"))
	assert.False(t, textutil.IsDomainName("my_host.com"))
	assert.True(t, textutil.IsDomainName("vmware.88"))
	assert.True(t, textutil.IsDomainName("subdomain-hostname.whatever88.vmware.com"))
}

func TestIsIPAddressOrDomainName(t *testing.T) {
	assert.True(t, textutil.IsIPAddressOrDomainName("com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("vmware.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("www.vmware.com"))
	assert.False(t, textutil.IsIPAddressOrDomainName("www.vmware,.com"))
	assert.False(t, textutil.IsIPAddressOrDomainName("-vmware.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("my-vmware.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("1.1.1.1"))
	assert.False(t, textutil.IsIPAddressOrDomainName("vmware-.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("a.b.c.d.vmware.com"))
	assert.False(t, textutil.IsIPAddressOrDomainName("my_host.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("vmware.88"))
	assert.True(t, textutil.IsIPAddressOrDomainName("subdomain-hostname.whatever88.vmware.com"))
	assert.True(t, textutil.IsIPAddressOrDomainName("2001:4860:4860::8888"))
	assert.True(t, textutil.IsIPAddressOrDomainName("2001:4860:4860:0:0:0:0:8888"))
}
