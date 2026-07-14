// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/universonic/turret/envutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/submodule"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/logging"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// RootCmd represents the root command.
var RootCmd = &cobra.Command{
	Use:     "mapd",
	Short:   "Daemon Utility of MAP",
	Long:    `Daemon Utility of MAP`,
	Version: mapd.SoftwareVersion(),
	Run: func(command *cobra.Command, args []string) {
		var (
			exitCode int
			module   *mapd.Module
		)
		logger, err := logging.NewLogger(mapd.GlobalConfig.Log.Level, mapd.GlobalConfig.Log.RotateSizeMiB, mapd.GlobalConfig.Log.File)
		if err != nil {
			exitCode = 5
			goto EXIT
		}
		logger = logger.Named("mapd")
		if mapd.GlobalConfig.Log.Level.Level < zapcore.InfoLevel {
			logger.Warn("DEBUG mode is enabled. DO NOT use it in production!")
		}

		if _, err = mapd.InitDatabase(); err != nil {
			exitCode = 2
			goto EXIT
		}
		if err = mapd.Database().Transaction(func(tx *gorm.DB) error {
			return model.Upgrade(tx, logger.Sugar())
		}); err != nil {
			exitCode = 3
			goto EXIT
		}

		module = mapd.NewModule(logger,
			submodule.NewConductor(logger),
			submodule.NewAPI(logger),
		)
		if err = module.Run(); err != nil {
			exitCode = 1
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v\n", exitCode, err)
		}
		os.Exit(exitCode)
		return

	EXIT:
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR(%d): %v\n", exitCode, err)
		}
		os.Exit(exitCode)
	},
}

var (
	help    bool
	version bool
)

func init() {
	flags := RootCmd.PersistentFlags()
	envs := envutil.NewNamespace("MAPD")
	mapd.GlobalConfig.Apply(flags, envs)

	flags.BoolVarP(
		&help, "help", "h", help, "Print this message.",
	)
	flags.BoolVarP(
		&version, "version", "v", version, "Print software version.",
	)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
