// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfmetrics is a middleware for GoFrame framework which record prometheus metrics for RPC
package rkgfmetrics

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"time"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		// start timer
		startTime := time.Now()

		ctx.Middleware.Next()

		// end timer
		elapsed := time.Now().Sub(startTime)

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if rkgfinter.ShouldLog(ctx) {
			if durationMetrics := GetServerDurationMetrics(ctx); durationMetrics != nil {
				durationMetrics.Observe(float64(elapsed.Nanoseconds()))
			}

			if resCodeMetrics := GetServerResCodeMetrics(ctx); resCodeMetrics != nil {
				resCodeMetrics.Inc()
			}
		}
	}
}
