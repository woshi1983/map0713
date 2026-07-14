// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil

import "github.com/Masterminds/semver/v3"

// CompareMajorMinorPatchVersion returns true if major/minor/patch parts of the versions match, else false
func CompareMajorMinorPatchVersion(version1, version2 string) bool {
	semVersion1, err := semver.NewVersion(version1)
	if err != nil {
		return false
	}
	semVersion2, err := semver.NewVersion(version2)
	if err != nil {
		return false
	}
	if semVersion1.Major() == semVersion2.Major() &&
		semVersion1.Minor() == semVersion2.Minor() &&
		semVersion1.Patch() == semVersion2.Patch() {
		return true
	}
	return false
}
