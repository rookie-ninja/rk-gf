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
	rkmidjwt "github.com/rookie-ninja/rk-entry/middleware/jwt"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	rkgfjwt "github.com/rookie-ninja/rk-gf/interceptor/jwt"
	"net/http"
	"time"
)

// In this example, we will start a new GoFrame server with JWT interceptor enabled.
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
		//rkgflog.Interceptor(),
		rkgfjwt.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkginjwt.WithEntryNameAndType("greeter", "gin"),
			//
			// Required, provide signing key.
			rkmidjwt.WithSigningKey([]byte("my-secret")),
			//
			// rkmidjwt.WithIgnorePrefix(""),
			//
			// Optional, provide skipper function
			//rkmidjwt.WithSkipper(func(e *gin.Context) bool {
			//	return true
			//}),
			//
			// Optional, provide token parse function, default one will be assigned.
			//rkmidjwt.WithParseTokenFunc(func(auth string, ctx *gin.Context) (*jwt.Token, error) {
			//	return nil, nil
			//}),
			//
			// Optional, provide key function, default one will be assigned.
			//rkmidjwt.WithKeyFunc(func(token *jwt.Token) (interface{}, error) {
			//	return nil, nil
			//}),
			//
			// Optional, default is Bearer
			//rkmidjwt.WithAuthScheme("Bearer"),
			//
			// Optional
			//rkmidjwt.WithTokenLookup("header:my-jwt-header-key"),
			//
			// Optional, default is HS256
			//rkmidjwt.WithSigningAlgorithm(rkginjwt.AlgorithmHS256),
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
		Message: fmt.Sprintf("Is token valid:%v!", rkgfctx.GetJwtToken(ctx).Valid),
	})
}
