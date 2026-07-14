// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package submodule

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	// Router is a global HTTP router.
	Router = echo.New()
)

func init() {
	Router.StdLogger = log.New(newNoOpWriter(), "", 0)
	Router.HideBanner = true
	Router.HidePort = true
	Router.GET("/", func(ctx echo.Context) (err error) {
		return ctx.Redirect(http.StatusSeeOther, "/devtool/graphql")
	})
}
