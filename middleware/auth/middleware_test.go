// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfauth

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/rookie-ninja/rk-gf/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	// with ignoring path
	handler := func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"),
		rkmidauth.WithPathToIgnore("/ut"))
	server := startServer(t, handler, inter)

	client := getClient()
	client.SetHeader(rkmid.HeaderAuthorization, "invalid")
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid basic auth format
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderAuthorization, "invalid")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid basic auth cred
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderAuthorization, fmt.Sprintf("%s invalid", "Basic"))
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid api key
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderApiKey, "invalid")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with missing auth header
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func startServer(t *testing.T, usherHandler ghttp.HandlerFunc, inters ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(rkmid.GenerateRequestId(nil))
	server.SetPort(8080)
	server.SetDumpRouterMap(false)
	server.BindMiddlewareDefault(inters...)
	server.BindHandler("/ut", usherHandler)
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
