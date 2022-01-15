// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/middleware/csrf"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-gf/interceptor/csrf"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with CSRF interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter.
func main() {
	// ******************************************************
	// ********** Override App name and version *************
	// ******************************************************
	//
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "demo-app"
	// rkentry.GlobalAppCtx.GetAppInfoEntry().Version = "demo-version"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []ghttp.HandlerFunc{
		rkgfcsrf.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidcsrf.WithEntryNameAndType("greeter", "gf"),
			// rkmidcsrf.WithIgnorePrefix(""),
			// WithTokenLength the length of the generated token.
			// Optional. Default value 32.
			//rkmidcsrf.WithTokenLength(10),
			//
			// WithTokenLookup a string in the form of "<source>:<key>" that is used
			// to extract token from the request.
			// Optional. Default value "header:X-CSRF-Token".
			// Possible values:
			// - "header:<name>"
			// - "form:<name>"
			// - "query:<name>"
			// Optional. Default value "header:X-CSRF-Token".
			//rkmidcsrf.WithTokenLookup("header:X-CSRF-Token"),
			//
			// WithCookieName provide name of the CSRF cookie. This cookie will store CSRF token.
			// Optional. Default value "csrf".
			//rkmidcsrf.WithCookieName("csrf"),
			//
			// WithCookieDomain provide domain of the CSRF cookie.
			// Optional. Default value "".
			//rkmidcsrf.WithCookieDomain(""),
			//
			// WithCookiePath provide path of the CSRF cookie.
			// Optional. Default value "".
			//rkmidcsrf.WithCookiePath(""),
			//
			// WithCookieMaxAge provide max age (in seconds) of the CSRF cookie.
			// Optional. Default value 86400 (24hr).
			//rkmidcsrf.WithCookieMaxAge(10),
			//
			// WithCookieHTTPOnly indicates if CSRF cookie is HTTP only.
			// Optional. Default value false.
			//rkmidcsrf.WithCookieHTTPOnly(false),
			//
			// WithCookieSameSite indicates SameSite mode of the CSRF cookie.
			// Optional. Default value SameSiteDefaultMode.
			//rkmidcsrf.WithCookieSameSite(http.SameSiteStrictMode),
		),
	}

	// 1: Create gf server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown()

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start gf server.
func startGreeterServer(interceptors ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(time.Now())
	server.SetPort(8080)
	server.SetDumpRouterMap(false)
	server.SetLogger(rkgfinter.NewNoopGLogger())
	server.Use(interceptors...)
	server.BindHandler("/rk/v1/greeter", Greeter)
	glog.SetStdoutPrint(false)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server
}

// GreeterResponse Response of Greeter.
type GreeterResponse struct {
	Message string
}

// Greeter Handler.
func Greeter(ctx *ghttp.Request) {
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	rkgfctx.GetLogger(ctx).Info("Received request from client.")

	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(&GreeterResponse{
		Message: fmt.Sprintf("CSRF token:%v", rkgfctx.GetCsrfToken(ctx)),
	})
}
