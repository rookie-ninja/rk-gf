// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfcsrf is a middleware for GoFrame framework which validating csrf token for RPC
package rkgfcsrf

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/middleware/csrf"
	"net/http"
)

// Interceptor Add CSRF interceptors.
func Interceptor(opts ...rkmidcsrf.Option) ghttp.HandlerFunc {
	set := rkmidcsrf.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			ctx.Response.WriteStatus(beforeCtx.Output.ErrResp.Err.Code, beforeCtx.Output.ErrResp)
			return
		}

		for _, v := range beforeCtx.Output.VaryHeaders {
			ctx.Response.Header().Add(rkmid.HeaderVary, v)
		}

		if beforeCtx.Output.Cookie != nil {
			http.SetCookie(ctx.Response.Writer, beforeCtx.Output.Cookie)
		}

		// store token in the context
		ctx.SetCtxVar(rkmid.CsrfTokenKey, beforeCtx.Input.Token)

		ctx.Middleware.Next()
	}
}
