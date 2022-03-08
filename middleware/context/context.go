// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfctx defines utility functions and variables used by GoFrame middleware
package rkgfctx

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

const (
	// RequestIdKey is the header key sent to client
	RequestIdKey = "X-Request-Id"
	// TraceIdKey is the header sent to client
	TraceIdKey = "X-Trace-Id"
)

var (
	noopTracerProvider = trace.NewNoopTracerProvider()
	noopEvent          = rkquery.NewEventFactory().CreateEventNoop()
)

// GetIncomingHeaders extract call-scoped incoming headers
func GetIncomingHeaders(ctx *ghttp.Request) http.Header {
	return ctx.Request.Header
}

// AddHeaderToClient headers that would be sent to client.
// Values would be merged.
func AddHeaderToClient(ctx *ghttp.Request, key, value string) {
	if ctx == nil || ctx.Response.Writer == nil {
		return
	}

	header := ctx.Response.Writer.Header()
	header.Add(key, value)
}

// SetHeaderToClient headers that would be sent to client.
// Values would be overridden.
func SetHeaderToClient(ctx *ghttp.Request, key, value string) {
	if ctx == nil || ctx.Response.Writer == nil {
		return
	}
	header := ctx.Response.Writer.Header()
	header.Set(key, value)
}

// GetEvent extract takes the call-scoped EventData from middleware.
func GetEvent(ctx *ghttp.Request) rkquery.Event {
	if raw := ctx.GetCtxVar(rkmid.EventKey).Interface(); raw != nil {
		return raw.(rkquery.Event)
	}

	return noopEvent
}

// GetLogger extract takes the call-scoped zap logger from middleware.
func GetLogger(ctx *ghttp.Request) *zap.Logger {
	if raw := ctx.GetCtxVar(rkmid.LoggerKey).Interface(); raw != nil {
		requestId := GetRequestId(ctx)
		traceId := GetTraceId(ctx)
		fields := make([]zap.Field, 0)
		if len(requestId) > 0 {
			fields = append(fields, zap.String("requestId", requestId))
		}
		if len(traceId) > 0 {
			fields = append(fields, zap.String("traceId", traceId))
		}

		return raw.(*zap.Logger).With(fields...)
	}

	return rklogger.NoopLogger
}

// GetRequestId extract request id from context.
// If user enabled meta interceptor, then a random request Id would e assigned and set to context as value.
// If user called AddHeaderToClient() with key of RequestIdKey, then a new request id would be updated.
func GetRequestId(ctx *ghttp.Request) string {
	if ctx == nil || ctx.Response.Writer == nil {
		return ""
	}

	return ctx.Response.Writer.Header().Get(RequestIdKey)
}

// GetTraceId extract trace id from context.
func GetTraceId(ctx *ghttp.Request) string {
	if ctx == nil || ctx.Response.Writer == nil {
		return ""
	}

	return ctx.Response.Writer.Header().Get(TraceIdKey)
}

// GetEntryName extract entry name from context.
func GetEntryName(ctx *ghttp.Request) string {
	if ctx == nil {
		return ""
	}

	if raw := ctx.GetCtxVar(rkmid.EntryNameKey).Interface(); raw != nil {
		return raw.(string)
	}

	return ""
}

// GetTraceSpan extract the call-scoped span from context.
func GetTraceSpan(ctx *ghttp.Request) trace.Span {
	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(context.TODO(), "noop-span")

	if ctx == nil || ctx.Request == nil {
		return span
	}

	_, span = noopTracerProvider.Tracer("rk-trace-noop").Start(ctx.Request.Context(), "noop-span")

	if raw := ctx.GetCtxVar(rkmid.SpanKey).Interface(); raw != nil {
		return raw.(trace.Span)
	}

	return span
}

// GetTracer extract the call-scoped tracer from context.
func GetTracer(ctx *ghttp.Request) trace.Tracer {
	if ctx == nil {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	if raw := ctx.GetCtxVar(rkmid.TracerKey).Interface(); raw != nil {
		return raw.(trace.Tracer)
	}

	return noopTracerProvider.Tracer("rk-trace-noop")
}

// GetTracerProvider extract the call-scoped tracer provider from context.
func GetTracerProvider(ctx *ghttp.Request) trace.TracerProvider {
	if ctx == nil {
		return noopTracerProvider
	}

	if raw := ctx.GetCtxVar(rkmid.TracerProviderKey).Interface(); raw != nil {
		return raw.(trace.TracerProvider)
	}

	return noopTracerProvider
}

// GetTracerPropagator extract takes the call-scoped propagator from middleware.
func GetTracerPropagator(ctx *ghttp.Request) propagation.TextMapPropagator {
	if ctx == nil {
		return nil
	}

	if raw := ctx.GetCtxVar(rkmid.PropagatorKey).Interface(); raw != nil {
		return raw.(propagation.TextMapPropagator)
	}

	return nil
}

// InjectSpanToHttpRequest inject span to http request
func InjectSpanToHttpRequest(ctx *ghttp.Request, req *http.Request) {
	if req == nil {
		return
	}

	newCtx := trace.ContextWithRemoteSpanContext(req.Context(), GetTraceSpan(ctx).SpanContext())

	if propagator := GetTracerPropagator(ctx); propagator != nil {
		propagator.Inject(newCtx, propagation.HeaderCarrier(req.Header))
	}
}

// NewTraceSpan start a new span
func NewTraceSpan(ctx *ghttp.Request, name string) trace.Span {
	tracer := GetTracer(ctx)
	newCtx, span := tracer.Start(ctx.Request.Context(), name)

	ctx.Request = ctx.Request.WithContext(newCtx)

	GetEvent(ctx).StartTimer(name)

	return span
}

// EndTraceSpan end span
func EndTraceSpan(ctx *ghttp.Request, span trace.Span, success bool) {
	if success {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}

	span.End()
}

// GetJwtToken return jwt.Token if exists
func GetJwtToken(ctx *ghttp.Request) *jwt.Token {
	if ctx == nil {
		return nil
	}

	if raw := ctx.GetCtxVar(rkmid.JwtTokenKey); raw != nil {
		if res, ok := raw.Interface().(*jwt.Token); ok {
			return res
		}
	}

	return nil
}

// GetCsrfToken return csrf token if exists
func GetCsrfToken(ctx *ghttp.Request) string {
	if ctx == nil {
		return ""
	}

	if raw := ctx.GetCtxVar(rkmid.CsrfTokenKey); raw != nil {
		return raw.String()
	}

	return ""
}
