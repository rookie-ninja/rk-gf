// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgflimit

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

func TestInterceptor_WithoutOptions(t *testing.T) {
	defer assertNotPanic(t)

	inter := Interceptor()
	server := startServer(t, func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func TestInterceptor_WithTokenBucket(t *testing.T) {
	defer assertNotPanic(t)

	inter := Interceptor(
		WithAlgorithm(TokenBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))
	server := startServer(t, func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func TestInterceptor_WithLeakyBucket(t *testing.T) {
	defer assertNotPanic(t)

	inter := Interceptor(
		WithAlgorithm(LeakyBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))
	server := startServer(t, func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}

func TestInterceptor_WithUserLimiter(t *testing.T) {
	defer assertNotPanic(t)

	inter := Interceptor(
		WithGlobalLimiter(func(ctx *ghttp.Request) error {
			return fmt.Errorf("ut-error")
		}),
		WithLimiterByPath("/ut-path", func(ctx *ghttp.Request) error {
			return fmt.Errorf("ut-error")
		}))
	server := startServer(t, func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
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
