// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfmeta is a middleware of GoFrame framework for adding metadata in RPC response
package rkgfmeta

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	"github.com/rookie-ninja/rk-gf/middleware/context"
)

// Middleware will add common headers as extension style in http response.
func Middleware(opts ...rkmidmeta.Option) ghttp.HandlerFunc {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request, rkgfctx.GetEvent(ctx))
		set.Before(beforeCtx)

		ctx.SetCtxVar(rkmid.HeaderRequestId, beforeCtx.Output.RequestId)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response.Header().Set(k, v)
		}

		ctx.Middleware.Next()
	}
}
