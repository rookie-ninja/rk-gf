// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestWithPath_HappyCase(t *testing.T) {
	entry := NewSwEntry(WithPathSw("ut-path"))
	assert.Equal(t, "/ut-path/", entry.Path)
}

func TestWithHeaders_HappyCase(t *testing.T) {
	headers := map[string]string{
		"key": "value",
	}
	entry := NewSwEntry(WithHeadersSw(headers))
	assert.Len(t, entry.Headers, 1)
}

func TestNewSwEntry(t *testing.T) {
	entry := NewSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	assert.Equal(t, uint64(1234), entry.Port)
	assert.Equal(t, "/ut-path/", entry.Path)
	assert.Equal(t, "ut-json-path", entry.JsonPath)
	assert.Len(t, entry.Headers, 1)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.True(t, entry.EnableCommonService)
}

func TestSwEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := NewSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	entry.Bootstrap(context.TODO())
}

func TestSwEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := NewSwEntry(
		WithPortSw(1234),
		WithNameSw("ut-name"),
		WithPathSw("ut-path"),
		WithJsonPathSw("ut-json-path"),
		WithHeadersSw(map[string]string{
			"key": "value",
		}),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()),
		WithEnableCommonServiceSw(true))

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestSwEntry_UnmarshalJSON(t *testing.T) {
	entry := NewSwEntry()
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestSwEntry(t *testing.T) {
	entry := NewSwEntry()

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestSwEntry_AssetsFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewSwEntry()

	server := startServer(t, "/rk/v1/assets/*", entry.AssetsFileHandler())

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/assets/sw/index.html")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}

func TestSwEntry_ConfigFileHandler(t *testing.T) {
	defer assertNotPanic(t)
	entry := NewSwEntry()

	server := startServer(t, "/sw/*", entry.ConfigFileHandler())

	client := getClient()
	resp, err := client.Get(context.TODO(), "/sw")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, bodyToString(resp.Body))
	assert.Nil(t, server.Shutdown())
}
