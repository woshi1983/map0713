// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config command.
var ConfigCmd = &cobra.Command{
	Use:    "config",
	Short:  "Config utilities for YAML config file.",
	Long:   "Config utilities for YAML config file.",
	Hidden: true,
}

var (
	configPath string
)

func init() {
	RootCmd.AddCommand(ConfigCmd)

	ConfigCmd.PersistentFlags().StringVarP(&configPath, "config-file", "f", "", "Path of the config file.")
}
