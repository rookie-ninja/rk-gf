// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgfsec is a secure middleware for GoFrame framework
package rkgfsec

import (
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-gf/interceptor"
)

// Interceptor Add Secure interceptors.
func Interceptor(opts ...Option) ghttp.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *ghttp.Request) {
		ctx.SetCtxVar(rkgfinter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Middleware.Next()
			return
		}

		req := ctx.Request
		res := ctx.Response

		// Add X-XSS-Protection header
		if set.XSSProtection != "" {
			res.Header().Set(headerXXSSProtection, set.XSSProtection)
		}

		// Add X-Content-Type-Options header
		if set.ContentTypeNosniff != "" {
			res.Header().Set(headerXContentTypeOptions, set.ContentTypeNosniff)
		}

		// Add X-Frame-Options header
		if set.XFrameOptions != "" {
			res.Header().Set(headerXFrameOptions, set.XFrameOptions)
		}

		// Add Strict-Transport-Security header
		if (req.TLS != nil || (req.Header.Get(headerXForwardedProto) == "https")) && set.HSTSMaxAge != 0 {
			subdomains := ""
			if !set.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if set.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			res.Header().Set(headerStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", set.HSTSMaxAge, subdomains))
		}

		// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
		if set.ContentSecurityPolicy != "" {
			if set.CSPReportOnly {
				res.Header().Set(headerContentSecurityPolicyReportOnly, set.ContentSecurityPolicy)
			} else {
				res.Header().Set(headerContentSecurityPolicy, set.ContentSecurityPolicy)
			}
		}

		// Add Referrer-Policy header
		if set.ReferrerPolicy != "" {
			res.Header().Set(headerReferrerPolicy, set.ReferrerPolicy)
		}

		ctx.Middleware.Next()
		return
	}
}
