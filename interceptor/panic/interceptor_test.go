// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfpanic

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// panic
	handler := func(ctx *ghttp.Request) {
		panic(errors.New("error"))
	}
	inter := Interceptor(rkmidpanic.WithEntryNameAndType("ut-entry", "ut-type"))
	server := startServer(t, handler, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Nil(t, server.Shutdown())

	// without panic
	handler = func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}
	inter = Interceptor(rkmidpanic.WithEntryNameAndType("ut-entry", "ut-type"))
	server = startServer(t, handler, inter)

	client = getClient()
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

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
