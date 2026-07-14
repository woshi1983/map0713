// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
)

func Context(ctx context.Context, info *SessionInfo) context.Context {
	return context.WithValue(ctx, sessCtxKey, info)
}

func FromContext(ctx context.Context) (*SessionInfo, bool) {
	v, ok := ctx.Value(sessCtxKey).(*SessionInfo)
	return v, ok
}

var (
	sessCtxKey = &contextKey{"sess"}
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
type contextKey struct {
	name string
}
