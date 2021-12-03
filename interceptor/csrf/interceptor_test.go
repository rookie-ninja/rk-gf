// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfcsrf

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
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

func TestInterceptor(t *testing.T) {
	// match 1
	inter := Interceptor(
		WithSkipper(func(ctx *ghttp.Request) bool {
			return true
		}))
	server := startServer(t, userHandler, inter)
	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 2.1
	inter = Interceptor()
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 2.2
	inter = Interceptor()
	server = startServer(t, userHandler, inter)
	client = getClient()
	client.SetCookie("_csrf", "ut-csrf-token")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 3.1
	inter = Interceptor()
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3.2
	inter = Interceptor()
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Post(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3.3
	inter = Interceptor()
	server = startServer(t, userHandler, inter)
	client = getClient()
	client.SetHeader(headerXCSRFToken, "ut-csrf-token")
	resp, err = client.Post(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 4.1
	inter = Interceptor(
		WithCookiePath("ut-path"))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-path")
	assert.Nil(t, server.Shutdown())

	// match 4.2
	inter = Interceptor(
		WithCookieDomain("ut-domain"))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-domain")
	assert.Nil(t, server.Shutdown())

	// match 4.3
	inter = Interceptor(
		WithCookieSameSite(http.SameSiteStrictMode))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "Strict")
	assert.Nil(t, server.Shutdown())
}
