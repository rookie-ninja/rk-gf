// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/entry"
	rkgflog "github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	rkgfmetrics "github.com/rookie-ninja/rk-gf/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
gf:
  - name: greeter
    port: 8080
    enabled: true
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    prom:
      enabled: true
      pusher:
        enabled: false
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
      meta:
        enabled: true
      tracingTelemetry:
        enabled: true
      ratelimit:
        enabled: true
      cors:
        enabled: true
      secure:
        enabled: true
  - name: greeter2
    port: 2008
    enabled: true
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
`
)

func TestWithZapLoggerEntryGf_HappyCase(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterGfEntry()

	option := WithZapLoggerEntryGf(loggerEntry)
	option(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()

	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	option := WithEventLoggerEntryGf(eventLoggerEntry)
	option(entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithInterceptorsGf_WithNilInterceptorList(t *testing.T) {
	entry := RegisterGfEntry()

	option := WithInterceptorsGf(nil)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
}

func TestWithInterceptorsGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()

	loggingInterceptor := rkgflog.Interceptor()
	metricsInterceptor := rkgfmetrics.Interceptor()

	interceptors := []ghttp.HandlerFunc{
		loggingInterceptor,
		metricsInterceptor,
	}

	option := WithInterceptorsGf(interceptors...)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
	// should contains logging, metrics and panic interceptor
	// where panic interceptor is inject by default
	assert.Len(t, entry.Interceptors, 3)
}

func TestWithCommonServiceEntryGf_WithEntry(t *testing.T) {
	entry := RegisterGfEntry()

	option := WithCommonServiceEntryGf(NewCommonServiceEntry())
	option(entry)

	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestWithCommonServiceEntryGf_WithoutEntry(t *testing.T) {
	entry := RegisterGfEntry()

	assert.Nil(t, entry.CommonServiceEntry)
}

func TestWithTVEntryGf_WithEntry(t *testing.T) {
	entry := RegisterGfEntry()

	option := WithTVEntryGf(NewTvEntry())
	option(entry)

	assert.NotNil(t, entry.TvEntry)
}

func TestWithTVEntry_WithoutEntry(t *testing.T) {
	entry := RegisterGfEntry()

	assert.Nil(t, entry.TvEntry)
}

func TestWithCertEntryGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()
	certEntry := &rkentry.CertEntry{}

	option := WithCertEntryGf(certEntry)
	option(entry)

	assert.Equal(t, entry.CertEntry, certEntry)
}

func TestWithSWEntryGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()
	sw := NewSwEntry()

	option := WithSwEntryGf(sw)
	option(entry)

	assert.Equal(t, entry.SwEntry, sw)
}

func TestWithPortGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()
	port := uint64(1111)

	option := WithPortGf(port)
	option(entry)

	assert.Equal(t, entry.Port, port)
}

func TestWithNameGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()
	name := "unit-test-entry"

	option := WithNameGf(name)
	option(entry)

	assert.Equal(t, entry.EntryName, name)
}

func TestRegisterGfEntriesWithConfig_WithInvalidConfigFilePath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterGfEntriesWithConfig("/invalid-path")
}

func TestRegisterGfEntriesWithConfig_WithNilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGfEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
}

func TestRegisterGfEntriesWithConfig_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGfEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*GfEntry)
	assert.NotNil(t, greeter)
	assert.Equal(t, uint64(8080), greeter.Port)
	assert.NotNil(t, greeter.SwEntry)
	assert.NotNil(t, greeter.CommonServiceEntry)
	assert.NotNil(t, greeter.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.True(t, len(greeter.Interceptors) > 0)

	greeter2 := entries["greeter2"].(*GfEntry)
	assert.NotNil(t, greeter2)
	assert.Equal(t, uint64(2008), greeter2.Port)
	assert.NotNil(t, greeter2.SwEntry)
	assert.NotNil(t, greeter2.CommonServiceEntry)
	assert.NotNil(t, greeter2.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.Len(t, greeter2.Interceptors, 4)
}

func TestRegisterGfEntry_WithZapLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterGfEntry(WithZapLoggerEntryGf(loggerEntry))
	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestRegisterGfEntry_WithEventLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopEventLoggerEntry()

	entry := RegisterGfEntry(WithEventLoggerEntryGf(loggerEntry))
	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestNewGfEntry_WithInterceptors(t *testing.T) {
	loggingInterceptor := rkgflog.Interceptor()
	entry := RegisterGfEntry(WithInterceptorsGf(loggingInterceptor))
	assert.Len(t, entry.Interceptors, 2)
}

func TestNewGfEntry_WithCommonServiceEntry(t *testing.T) {
	entry := RegisterGfEntry(WithCommonServiceEntryGf(NewCommonServiceEntry()))
	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestNewGfEntry_WithTVEntry(t *testing.T) {
	entry := RegisterGfEntry(WithTVEntryGf(NewTvEntry()))
	assert.NotNil(t, entry.TvEntry)
}

func TestNewGfEntry_WithCertStore(t *testing.T) {
	certEntry := &rkentry.CertEntry{}

	entry := RegisterGfEntry(WithCertEntryGf(certEntry))
	assert.Equal(t, certEntry, entry.CertEntry)
}

func TestNewGfEntry_WithSwEntry(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterGfEntry(WithSwEntryGf(sw))
	assert.Equal(t, sw, entry.SwEntry)
}

func TestNewGfEntry_WithPort(t *testing.T) {
	entry := RegisterGfEntry(WithPortGf(8080))
	assert.Equal(t, uint64(8080), entry.Port)
}

func TestNewGfEntry_WithName(t *testing.T) {
	entry := RegisterGfEntry(WithNameGf("unit-test-greeter"))
	assert.Equal(t, "unit-test-greeter", entry.GetName())
}

func TestNewGfEntry_WithDefaultValue(t *testing.T) {
	entry := RegisterGfEntry()
	assert.True(t, strings.HasPrefix(entry.GetName(), "GfServer-"))
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Len(t, entry.Interceptors, 1)
	assert.NotNil(t, entry.Server)
	assert.Nil(t, entry.SwEntry)
	assert.Nil(t, entry.CertEntry)
	assert.False(t, entry.IsSwEnabled())
	assert.False(t, entry.IsTlsEnabled())
	assert.Nil(t, entry.CommonServiceEntry)
	assert.Nil(t, entry.TvEntry)
	assert.Equal(t, "GfEntry", entry.GetType())
}

func TestGfEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterGfEntry(WithNameGf("unit-test-entry"))
	assert.Equal(t, "unit-test-entry", entry.GetName())
}

func TestGfEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, "GfEntry", RegisterGfEntry().GetType())
}

func TestGfEntry_String_HappyCase(t *testing.T) {
	assert.NotEmpty(t, RegisterGfEntry().String())
}

func TestGfEntry_IsSwEnabled_ExpectTrue(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterGfEntry(WithSwEntryGf(sw))
	assert.True(t, entry.IsSwEnabled())
}

func TestGfEntry_IsSwEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterGfEntry()
	assert.False(t, entry.IsSwEnabled())
}

func TestGfEntry_IsTlsEnabled_ExpectTrue(t *testing.T) {
	certEntry := &rkentry.CertEntry{
		Store: &rkentry.CertStore{},
	}

	entry := RegisterGfEntry(WithCertEntryGf(certEntry))
	assert.True(t, entry.IsTlsEnabled())
}

func TestGfEntry_IsTlsEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterGfEntry()
	assert.False(t, entry.IsTlsEnabled())
}

func TestGfEntry_GetGf_HappyCase(t *testing.T) {
	entry := RegisterGfEntry()
	assert.NotNil(t, entry.Server)
}

func TestGfEntry_Bootstrap_WithSwagger(t *testing.T) {
	sw := NewSwEntry(
		WithPathSw("sw"),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()))
	entry := RegisterGfEntry(
		WithNameGf(time.Now().String()),
		WithPortGf(8080),
		WithZapLoggerEntryGf(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGf(rkentry.NoopEventLoggerEntry()),
		WithSwEntryGf(sw))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.True(t, len(entry.Server.GetRoutes()) > 0)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestGfEntry_Bootstrap_WithoutSwagger(t *testing.T) {
	entry := RegisterGfEntry(
		WithNameGf(time.Now().String()),
		WithPortGf(8080),
		WithZapLoggerEntryGf(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGf(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestGfEntry_Bootstrap_WithoutTLS(t *testing.T) {
	entry := RegisterGfEntry(
		WithNameGf(time.Now().String()),
		WithPortGf(8080),
		WithZapLoggerEntryGf(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGf(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestGfEntry_Shutdown_WithBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithNameGf(time.Now().String()),
		WithPortGf(8080),
		WithZapLoggerEntryGf(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGf(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestGfEntry_Shutdown_WithoutBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithNameGf(time.Now().String()),
		WithPortGf(8080),
		WithZapLoggerEntryGf(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGf(rkentry.NoopEventLoggerEntry()))

	entry.Interrupt(context.Background())
}

func validateServerIsUp(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	if conn != nil {
		assert.Nil(t, conn.Close())
	}
}
