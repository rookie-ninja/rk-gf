// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-gf/boot"
	"github.com/rookie-ninja/rk-gf/middleware/context"
	"net/http"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgf.RegisterGfEntryYAML(boot)

	entry := res["greeter"].(*rkgf.GfEntry)
	entry.Server.BindHandler("/rk/v1/greeter", Greeter)

	// Bootstrap gf entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gf entry
	res["greeter"].Interrupt(context.Background())
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
