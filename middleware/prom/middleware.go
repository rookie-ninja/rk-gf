// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfprom is a middleware for GoFrame framework which record prometheus metrics for RPC
package rkgfprom

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"strconv"
)

// Middleware create a new prometheus metrics interceptor with options.
func Middleware(opts ...rkmidprom.Option) ghttp.HandlerFunc {
	set := rkmidprom.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		ctx.Middleware.Next()

		afterCtx := set.AfterCtx(strconv.Itoa(ctx.Response.Status))
		set.After(beforeCtx, afterCtx)
	}
}
