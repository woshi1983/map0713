// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import "strings"

const (
	defNamespace                     = "vmware-system-map"
	defRegistrySecret                = "vmware-system-map-registry"
	defExtensionEndpointDeadlineSecs = 900

	defLabelComponent = "map.tanzu.vmware.com/component"
	defLabelManaged   = "map.tanzu.vmware.com/managed"
	defLabelInternal  = "map.tanzu.vmware.com/internal"
	defLabelRestart   = "map.tanzu.vmware.com/restart"

	defAVIVirtualServiceApplicationProfile = "System-L4-Application"
)

func standardizeString(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "_", "-"), "/", "-"), " ", "-"))
}
