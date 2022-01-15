// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfjwt

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-common/common"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/jwt"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

var userHandler = func(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
}

func TestInterceptor(t *testing.T) {
	// without options
	inter := Interceptor()
	server := startServer(t, userHandler, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// with parse token error
	parseTokenErrFunc := func(auth string) (*jwt.Token, error) {
		return nil, errors.New("ut-error")
	}
	inter = Interceptor(
		rkmidjwt.WithParseTokenFunc(parseTokenErrFunc))
	server = startServer(t, userHandler, inter)

	client = getClient()
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// happy case
	parseTokenErrFunc = func(auth string) (*jwt.Token, error) {
		return &jwt.Token{}, nil
	}
	inter = Interceptor(
		rkmidjwt.WithParseTokenFunc(parseTokenErrFunc))
	server = startServer(t, userHandler, inter)

	client = getClient()
	client.SetHeader(rkmid.HeaderAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	resp, err = client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
