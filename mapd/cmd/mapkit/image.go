// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/spf13/cobra"

// ImageCmd represents the config command.
var ImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image utilities for uploading/download images.",
	Long:  "Image utilities for uploading/download images.",
}

var (
	tarFilePath                        string
	registryCACertPaths                []string
	registryVerifyCerts                = true
	registryInsecure                   bool
	registryUsername, registryPassword string
	registryToken                      string
)

func init() {
	RootCmd.AddCommand(ImageCmd)
	ImageCmd.AddCommand(DownloadImageCmd)
	ImageCmd.AddCommand(UploadImageCmd)
	ImageCmd.AddCommand(IndexImageCmd)
}

type ImageEntry struct {
	ID   string   `json:"id,omitempty"`
	Path string   `json:"path,omitempty"`
	Type string   `json:"type,omitempty"`
	Refs []string `json:"refs,omitempty"`
}
