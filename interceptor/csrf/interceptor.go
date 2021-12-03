// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfcsrf is a middleware for GoFrame framework which validating csrf token for RPC
package rkgfcsrf

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"net/http"
	"time"
)

// Interceptor Add CSRF interceptors.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Middleware.Next()
			return
		}

		req := ctx.Request
		k, err := ctx.Request.Cookie(set.CookieName)
		token := ""

		// 2.1: generate token if failed to get cookie from context
		if err != nil {
			token = rkcommon.RandString(set.TokenLength)
		} else {
			// 2.2: reuse token if exists
			token = k.Value
		}

		// 3.1: do not check http methods of GET, HEAD, OPTIONS and TRACE
		switch req.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// 3.2: validate token only for requests which are not defined as 'safe' by RFC7231
			clientToken, err := set.extractor(ctx)
			if err != nil {
				ctx.Response.WriteHeader(http.StatusBadRequest)
				ctx.Response.WriteJson(rkerror.New(
					rkerror.WithHttpCode(http.StatusBadRequest),
					rkerror.WithMessage("failed to extract client token"),
					rkerror.WithDetails(err)))
				return
			}

			// 3.3: return 403 to client if token is not matched
			if !isValidToken(token, clientToken) {
				ctx.Response.WriteHeader(http.StatusForbidden)
				ctx.Response.WriteJson(rkerror.New(
					rkerror.WithHttpCode(http.StatusForbidden),
					rkerror.WithMessage("invalid csrf token"),
					rkerror.WithDetails(err)))
				return
			}
		}

		// set CSRF cookie
		cookie := new(http.Cookie)
		cookie.Name = set.CookieName
		cookie.Value = token
		// 4.1
		if set.CookiePath != "" {
			cookie.Path = set.CookiePath
		}
		// 4.2
		if set.CookieDomain != "" {
			cookie.Domain = set.CookieDomain
		}
		// 4.3
		if set.CookieSameSite != http.SameSiteDefaultMode {
			cookie.SameSite = set.CookieSameSite
		}
		cookie.Expires = time.Now().Add(time.Duration(set.CookieMaxAge) * time.Second)
		cookie.Secure = set.CookieSameSite == http.SameSiteNoneMode
		cookie.HttpOnly = set.CookieHTTPOnly
		http.SetCookie(ctx.Response.Writer, cookie)

		// store token in the context
		ctx.SetCtxVar(rkgfinter.RpcCsrfTokenKey, token)

		// protect clients from caching the response
		ctx.Response.Header().Add(headerVary, headerCookie)

		ctx.Middleware.Next()
	}
}
