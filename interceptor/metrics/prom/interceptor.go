// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfmetrics is a middleware for GoFrame framework which record prometheus metrics for RPC
package rkgfmetrics

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	"strconv"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...rkmidmetrics.Option) ghttp.HandlerFunc {
	set := rkmidmetrics.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		ctx.Middleware.Next()

		afterCtx := set.AfterCtx(strconv.Itoa(ctx.Response.Status))
		set.After(beforeCtx, afterCtx)
	}
}
