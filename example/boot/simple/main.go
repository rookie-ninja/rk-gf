// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-gf/boot"
	"net/http"
)

// How to use embed.FS for:
//
// - boot.yaml
// - rkentry.DocsEntryType
// - rkentry.SWEntryType
// - rkentry.StaticFileHandlerEntryType
// - rkentry.CertEntry
//
// If we use embed.FS, then we only need one single binary file while packing.
// We suggest use embed.FS to pack swagger local file since rk-entry would use os.Getwd() to look for files
// if relative path was provided.
//
//go:embed docs
var docsFS embed.FS

func init() {
	rkentry.GlobalAppCtx.AddEmbedFS(rkentry.SWEntryType, "greeter", &docsFS)
}

//go:embed boot.yaml
var boot []byte

// @title RK Swagger for Mux
// @version 1.0
// @description This is a greeter service with rk-boot.
func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgf.RegisterGfEntryYAML(boot)

	// Get rkgf.GfEntry
	gfEntry := res["greeter"].(*rkgf.GfEntry)
	gfEntry.Server.BindHandler("/v1/greeter", Greeter)

	fmt.Println("I amhere")

	// Bootstrap gf entry
	gfEntry.Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gf entry
	gfEntry.Interrupt(context.Background())
}

// Greeter handler
// @Summary Greeter service
// @Id 1
// @version 1.0
// @produce application/json
// @Param name query string true "Input name"
// @Success 200 {object} GreeterResponse
// @Router /v1/greeter [get]
func Greeter(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.GetQuery("name").String()),
	})
}

type GreeterResponse struct {
	Message string
}
