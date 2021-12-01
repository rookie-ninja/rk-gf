// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfinter provides common utility functions for middleware of GoFrame framework
package rkgfinter

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"net"
	"strings"
)

var (
	// Realm environment variable
	Realm = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	// Region environment variable
	Region = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	// AZ environment variable
	AZ = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	// Domain environment variable
	Domain = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	// LocalIp read local IP from localhost
	LocalIp = zap.String("localIp", rkcommon.GetLocalIP())
	// LocalHostname read hostname from localhost
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())
)

const (
	// RpcEntryNameKey entry name key
	RpcEntryNameKey = "gfEntryName"
	// RpcEntryNameValue entry name
	RpcEntryNameValue = "gf"
	// RpcEntryTypeValue entry type
	RpcEntryTypeValue = "gf"
	// RpcEventKey event key
	RpcEventKey = "gfEvent"
	// RpcLoggerKey logger key
	RpcLoggerKey = "gfLogger"
	// RpcTracerKey tracer key
	RpcTracerKey = "gfTracer"
	// RpcSpanKey span key
	RpcSpanKey = "gfSpan"
	// RpcTracerProviderKey trace provider key
	RpcTracerProviderKey = "gfTracerProvider"
	// RpcPropagatorKey propagator key
	RpcPropagatorKey = "gfPropagator"
	// RpcAuthorizationHeaderKey auth key
	RpcAuthorizationHeaderKey = "authorization"
	// RpcApiKeyHeaderKey api auth key
	RpcApiKeyHeaderKey = "X-API-Key"
	// RpcJwtTokenKey key of jwt token in context
	RpcJwtTokenKey = "gfJwt"
)

// GetRemoteAddressSet returns remote endpoint information set including IP, Port.
// We will do as best as we can to determine it.
// If fails, then just return default ones.
func GetRemoteAddressSet(ctx *ghttp.Request) (remoteIp, remotePort string) {
	remoteIp, remotePort = "0.0.0.0", "0"

	if ctx == nil || ctx.Request == nil {
		return
	}

	var err error
	if remoteIp, remotePort, err = net.SplitHostPort(ctx.Request.RemoteAddr); err != nil {
		return
	}

	forwardedRemoteIp := ctx.Request.Header.Get("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIp) > 0 {
		if forwardedRemoteIp == "::1" {
			forwardedRemoteIp = "localhost"
		}

		remoteIp = forwardedRemoteIp
	}

	if remoteIp == "::1" {
		remoteIp = "localhost"
	}

	return remoteIp, remotePort
}

// ShouldLog determines whether should log the RPC
func ShouldLog(ctx *ghttp.Request) bool {
	if ctx == nil || ctx.Request == nil {
		return false
	}

	// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
	if strings.HasPrefix(ctx.Request.URL.Path, "/rk/v1/assets") ||
		strings.HasPrefix(ctx.Request.URL.Path, "/rk/v1/tv") ||
		strings.HasPrefix(ctx.Request.URL.Path, "/sw/") {
		return false
	}

	return true
}

type noopWriter struct{}

func (w noopWriter) Write([]byte) (n int, err error) {
	return 0, nil
}

func NewNoopGLogger() *glog.Logger {
	return glog.NewWithWriter(noopWriter{})
}
