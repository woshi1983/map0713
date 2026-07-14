// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kvstore

import (
	"context"
	"sync"
	"time"
)

type memoryKVStore struct {
	mux sync.RWMutex
	val map[string][]byte
	del map[string]context.CancelFunc
	exp map[string]context.CancelFunc
}

func (s *memoryKVStore) Set(key string, value []byte) {
	s.SetTTL(key, value, -1)
}

func (s *memoryKVStore) Get(key string) []byte {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.val[key]
}

func (s *memoryKVStore) SetTTL(key string, value []byte, ttl time.Duration) {
	s.mux.Lock()
	defer s.mux.Unlock()

	cancel, ok := s.del[key]
	if ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	var (
		expCtx    context.Context
		expCancel context.CancelFunc
	)
	if ttl <= 0 {
		expCtx, expCancel = context.WithCancel(context.Background())
	} else {
		expCtx, expCancel = context.WithTimeout(context.Background(), ttl)
	}
	s.val[key] = value
	s.del[key] = cancel
	s.exp[key] = expCancel

	go func() {
		defer expCancel()

		select {
		case <-ctx.Done():
		case <-expCtx.Done():
			s.mux.Lock()
			defer s.mux.Unlock()

			cancel()

			delete(s.val, key)
			delete(s.del, key)
			delete(s.exp, key)
		}
	}()
}

func (s *memoryKVStore) Delete(key string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	cancel, ok := s.del[key]
	if ok {
		cancel()
	}

	delete(s.val, key)
	delete(s.del, key)
	delete(s.exp, key)
}

func (s *memoryKVStore) Len() int {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return len(s.val)
}

func NewMemoryKVStore() *memoryKVStore {
	return &memoryKVStore{
		val: make(map[string][]byte),
		del: make(map[string]context.CancelFunc),
		exp: make(map[string]context.CancelFunc),
	}
}
