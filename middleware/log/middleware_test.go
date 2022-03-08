// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgflog

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"github.com/rookie-ninja/rk-gf/middleware"
	"github.com/rookie-ninja/rk-gf/middleware/context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
}

func TestMiddleware_WithShouldNotLog(t *testing.T) {
	defer assertNotPanic(t)

	inter := Middleware(
		rkmidlog.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidlog.WithLoggerEntry(rkentry.LoggerEntryNoop),
		rkmidlog.WithEventEntry(rkentry.EventEntryNoop))
	server := startServer(t, userHandler, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/assets")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
	time.Sleep(10 * time.Millisecond)
}

func TestMiddleware_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	inter := Middleware(
		rkmidlog.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidlog.WithLoggerEntry(rkentry.LoggerEntryNoop),
		rkmidlog.WithEventEntry(rkentry.EventEntryNoop))
	server := startServer(t, func(ctx *ghttp.Request) {
		event := rkgfctx.GetEvent(ctx)
		assert.NotNil(t, event)
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func startServer(t *testing.T, usherHandler ghttp.HandlerFunc, inters ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(rkmid.GenerateRequestId())
	server.SetPort(8080)
	server.SetDumpRouterMap(false)
	server.BindMiddlewareDefault(inters...)
	server.BindHandler("/ut", usherHandler)
	server.BindHandler("/rk/v1/assets", usherHandler)
	server.SetLogger(rkgfinter.NewNoopGLogger())
	assert.Nil(t, server.Start())

	return server
}

func getClient() *gclient.Client {
	time.Sleep(100 * time.Millisecond)
	client := g.Client()
	client.SetBrowserMode(true)
	client.SetPrefix("http://127.0.0.1:8080")

	return client
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
