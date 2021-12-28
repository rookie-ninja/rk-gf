// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gf/boot"
	"net/http"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

	// Bootstrap gf entry from boot config
	res := rkgf.RegisterGfEntriesWithConfig("example/boot/simple/boot.yaml")

	// Get rkgf.GfEntry
	gfEntry := res["greeter"].(*rkgf.GfEntry)
	gfEntry.Server.BindHandler("/v1/greeter", func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.WriteJson(map[string]string{
			"message": "Hello!",
		})
	})

	// Bootstrap gf entry
	gfEntry.Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gf entry
	gfEntry.Interrupt(context.Background())
}
