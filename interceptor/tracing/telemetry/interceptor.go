// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgftrace is aa middleware of GoFrame framework for recording trace info of RPC
package rkgftrace

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)
		ctx.SetCtxVar(rkgfinter.RpcTracerKey, set.Tracer)
		ctx.SetCtxVar(rkgfinter.RpcTracerProviderKey, set.Provider)
		ctx.SetCtxVar(rkgfinter.RpcPropagatorKey, set.Propagator)

		span := before(ctx, set)
		defer span.End()

		ctx.Middleware.Next()

		after(ctx, span)
	}
}

func before(ctx *ghttp.Request, set *optionSet) oteltrace.Span {
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", ctx.Request)...),
		oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(ctx.Request)...),
		oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, ctx.RequestURI, ctx.Request)...),
		oteltrace.WithAttributes(localeToAttributes()...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}

	// 1: extract tracing info from request header
	spanCtx := oteltrace.SpanContextFromContext(
		set.Propagator.Extract(ctx.Request.Context(), propagation.HeaderCarrier(ctx.Request.Header)))

	spanName := ctx.RequestURI
	if len(spanName) < 1 {
		spanName = "rk-span-default"
	}

	// 2: start new span
	newRequestCtx, span := set.Tracer.Start(
		oteltrace.ContextWithRemoteSpanContext(ctx.Request.Context(), spanCtx),
		spanName, opts...)
	// 2.1: pass the span through the request context
	ctx.Request = ctx.Request.WithContext(newRequestCtx)

	// 3: read trace id, tracer, traceProvider, propagator and logger into event data and echo context
	rkgfctx.GetEvent(ctx).SetTraceId(span.SpanContext().TraceID().String())
	ctx.Response.Header().Set(rkgfctx.TraceIdKey, span.SpanContext().TraceID().String())

	ctx.SetCtxVar(rkgfinter.RpcSpanKey, span)
	return span
}

func after(ctx *ghttp.Request, span oteltrace.Span) {
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(ctx.Response.Status)
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(ctx.Response.Status)
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)
}

// Convert locale information into attributes.
func localeToAttributes() []attribute.KeyValue {
	res := []attribute.KeyValue{
		attribute.String(rkgfinter.Realm.Key, rkgfinter.Realm.String),
		attribute.String(rkgfinter.Region.Key, rkgfinter.Region.String),
		attribute.String(rkgfinter.AZ.Key, rkgfinter.AZ.String),
		attribute.String(rkgfinter.Domain.Key, rkgfinter.Domain.String),
	}

	return res
}
