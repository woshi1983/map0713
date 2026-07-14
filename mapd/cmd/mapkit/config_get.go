// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetConfigCmd represents the config get command.
var GetConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Get value of given keys from YAML config file.",
	Long:  "Get value of given keys from YAML config file.",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			obj      = make(map[string]interface{})
			val      interface{}
			found    bool
			exitCode int
			err      error
		)
		b, err := os.ReadFile(configPath)
		if err != nil {
			exitCode = 1
			goto EXIT
		}
		if err = yaml.Unmarshal(b, &obj); err != nil {
			exitCode = 1
			goto EXIT
		}

		val, found, err = unstructured.NestedFieldNoCopy(obj, args...)
		if !found || err != nil {
			fmt.Fprintln(os.Stdout, "")
			return
		}

		switch val := val.(type) {
		case bool:
			fmt.Fprintln(os.Stdout, val)
		case string:
			fmt.Fprintln(os.Stdout, val)
		case []byte:
			fmt.Fprintln(os.Stdout, string(val))
		case int:
			fmt.Fprintln(os.Stdout, val)
		case int64:
			fmt.Fprintln(os.Stdout, val)
		case float32:
			fmt.Fprintln(os.Stdout, val)
		case float64:
			fmt.Fprintln(os.Stdout, val)
		default:
			b, err = yaml.Marshal(val)
			if err != nil {
				exitCode = 2
				goto EXIT
			}
			fmt.Fprintln(os.Stdout, string(b))
		}
		return

	EXIT:
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
		}
		os.Exit(exitCode)
	},
}

func init() {
	ConfigCmd.AddCommand(GetConfigCmd)
}
