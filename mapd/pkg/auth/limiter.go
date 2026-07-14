// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net"
	"sync"
	"time"

	echo "github.com/labstack/echo/v4"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
)

// Limiter provides protection by limiting
type Limiter struct {
	mux    sync.RWMutex
	store  map[string]*Limit
	cancel map[string]context.CancelFunc
}

func (l *Limiter) IsBanned(ctx echo.Context) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	ip, _, _ := net.SplitHostPort(ctx.Request().RemoteAddr)
	if v, ok := l.store[ip]; ok {
		if v.Value() < 1 {
			return echo.ErrForbidden
		}
	}
	return nil
}

func (l *Limiter) Succeed(ctx echo.Context) {
	l.mux.Lock()
	defer l.mux.Unlock()

	ip, _, _ := net.SplitHostPort(ctx.Request().RemoteAddr)
	if cancel, ok := l.cancel[ip]; ok {
		cancel()
	}
	delete(l.store, ip)
}

func (l *Limiter) Fail(ctx echo.Context) {
	l.mux.Lock()
	defer l.mux.Unlock()

	ip, _, _ := net.SplitHostPort(ctx.Request().RemoteAddr)
	if cancel, ok := l.cancel[ip]; ok {
		cancel()
	}

	var (
		v      *Limit
		exists bool
	)
	if v, exists = l.store[ip]; !exists {
		v = NewLimit()
		v.Reset()
	}
	v.Decrease()
	l.store[ip] = v

	timer := time.NewTimer(mapd.GlobalConfig.Auth.MaxRequestsDuration)
	cTx, cancel := context.WithCancel(context.Background())
	go func() {
		defer timer.Stop()
		defer cancel()

		select {
		case <-timer.C:
		case <-cTx.Done():
			return
		}

		l.mux.Lock()
		defer l.mux.Unlock()

		delete(l.store, ip)
		delete(l.cancel, ip)
	}()
}

func NewLimiter() *Limiter {
	return &Limiter{store: make(map[string]*Limit), cancel: make(map[string]context.CancelFunc)}
}

type Limit struct {
	mux sync.RWMutex
	val int64
}

func (l *Limit) Value() int64 {
	l.mux.RLock()
	defer l.mux.RUnlock()

	return l.val
}

func (l *Limit) Reset() int64 {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.val = mapd.GlobalConfig.Auth.MaxRequestsPerIP
	return l.val
}

func (l *Limit) Decrease() int64 {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.val--
	return l.val
}

func NewLimit() *Limit {
	return &Limit{}
}
