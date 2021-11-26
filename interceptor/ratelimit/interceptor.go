// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgflimit is a middleware of GoFrame framework for adding rate limit in RPC response
package rkgflimit

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"net/http"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		event := rkgfctx.GetEvent(ctx)

		if duration, err := set.Wait(ctx); err != nil {
			event.SetCounter("rateLimitWaitMs", duration.Milliseconds())
			event.AddErr(err)

			ctx.Response.WriteHeader(http.StatusTooManyRequests)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusTooManyRequests),
				rkerror.WithDetails(err)))
			return
		}

		ctx.Middleware.Next()
	}
}
