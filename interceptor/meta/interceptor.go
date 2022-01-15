// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfmeta is a middleware of GoFrame framework for adding metadata in RPC response
package rkgfmeta

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidmeta "github.com/rookie-ninja/rk-entry/middleware/meta"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
)

// Interceptor will add common headers as extension style in http response.
func Interceptor(opts ...rkmidmeta.Option) ghttp.HandlerFunc {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(rkgfctx.GetEvent(ctx))
		set.Before(beforeCtx)

		ctx.SetCtxVar(rkmid.HeaderRequestId, beforeCtx.Output.RequestId)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response.Header().Set(k, v)
		}

		ctx.Middleware.Next()
	}
}
