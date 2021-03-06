// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgflimit is a middleware of GoFrame framework for adding rate limit in RPC response
package rkgflimit

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
)

// Middleware Add rate limit interceptors.
func Middleware(opts ...rkmidlimit.Option) ghttp.HandlerFunc {
	set := rkmidlimit.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			ctx.Response.WriteStatus(beforeCtx.Output.ErrResp.Code(), beforeCtx.Output.ErrResp)
			return
		}

		ctx.Middleware.Next()
	}
}
