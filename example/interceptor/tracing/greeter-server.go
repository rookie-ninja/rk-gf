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
	"github.com/rookie-ninja/rk-gf/interceptor/tracing/telemetry"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with tracing interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ****************************************
	// ********** Create Exporter *************
	// ****************************************

	// Export trace to stdout
	exporter := rkgftrace.CreateFileExporter("stdout")

	// Export trace to local file system
	//exporter := rkgftrace.CreateFileExporter("logs/trace.log")

	// Export trace to jaeger agent
	//exporter := rkgftrace.CreateJaegerExporter(jaeger.WithAgentEndpoint())

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []ghttp.HandlerFunc{
		rkgflog.Interceptor(),
		rkgftrace.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			//rkgftrace.WithEntryNameAndType("greeter", "gf"),
			//
			// Provide an exporter.
			rkgftrace.WithExporter(exporter),
			//
			// Provide propagation.TextMapPropagator
			// rkgftrace.WithPropagator(<propagator>),
			//
			// Provide SpanProcessor
			// rkgftrace.WithSpanProcessor(<span processor>),
			//
			// Provide TracerProvider
			// rkgftrace.WithTracerProvider(<trace provider>),
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
		Message: fmt.Sprintf("Hello %s!", ctx.GetQuery("name")),
	})
}
