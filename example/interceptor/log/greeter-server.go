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
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with log interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []ghttp.HandlerFunc{
		//rkgfmeta.Interceptor(),
		rkgflog.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidlog.WithEntryNameAndType("greeter", "gin"),
			//
			// Zap logger would be logged as JSON format.
			// rkmidlog.WithZapLoggerEncoding("json"),
			//
			// Event logger would be logged as JSON format.
			// rkmidlog.WithEventLoggerEncoding("json"),
			//
			// Zap logger would be logged to specified path.
			// rkmidlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
			//
			// Event logger would be logged to specified path.
			// rkmidlog.WithEventLoggerOutputPaths("logs/server-event.log"),
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

// Greeter Handler for greeter.
func Greeter(ctx *ghttp.Request) {
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkgfmeta.Interceptor() in server side.
	// 2: rkgfctx.SetHeaderToClient(ctx, rkgfctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkgfctx.GetLogger(ctx).Info("Received request from client.")

	// *******************************************
	// ********** rpc-scoped event  *************
	// *******************************************
	//
	// Get rkquery.Event which would be printed as soon as request finish.
	// User can call any Add/Set/Get functions on rkquery.Event
	//
	// rkgfctx.GetEvent(ctx).AddPair("rk-key", "rk-value")

	// *********************************************
	// ********** Get incoming headers *************
	// *********************************************
	//
	// Read headers sent from client.
	//
	//for k, v := range rkgfctx.GetIncomingHeaders(ctx) {
	//	 fmt.Println(fmt.Sprintf("%s: %s", k, v))
	//}

	// *********************************************************
	// ********** Add headers will send to client **************
	// *********************************************************
	//
	// Send headers to client with this function
	//
	//rkgfctx.AddHeaderToClient(ctx, "from-server", "value")

	// ***********************************************
	// ********** Get and log request id *************
	// ***********************************************
	//
	// RequestId will be printed on both client and server side.
	//
	//rkgfctx.SetHeaderToClient(ctx, rkgfctx.RequestIdKey, rkcommon.GenerateRequestId())

	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.GetQuery("name")),
	})
}
