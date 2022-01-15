// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgflog is a middleware for GoFrame framework for logging RPC.
package rkgflog

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"strconv"
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...rkmidlog.Option) ghttp.HandlerFunc {
	set := rkmidlog.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		// call before
		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		ctx.SetCtxVar(rkmid.EventKey, beforeCtx.Output.Event)
		ctx.SetCtxVar(rkmid.LoggerKey, beforeCtx.Output.Logger)

		ctx.Middleware.Next()

		// call after
		afterCtx := set.AfterCtx(
			rkgfctx.GetRequestId(ctx),
			rkgfctx.GetTraceId(ctx),
			strconv.Itoa(ctx.Response.Status))
		set.After(beforeCtx, afterCtx)
	}
}
