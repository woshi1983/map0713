// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
)

// jwtSecret initializes on program starting, for this reason, all generated
// token will be expired on program restart.
var jwtSecret = randutil.BytesN(16)

// JWTCodec is a codec for JWT.
type JWTCodec interface {
	Encode(jwt.Claims) (string, error)
	Decode(string) (jwt.Claims, error)
}

// JWTNewClaimsFunc is used for generating a new instance for unmarshalling
// token.
type JWTNewClaimsFunc func() jwt.Claims

type jwtHMACCodec struct {
	newFn JWTNewClaimsFunc
}

func (c jwtHMACCodec) Encode(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (c jwtHMACCodec) Decode(str string) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(str, c.newFn(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot verify JWT signature: %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT signature")
	}
	return token.Claims, nil
}

// NewHMACJWTCodec initializes a HMAC JWT codec.
func NewHMACJWTCodec(newfn JWTNewClaimsFunc) JWTCodec {
	return &jwtHMACCodec{newfn}
}
