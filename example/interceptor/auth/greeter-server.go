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
	rkgfinter "github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/auth"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with auth interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []ghttp.HandlerFunc{
		rkgflog.Interceptor(),
		rkgfauth.Interceptor(
			// rkgfauth.WithIgnorePrefix("/rk/v1/greeter"),
			rkgfauth.WithBasicAuth("", "rk-user:rk-pass"),
			rkgfauth.WithApiKeyAuth("rk-api-key"),
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
	validateCtx(ctx)

	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.GetQuery("name")),
	})
}

func validateCtx(ctx *ghttp.Request) {
	// 1: get incoming headers
	printIndex("[1]: get incoming headers")
	prettyHeader(rkgfctx.GetIncomingHeaders(ctx))

	// 2: add header to client
	printIndex("[2]: add header to client")
	rkgfctx.AddHeaderToClient(ctx, "add-key", "add-value")

	// 3: set header to client
	printIndex("[3]: set header to client")
	rkgfctx.SetHeaderToClient(ctx, "set-key", "set-value")

	// 4: get event
	printIndex("[4]: get event")
	rkgfctx.GetEvent(ctx).SetCounter("my-counter", 1)

	// 5: get logger
	printIndex("[5]: get logger")
	rkgfctx.GetLogger(ctx).Info("error msg")

	// 6: get request id
	printIndex("[6]: get request id")
	fmt.Println(rkgfctx.GetRequestId(ctx))

	// 7: get trace id
	printIndex("[7]: get trace id")
	fmt.Println(rkgfctx.GetTraceId(ctx))

	// 8: get entry name
	printIndex("[8]: get entry name")
	fmt.Println(rkgfctx.GetEntryName(ctx))

	// 9: get trace span
	printIndex("[9]: get trace span")
	fmt.Println(rkgfctx.GetTraceSpan(ctx))

	// 10: get tracer
	printIndex("[10]: get tracer")
	fmt.Println(rkgfctx.GetTracer(ctx))

	// 11: get trace provider
	printIndex("[11]: get trace provider")
	fmt.Println(rkgfctx.GetTracerProvider(ctx))

	// 12: get tracer propagator
	printIndex("[12]: get tracer propagator")
	fmt.Println(rkgfctx.GetTracerPropagator(ctx))

	// 13: inject span
	printIndex("[13]: inject span")
	req := &http.Request{}
	rkgfctx.InjectSpanToHttpRequest(ctx, req)

	// 14: new trace span
	printIndex("[14]: new trace span")
	fmt.Println(rkgfctx.NewTraceSpan(ctx, "my-span"))

	// 15: end trace span
	printIndex("[15]: end trace span")
	rkgfctx.EndTraceSpan(ctx, rkgfctx.GetTraceSpan(ctx), true)
}

func printIndex(key string) {
	fmt.Println(fmt.Sprintf("%s", key))
}

func prettyHeader(header http.Header) {
	for k, v := range header {
		fmt.Println(fmt.Sprintf("%s:%s", k, v))
	}
}
