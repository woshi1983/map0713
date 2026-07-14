// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
)

// RootCmd represents the root command.
var RootCmd = &cobra.Command{
	Use:     "mapkit",
	Short:   "Tool Kit of MAP",
	Long:    `Tool Kit of MAP`,
	Version: fmt.Sprintf("%s (%s@%s), built with %s %s/%s, %s", mapd.Version, mapd.Revision, mapd.Branch, runtime.Version(), runtime.GOOS, runtime.GOARCH, mapd.BuildDate),
}

var (
	help    bool
	version bool
)

func init() {
	RootCmd.PersistentFlags().BoolVarP(
		&help, "help", "h", help, "Print this message.",
	)
	RootCmd.PersistentFlags().BoolVarP(
		&version, "version", "v", version, "Print software version.",
	)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
