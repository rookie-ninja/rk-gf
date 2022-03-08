// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfsec is a secure middleware for GoFrame framework
package rkgfsec

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
)

// Middleware Add Secure interceptors.
func Middleware(opts ...rkmidsec.Option) ghttp.HandlerFunc {
	set := rkmidsec.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response.Header().Set(k, v)
		}

		ctx.Middleware.Next()
	}
}
