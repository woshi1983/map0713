// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ChartCmd represents the chart command.
var ChartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Helm chart utilities for transfering Helm charts between repositories.",
	Long:  `Helm chart utilities for transfering Helm charts between repositories.`,
}

func init() {
	wd, err := os.UserHomeDir()
	if err != nil {
		wd, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}
	downloadChartPath = filepath.Join(wd, ".mapkit", "charts")
	uploadChartPath = filepath.Join(wd, ".mapkit", "charts")

	RootCmd.AddCommand(ChartCmd)
	ChartCmd.AddCommand(DownloadChartCmd)
	ChartCmd.AddCommand(UploadChartCmd)

	DownloadChartCmd.Flags().StringVarP(&downloadChartPath, "output", "o", downloadChartPath, "Directory to save the downloaded charts.")
	downloadChart.SetFlags(DownloadChartCmd.Flags())

	UploadChartCmd.Flags().StringVarP(&uploadChartPath, "input", "i", uploadChartPath, "Directory of the charts to upload.")
	uploadChart.SetFlags(UploadChartCmd.Flags())
}
