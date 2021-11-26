// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfauth is auth middleware for GoFrame framework
package rkgfauth

import (
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"net/http"
	"strings"
)

const (
	typeBasic  = "Basic"
	typeApiKey = "X-API-Key"
)

// Interceptor validate bellow authorization.
//
// 1: Basic Auth: The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password.
// 2: Bearer Token: Commonly known as token authentication. It is an HTTP authentication scheme that involves security tokens called bearer tokens.
// 3: API key: An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)
		before(ctx, set)
	}
}

func before(ctx *ghttp.Request, set *optionSet) {
	if !set.ShouldAuth(ctx) {
		ctx.Middleware.Next()
		return
	}

	authHeader := ctx.Request.Header.Get(rkgfinter.RpcAuthorizationHeaderKey)
	apiKeyHeader := ctx.Request.Header.Get(rkgfinter.RpcApiKeyHeaderKey)

	if len(authHeader) > 0 {
		// Contains auth header
		// Basic auth type
		tokens := strings.SplitN(authHeader, " ", 2)
		if len(tokens) != 2 {
			ctx.Response.WriteHeader(http.StatusUnauthorized)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid Basic Auth format")))
			return
		}
		if !set.Authorized(tokens[0], tokens[1]) {
			if tokens[0] == typeBasic {
				ctx.Response.RawWriter().Header().Set("WWW-Authenticate", fmt.Sprintf(`%s realm="%s"`, typeBasic, set.BasicRealm))
			}

			ctx.Response.WriteHeader(http.StatusUnauthorized)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid credential")))
			return
		}
	} else if len(apiKeyHeader) > 0 {
		// Contains api key
		if !set.Authorized(typeApiKey, apiKeyHeader) {
			ctx.Response.WriteHeader(http.StatusUnauthorized)
			ctx.Response.Write(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid X-API-Key")))
			return
		}
	} else {
		authHeaders := []string{}
		if len(set.BasicAccounts) > 0 {
			ctx.Response.RawWriter().Header().Set("WWW-Authenticate", fmt.Sprintf(`%s realm="%s"`, typeBasic, set.BasicRealm))
			authHeaders = append(authHeaders, "Basic Auth")
		}
		if len(set.ApiKey) > 0 {
			authHeaders = append(authHeaders, "X-API-Key")
		}

		errMsg := fmt.Sprintf("Missing authorization, provide one of bellow auth header:[%s]", strings.Join(authHeaders, ","))

		ctx.Response.WriteHeader(http.StatusUnauthorized)
		ctx.Response.Write(rkerror.New(
			rkerror.WithHttpCode(http.StatusUnauthorized),
			rkerror.WithMessage(errMsg)))
		return
	}

	ctx.Middleware.Next()
}
