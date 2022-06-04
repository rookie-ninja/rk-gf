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
	"github.com/rookie-ninja/rk-entry/v2/entry"
	rkerror "github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/rookie-ninja/rk-entry/v2/middleware/cors"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
	"github.com/rookie-ninja/rk-entry/v2/middleware/tracing"
	rkgfinter "github.com/rookie-ninja/rk-gf/middleware"
	"github.com/rookie-ninja/rk-gf/middleware/auth"
	"github.com/rookie-ninja/rk-gf/middleware/cors"
	"github.com/rookie-ninja/rk-gf/middleware/csrf"
	"github.com/rookie-ninja/rk-gf/middleware/jwt"
	"github.com/rookie-ninja/rk-gf/middleware/log"
	"github.com/rookie-ninja/rk-gf/middleware/meta"
	"github.com/rookie-ninja/rk-gf/middleware/panic"
	"github.com/rookie-ninja/rk-gf/middleware/prom"
	"github.com/rookie-ninja/rk-gf/middleware/ratelimit"
	"github.com/rookie-ninja/rk-gf/middleware/secure"
	"github.com/rookie-ninja/rk-gf/middleware/tracing"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	"path"
	"strconv"
	"strings"
	"sync"
)

const (
	// GfEntryType type of entry
	GfEntryType = "GoFrameEntry"
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap GoFrame entry automatically from boot config file
func init() {
	rkentry.RegisterWebFrameRegFunc(RegisterGfEntryYAML)
}

// BootGf boot config which is for GoFrame entry.
type BootGf struct {
	Gf []struct {
		Enabled       bool                          `yaml:"enabled" json:"enabled"`
		Name          string                        `yaml:"name" json:"name"`
		Port          uint64                        `yaml:"port" json:"port"`
		Description   string                        `yaml:"description" json:"description"`
		CertEntry     string                        `yaml:"certEntry" json:"certEntry"`
		LoggerEntry   string                        `yaml:"loggerEntry" json:"loggerEntry"`
		EventEntry    string                        `yaml:"eventEntry" json:"eventEntry"`
		SW            rkentry.BootSW                `yaml:"sw" json:"sw"`
		Docs          rkentry.BootDocs              `yaml:"docs" json:"docs"`
		CommonService rkentry.BootCommonService     `yaml:"commonService" json:"commonService"`
		Prom          rkentry.BootProm              `yaml:"prom" json:"prom"`
		Static        rkentry.BootStaticFileHandler `yaml:"static" json:"static"`
		PProf         rkentry.BootPProf             `yaml:"pprof" json:"pprof"`
		Middleware    struct {
			Ignore     []string              `yaml:"ignore" json:"ignore"`
			ErrorModel string                `yaml:"errorModel" json:"errorModel"`
			Logging    rkmidlog.BootConfig   `yaml:"logging" json:"logging"`
			Prom       rkmidprom.BootConfig  `yaml:"prom" json:"prom"`
			Auth       rkmidauth.BootConfig  `yaml:"auth" json:"auth"`
			Cors       rkmidcors.BootConfig  `yaml:"cors" json:"cors"`
			Meta       rkmidmeta.BootConfig  `yaml:"meta" json:"meta"`
			Jwt        rkmidjwt.BootConfig   `yaml:"jwt" json:"jwt"`
			Secure     rkmidsec.BootConfig   `yaml:"secure" json:"secure"`
			RateLimit  rkmidlimit.BootConfig `yaml:"rateLimit" json:"rateLimit"`
			Csrf       rkmidcsrf.BootConfig  `yaml:"csrf" yaml:"csrf"`
			Trace      rkmidtrace.BootConfig `yaml:"trace" json:"trace"`
		} `yaml:"middleware" json:"middleware"`
	} `yaml:"gf" json:"gf"`
}

// GfEntry implements rkentry.Entry interface.
type GfEntry struct {
	entryName          string                          `json:"-" yaml:"-"`
	entryType          string                          `json:"-" yaml:"-"`
	entryDescription   string                          `json:"-" yaml:"-"`
	LoggerEntry        *rkentry.LoggerEntry            `json:"-" yaml:"-"`
	EventEntry         *rkentry.EventEntry             `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	Port               uint64                          `json:"-" yaml:"-"`
	Server             *ghttp.Server                   `json:"-" yaml:"-"`
	SwEntry            *rkentry.SWEntry                `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	DocsEntry          *rkentry.DocsEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	PProfEntry         *rkentry.PProfEntry             `json:"-" yaml:"-"`
	Middlewares        []ghttp.HandlerFunc             `json:"-" yaml:"-"`
	bootstrapLogOnce   sync.Once                       `json:"-" yaml:"-"`
}

// RegisterGfEntryYAML register GoFrame entries with provided config file (Must YAML file).
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
func RegisterGfEntryYAML(raw []byte) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootGf{}
	rkentry.UnmarshalBootYAML(raw, config)

	// 2: Init GoFrame entries with boot config
	for i := range config.Gf {
		element := config.Gf[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		// logger entry
		loggerEntry := rkentry.GlobalAppCtx.GetLoggerEntry(element.LoggerEntry)
		if loggerEntry == nil {
			loggerEntry = rkentry.LoggerEntryStdout
		}

		// event entry
		eventEntry := rkentry.GlobalAppCtx.GetEventEntry(element.EventEntry)
		if eventEntry == nil {
			eventEntry = rkentry.EventEntryStdout
		}

		// cert entry
		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.CertEntry)

		// Register swagger entry
		swEntry := rkentry.RegisterSWEntry(&element.SW, rkentry.WithNameSWEntry(element.Name))

		// Register docs entry
		docsEntry := rkentry.RegisterDocsEntry(&element.Docs, rkentry.WithNameDocsEntry(element.Name))

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntry(&element.Prom, rkentry.WithRegistryPromEntry(promRegistry))

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntry(&element.CommonService)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntry(&element.Static, rkentry.WithNameStaticFileHandlerEntry(element.Name))

		// Register pprof entry
		pprofEntry := rkentry.RegisterPProfEntry(&element.PProf, rkentry.WithNamePProfEntry(element.Name))

		inters := make([]ghttp.HandlerFunc, 0)

		// add global path ignorance
		rkmid.AddPathToIgnoreGlobal(element.Middleware.Ignore...)

		// set error builder based on error builder
		switch strings.ToLower(element.Middleware.ErrorModel) {
		case "", "google":
			rkmid.SetErrorBuilder(rkerror.NewErrorBuilderGoogle())
		case "amazon":
			rkmid.SetErrorBuilder(rkerror.NewErrorBuilderAMZN())
		}

		// logging middlewares
		if element.Middleware.Logging.Enabled {
			inters = append(inters, rkgflog.Middleware(
				rkmidlog.ToOptions(&element.Middleware.Logging, element.Name, GfEntryType,
					loggerEntry, eventEntry)...))
		}

		// insert panic interceptor
		inters = append(inters, rkgfpanic.Middleware(
			rkmidpanic.WithEntryNameAndType(element.Name, GfEntryType)))

		// metrics middleware
		if element.Middleware.Prom.Enabled {
			inters = append(inters, rkgfprom.Middleware(
				rkmidprom.ToOptions(&element.Middleware.Prom, element.Name, GfEntryType,
					promRegistry, rkmidprom.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Middleware.Trace.Enabled {
			inters = append(inters, rkgftrace.Middleware(
				rkmidtrace.ToOptions(&element.Middleware.Trace, element.Name, GfEntryType)...))
		}

		// jwt middleware
		if element.Middleware.Jwt.Enabled {
			inters = append(inters, rkgfjwt.Middleware(
				rkmidjwt.ToOptions(&element.Middleware.Jwt, element.Name, GfEntryType)...))
		}

		// secure middleware
		if element.Middleware.Secure.Enabled {
			inters = append(inters, rkgfsec.Middleware(
				rkmidsec.ToOptions(&element.Middleware.Secure, element.Name, GfEntryType)...))
		}

		// csrf middleware
		if element.Middleware.Csrf.Enabled {
			inters = append(inters, rkgfcsrf.Middleware(
				rkmidcsrf.ToOptions(&element.Middleware.Csrf, element.Name, GfEntryType)...))
		}

		// Did we enabled cors interceptor?
		if element.Middleware.Cors.Enabled {
			inters = append(inters, rkgfcors.Middleware(
				rkmidcors.ToOptions(&element.Middleware.Cors, element.Name, GfEntryType)...))
		}

		// meta middleware
		if element.Middleware.Meta.Enabled {
			inters = append(inters, rkgfmeta.Middleware(
				rkmidmeta.ToOptions(&element.Middleware.Meta, element.Name, GfEntryType)...))
		}

		// auth middlewares
		if element.Middleware.Auth.Enabled {
			inters = append(inters, rkgfauth.Middleware(
				rkmidauth.ToOptions(&element.Middleware.Auth, element.Name, GfEntryType)...))
		}

		// rate limit middleware
		if element.Middleware.RateLimit.Enabled {
			inters = append(inters, rkgflimit.Middleware(
				rkmidlimit.ToOptions(&element.Middleware.RateLimit, element.Name, GfEntryType)...))
		}

		entry := RegisterGfEntry(
			WithLoggerEntry(loggerEntry),
			WithEventEntry(eventEntry),
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithSwEntry(swEntry),
			WithPromEntry(promEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithCertEntry(certEntry),
			WithDocsEntry(docsEntry),
			WithPProfEntry(pprofEntry),
			WithStaticFileHandlerEntry(staticEntry),
			WithMiddlewares(inters...))

		entry.AddMiddleware(inters...)

		res[name] = entry
	}

	return res
}

// RegisterGfEntry register GfEntry with options.
func RegisterGfEntry(opts ...GfEntryOption) *GfEntry {
	entry := &GfEntry{
		entryType:        GfEntryType,
		entryDescription: "Internal RK entry which helps to bootstrap with GoFrame framework.",
		LoggerEntry:      rkentry.NewLoggerEntryStdout(),
		EventEntry:       rkentry.NewEventEntryStdout(),
		Middlewares:      make([]ghttp.HandlerFunc, 0),
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "gf-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Server == nil {
		entry.Server = g.Server(entry.entryName)
		entry.Server.SetDumpRouterMap(false)
		entry.Server.SetLogger(rkgfinter.NewGLogger(entry.LoggerEntry))
		glog.SetStdoutPrint(false)
	}

	if entry.Port != 0 {
		entry.Server.SetPort(int(entry.Port))
	}

	// add entry name and entry type into loki syncer if enabled
	entry.LoggerEntry.AddEntryLabelToLokiSyncer(entry)
	entry.EventEntry.AddEntryLabelToLokiSyncer(entry)

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *GfEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *GfEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *GfEntry) GetDescription() string {
	return entry.entryDescription
}

// Bootstrap GfEntry.
func (entry *GfEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap", ctx)

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Server.BindHandler(entry.CommonServiceEntry.ReadyPath, ghttp.WrapF(entry.CommonServiceEntry.Ready))
		entry.Server.BindHandler(entry.CommonServiceEntry.GcPath, ghttp.WrapF(entry.CommonServiceEntry.Gc))
		entry.Server.BindHandler(entry.CommonServiceEntry.InfoPath, ghttp.WrapF(entry.CommonServiceEntry.Info))
		entry.Server.BindHandler(entry.CommonServiceEntry.AlivePath, ghttp.WrapF(entry.CommonServiceEntry.Alive))

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Server.BindHandler(path.Join(entry.SwEntry.Path, "*any"), ghttp.WrapF(entry.SwEntry.ConfigFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is Docs enabled?
	if entry.IsDocsEnabled() {
		// Bootstrap Docs entry.
		entry.Server.BindHandler(path.Join(entry.DocsEntry.Path, "*any"), ghttp.WrapF(entry.DocsEntry.ConfigFileHandler()))
		entry.DocsEntry.Bootstrap(ctx)
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

	// Is pprof enabled?
	if entry.IsPProfEnabled() {
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path), ghttp.WrapF(pprof.Index))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "cmdline"), ghttp.WrapF(pprof.Cmdline))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "profile"), ghttp.WrapF(pprof.Profile))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "symbol"), ghttp.WrapF(pprof.Symbol))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "trace"), ghttp.WrapF(pprof.Trace))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "allocs"), ghttp.WrapF(pprof.Handler("allocs").ServeHTTP))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "block"), ghttp.WrapF(pprof.Handler("block").ServeHTTP))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "goroutine"), ghttp.WrapF(pprof.Handler("goroutine").ServeHTTP))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "heap"), ghttp.WrapF(pprof.Handler("heap").ServeHTTP))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "mutex"), ghttp.WrapF(pprof.Handler("mutex").ServeHTTP))
		entry.Server.BindHandler(path.Join(entry.PProfEntry.Path, "threadcreate"), ghttp.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}

	go entry.startServer(event, logger)

	entry.bootstrapLogOnce.Do(func() {
		// Print link and logging message
		scheme := "http"
		if entry.IsTlsEnabled() {
			scheme = "https"
		}

		if entry.IsSwEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("SwaggerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.SwEntry.Path))
		}
		if entry.IsDocsEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("DocsEntry: %s://localhost:%d%s", scheme, entry.Port, entry.DocsEntry.Path))
		}
		if entry.IsPromEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PromEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PromEntry.Path))
		}
		if entry.IsStaticFileHandlerEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("StaticFileHandlerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.StaticFileEntry.Path))
		}
		if entry.IsCommonServiceEnabled() {
			handlers := []string{
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.ReadyPath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.AlivePath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.InfoPath),
			}

			entry.LoggerEntry.Info(fmt.Sprintf("CommonSreviceEntry: %s", strings.Join(handlers, ", ")))
		}
		if entry.IsPProfEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PProfEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PProfEntry.Path))
		}
		entry.EventEntry.Finish(event)
	})
}

// Interrupt GfEntry.
func (entry *GfEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt", ctx)

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

	if entry.IsDocsEnabled() {
		// Interrupt common service entry
		entry.DocsEntry.Interrupt(ctx)
	}

	if entry.IsPProfEnabled() {
		entry.PProfEntry.Interrupt(ctx)
	}

	if entry.Server != nil {
		if err := entry.Server.Shutdown(); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gf-server.", event.ListPayloads()...)
		}
	}

	rkentry.GlobalAppCtx.RemoveEntry(entry)

	entry.EventEntry.Finish(event)
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
		"name":                   entry.entryName,
		"type":                   entry.entryType,
		"description":            entry.entryDescription,
		"port":                   entry.Port,
		"swEntry":                entry.SwEntry,
		"docsEntry":              entry.DocsEntry,
		"commonServiceEntry":     entry.CommonServiceEntry,
		"promEntry":              entry.PromEntry,
		"staticFileHandlerEntry": entry.StaticFileEntry,
		"pprofEntry":             entry.PProfEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
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
	entryRaw := rkentry.GlobalAppCtx.GetEntry(GfEntryType, name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GfEntry)
	return entry
}

// AddMiddleware Add middleware.
// This function should be called before Bootstrap() called.
func (entry *GfEntry) AddMiddleware(inters ...ghttp.HandlerFunc) {
	entry.Server.Use(inters...)
}

// IsTlsEnabled Is TLS enabled?
func (entry *GfEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Certificate != nil
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

// IsDocsEnabled Is Docs entry enabled?
func (entry *GfEntry) IsDocsEnabled() bool {
	return entry.DocsEntry != nil
}

// IsPromEnabled Is prometheus entry enabled?
func (entry *GfEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// IsPProfEnabled Is pprof entry enabled?
func (entry *GfEntry) IsPProfEnabled() bool {
	return entry.PProfEntry != nil
}

// ***************** Helper function *****************

// Add basic fields into event.
func (entry *GfEntry) logBasicInfo(operation string, ctx context.Context) (rkquery.Event, *zap.Logger) {
	event := entry.EventEntry.Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))

	// extract eventId if exists
	if val := ctx.Value("eventId"); val != nil {
		if id, ok := val.(string); ok {
			event.SetEventId(id)
		}
	}

	logger := entry.LoggerEntry.With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.entryName),
		zap.String("entryType", entry.entryType))

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

	// add DocsEntry info
	if entry.IsDocsEnabled() {
		event.AddPayloads(
			zap.Bool("docsEnabled", true),
			zap.String("docsPath", entry.DocsEntry.Path))
	}

	// add pprofEntry info
	if entry.IsPProfEnabled() {
		event.AddPayloads(
			zap.Bool("pprofEnabled", true),
			zap.String("pprofPath", entry.PProfEntry.Path))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.Port),
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
			entry.Server.SetTLSConfig(&tls.Config{Certificates: []tls.Certificate{*entry.CertEntry.Certificate}})
		}

		err := entry.Server.Start()

		if err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Error("Error occurs while starting GoFrame server.", event.ListPayloads()...)
			rkentry.ShutdownWithError(err)
		}
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
		entry.entryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) GfEntryOption {
	return func(entry *GfEntry) {
		entry.entryDescription = description
	}
}

// WithLoggerEntry provide rkentry.LoggerEntry.
func WithLoggerEntry(zapLogger *rkentry.LoggerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.LoggerEntry = zapLogger
	}
}

// WithEventEntry provide rkentry.EventEntry.
func WithEventEntry(eventLogger *rkentry.EventEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EventEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SWEntry) GfEntryOption {
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

// WithMiddlewares provide user interceptors.
func WithMiddlewares(inters ...ghttp.HandlerFunc) GfEntryOption {
	return func(entry *GfEntry) {
		if entry.Middlewares == nil {
			entry.Middlewares = make([]ghttp.HandlerFunc, 0)
		}

		entry.Middlewares = append(entry.Middlewares, inters...)
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

// WithDocsEntry provide rkentry.DocsEntry.
func WithDocsEntry(docsEntry *rkentry.DocsEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.DocsEntry = docsEntry
	}
}

// WithPProfEntry provide rkentry.PProfEntry.
func WithPProfEntry(p *rkentry.PProfEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.PProfEntry = p
	}
}
