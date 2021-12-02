// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfsec is a secure middleware for GoFrame framework
package rkgfsec

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

	// without options
	inter = Interceptor()
	server = startServer(t, userHandler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp.Header,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions)
	assert.Nil(t, server.Shutdown())

	// with options
	inter = Interceptor(
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(headerXForwardedProto, "https")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp.Header,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions,
		headerStrictTransportSecurity,
		headerContentSecurityPolicyReportOnly,
		headerReferrerPolicy)
	assert.Nil(t, server.Shutdown())

}

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
}

func containsHeader(t *testing.T, resp http.Header, headers ...string) {
	for _, v := range headers {
		assert.Contains(t, resp, v)
	}
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
