// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgfcors

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/common"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const originHeaderValue = "http://ut-origin"

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
}

func TestInterceptor(t *testing.T) {
	// with empty option, all request will be passed
	inter := Interceptor()
	server := startServer(t, userHandler, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 1.1
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 1.2
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 2.1
	inter = Interceptor(rkmidcors.WithAllowOrigins("http://do-not-pass-through"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.Nil(t, server.Shutdown())

	// match 3.1
	inter = Interceptor(rkmidcors.WithAllowCredentials(true))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(rkmid.HeaderAccessControlAllowCredentials))
	assert.Nil(t, server.Shutdown())

	// match 3.2
	inter = Interceptor(
		rkmidcors.WithAllowCredentials(true),
		rkmidcors.WithExposeHeaders("expose"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(rkmid.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "expose", resp.Header.Get(rkmid.HeaderAccessControlExposeHeaders))
	assert.Nil(t, server.Shutdown())

	// match 4
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		rkmid.HeaderAccessControlRequestMethod,
		rkmid.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(rkmid.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(rkmid.HeaderAccessControlAllowMethods))
	assert.Nil(t, server.Shutdown())

	// match 4.1
	inter = Interceptor(rkmidcors.WithAllowCredentials(true))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		rkmid.HeaderAccessControlRequestMethod,
		rkmid.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(rkmid.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(rkmid.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", resp.Header.Get(rkmid.HeaderAccessControlAllowCredentials))
	assert.Nil(t, server.Shutdown())

	// match 4.2
	inter = Interceptor(rkmidcors.WithAllowHeaders("ut-header"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		rkmid.HeaderAccessControlRequestMethod,
		rkmid.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(rkmid.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(rkmid.HeaderAccessControlAllowMethods))
	assert.Equal(t, "ut-header", resp.Header.Get(rkmid.HeaderAccessControlAllowHeaders))
	assert.Nil(t, server.Shutdown())

	// match 4.3
	inter = Interceptor(rkmidcors.WithMaxAge(1))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		rkmid.HeaderAccessControlRequestMethod,
		rkmid.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(rkmid.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(rkmid.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(rkmid.HeaderAccessControlAllowMethods))
	assert.Equal(t, "1", resp.Header.Get(rkmid.HeaderAccessControlMaxAge))
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
