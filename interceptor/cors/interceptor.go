// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfcors is a CORS middleware for GoFrame framework
package rkgfcors

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"net/http"
	"strconv"
	"strings"
)

// Interceptor Add CORS interceptors.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	allowMethods := strings.Join(set.AllowMethods, ",")
	allowHeaders := strings.Join(set.AllowHeaders, ",")
	exposeHeaders := strings.Join(set.ExposeHeaders, ",")

	maxAge := strconv.Itoa(set.MaxAge)
	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Middleware.Next()
			return
		}

		originHeader := ctx.Request.Header.Get(headerOrigin)
		preflight := ctx.Request.Method == http.MethodOptions

		// 1: if no origin header was provided, we will return 204 if request is not a OPTION method
		if originHeader == "" {
			// 1.1: if not a preflight request, then pass through
			if !preflight {
				ctx.Middleware.Next()
				return
			}

			// 1.2: if it is a preflight request, then return with 204
			ctx.Response.WriteHeader(http.StatusNoContent)
			return
		}

		// 2: origin not allowed, we will return 204 if request is not a OPTION method
		if !set.isOriginAllowed(originHeader) {
			// 2.1: if not a preflight request, then pass through
			if !preflight {
				ctx.Response.WriteHeader(http.StatusNoContent)
				return
			}

			// 2.2: if it is a preflight request, then return with 204
			ctx.Response.WriteHeader(http.StatusNoContent)
			return
		}

		// 3: not a OPTION method
		if !preflight {
			ctx.Response.Writer.Header().Set(headerAccessControlAllowOrigin, originHeader)
			// 3.1: add Access-Control-Allow-Credentials
			if set.AllowCredentials {
				ctx.Response.Writer.Header().Set(headerAccessControlAllowCredentials, "true")
			}
			// 3.2: add Access-Control-Expose-Headers
			if exposeHeaders != "" {
				ctx.Response.Writer.Header().Set(headerAccessControlExposeHeaders, exposeHeaders)
			}
			ctx.Middleware.Next()
			return
		}

		// 4: preflight request, return 204
		// add related headers including:
		//
		// - Vary
		// - Access-Control-Allow-Origin
		// - Access-Control-Allow-Methods
		// - Access-Control-Allow-Credentials
		// - Access-Control-Allow-Headers
		// - Access-Control-Max-Age
		ctx.Response.Writer.Header().Add(headerVary, headerAccessControlRequestMethod)
		ctx.Response.Writer.Header().Add(headerVary, headerAccessControlRequestHeaders)
		ctx.Response.Writer.Header().Set(headerAccessControlAllowOrigin, originHeader)
		ctx.Response.Writer.Header().Set(headerAccessControlAllowMethods, allowMethods)

		// 4.1: Access-Control-Allow-Credentials
		if set.AllowCredentials {
			ctx.Response.Writer.Header().Set(headerAccessControlAllowCredentials, "true")
		}

		// 4.2: Access-Control-Allow-Headers
		if allowHeaders != "" {
			ctx.Response.Writer.Header().Set(headerAccessControlAllowHeaders, allowHeaders)
		} else {
			h := ctx.Request.Header.Get(headerAccessControlRequestHeaders)
			if h != "" {
				ctx.Response.Writer.Header().Set(headerAccessControlAllowHeaders, h)
			}
		}
		if set.MaxAge > 0 {
			// 4.3: Access-Control-Max-Age
			ctx.Response.Writer.Header().Set(headerAccessControlMaxAge, maxAge)
		}

		ctx.Response.WriteHeader(http.StatusNoContent)
		return
	}
}
