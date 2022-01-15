// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgf

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	rkgfmeta "github.com/rookie-ninja/rk-gf/interceptor/meta"
	rkgfmetrics "github.com/rookie-ninja/rk-gf/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
gf:
 - name: greeter
   port: 1949
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
     timeout:
       enabled: true
     cors:
       enabled: true
     jwt:
       enabled: true
     secure:
       enabled: true
     csrf:
       enabled: true
     gzip:
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
 - name: greeter3
   port: 2022
   enabled: false
`
)

func TestGetGfEntry(t *testing.T) {
	// expect nil
	assert.Nil(t, GetGfEntry("entry-name"))

	// happy case
	ginEntry := RegisterGfEntry(WithName("ut-gin"))
	assert.Equal(t, ginEntry, GetGfEntry("ut-gin"))

	rkentry.GlobalAppCtx.RemoveEntry("ut-gin")
}

func TestRegisterGfEntry(t *testing.T) {
	// without options
	entry := RegisterGfEntry()
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())

	// with options
	entry = RegisterGfEntry(
		WithZapLoggerEntry(nil),
		WithEventLoggerEntry(nil),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(rkentry.RegisterCertEntry()),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPort(8080),
		WithName("ut-entry"),
		WithDescription("ut-desc"),
		WithPromEntry(rkentry.RegisterPromEntry()))

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.True(t, entry.IsSwEnabled())
	assert.True(t, entry.IsStaticFileHandlerEnabled())
	assert.True(t, entry.IsPromEnabled())
	assert.True(t, entry.IsCommonServiceEnabled())
	assert.True(t, entry.IsTvEnabled())
	assert.True(t, entry.IsTlsEnabled())

	bytes, err := entry.MarshalJSON()
	assert.NotEmpty(t, bytes)
	assert.Nil(t, err)
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestGfEntry_AddInterceptor(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterGfEntry()
	inter := rkgfmeta.Interceptor()
	entry.AddInterceptor(inter)
}

func TestGinEntry_Bootstrap(t *testing.T) {
	//defer assertNotPanic(t)

	// without enable sw, static, prom, common, tv, tls
	entry := RegisterGfEntry(WithPort(8080))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.Server.GetRoutes())

	entry.Interrupt(context.TODO())

	// with enable sw, static, prom, common, tv, tls
	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	entry = RegisterGfEntry(
		WithPort(8081),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(certEntry),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPromEntry(rkentry.RegisterPromEntry()))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8081, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.Server.GetRoutes())

	entry.Interrupt(context.TODO())
}

func TestGfEntry_startServer_InvalidTls(t *testing.T) {
	defer assertPanic(t)

	// with invalid tls
	entry := RegisterGfEntry(
		WithPort(8080),
		WithCertEntry(rkentry.RegisterCertEntry()))
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestRegisterGfEntriesWithConfig(t *testing.T) {
	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGfEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*GfEntry)
	assert.NotNil(t, greeter)

	greeter2 := entries["greeter2"].(*GfEntry)
	assert.NotNil(t, greeter2)

	greeter3 := entries["greeter3"]
	assert.Nil(t, greeter3)
}

func TestGfEntry_constructSwUrl(t *testing.T) {
	// happy case
	ctx := &ghttp.Request{}
	ctx.Request = &http.Request{
		Host: "8.8.8.8:1111",
	}

	path := "ut-sw"
	port := 1111

	sw := rkentry.RegisterSwEntry(rkentry.WithPathSw(path), rkentry.WithPortSw(uint64(port)))
	entry := RegisterGfEntry(WithSwEntry(sw), WithPort(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// with tls
	ctx.Request.TLS = &tls.ConnectionState{}
	assert.Equal(t, fmt.Sprintf("https://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// without swagger
	entry = RegisterGfEntry(WithPort(uint64(port)))
	assert.Equal(t, "N/A", entry.constructSwUrl(ctx))
}

func TestGfEntry_API(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithPort(8080),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithName("unit-test"))

	entry.Bootstrap(context.TODO())

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/apis")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
}

func TestGfEntry_Req_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithPort(8080),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithName("unit-test-req"))

	entry.Bootstrap(context.TODO())

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/req")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
}

func TestGfEntry_Req_WithEmpty(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithPort(8080),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithName("unit-test-req-empty"))

	entry.Bootstrap(context.TODO())

	entry.AddInterceptor(rkgfmetrics.Interceptor(
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/req")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
}

func TestGfEntry_TV(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGfEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithPort(8080),
		WithName("ut-gf"))

	entry.AddInterceptor(rkgfmetrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut-gf", "Gf")))

	entry.Bootstrap(context.TODO())

	// for /api
	client := getClient()
	resp, err := client.Get(context.TODO(), "/rk/v1/tv/apis")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// for default
	client = getClient()
	resp, err = client.Get(context.TODO(), "/rk/v1/tv/other")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
}

func getClient() *gclient.Client {
	time.Sleep(100 * time.Millisecond)
	client := g.Client()
	client.SetBrowserMode(true)
	client.SetPrefix("http://127.0.0.1:8080")

	return client
}

func generateCerts() ([]byte, []byte) {
	// Create certs and return as []byte
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Fake cert."},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, _ := rsa.GenerateKey(rand.Reader, 4096)

	// Use CA Cert to sign a CSR and create a Public Cert
	csr := &key.PublicKey
	cert, _ := x509.CreateCertificate(rand.Reader, ca, ca, csr, key)

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(c), pem.EncodeToMemory(k)
}

func validateServerIsUp(t *testing.T, port uint64, isTls bool) {
	// sleep for 2 seconds waiting server startup
	time.Sleep(2 * time.Second)

	if !isTls {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
		if conn != nil {
			assert.Nil(t, conn.Close())
		}
		return
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	tlsConn, err := tls.Dial("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), tlsConf)
	assert.Nil(t, err)
	assert.NotNil(t, tlsConn)
	if tlsConn != nil {
		assert.Nil(t, tlsConn.Close())
	}
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}
