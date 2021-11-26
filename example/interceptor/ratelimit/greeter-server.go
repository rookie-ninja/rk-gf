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
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gf/interceptor/ratelimit"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with rate limit interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
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
		rkgflog.Interceptor(),
		rkgflimit.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		//rkgflimit.WithEntryNameAndType("greeter", "gf"),
		//
		// Provide algorithm, rkgflimit.LeakyBucket and rkgflimit.TokenBucket was available, default is TokenBucket.
		//rkgflimit.WithAlgorithm(rkgflimit.LeakyBucket),
		//
		// Provide request per second, if provide value of zero, then no requests will be pass through and user will receive an error with
		// resource exhausted.
		//rkgflimit.WithReqPerSec(10),
		//
		// Provide request per second with path name.
		// The name should be full path name. if provide value of zero,
		// then no requests will be pass through and user will receive an error with resource exhausted.
		//rkgflimit.WithReqPerSecByPath("/rk/v1/greeter", 10),
		//
		// Provide user function of limiter. Returns error if you want to limit the request.
		// Please do not try to set response code since it will be overridden by middleware.
		//rkgflimit.WithGlobalLimiter(func(ctx *ghttp.Request) error {
		//	return fmt.Errorf("limited by custom limiter")
		//}),
		//
		// Provide user function of limiter by path name.
		// The name should be full path name.
		//rkgflimit.WithLimiterByPath("/rk/v1/greeter", func(ctx *ghttp.Request) error {
		//	 return nil
		//}),
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
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkgfmeta.Interceptor() in server side.
	// 2: rkgfctx.AddHeaderToClient(ctx, rkgfctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkgfctx.GetLogger(ctx).Info("Received request from client.")

	// Set request id with X-Request-Id to outgoing headers.
	// rkgfctx.SetHeaderToClient(ctx, rkgfctx.RequestIdKey, "this-is-my-request-id-overridden")

	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.GetQuery("name")),
	})
}
