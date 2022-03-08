// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfcors is a CORS middleware for GoFrame framework
package rkgfcors

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/cors"
	"net/http"
)

// Middleware Add CORS interceptors.
func Middleware(opts ...rkmidcors.Option) ghttp.HandlerFunc {
	set := rkmidcors.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response.Header().Set(k, v)
		}

		for _, v := range beforeCtx.Output.HeaderVary {
			ctx.Response.Header().Add(rkmid.HeaderVary, v)
		}

		// case 1: with abort
		if beforeCtx.Output.Abort {
			ctx.Response.WriteHeader(http.StatusNoContent)
			return
		}

		ctx.Middleware.Next()
	}
}
