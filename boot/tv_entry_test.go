// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"bytes"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewTvEntry(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	assert.Equal(t, TvEntryNameDefault, entry.GetName())
	assert.Equal(t, TvEntryType, entry.GetType())
	assert.Equal(t, TvEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestTvEntry_Bootstrap(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	ctx := context.WithValue(context.Background(), bootstrapEventIdKey, "ut")
	entry.Bootstrap(ctx)
}

func TestTvEntry_Interrupt(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	ctx := context.WithValue(context.Background(), bootstrapEventIdKey, "ut")
	entry.Interrupt(ctx)
}

func TestTvEntry_TV(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))
	entry.Bootstrap(context.TODO())

	defer assertNotPanic(t)
	// With nil context
	entry.TV(nil)

	// With all paths
	server := startServer(t, "/rk/v1/tv/*item", entry.TV)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/tv/")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// apis
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/apis")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// entries
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/entries")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// configs
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/configs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// certs
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/certs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// os
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/os")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// env
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/env")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// prometheus
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/prometheus")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// logs
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/logs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// deps
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/deps")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// license
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/license")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// info
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/info")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// git
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/git")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	// unknown
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/unknown")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))

	assert.Nil(t, server.Shutdown())
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

func startServer(t *testing.T, path string, usherHandler ghttp.HandlerFunc, inters ...ghttp.HandlerFunc) *ghttp.Server {
	server := g.Server(rkcommon.GenerateRequestId())
	server.SetPort(8080)
	server.SetDumpRouterMap(false)
	server.BindMiddlewareDefault(inters...)
	server.BindHandler(path, usherHandler)
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

func bodyToString(body io.ReadCloser) string {
	buffer := bytes.Buffer{}
	buffer.ReadFrom(body)

	return buffer.String()
}
