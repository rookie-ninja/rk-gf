// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfpanic is a middleware of GoFrame framework for recovering from panic
package rkgfpanic

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/error"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
)

// Interceptor returns a ghttp.HandlerFunc (middleware)
func Interceptor(opts ...rkmidpanic.Option) ghttp.HandlerFunc {
	set := rkmidpanic.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		handlerFunc := func(resp *rkerror.ErrorResp) {
			ctx.Response.ClearBuffer()
			ctx.Response.WriteHeader(resp.Err.Code)
			ctx.Response.WriteJson(resp)
		}
		beforeCtx := set.BeforeCtx(rkgfctx.GetEvent(ctx), rkgfctx.GetLogger(ctx), handlerFunc)
		set.Before(beforeCtx)

		defer beforeCtx.Output.DeferFunc()

		ctx.Middleware.Next()
	}
}
