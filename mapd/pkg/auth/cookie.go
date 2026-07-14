// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	echo "github.com/labstack/echo/v4"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
)

func SetSecuredCookie(ctx echo.Context, name, value string, expiry time.Time) {
	setSecuredCookie(ctx, name, value, true, expiry)
}

func SetBrowserReadableSecuredCookie(ctx echo.Context, name, value string, expiry time.Time) {
	setSecuredCookie(ctx, name, value, false, expiry)
}

func setSecuredCookie(ctx echo.Context, name, value string, httpOnly bool, expiry time.Time) {
	// signature of HMAC
	s := base64.StdEncoding.EncodeToString([]byte(value))
	h := hmac.New(sha1.New, []byte(signatureKey))
	h.Write([]byte(s))

	cookie := &http.Cookie{
		Name:     name,
		Value:    fmt.Sprintf("%s.%x", s, h.Sum(nil)),
		Path:     "/",
		Domain:   mapd.GlobalConfig.FQDN,
		Secure:   ctx.Request().URL.Scheme == "https",
		HttpOnly: httpOnly,
		SameSite: http.SameSiteStrictMode,
	}
	if !expiry.IsZero() {
		cookie.Expires = expiry
		d := time.Until(cookie.Expires)
		if d < 0 {
			cookie.MaxAge = -1
		} else {
			cookie.MaxAge = int(math.Round(d.Seconds())) + 1
		}
	}
	ctx.SetCookie(cookie)
}

func GetSecuredCookie(ctx echo.Context, name string, validate bool) (value string, err error) {
	var cookie *http.Cookie
	cookie, err = ctx.Cookie(name)
	if err != nil {
		return
	}

	ss := strings.SplitN(cookie.Value, ".", 2)
	if len(ss) < 2 {
		return value, echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	s := ss[0]

	// signature of HMAC
	h := hmac.New(sha1.New, []byte(signatureKey))
	h.Write([]byte(s))
	if ss[1] != fmt.Sprintf("%x", h.Sum(nil)) {
		return value, echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return value, echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return string(b), nil
}
