// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gf/interceptor/metrics/prom"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func TestWithNameCommonService_WithEmptyString(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService(""))

	assert.NotEmpty(t, entry.GetName())
}

func TestWithNameCommonService_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("unit-test"))

	assert.Equal(t, "unit-test", entry.GetName())
}

func TestWithEventLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCommonService_HappyCase(t *testing.T) {
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(eventLoggerEntry))

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
}

func TestNewCommonServiceEntry_WithoutOptions(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestNewCommonServiceEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry),
		WithEventLoggerEntryCommonService(eventLoggerEntry),
		WithNameCommonService("ut"))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, "ut", entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestCommonServiceEntry_Bootstrap_WithoutRouter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Bootstrap_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Interrupt_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Interrupt(context.Background())
}

func TestCommonServiceEntry_GetName_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("ut"))

	assert.Equal(t, "ut", entry.GetName())
}

func TestCommonServiceEntry_GetType_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.Equal(t, "CommonServiceEntry", entry.GetType())
}

func TestCommonServiceEntry_String_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotEmpty(t, entry.String())
}

func TestCommonServiceEntry_Healthy_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)

	entry := NewCommonServiceEntry()
	entry.Healthy(nil)
}

func TestCommonServiceEntry_Healthy_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/healthy", entry.Healthy)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/healthy")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_GC_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	entry.Gc(nil)
}

func TestCommonServiceEntry_GC_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/gc", entry.Gc)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/gc")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Info_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	entry.Info(nil)
}

func TestCommonServiceEntry_Info_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/info", entry.Info)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/info")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Config_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Configs(nil)
}

func TestCommonServiceEntry_Config_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	vp := viper.New()
	vp.Set("unit-test-key", "unit-test-value")

	viperEntry := rkentry.RegisterConfigEntry(
		rkentry.WithNameConfig("unit-test"),
		rkentry.WithViperInstanceConfig(vp))

	rkentry.GlobalAppCtx.AddConfigEntry(viperEntry)

	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/configs", entry.Configs)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/configs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	respStr := bodyToString(resp.Body)
	assert.NotEmpty(t, respStr)
	assert.Contains(t, respStr, "unit-test-key")
	assert.Contains(t, respStr, "unit-test-value")
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_APIs_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Apis(nil)
}

func TestCommonServiceEntry_APIs_WithEmptyEntries(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/apis", entry.Apis)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/apis")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_APIs_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-echo"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	gfEntry.Server.BindHandler("/ut-test", func(request *ghttp.Request) {})

	server := startServer(t, "/rk/v1/apis", entry.Apis)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/apis")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Sys_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Sys(nil)
}

func TestCommonServiceEntry_Sys_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	server := startServer(t, "/rk/v1/sys", entry.Sys)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/sys")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Req_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-echo"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	gfEntry.AddInterceptor(rkgfmetrics.Interceptor(
		rkgfmetrics.WithRegisterer(prometheus.NewRegistry())))

	server := startServer(t, "/rk/v1/req", entry.Req)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/req")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Req_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Req(nil)
}

func TestCommonServiceEntry_Entries_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Entries(nil)
}

func TestCommonServiceEntry_Entries_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	server := startServer(t, "/rk/v1/entries", entry.Entries)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/entries")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Certs_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Certs(nil)
}

func TestCommonServiceEntry_Certs_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)
	rkentry.RegisterCertEntry(rkentry.WithNameCert("ut-cert"))
	certEntry := rkentry.GlobalAppCtx.GetCertEntry("ut-cert")
	certEntry.Retriever = &rkentry.CredRetrieverLocalFs{}
	certEntry.Store = &rkentry.CertStore{}

	server := startServer(t, "/rk/v1/certs", entry.Certs)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/certs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Logs_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Logs(nil)
}

func TestCommonServiceEntry_Logs_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	server := startServer(t, "/rk/v1/logs", entry.Logs)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/logs")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Deps_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Deps(nil)
}

func TestCommonServiceEntry_Deps_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	server := startServer(t, "/rk/v1/deps", entry.Deps)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/deps")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_License_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.License(nil)
}

func TestCommonServiceEntry_License_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	server := startServer(t, "/rk/v1/license", entry.License)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/license")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Readme_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Readme(nil)
}

func TestCommonServiceEntry_Readme_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)

	server := startServer(t, "/rk/v1/readme", entry.Readme)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/readme")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestCommonServiceEntry_Git_WithNilContext(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()
	entry.Git(nil)
}

func TestCommonServiceEntry_Git_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewCommonServiceEntry()

	gfEntry := RegisterGfEntry(
		WithCommonServiceEntryGf(entry),
		WithNameGf("unit-test-gf"))
	rkentry.GlobalAppCtx.AddEntry(gfEntry)
	rkentry.GlobalAppCtx.SetRkMetaEntry(&rkentry.RkMetaEntry{
		RkMeta: &rkcommon.RkMeta{
			Git: &rkcommon.Git{
				Commit: &rkcommon.Commit{
					Committer: &rkcommon.Committer{},
				},
			},
		},
	})

	server := startServer(t, "/rk/v1/git", entry.Git)

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/git")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestGetEntry_WithNilContext(t *testing.T) {
	assert.Nil(t, getEntry(nil))
}

func TestConstructSwUrl_WithNilEntry(t *testing.T) {
	assert.Equal(t, "N/A", constructSwUrl(nil, nil))
}

func TestConstructSwUrl_WithNilContext(t *testing.T) {
	path := "ut-sw"
	port := 1111
	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGfEntry(WithSwEntryGf(sw), WithPortGf(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, nil))
}

func TestConstructSwUrl_WithNilRequest(t *testing.T) {
	path := "ut-sw"
	port := 1111

	ctx := &ghttp.Request{
		Request: &http.Request{
			Host: fmt.Sprintf("localhost:%d", port),
		},
	}

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGfEntry(WithSwEntryGf(sw), WithPortGf(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_WithEmptyHost(t *testing.T) {
	path := "ut-sw"
	port := 1111

	ctx := &ghttp.Request{
		Request: &http.Request{
			Host: fmt.Sprintf("localhost:%d", port),
		},
	}

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGfEntry(WithSwEntryGf(sw), WithPortGf(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_HappyCase(t *testing.T) {
	ctx := &ghttp.Request{
		Request: &http.Request{
			Host: "8.8.8.8:1111",
		},
	}

	path := "ut-sw"
	port := 1111

	sw := NewSwEntry(WithPathSw(path), WithPortSw(uint64(port)))
	entry := RegisterGfEntry(WithSwEntryGf(sw), WithPortGf(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestContainsMetrics_ExpectFalse(t *testing.T) {
	api := "/rk/v1/non-exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: "/rk/v1/exist",
	})

	assert.False(t, containsMetrics(api, metrics))
}

func TestContainsMetrics_ExpectTrue(t *testing.T) {
	api := "/rk/v1/exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: api,
	})

	assert.True(t, containsMetrics(api, metrics))
}
