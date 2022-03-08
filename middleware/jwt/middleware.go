// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfjwt is a JWT middleware for GoFrame framework
package rkgfjwt

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	rkmidjwt "github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
)

// Middleware Add CORS interceptors.
func Middleware(opts ...rkmidjwt.Option) ghttp.HandlerFunc {
	set := rkmidjwt.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request, nil)
		set.Before(beforeCtx)

		// case 1: error response
		if beforeCtx.Output.ErrResp != nil {
			ctx.Response.WriteStatus(beforeCtx.Output.ErrResp.Err.Code, beforeCtx.Output.ErrResp)
			return
		}

		// insert into context
		ctx.SetCtxVar(rkmid.JwtTokenKey, beforeCtx.Output.JwtToken)

		ctx.Middleware.Next()
	}
}
