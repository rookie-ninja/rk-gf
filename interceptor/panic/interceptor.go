// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfpanic is a middleware of GoFrame framework for recovering from panic
package rkgfpanic

import (
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/context"
	"go.uber.org/zap"
	"net/http"
)

// Interceptor returns a ghttp.HandlerFunc (middleware)
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		defer func() {
			if err := ctx.GetError(); err != nil {
				var res *rkerror.ErrorResp
				if re, ok := err.(error); ok {
					res = rkerror.FromError(re)
				} else {
					res = rkerror.New(rkerror.WithMessage(fmt.Sprintf("%v", err)))
				}

				rkgfctx.GetEvent(ctx).SetCounter("panic", 1)
				rkgfctx.GetEvent(ctx).AddErr(res.Err)
				rkgfctx.GetLogger(ctx).Error(fmt.Sprintf("panic occurs:\n%+v", err), zap.Error(res.Err))

				ctx.Response.ClearBuffer()
				ctx.Response.WriteHeader(http.StatusInternalServerError)
				ctx.Response.Write(res)
			}
		}()

		ctx.Middleware.Next()
	}
}
