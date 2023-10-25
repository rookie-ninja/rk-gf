// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfcsrf

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/rookie-ninja/rk-gf/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
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

func TestMiddleware(t *testing.T) {
	// match 2.1
	inter := Middleware()
	server := startServer(t, userHandler, inter)
	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 2.2
	inter = Middleware()
	server = startServer(t, userHandler, inter)
	client = getClient()
	client.SetCookie("_csrf", "ut-csrf-token")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 3.1
	inter = Middleware()
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3.2
	inter = Middleware()
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Post(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3.3
	inter = Middleware()
	server = startServer(t, userHandler, inter)
	client = getClient()
	client.SetHeader(rkmid.HeaderXCSRFToken, "ut-csrf-token")
	resp, err = client.Post(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 4.1
	inter = Middleware(
		rkmidcsrf.WithCookiePath("ut-path"))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-path")
	assert.Nil(t, server.Shutdown())

	// match 4.2
	inter = Middleware(
		rkmidcsrf.WithCookieDomain("ut-domain"))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-domain")
	assert.Nil(t, server.Shutdown())

	// match 4.3
	inter = Middleware(
		rkmidcsrf.WithCookieSameSite(http.SameSiteStrictMode))
	server = startServer(t, userHandler, inter)
	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "Strict")
	assert.Nil(t, server.Shutdown())
}
