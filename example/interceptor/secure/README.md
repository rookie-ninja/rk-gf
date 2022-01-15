# Secure interceptor
In this example, we will try to create GoFrame server with secure interceptor enabled.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Options](#options)
- [Example](#example)
  - [Start server](#start-server)
  - [Send request](#send-request)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-gf package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-gf
```

### Code
Add rkgfsec.Interceptor() CORS with option.

```go
import     "github.com/rookie-ninja/rk-gf/interceptor/secure"
```

```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []ghttp.HandlerFunc{
        rkgfsec.Interceptor(),
    }
```

## Options

| Name | Description | Default Values |
| ---- | ---- | ---- |
| rkmidsec.WithEntryNameAndType(entryName, entryType string) | Optional. Provide entry name and type if there are multiple secure interceptors needs to be used. | gin, gin |
| rkmidsec.WithXSSProtection(string) | Optional. X-XSS-Protection header value | "1; mode=block" |
| rkmidsec.WithContentTypeNosniff(string) | Optional. X-Content-Type-Options header value | nosniff |
| rkmidsec.WithXFrameOptions(string) | Optional. X-Frame-Options header value | SAMEORIGIN |
| rkmidsec.WithHSTSMaxAge(int) | Optional, Strict-Transport-Security header value | 0 |
| rkmidsec.WithHSTSExcludeSubdomains(bool) | Optional, excluding subdomains of HSTS | false |
| rkmidsec.WithHSTSPreloadEnabled(bool) | Optional, enabling HSTS preload | false |
| rkmidsec.WithContentSecurityPolicy(string) | Optional, Content-Security-Policy header value | "" |
| rkmidsec.WithCSPReportOnly(bool) | Optional, Content-Security-Policy-Report-Only header value | false |
| rkmidsec.WithReferrerPolicy(string) | Optional, Referrer-Policy header value | "" | 
| rkmidsec.WithIgnorePrefix([]string) | Optional, provide ignoring path prefix. | [] |

```go
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []ghttp.HandlerFunc{
		rkgfsec.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidsec.WithEntryNameAndType("greeter", "gin"),
			//
			// X-XSS-Protection header value.
			// Optional. Default value "1; mode=block".
			//rkmidsec.WithXSSProtection("my-value"),
			//
			// X-Content-Type-Options header value.
			// Optional. Default value "nosniff".
			//rkmidsec.WithContentTypeNosniff("my-value"),
			//
			// X-Frame-Options header value.
			// Optional. Default value "SAMEORIGIN".
			//rkmidsec.WithXFrameOptions("my-value"),
			//
			// Optional, Strict-Transport-Security header value.
			//rkmidsec.WithHSTSMaxAge(1),
			//
			// Optional, excluding subdomains of HSTS, default is false
			//rkmidsec.WithHSTSExcludeSubdomains(true),
			//
			// Optional, enabling HSTS preload, default is false
			//rkmidsec.WithHSTSPreloadEnabled(true),
			//
			// Content-Security-Policy header value.
			// Optional. Default value "".
			//rkmidsec.WithContentSecurityPolicy("my-value"),
			//
			// Content-Security-Policy-Report-Only header value.
			// Optional. Default value false.
			//rkmidsec.WithCSPReportOnly(true),
			//
			// Referrer-Policy header value.
			// Optional. Default value "".
			//rkmidsec.WithReferrerPolicy("my-value"),
			//
			// Ignoring path prefix.
			//rkmidsec.WithIgnorePrefix("/rk/v1"),
		),
	}
```

## Example
### Start server
```shell script
$ go run greeter-server.go
```

### Send request
```shell script
$ curl -vs localhost:8080/rk/v1/greeter
...
< HTTP/1.1 200 OK
< Content-Type: application/json
< Server: GoFrame HTTP Server
< X-Content-Type-Options: nosniff
< X-Frame-Options: SAMEORIGIN
< X-Xss-Protection: 1; mode=block
< Date: Thu, 02 Dec 2021 11:04:10 GMT
< Content-Length: 31
...
{"Message":"Received request!"}
```

### Code
- [greeter-server.go](greeter-server.go)
