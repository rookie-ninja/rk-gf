// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgfcors

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	rkcommon "github.com/rookie-ninja/rk-common/common"
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
	// with skipper
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

	// with empty option, all request will be passed
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
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
	inter = Interceptor(WithAllowOrigins("http://do-not-pass-through"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// match 3
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.Nil(t, server.Shutdown())

	// match 3.1
	inter = Interceptor(WithAllowCredentials(true))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(headerAccessControlAllowCredentials))
	assert.Nil(t, server.Shutdown())

	// match 3.2
	inter = Interceptor(
		WithAllowCredentials(true),
		WithExposeHeaders("expose"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(headerAccessControlAllowCredentials))
	assert.Equal(t, "expose", resp.Header.Get(headerAccessControlExposeHeaders))
	assert.Nil(t, server.Shutdown())

	// match 4
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header.Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(headerAccessControlAllowMethods))
	assert.Nil(t, server.Shutdown())

	// match 4.1
	inter = Interceptor(WithAllowCredentials(true))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header.Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(headerAccessControlAllowMethods))
	assert.Equal(t, "true", resp.Header.Get(headerAccessControlAllowCredentials))
	assert.Nil(t, server.Shutdown())

	// match 4.2
	inter = Interceptor(WithAllowHeaders("ut-header"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header.Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(headerAccessControlAllowMethods))
	assert.Equal(t, "ut-header", resp.Header.Get(headerAccessControlAllowHeaders))
	assert.Nil(t, server.Shutdown())

	// match 4.3
	inter = Interceptor(WithMaxAge(1))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerOrigin, originHeaderValue)
	resp, err = client.Options(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header.Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(headerAccessControlAllowMethods))
	assert.Equal(t, "1", resp.Header.Get(headerAccessControlMaxAge))
	assert.Nil(t, server.Shutdown())
}

func startServer(t *testing.T, usherHandler ghttp.HandlerFunc, inters ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(rkcommon.GenerateRequestId())
	server.SetPort(8081)
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
	client.SetPrefix("http://127.0.0.1:8081")

	return client
}
