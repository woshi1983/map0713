// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kvstore provides a generic key-value store interface.
package kvstore

import "time"

type KVStore interface {
	Set(key string, value []byte)
	Get(key string) []byte
	SetTTL(key string, value []byte, ttl time.Duration)
	Delete(key string)
	Len() int
}
