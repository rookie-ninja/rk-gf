// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfauth

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestInterceptor(t *testing.T) {
	// with ignoring path
	handler := func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"),
		WithIgnorePrefix("ut"))
	server := startServer(t, handler, inter)

	client := getClient()
	client.SetHeader(rkgfinter.RpcAuthorizationHeaderKey, "invalid")
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid basic auth format
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkgfinter.RpcAuthorizationHeaderKey, "invalid")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid basic auth cred
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkgfinter.RpcAuthorizationHeaderKey, fmt.Sprintf("%s invalid", typeBasic))
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with invalid api key
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	client.SetHeader(rkgfinter.RpcApiKeyHeaderKey, "invalid")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with missing auth header
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))
	server = startServer(t, handler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func startServer(t *testing.T, usherHandler ghttp.HandlerFunc, inters ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(rkcommon.GenerateRequestId())
	server.SetPort(8080)
	server.SetDumpRouterMap(false)
	server.BindMiddlewareDefault(inters...)
	server.BindHandler("/ut", usherHandler)
	server.SetLogger(rkgfinter.NewNoopGLogger())
	assert.Nil(t, server.Start())

	return server
}

func getClient() *ghttp.Client {
	time.Sleep(100 * time.Millisecond)
	client := g.Client()
	client.SetBrowserMode(true)
	client.SetPrefix("http://127.0.0.1:8080")

	return client
}
