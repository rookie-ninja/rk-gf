// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgflimit

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	rkmidlimit "github.com/rookie-ninja/rk-entry/middleware/ratelimit"
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
		rkmidlimit.WithAlgorithm(rkmidlimit.TokenBucket),
		rkmidlimit.WithReqPerSec(1),
		rkmidlimit.WithReqPerSecByPath("ut-path", 1))
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
		rkmidlimit.WithAlgorithm(rkmidlimit.LeakyBucket),
		rkmidlimit.WithReqPerSec(1),
		rkmidlimit.WithReqPerSecByPath("ut-path", 1))
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
		rkmidlimit.WithGlobalLimiter(func() error {
			return fmt.Errorf("ut-error")
		}),
		rkmidlimit.WithLimiterByPath("/ut-path", func() error {
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
