// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgflog is a middleware for GoFrame framework for logging RPC.
package rkgflog

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		before(ctx, set)

		ctx.Middleware.Next()

		after(ctx)
	}
}

func before(ctx *ghttp.Request, set *optionSet) {
	var event rkquery.Event
	if rkgfinter.ShouldLog(ctx) {
		event = set.eventLoggerEntry.GetEventFactory().CreateEvent(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.EntryName),
			rkquery.WithEntryType(set.EntryType))
	} else {
		event = set.eventLoggerEntry.GetEventFactory().CreateEventNoop()
	}

	event.SetStartTime(time.Now())

	remoteIp, remotePort := rkgfinter.GetRemoteAddressSet(ctx)
	// handle remote address
	event.SetRemoteAddr(remoteIp + ":" + remotePort)

	payloads := []zap.Field{
		zap.String("apiPath", ctx.Request.URL.Path),
		zap.String("apiMethod", ctx.Request.Method),
		zap.String("apiQuery", ctx.Request.URL.RawQuery),
		zap.String("apiProtocol", ctx.Request.Proto),
		zap.String("userAgent", ctx.Request.UserAgent()),
	}

	// handle payloads
	event.AddPayloads(payloads...)

	// handle operation
	event.SetOperation(ctx.Request.URL.Path)

	ctx.SetCtxVar(rkgfinter.RpcEventKey, event)
	ctx.SetCtxVar(rkgfinter.RpcLoggerKey, set.ZapLogger)
}

func after(ctx *ghttp.Request) {
	event := rkgfctx.GetEvent(ctx)

	if requestId := rkgfctx.GetRequestId(ctx); len(requestId) > 0 {
		event.SetEventId(requestId)
		event.SetRequestId(requestId)
	}

	if traceId := rkgfctx.GetTraceId(ctx); len(traceId) > 0 {
		event.SetTraceId(traceId)
	}

	event.SetResCode(strconv.Itoa(ctx.Response.Status))
	event.SetEndTime(time.Now())
	event.Finish()
}
