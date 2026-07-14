// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package provides useful helpers and utilities.
package util

import "time"

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func UintPtr(i uint) *uint {
	return &i
}

func Uint32Ptr(i uint32) *uint32 {
	return &i
}

func Uint64Ptr(i uint64) *uint64 {
	return &i
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func StringPtr(s string) *string {
	return &s
}

func Int32PtrToInt64Ptr(v *int32) *int64 {
	val := int64(*v)
	return &val
}

func Int64PtrToInt32Ptr(v *int64) *int32 {
	val := int32(*v)
	return &val
}

func Retry(fn func() error, n int, interval time.Duration) error {
	err := fn()
	for i := 0; err != nil && (n < 0 || n > -1 && i < n); i++ {
		time.Sleep(interval)
		err = fn()
	}
	return err
}
