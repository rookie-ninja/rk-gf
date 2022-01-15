// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgftrace is aa middleware of GoFrame framework for recording trace info of RPC
package rkgftrace

import (
	"github.com/gogf/gf/v2/net/ghttp"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...rkmidtrace.Option) ghttp.HandlerFunc {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkmid.EntryNameKey, set.GetEntryName())
		ctx.SetCtxVar(rkmid.TracerKey, set.GetTracer())
		ctx.SetCtxVar(rkmid.TracerProviderKey, set.GetProvider())
		ctx.SetCtxVar(rkmid.PropagatorKey, set.GetPropagator())

		beforeCtx := set.BeforeCtx(ctx.Request, false)
		set.Before(beforeCtx)

		// create request with new context
		ctx.Request = ctx.Request.WithContext(beforeCtx.Output.NewCtx)

		// add to context
		if beforeCtx.Output.Span != nil {
			traceId := beforeCtx.Output.Span.SpanContext().TraceID().String()
			rkgfctx.GetEvent(ctx).SetTraceId(traceId)
			ctx.Response.Header().Set(rkmid.HeaderTraceId, traceId)
			ctx.SetCtxVar(rkmid.SpanKey, beforeCtx.Output.Span)
		}

		ctx.Middleware.Next()

		afterCtx := set.AfterCtx(ctx.Response.Status, "")
		set.After(beforeCtx, afterCtx)
	}
}
