// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidsec "github.com/rookie-ninja/rk-entry/middleware/secure"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	rkgfsec "github.com/rookie-ninja/rk-gf/interceptor/secure"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with secure interceptor enabled.
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
		rkgfsec.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidsec.WithEntryNameAndType("greeter", "gf"),
			//
			// X-XSS-Protection header value.
			// Optional. Default value "1; mode=block".
			//rkmidsec.WithXSSProtection("my-value"),
			//
			// X-Content-Type-Options header value.
			// Optional. Default value "nosniff".
			//rkmidsec.WithContentTypeNosniff("my-value"),
			//
			// X-Frame-Options header value.
			// Optional. Default value "SAMEORIGIN".
			//rkmidsec.WithXFrameOptions("my-value"),
			//
			// Optional, Strict-Transport-Security header value.
			//rkmidsec.WithHSTSMaxAge(1),
			//
			// Optional, excluding subdomains of HSTS, default is false
			//rkmidsec.WithHSTSExcludeSubdomains(true),
			//
			// Optional, enabling HSTS preload, default is false
			//rkmidsec.WithHSTSPreloadEnabled(true),
			//
			// Content-Security-Policy header value.
			// Optional. Default value "".
			//rkmidsec.WithContentSecurityPolicy("my-value"),
			//
			// Content-Security-Policy-Report-Only header value.
			// Optional. Default value false.
			//rkmidsec.WithCSPReportOnly(true),
			//
			// Referrer-Policy header value.
			// Optional. Default value "".
			//rkmidsec.WithReferrerPolicy("my-value"),
			//
			// Ignoring path prefix.
			//rkmidsec.WithIgnorePrefix("/rk/v1"),
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
		Message: "Received request!",
	})
}
