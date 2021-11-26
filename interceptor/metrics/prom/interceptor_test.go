// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfmetrics

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	inter := Interceptor(WithEntryNameAndType("ut-entry", "ut-type"))
	server := startServer(t, func(ctx *ghttp.Request) {
		ctx.Response.WriteHeader(http.StatusOK)
	}, inter)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/ut")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, server.Shutdown())
}
