// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestNewStaticFileHandlerEntry(t *testing.T) {
	// without options
	entry := NewStaticFileHandlerEntry()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Equal(t, "/rk/v1/static/", entry.Path)
	assert.NotNil(t, entry.Fs)
	assert.NotNil(t, entry.Template)

	// with options
	utFs := http.Dir("")
	utPath := "/ut-path/"
	utZapLogger := rkentry.NoopZapLoggerEntry()
	utEventLogger := rkentry.NoopEventLoggerEntry()
	utName := "ut-entry"

	entry = NewStaticFileHandlerEntry(
		WithPathStatic(utPath),
		WithEventLoggerEntryStatic(utEventLogger),
		WithZapLoggerEntryStatic(utZapLogger),
		WithNameStatic(utName),
		WithFileSystemStatic(utFs))

	assert.NotNil(t, entry)
	assert.Equal(t, utZapLogger, entry.ZapLoggerEntry)
	assert.Equal(t, utEventLogger, entry.EventLoggerEntry)
	assert.Equal(t, utPath, entry.Path)
	assert.Equal(t, utFs, entry.Fs)
	assert.NotNil(t, entry.Template)
	assert.Equal(t, utName, entry.EntryName)
}

func TestStaticFileHandlerEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := NewStaticFileHandlerEntry()
	entry.Bootstrap(context.TODO())
}

func TestStaticFileHandlerEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := NewStaticFileHandlerEntry()
	entry.Interrupt(context.TODO())
}

func TestStaticFileHandlerEntry_EntryFunctions(t *testing.T) {
	entry := NewStaticFileHandlerEntry()
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestStaticFileHandlerEntry_GetFileHandler(t *testing.T) {
	currDir := t.TempDir()
	os.MkdirAll(path.Join(currDir, "ut-dir"), os.ModePerm)
	os.WriteFile(path.Join(currDir, "ut-file"), []byte("ut content"), os.ModePerm)

	entry := NewStaticFileHandlerEntry(
		WithPathStatic("/rk/v1/static"),
		WithFileSystemStatic(http.Dir(currDir)))
	entry.Bootstrap(context.TODO())

	// expect to get list of files
	server := startServer(t, "/rk/v1/static/*", entry.GetFileHandler())
	server.BindHandler("/rk/v1/static", func(ctx *ghttp.Request) {
		ctx.Response.RedirectTo("/rk/v1/static/", http.StatusTemporaryRedirect)
	})
	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/static/")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, bodyToString(resp.Body), "Index of")
	assert.Nil(t, server.Shutdown())

	// expect to get files to download
	server = startServer(t, "/rk/v1/static/*", entry.GetFileHandler())
	client = getClient()
	resp, err = client.Get(context.TODO(), "/rk/v1/static/ut-file")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
	assert.NotEmpty(t, resp.Header.Get("Content-Type"))
	assert.Contains(t, bodyToString(resp.Body), "ut content")
	assert.Nil(t, server.Shutdown())
}
