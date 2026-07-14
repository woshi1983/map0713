// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package mapd defines global variables or functions that are used by the
// daemon utility.
package mapd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// The following variables' value shall not be modified unless during
	// compilation

	// Version is part of the build information.
	Version string
	// Branch is part of the build information.
	Branch string
	// Revision is part of the build information.
	Revision string
	// BuildDate is part of the build information.
	BuildDate string
)

var (
	// SelfPath is the path of binary of this program.
	SelfPath string
	// CurrentDir presents the current working directory.
	CurrentDir string
)

var (
	PublicImageRegistry string
)

func SoftwareVersion() string {
	return fmt.Sprintf("%s (%s@%s), built with %s for %s/%s on %s", Version, Revision, Branch, runtime.Version(), runtime.GOOS, runtime.GOARCH, BuildDate)
}

func init() {
	dir, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	SelfPath = dir

	CurrentDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}
