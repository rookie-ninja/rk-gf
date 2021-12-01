// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfjwt is a JWT middleware for GoFrame framework
package rkgfjwt

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"net/http"
)

// Interceptor Add CORS interceptors.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Middleware.Next()
			return
		}

		// extract token from extractor
		var auth string
		var err error
		for _, extractor := range set.extractors {
			// Extract token from extractor, if it's not fail break the loop and
			// set auth
			auth, err = extractor(ctx)
			if err == nil {
				break
			}
		}

		if err != nil {
			ctx.Response.WriteHeader(http.StatusUnauthorized)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return
		}

		// parse token
		token, err := set.ParseTokenFunc(auth, ctx)

		if err != nil {
			ctx.Response.WriteHeader(http.StatusUnauthorized)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return
		}

		// insert into context
		ctx.SetCtxVar(rkgfinter.RpcJwtTokenKey, token)

		ctx.Middleware.Next()
	}
}
