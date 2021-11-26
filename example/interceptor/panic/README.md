# Panic interceptor
In this example, we will try to create GoFrame server with panic interceptor enabled.

Panic interceptor will add do the bellow actions.
- Recover from panic
- Convert interface to standard rkerror.ErrorResp style of error
- Set resCode to 500
- Print stacktrace
- Set [panic:1] into event as counters
- Add error into event

**Please make sure panic interceptor to be added at last in chain of interceptors.**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Example](#example)
  - [Start server](#start-server)
  - [Output](#output)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-gf package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-gf
```
### Code
```go
import     "github.com/rookie-ninja/rk-gf/interceptor/panic"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []ghttp.HandlerFunc{
        rkgfpanic.Interceptor(),
    }
```

## Example
We will enable log interceptor to monitor RPC.

### Start server
```shell script
$ go run greeter-server.go
```

### Output
- Server side log (zap & event)
```shell script
2021-11-26T04:59:55.551+0800    ERROR   panic/interceptor.go:37 panic occurs:
Panic manually!
1. Panic manually!
   1).  main.Greeter
        /Users/dongxuny/workspace/rk/rk-gf/example/interceptor/panic/greeter-server.go:76
   2).  github.com/rookie-ninja/rk-gf/interceptor/panic.Interceptor.func1
        /Users/dongxuny/workspace/rk/rk-gf/interceptor/panic/interceptor.go:45
   3).  github.com/rookie-ninja/rk-gf/interceptor/log/zap.Interceptor.func1
        /Users/dongxuny/workspace/rk/rk-gf/interceptor/log/zap/interceptor.go:29
        {"error": "[Internal Server Error] Panic manually!"}
```

```shell script
------------------------------------------------------------------------
endTime=2021-11-26T04:48:41.399958+08:00
startTime=2021-11-26T04:48:41.39935+08:00
elapsedNano=608128
timezone=CST
ids={"eventId":"cd458d5e-d2b0-489a-90b9-bfed844881f7"}
app={"appName":"rk","appVersion":"","entryName":"gf","entryType":"gf"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={"[Internal Server Error] Panic manually!":1}
counters={"panic":1}
pairs={}
timing={}
remoteAddr=localhost:58918
operation=/rk/v1/greeter
resCode=500
eventStatus=Ended
EOE
```
- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}
```

### Code
- [greeter-server.go](greeter-server.go)
