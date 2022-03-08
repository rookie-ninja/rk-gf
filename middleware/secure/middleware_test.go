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
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
	"github.com/rookie-ninja/rk-gf/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	// without options
	inter := Middleware()
	server := startServer(t, userHandler, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp.Header,
		rkmid.HeaderXXSSProtection,
		rkmid.HeaderXContentTypeOptions,
		rkmid.HeaderXFrameOptions)
	assert.Nil(t, server.Shutdown())

	// with options
	inter = Middleware(
		rkmidsec.WithXSSProtection("ut-xss"),
		rkmidsec.WithContentTypeNosniff("ut-sniff"),
		rkmidsec.WithXFrameOptions("ut-frame"),
		rkmidsec.WithHSTSMaxAge(10),
		rkmidsec.WithHSTSExcludeSubdomains(true),
		rkmidsec.WithHSTSPreloadEnabled(true),
		rkmidsec.WithContentSecurityPolicy("ut-policy"),
		rkmidsec.WithCSPReportOnly(true),
		rkmidsec.WithReferrerPolicy("ut-ref"),
		rkmidsec.WithPathToIgnore("ut-prefix"))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderXForwardedProto, "https")
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp.Header,
		rkmid.HeaderXXSSProtection,
		rkmid.HeaderXContentTypeOptions,
		rkmid.HeaderXFrameOptions,
		rkmid.HeaderStrictTransportSecurity,
		rkmid.HeaderContentSecurityPolicyReportOnly,
		rkmid.HeaderReferrerPolicy)
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
	server := g.Server(rkmid.GenerateRequestId())
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
