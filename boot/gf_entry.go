package rkgf

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidauth "github.com/rookie-ninja/rk-entry/middleware/auth"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/middleware/csrf"
	rkmidjwt "github.com/rookie-ninja/rk-entry/middleware/jwt"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	rkmidmeta "github.com/rookie-ninja/rk-entry/middleware/meta"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	rkmidlimit "github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	rkmidsec "github.com/rookie-ninja/rk-entry/middleware/secure"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	rkgfinter "github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/auth"
	rkgfctx "github.com/rookie-ninja/rk-gf/interceptor/context"
	"github.com/rookie-ninja/rk-gf/interceptor/cors"
	"github.com/rookie-ninja/rk-gf/interceptor/csrf"
	"github.com/rookie-ninja/rk-gf/interceptor/jwt"
	"github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gf/interceptor/meta"
	"github.com/rookie-ninja/rk-gf/interceptor/metrics/prom"
	rkgfpanic "github.com/rookie-ninja/rk-gf/interceptor/panic"
	"github.com/rookie-ninja/rk-gf/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-gf/interceptor/secure"
	"github.com/rookie-ninja/rk-gf/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	// GfEntryType type of entry
	GfEntryType = "GfEntry"
	// GfEntryDescription description of entry
	GfEntryDescription = "Internal RK entry which helps to bootstrap with GoFrame framework."
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap GoFrame entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterGfEntriesWithConfig)
}

// BootConfigGf boot config which is for GoFrame entry.
type BootConfigGf struct {
	Gf []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            rkentry.BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService rkentry.BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            rkentry.BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          rkentry.BootConfigProm          `yaml:"prom" json:"prom"`
		Static        rkentry.BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap       rkmidlog.BootConfig     `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm      rkmidmetrics.BootConfig `yaml:"metricsProm" json:"metricsProm"`
			Auth             rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors             rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta             rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt              rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure           rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			RateLimit        rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Csrf             rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			TracingTelemetry rkmidtrace.BootConfig   `yaml:"tracingTelemetry" json:"tracingTelemetry"`
		} `yaml:"interceptors" json:"interceptors"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"gf" json:"gf"`
}

// GfEntry implements rkentry.Entry interface.
type GfEntry struct {
	EntryName          string                          `json:"entryName" yaml:"entryName"`
	EntryType          string                          `json:"entryType" yaml:"entryType"`
	EntryDescription   string                          `json:"-" yaml:"-"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry         `json:"-" yaml:"-"`
	EventLoggerEntry   *rkentry.EventLoggerEntry       `json:"-" yaml:"-"`
	Port               uint64                          `json:"port" yaml:"port"`
	Server             *ghttp.Server                   `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	SwEntry            *rkentry.SwEntry                `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	Interceptors       []ghttp.HandlerFunc             `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	TvEntry            *rkentry.TvEntry                `json:"-" yaml:"-"`
}

// RegisterGfEntriesWithConfig register GoFrame entries with provided config file (Must YAML file).
//
// Currently, support two ways to provide config file path.
// 1: With function parameters
// 2: With command line flag "--rkboot" described in rkcommon.BootConfigPathFlagKey (Will override function parameter if exists)
// Command line flag has high priority which would override function parameter
//
// Error handling:
// Process will shutdown if any errors occur with rkcommon.ShutdownWithError function
//
// Override elements in config file:
// We learned from HELM source code which would override elements in YAML file with "--set" flag followed with comma
// separated key/value pairs.
//
// We are using "--rkset" described in rkcommon.BootConfigOverrideKey in order to distinguish with user flags
// Example of common usage: ./binary_file --rkset "key1=val1,key2=val2"
// Example of nested map:   ./binary_file --rkset "outer.inner.key=val"
// Example of slice:        ./binary_file --rkset "outer[0].key=val"
func RegisterGfEntriesWithConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootConfigGf{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init GoFrame entries with boot config
	for i := range config.Gf {
		element := config.Gf[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		// Register swagger entry
		swEntry := rkentry.RegisterSwEntryWithConfig(&element.SW, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, element.CommonService.Enabled)

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntryWithConfig(&element.Prom, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, promRegistry)

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntryWithConfig(&element.CommonService, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register TV entry
		tvEntry := rkentry.RegisterTvEntryWithConfig(&element.TV, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntryWithConfig(&element.Static, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		inters := make([]ghttp.HandlerFunc, 0)

		// logging middlewares
		if element.Interceptors.LoggingZap.Enabled {
			inters = append(inters, rkgflog.Interceptor(
				rkmidlog.ToOptions(&element.Interceptors.LoggingZap, element.Name, GfEntryType,
					zapLoggerEntry, eventLoggerEntry)...))
		}

		// metrics middleware
		if element.Interceptors.MetricsProm.Enabled {
			inters = append(inters, rkgfmetrics.Interceptor(
				rkmidmetrics.ToOptions(&element.Interceptors.MetricsProm, element.Name, GfEntryType,
					promRegistry, rkmidmetrics.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Interceptors.TracingTelemetry.Enabled {
			inters = append(inters, rkgftrace.Interceptor(
				rkmidtrace.ToOptions(&element.Interceptors.TracingTelemetry, element.Name, GfEntryType)...))
		}

		// jwt middleware
		if element.Interceptors.Jwt.Enabled {
			inters = append(inters, rkgfjwt.Interceptor(
				rkmidjwt.ToOptions(&element.Interceptors.Jwt, element.Name, GfEntryType)...))
		}

		// secure middleware
		if element.Interceptors.Secure.Enabled {
			inters = append(inters, rkgfsec.Interceptor(
				rkmidsec.ToOptions(&element.Interceptors.Secure, element.Name, GfEntryType)...))
		}

		// csrf middleware
		if element.Interceptors.Csrf.Enabled {
			inters = append(inters, rkgfcsrf.Interceptor(
				rkmidcsrf.ToOptions(&element.Interceptors.Csrf, element.Name, GfEntryType)...))
		}

		// Did we enabled cors interceptor?
		if element.Interceptors.Cors.Enabled {
			inters = append(inters, rkgfcors.Interceptor(
				rkmidcors.ToOptions(&element.Interceptors.Cors, element.Name, GfEntryType)...))
		}

		// meta middleware
		if element.Interceptors.Meta.Enabled {
			inters = append(inters, rkgfmeta.Interceptor(
				rkmidmeta.ToOptions(&element.Interceptors.Meta, element.Name, GfEntryType)...))
		}

		// auth middlewares
		if element.Interceptors.Auth.Enabled {
			inters = append(inters, rkgfauth.Interceptor(
				rkmidauth.ToOptions(&element.Interceptors.Auth, element.Name, GfEntryType)...))
		}

		// rate limit middleware
		if element.Interceptors.RateLimit.Enabled {
			inters = append(inters, rkgflimit.Interceptor(
				rkmidlimit.ToOptions(&element.Interceptors.RateLimit, element.Name, GfEntryType)...))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterGfEntry(
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry),
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithSwEntry(swEntry),
			WithPromEntry(promEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithCertEntry(certEntry),
			WithTvEntry(tvEntry),
			WithStaticFileHandlerEntry(staticEntry),
			WithInterceptors(inters...))

		entry.AddInterceptor(inters...)

		res[name] = entry
	}

	return res
}

// RegisterGfEntry register GfEntry with options.
func RegisterGfEntry(opts ...GfEntryOption) *GfEntry {
	entry := &GfEntry{
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		EntryType:        GfEntryType,
		EntryDescription: GfEntryDescription,
		Interceptors:     make([]ghttp.HandlerFunc, 0),
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "GfServer-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Server == nil {
		entry.Server = g.Server(entry.EntryName)
		entry.Server.SetDumpRouterMap(false)
		entry.Server.SetLogger(rkgfinter.NewNoopGLogger())
		glog.SetStdoutPrint(false)
	}

	if entry.Port != 0 {
		entry.Server.SetPort(int(entry.Port))
	}

	// insert panic interceptor
	entry.Server.Use(rkgfpanic.Interceptor(
		rkmidpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *GfEntry) GetName() string {
	return entry.EntryName
}

// GetType Get entry type.
func (entry *GfEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry.
func (entry *GfEntry) GetDescription() string {
	return entry.EntryDescription
}

// Bootstrap GfEntry.
func (entry *GfEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap")

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Server.BindHandler(path.Join(entry.SwEntry.Path, "*any"), ghttp.WrapF(entry.SwEntry.ConfigFileHandler()))
		entry.Server.BindHandler(path.Join(entry.SwEntry.AssetsFilePath, "*"), ghttp.WrapF(entry.SwEntry.AssetsFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		// Register path into Router.
		entry.Server.BindHandler(path.Join(entry.StaticFileEntry.Path, "*any"), ghttp.WrapF(entry.StaticFileEntry.GetFileHandler()))
		entry.StaticFileEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.Server.BindHandler(entry.PromEntry.Path, ghttp.WrapH(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Server.BindHandler(entry.CommonServiceEntry.HealthyPath, ghttp.WrapF(entry.CommonServiceEntry.Healthy))
		entry.Server.BindHandler(entry.CommonServiceEntry.GcPath, ghttp.WrapF(entry.CommonServiceEntry.Gc))
		entry.Server.BindHandler(entry.CommonServiceEntry.InfoPath, ghttp.WrapF(entry.CommonServiceEntry.Info))
		entry.Server.BindHandler(entry.CommonServiceEntry.ConfigsPath, ghttp.WrapF(entry.CommonServiceEntry.Configs))
		entry.Server.BindHandler(entry.CommonServiceEntry.SysPath, ghttp.WrapF(entry.CommonServiceEntry.Sys))
		entry.Server.BindHandler(entry.CommonServiceEntry.EntriesPath, ghttp.WrapF(entry.CommonServiceEntry.Entries))
		entry.Server.BindHandler(entry.CommonServiceEntry.CertsPath, ghttp.WrapF(entry.CommonServiceEntry.Certs))
		entry.Server.BindHandler(entry.CommonServiceEntry.LogsPath, ghttp.WrapF(entry.CommonServiceEntry.Logs))
		entry.Server.BindHandler(entry.CommonServiceEntry.DepsPath, ghttp.WrapF(entry.CommonServiceEntry.Deps))
		entry.Server.BindHandler(entry.CommonServiceEntry.LicensePath, ghttp.WrapF(entry.CommonServiceEntry.License))
		entry.Server.BindHandler(entry.CommonServiceEntry.ReadmePath, ghttp.WrapF(entry.CommonServiceEntry.Readme))
		entry.Server.BindHandler(entry.CommonServiceEntry.GitPath, ghttp.WrapF(entry.CommonServiceEntry.Git))

		// swagger doc already generated at rkentry.CommonService
		// follow bellow actions
		entry.Server.BindHandler(entry.CommonServiceEntry.ApisPath, entry.Apis)
		entry.Server.BindHandler(entry.CommonServiceEntry.ReqPath, entry.Req)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Server.BindHandler(strings.TrimSuffix(entry.TvEntry.BasePath, "/"), func(ctx *ghttp.Request) {
			ctx.Response.RedirectTo(entry.SwEntry.Path, http.StatusTemporaryRedirect)
		})
		entry.Server.BindHandler(path.Join(entry.TvEntry.BasePath, "*item"), entry.TV)
		entry.Server.BindHandler(path.Join(entry.TvEntry.AssetsFilePath, "*"), ghttp.WrapF(entry.TvEntry.AssetsFileHandler()))

		entry.TvEntry.Bootstrap(ctx)
	}

	go entry.startServer(event, logger)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt GfEntry.
func (entry *GfEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt")

	if entry.IsSwEnabled() {
		// Interrupt swagger entry
		entry.SwEntry.Interrupt(ctx)
	}

	if entry.IsStaticFileHandlerEnabled() {
		// Interrupt entry
		entry.StaticFileEntry.Interrupt(ctx)
	}

	if entry.IsPromEnabled() {
		// Interrupt prometheus entry
		entry.PromEntry.Interrupt(ctx)
	}

	if entry.IsCommonServiceEnabled() {
		// Interrupt common service entry
		entry.CommonServiceEntry.Interrupt(ctx)
	}

	if entry.IsTvEnabled() {
		// Interrupt common service entry
		entry.TvEntry.Interrupt(ctx)
	}

	if entry.Server != nil {
		if err := entry.Server.Shutdown(); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gf-server.", event.ListPayloads()...)
		}
	}

	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// String Stringfy gf entry.
func (entry *GfEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// ***************** Stringfy *****************

// MarshalJSON Marshal entry.
func (entry *GfEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":          entry.EntryName,
		"entryType":          entry.EntryType,
		"entryDescription":   entry.EntryDescription,
		"eventLoggerEntry":   entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":     entry.ZapLoggerEntry.GetName(),
		"port":               entry.Port,
		"swEntry":            entry.SwEntry,
		"commonServiceEntry": entry.CommonServiceEntry,
		"promEntry":          entry.PromEntry,
		"tvEntry":            entry.TvEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
	}

	interceptorsStr := make([]string, 0)
	m["interceptors"] = &interceptorsStr

	for i := range entry.Interceptors {
		element := entry.Interceptors[i]
		interceptorsStr = append(interceptorsStr,
			path.Base(runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name()))
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *GfEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ***************** Public functions *****************

// GetGfEntry Get GfEntry from rkentry.GlobalAppCtx.
func GetGfEntry(name string) *GfEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GfEntry)
	return entry
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *GfEntry) AddInterceptor(inters ...ghttp.HandlerFunc) {
	entry.Server.Use(inters...)
}

// IsTlsEnabled Is TLS enabled?
func (entry *GfEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *GfEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsStaticFileHandlerEnabled Is static file handler entry enabled?
func (entry *GfEntry) IsStaticFileHandlerEnabled() bool {
	return entry.StaticFileEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *GfEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsTvEnabled Is TV entry enabled?
func (entry *GfEntry) IsTvEnabled() bool {
	return entry.TvEntry != nil
}

// IsPromEnabled Is prometheus entry enabled?
func (entry *GfEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// ***************** Helper function *****************

// Add basic fields into event.
func (entry *GfEntry) logBasicInfo(operation string) (rkquery.Event, *zap.Logger) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	logger := entry.ZapLoggerEntry.GetLogger().With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.EntryName))

	// add general info
	event.AddPayloads(
		zap.Uint64("gfPort", entry.Port))

	// add SwEntry info
	if entry.IsSwEnabled() {
		event.AddPayloads(
			zap.Bool("swEnabled", true),
			zap.String("swPath", entry.SwEntry.Path))
	}

	// add CommonServiceEntry info
	if entry.IsCommonServiceEnabled() {
		event.AddPayloads(
			zap.Bool("commonServiceEnabled", true),
			zap.String("commonServicePathPrefix", "/rk/v1/"))
	}

	// add TvEntry info
	if entry.IsTvEnabled() {
		event.AddPayloads(
			zap.Bool("tvEnabled", true),
			zap.String("tvPath", "/rk/v1/tv/"))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.PromEntry.Port),
			zap.String("promPath", entry.PromEntry.Path))
	}

	// add StaticFileHandlerEntry info
	if entry.IsStaticFileHandlerEnabled() {
		event.AddPayloads(
			zap.Bool("staticFileHandlerEnabled", true),
			zap.String("staticFileHandlerPath", entry.StaticFileEntry.Path))
	}

	// add tls info
	if entry.IsTlsEnabled() {
		event.AddPayloads(
			zap.Bool("tlsEnabled", true))
	}

	logger.Info(fmt.Sprintf("%s gfEntry", operation))

	return event, logger
}

// Start server
// We move the code here for testability
func (entry *GfEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.Server != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			if cert, err := tls.X509KeyPair(entry.CertEntry.Store.ServerCert, entry.CertEntry.Store.ServerKey); err != nil {
				event.AddErr(err)
				logger.Error("Error occurs while parsing TLS.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			} else {
				entry.Server.SetTLSConfig(&tls.Config{Certificates: []tls.Certificate{cert}})
			}
		}

		err := entry.Server.Start()

		if err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Error("Error occurs while starting GoFrame server.", event.ListPayloads()...)
			rkcommon.ShutdownWithError(err)
		}
	}
}

// ***************** Common Service Extension API *****************

// Apis list apis from gin.Router
func (entry *GfEntry) Apis(ctx *ghttp.Request) {
	ctx.Response.Header().Set("Access-Control-Allow-Origin", "*")

	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(entry.doApis(ctx))
}

// Req handler
func (entry *GfEntry) Req(ctx *ghttp.Request) {
	ctx.Response.WriteHeader(http.StatusOK)
	ctx.Response.WriteJson(entry.doReq(ctx))
}

// Construct swagger URL based on IP and scheme
func (entry *GfEntry) constructSwUrl(ctx *ghttp.Request) string {
	if entry.SwEntry == nil {
		return "N/A"
	}

	originalURL := fmt.Sprintf("localhost:%d", entry.Port)
	if ctx != nil && ctx.Request != nil && len(ctx.Request.Host) > 0 {
		originalURL = ctx.Request.Host
	}

	scheme := "http"
	if ctx != nil && ctx.Request != nil && ctx.Request.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, originalURL, entry.SwEntry.Path)
}

// Helper function for APIs call
func (entry *GfEntry) doApis(ctx *ghttp.Request) *rkentry.ApisResponse {
	res := &rkentry.ApisResponse{
		Entries: make([]*rkentry.ApisResponseElement, 0),
	}

	routes := entry.Server.GetRoutes()
	for j := range routes {
		info := routes[j]

		entry := &rkentry.ApisResponseElement{
			EntryName: entry.GetName(),
			Method:    info.Method,
			Path:      info.Route,
			Port:      entry.Port,
			SwUrl:     entry.constructSwUrl(ctx),
		}
		res.Entries = append(res.Entries, entry)
	}

	return res
}

// Is metrics from prometheus contains particular api?
func (entry *GfEntry) containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].RestPath == api {
			return true
		}
	}

	return false
}

// Helper function for Req call
func (entry *GfEntry) doReq(ctx *ghttp.Request) *rkentry.ReqResponse {
	metricsSet := rkmidmetrics.GetServerMetricsSet(entry.GetName())
	if metricsSet == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	vector := metricsSet.GetSummary(rkmidmetrics.MetricsNameElapsedNano)
	if vector == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// Fill missed metrics
	apis := make([]string, 0)
	apisMethod := make([]string, 0)

	routes := entry.Server.GetRoutes()
	for j := range routes {
		info := routes[j]
		apis = append(apis, info.Route)
		apisMethod = append(apisMethod, info.Method)
	}

	// Add empty metrics into result
	for i := range apis {
		if !entry.containsMetrics(apis[i], reqMetrics) {
			reqMetrics = append(reqMetrics, &rkentry.ReqMetricsRK{
				RestMethod: apisMethod[i],
				RestPath:   apis[i],
				ResCode:    make([]*rkentry.ResCodeRK, 0),
			})
		}
	}

	return &rkentry.ReqResponse{
		Metrics: reqMetrics,
	}
}

// TV handler
func (entry *GfEntry) TV(ctx *ghttp.Request) {
	logger := rkgfctx.GetLogger(ctx)

	switch item := ctx.Get("item").String(); item {
	case "apis":
		buf := entry.TvEntry.ExecuteTemplate("apis", entry.doApis(ctx), logger)
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write(buf.Bytes())
	default:
		buf := entry.TvEntry.Action(item, logger)
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write(buf.Bytes())
	}
}

// ***************** Options *****************

// GfEntryOption Gf entry option.
type GfEntryOption func(*GfEntry)

// WithPort provide port.
func WithPort(port uint64) GfEntryOption {
	return func(entry *GfEntry) {
		entry.Port = port
	}
}

// WithName provide name.
func WithName(name string) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EntryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EntryDescription = description
	}
}

// WithZapLoggerEntry provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntry(zapLogger *rkentry.ZapLoggerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntry provide rkentry.EventLoggerEntry.
func WithEventLoggerEntry(eventLogger *rkentry.EventLoggerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SwEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntry provide CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithInterceptors provide user interceptors.
func WithInterceptors(inters ...ghttp.HandlerFunc) GfEntryOption {
	return func(entry *GfEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]ghttp.HandlerFunc, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// WithPromEntry provide PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntry provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntry(staticEntry *rkentry.StaticFileHandlerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithTvEntry provide TvEntry.
func WithTvEntry(tvEntry *rkentry.TvEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.TvEntry = tvEntry
	}
}
