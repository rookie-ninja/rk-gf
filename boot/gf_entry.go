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
	"github.com/rookie-ninja/rk-gf/interceptor"
	"github.com/rookie-ninja/rk-gf/interceptor/auth"
	"github.com/rookie-ninja/rk-gf/interceptor/cors"
	rkgfjwt "github.com/rookie-ninja/rk-gf/interceptor/jwt"
	"github.com/rookie-ninja/rk-gf/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gf/interceptor/meta"
	"github.com/rookie-ninja/rk-gf/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-gf/interceptor/panic"
	"github.com/rookie-ninja/rk-gf/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-gf/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// GfEntryType type of entry
	GfEntryType = "GfEntry"
	// GfEntryDescription description of entry
	GfEntryDescription = "Internal RK entry which helps to bootstrap with GoFrame framework."
)

var bootstrapEventIdKey = eventIdKey{}

type eventIdKey struct{}

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap GoFrame entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterGfEntriesWithConfig)
}

// BootConfigGf boot config which is for GoFrame entry.
//
// 1: gf.Enabled: Enable GoFrame entry, default is true.
// 2: gf.Name: Name of GoFrame entry, should be unique globally.
// 3: gf.Port: Port of GoFrame entry.
// 4: gf.Cert.Ref: Reference of rkentry.CertEntry.
// 5: gf.SW: See BootConfigSW for details.
// 6: gf.CommonService: See BootConfigCommonService for details.
// 7: gf.TV: See BootConfigTv for details.
// 8: gf.Prom: See BootConfigProm for details.
// 9: gf.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 10: gf.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 11: gf.Interceptors.auth.Enabled: Enable basic auth.
// 12: gf.Interceptors.auth.Basic: Credential for basic auth, scheme: <user:pass>
// 13: gf.Interceptors.auth.ApiKey: Credential for X-API-Key.
// 14: gf.Interceptors.auth.igorePrefix: List of paths that will be ignored.
// 15: gf.Interceptors.Extension.Enabled: Enable extension interceptor.
// 16: gf.Interceptors.Extension.Prefix: Prefix of extension header key.
// 17: gf.Interceptors.TracingTelemetry.Enabled: Enable tracing interceptor with opentelemetry.
// 18: gf.Interceptors.TracingTelemetry.Exporter.File.Enabled: Enable file exporter which support type of stdout and local file.
// 19: gf.Interceptors.TracingTelemetry.Exporter.File.OutputPath: Output path of file exporter, stdout and file path is supported.
// 20: gf.Interceptors.TracingTelemetry.Exporter.Jaeger.Enabled: Enable jaeger exporter.
// 21: gf.Interceptors.TracingTelemetry.Exporter.Jaeger.AgentEndpoint: Specify jeager agent endpoint, localhost:6832 would be used by default.
// 22: gf.Interceptors.RateLimit.Enabled: Enable rate limit interceptor.
// 23: gf.Interceptors.RateLimit.Algorithm: Algorithm of rate limiter.
// 24: gf.Interceptors.RateLimit.ReqPerSec: Request per second.
// 25: gf.Interceptors.RateLimit.Paths.path: Name of full path.
// 26: gf.Interceptors.RateLimit.Paths.ReqPerSec: Request per second by path.
// 27: gf.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 28: gf.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
type BootConfigGf struct {
	Gf []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          BootConfigProm          `yaml:"prom" json:"prom"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled                bool     `yaml:"enabled" json:"enabled"`
				ZapLoggerEncoding      string   `yaml:"zapLoggerEncoding" json:"zapLoggerEncoding"`
				ZapLoggerOutputPaths   []string `yaml:"zapLoggerOutputPaths" json:"zapLoggerOutputPaths"`
				EventLoggerEncoding    string   `yaml:"eventLoggerEncoding" json:"eventLoggerEncoding"`
				EventLoggerOutputPaths []string `yaml:"eventLoggerOutputPaths" json:"eventLoggerOutputPaths"`
			} `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"metricsProm" json:"metricsProm"`
			Auth struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				Basic        []string `yaml:"basic" json:"basic"`
				ApiKey       []string `yaml:"apiKey" json:"apiKey"`
			} `yaml:"auth" json:"auth"`
			Cors struct {
				Enabled          bool     `yaml:"enabled" json:"enabled"`
				AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
				AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
				AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
				AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
				ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
				MaxAge           int      `yaml:"maxAge" json:"maxAge"`
			} `yaml:"cors" json:"cors"`
			Meta struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Prefix  string `yaml:"prefix" json:"prefix"`
			} `yaml:"meta" json:"meta"`
			Jwt struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				SigningKey   string   `yaml:"signingKey" json:"signingKey"`
				SigningKeys  []string `yaml:"signingKeys" json:"signingKeys"`
				SigningAlgo  string   `yaml:"signingAlgo" json:"signingAlgo"`
				TokenLookup  string   `yaml:"tokenLookup" json:"tokenLookup"`
				AuthScheme   string   `yaml:"authScheme" json:"authScheme"`
			} `yaml:"jwt" json:"jwt"`
			RateLimit struct {
				Enabled   bool   `yaml:"enabled" json:"enabled"`
				Algorithm string `yaml:"algorithm" json:"algorithm"`
				ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"rateLimit" json:"rateLimit"`
			TracingTelemetry struct {
				Enabled  bool `yaml:"enabled" json:"enabled"`
				Exporter struct {
					File struct {
						Enabled    bool   `yaml:"enabled" json:"enabled"`
						OutputPath string `yaml:"outputPath" json:"outputPath"`
					} `yaml:"file" json:"file"`
					Jaeger struct {
						Agent struct {
							Enabled bool   `yaml:"enabled" json:"enabled"`
							Host    string `yaml:"host" json:"host"`
							Port    int    `yaml:"port" json:"port"`
						} `yaml:"agent" json:"agent"`
						Collector struct {
							Enabled  bool   `yaml:"enabled" json:"enabled"`
							Endpoint string `yaml:"endpoint" json:"endpoint"`
							Username string `yaml:"username" json:"username"`
							Password string `yaml:"password" json:"password"`
						} `yaml:"collector" json:"collector"`
					} `yaml:"jaeger" json:"jaeger"`
				} `yaml:"exporter" json:"exporter"`
			} `yaml:"tracingTelemetry" json:"tracingTelemetry"`
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
	EntryName          string                    `json:"entryName" yaml:"entryName"`
	EntryType          string                    `json:"entryType" yaml:"entryType"`
	EntryDescription   string                    `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry   *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	Port               uint64                    `json:"port" yaml:"port"`
	Server             *ghttp.Server             `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry        `json:"certEntry" yaml:"certEntry"`
	SwEntry            *SwEntry                  `json:"swEntry" yaml:"swEntry"`
	CommonServiceEntry *CommonServiceEntry       `json:"commonServiceEntry" yaml:"commonServiceEntry"`
	Interceptors       []ghttp.HandlerFunc       `json:"-" yaml:"-"`
	PromEntry          *PromEntry                `json:"promEntry" yaml:"promEntry"`
	TvEntry            *TvEntry                  `json:"tvEntry" yaml:"tvEntry"`
}

// GfEntryOption Gf entry option.
type GfEntryOption func(*GfEntry)

// GetGfEntry Get GfEntry from rkentry.GlobalAppCtx.
func GetGfEntry(name string) *GfEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GfEntry)
	return entry
}

// WithPortGf provide port.
func WithPortGf(port uint64) GfEntryOption {
	return func(entry *GfEntry) {
		entry.Port = port
	}
}

// WithNameGf provide name.
func WithNameGf(name string) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionGf provide name.
func WithDescriptionGf(description string) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EntryDescription = description
	}
}

// WithZapLoggerEntryGf provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryGf(zapLogger *rkentry.ZapLoggerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntryGf provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryGf(eventLogger *rkentry.EventLoggerEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntryGf provide rkentry.CertEntry.
func WithCertEntryGf(certEntry *rkentry.CertEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntryGf provide SwEntry.
func WithSwEntryGf(sw *SwEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntryGf provide CommonServiceEntry.
func WithCommonServiceEntryGf(commonServiceEntry *CommonServiceEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithInterceptorsGf provide user interceptors.
func WithInterceptorsGf(inters ...ghttp.HandlerFunc) GfEntryOption {
	return func(entry *GfEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]ghttp.HandlerFunc, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// WithPromEntryGf provide PromEntry.
func WithPromEntryGf(prom *PromEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.PromEntry = prom
	}
}

// WithTVEntryGf provide TvEntry.
func WithTVEntryGf(tvEntry *TvEntry) GfEntryOption {
	return func(entry *GfEntry) {
		entry.TvEntry = tvEntry
	}
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

		promRegistry := prometheus.NewRegistry()
		// Did we enabled swagger?
		var swEntry *SwEntry
		if element.SW.Enabled {
			// Init swagger custom headers from config
			headers := make(map[string]string, 0)
			for i := range element.SW.Headers {
				header := element.SW.Headers[i]
				tokens := strings.Split(header, ":")
				if len(tokens) == 2 {
					headers[tokens[0]] = tokens[1]
				}
			}

			swEntry = NewSwEntry(
				WithNameSw(fmt.Sprintf("%s-sw", element.Name)),
				WithZapLoggerEntrySw(zapLoggerEntry),
				WithEventLoggerEntrySw(eventLoggerEntry),
				WithEnableCommonServiceSw(element.CommonService.Enabled),
				WithPortSw(element.Port),
				WithPathSw(element.SW.Path),
				WithJsonPathSw(element.SW.JsonPath),
				WithHeadersSw(headers))
		}

		// Did we enabled prometheus?
		var promEntry *PromEntry
		if element.Prom.Enabled {
			var pusher *rkprom.PushGatewayPusher
			if element.Prom.Pusher.Enabled {
				certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Prom.Pusher.Cert.Ref)
				var certStore *rkentry.CertStore

				if certEntry != nil {
					certStore = certEntry.Store
				}

				pusher, _ = rkprom.NewPushGatewayPusher(
					rkprom.WithIntervalMSPusher(time.Duration(element.Prom.Pusher.IntervalMs)*time.Millisecond),
					rkprom.WithRemoteAddressPusher(element.Prom.Pusher.RemoteAddress),
					rkprom.WithJobNamePusher(element.Prom.Pusher.JobName),
					rkprom.WithBasicAuthPusher(element.Prom.Pusher.BasicAuth),
					rkprom.WithZapLoggerEntryPusher(zapLoggerEntry),
					rkprom.WithEventLoggerEntryPusher(eventLoggerEntry),
					rkprom.WithCertStorePusher(certStore))
			}

			promRegistry.Register(prometheus.NewGoCollector())
			promEntry = NewPromEntry(
				WithNameProm(fmt.Sprintf("%s-prom", element.Name)),
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithPromRegistryProm(promRegistry),
				WithEventLoggerEntryProm(eventLoggerEntry),
				WithPusherProm(pusher))

			if promEntry.Pusher != nil {
				promEntry.Pusher.SetGatherer(promEntry.Gatherer)
			}
		}

		inters := make([]ghttp.HandlerFunc, 0)

		// Did we enabled logging interceptor?
		if element.Interceptors.LoggingZap.Enabled {
			opts := []rkgflog.Option{
				rkgflog.WithEntryNameAndType(element.Name, GfEntryType),
				rkgflog.WithEventLoggerEntry(eventLoggerEntry),
				rkgflog.WithZapLoggerEntry(zapLoggerEntry),
			}

			if strings.ToLower(element.Interceptors.LoggingZap.ZapLoggerEncoding) == "json" {
				opts = append(opts, rkgflog.WithZapLoggerEncoding(rkgflog.ENCODING_JSON))
			}

			if strings.ToLower(element.Interceptors.LoggingZap.EventLoggerEncoding) == "json" {
				opts = append(opts, rkgflog.WithEventLoggerEncoding(rkgflog.ENCODING_JSON))
			}

			if len(element.Interceptors.LoggingZap.ZapLoggerOutputPaths) > 0 {
				opts = append(opts, rkgflog.WithZapLoggerOutputPaths(element.Interceptors.LoggingZap.ZapLoggerOutputPaths...))
			}

			if len(element.Interceptors.LoggingZap.EventLoggerOutputPaths) > 0 {
				opts = append(opts, rkgflog.WithEventLoggerOutputPaths(element.Interceptors.LoggingZap.EventLoggerOutputPaths...))
			}

			inters = append(inters, rkgflog.Interceptor(opts...))
		}

		// Did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkgfmetrics.Option{
				rkgfmetrics.WithRegisterer(promRegistry),
				rkgfmetrics.WithEntryNameAndType(element.Name, GfEntryType),
			}

			inters = append(inters, rkgfmetrics.Interceptor(opts...))
		}

		// Did we enabled tracing interceptor?
		if element.Interceptors.TracingTelemetry.Enabled {
			var exporter trace.SpanExporter

			if element.Interceptors.TracingTelemetry.Exporter.File.Enabled {
				exporter = rkgftrace.CreateFileExporter(element.Interceptors.TracingTelemetry.Exporter.File.OutputPath)
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Enabled {
				opts := make([]jaeger.AgentEndpointOption, 0)
				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host) > 0 {
					opts = append(opts,
						jaeger.WithAgentHost(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host))
				}
				if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port > 0 {
					opts = append(opts,
						jaeger.WithAgentPort(
							fmt.Sprintf("%d", element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port)))
				}

				exporter = rkgftrace.CreateJaegerExporter(jaeger.WithAgentEndpoint(opts...))
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Enabled {
				opts := []jaeger.CollectorEndpointOption{
					jaeger.WithUsername(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Username),
					jaeger.WithPassword(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Password),
				}

				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint) > 0 {
					opts = append(opts, jaeger.WithEndpoint(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint))
				}

				exporter = rkgftrace.CreateJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
			}

			opts := []rkgftrace.Option{
				rkgftrace.WithEntryNameAndType(element.Name, GfEntryType),
				rkgftrace.WithExporter(exporter),
			}

			inters = append(inters, rkgftrace.Interceptor(opts...))
		}

		// Did we enabled jwt interceptor?
		if element.Interceptors.Jwt.Enabled {
			var signingKey []byte
			if len(element.Interceptors.Jwt.SigningKey) > 0 {
				signingKey = []byte(element.Interceptors.Jwt.SigningKey)
			}

			opts := []rkgfjwt.Option{
				rkgfjwt.WithEntryNameAndType(element.Name, GfEntryType),
				rkgfjwt.WithSigningKey(signingKey),
				rkgfjwt.WithSigningAlgorithm(element.Interceptors.Jwt.SigningAlgo),
				rkgfjwt.WithTokenLookup(element.Interceptors.Jwt.TokenLookup),
				rkgfjwt.WithAuthScheme(element.Interceptors.Jwt.AuthScheme),
				rkgfjwt.WithIgnorePrefix(element.Interceptors.Jwt.IgnorePrefix...),
			}

			for _, v := range element.Interceptors.Jwt.SigningKeys {
				tokens := strings.SplitN(v, ":", 2)
				if len(tokens) == 2 {
					opts = append(opts, rkgfjwt.WithSigningKeys(tokens[0], tokens[1]))
				}
			}

			inters = append(inters, rkgfjwt.Interceptor(opts...))
		}

		// Did we enabled cors interceptor?
		if element.Interceptors.Cors.Enabled {
			opts := []rkgfcors.Option{
				rkgfcors.WithEntryNameAndType(element.Name, GfEntryType),
				rkgfcors.WithAllowOrigins(element.Interceptors.Cors.AllowOrigins...),
				rkgfcors.WithAllowCredentials(element.Interceptors.Cors.AllowCredentials),
				rkgfcors.WithExposeHeaders(element.Interceptors.Cors.ExposeHeaders...),
				rkgfcors.WithMaxAge(element.Interceptors.Cors.MaxAge),
				rkgfcors.WithAllowHeaders(element.Interceptors.Cors.AllowHeaders...),
				rkgfcors.WithAllowMethods(element.Interceptors.Cors.AllowMethods...),
			}

			inters = append(inters, rkgfcors.Interceptor(opts...))
		}

		// Did we enabled meta interceptor?
		if element.Interceptors.Meta.Enabled {
			opts := []rkgfmeta.Option{
				rkgfmeta.WithEntryNameAndType(element.Name, GfEntryType),
				rkgfmeta.WithPrefix(element.Interceptors.Meta.Prefix),
			}

			inters = append(inters, rkgfmeta.Interceptor(opts...))
		}

		// Did we enabled auth interceptor?
		if element.Interceptors.Auth.Enabled {
			opts := make([]rkgfauth.Option, 0)
			opts = append(opts,
				rkgfauth.WithEntryNameAndType(element.Name, GfEntryType),
				rkgfauth.WithBasicAuth(element.Name, element.Interceptors.Auth.Basic...),
				rkgfauth.WithApiKeyAuth(element.Interceptors.Auth.ApiKey...))

			// Add exceptional path
			if swEntry != nil {
				opts = append(opts, rkgfauth.WithIgnorePrefix(strings.TrimSuffix(swEntry.Path, "/")))
			}

			opts = append(opts, rkgfauth.WithIgnorePrefix("/rk/v1/assets"))
			opts = append(opts, rkgfauth.WithIgnorePrefix(element.Interceptors.Auth.IgnorePrefix...))

			inters = append(inters, rkgfauth.Interceptor(opts...))
		}

		// Did we enabled rate limit interceptor?
		if element.Interceptors.RateLimit.Enabled {
			opts := make([]rkgflimit.Option, 0)
			opts = append(opts,
				rkgflimit.WithEntryNameAndType(element.Name, GfEntryType))

			if len(element.Interceptors.RateLimit.Algorithm) > 0 {
				opts = append(opts, rkgflimit.WithAlgorithm(element.Interceptors.RateLimit.Algorithm))
			}
			opts = append(opts, rkgflimit.WithReqPerSec(element.Interceptors.RateLimit.ReqPerSec))

			for i := range element.Interceptors.RateLimit.Paths {
				e := element.Interceptors.RateLimit.Paths[i]
				opts = append(opts, rkgflimit.WithReqPerSecByPath(e.Path, e.ReqPerSec))
			}

			inters = append(inters, rkgflimit.Interceptor(opts...))
		}

		// Did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("%s-commonService", element.Name)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry))
		}

		// Did we enabled tv?
		var tvEntry *TvEntry
		if element.TV.Enabled {
			tvEntry = NewTvEntry(
				WithNameTv(fmt.Sprintf("%s-tv", element.Name)),
				WithZapLoggerEntryTv(zapLoggerEntry),
				WithEventLoggerEntryTv(eventLoggerEntry))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterGfEntry(
			WithZapLoggerEntryGf(zapLoggerEntry),
			WithEventLoggerEntryGf(eventLoggerEntry),
			WithNameGf(name),
			WithDescriptionGf(element.Description),
			WithPortGf(element.Port),
			WithSwEntryGf(swEntry),
			WithPromEntryGf(promEntry),
			WithCommonServiceEntryGf(commonServiceEntry),
			WithCertEntryGf(certEntry),
			WithTVEntryGf(tvEntry),
			WithInterceptorsGf(inters...))

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

	// insert panic interceptor
	entry.Interceptors = append(entry.Interceptors, rkgfpanic.Interceptor(
		rkgfpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

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

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Bootstrap GfEntry.
func (entry *GfEntry) Bootstrap(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Server.BindHandler(path.Join(entry.SwEntry.Path, "*any"), entry.SwEntry.ConfigFileHandler())
		entry.Server.BindHandler("/rk/v1/assets/sw/*", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.Server.BindHandler(entry.PromEntry.Path, ghttp.WrapH(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))

		// don't start with http handler, we will handle it by ourselves
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Server.BindHandler("/rk/v1/healthy", entry.CommonServiceEntry.Healthy)
		entry.Server.BindHandler("/rk/v1/gc", entry.CommonServiceEntry.Gc)
		entry.Server.BindHandler("/rk/v1/info", entry.CommonServiceEntry.Info)
		entry.Server.BindHandler("/rk/v1/configs", entry.CommonServiceEntry.Configs)
		entry.Server.BindHandler("/rk/v1/apis", entry.CommonServiceEntry.Apis)
		entry.Server.BindHandler("/rk/v1/sys", entry.CommonServiceEntry.Sys)
		entry.Server.BindHandler("/rk/v1/req", entry.CommonServiceEntry.Req)
		entry.Server.BindHandler("/rk/v1/entries", entry.CommonServiceEntry.Entries)
		entry.Server.BindHandler("/rk/v1/certs", entry.CommonServiceEntry.Certs)
		entry.Server.BindHandler("/rk/v1/logs", entry.CommonServiceEntry.Logs)
		entry.Server.BindHandler("/rk/v1/deps", entry.CommonServiceEntry.Deps)
		entry.Server.BindHandler("/rk/v1/license", entry.CommonServiceEntry.License)
		entry.Server.BindHandler("/rk/v1/readme", entry.CommonServiceEntry.Readme)
		entry.Server.BindHandler("/rk/v1/git", entry.CommonServiceEntry.Git)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}
	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Server.BindHandler("/rk/v1/tv", func(ctx *ghttp.Request) {
			ctx.Response.RedirectTo("/rk/v1/tv/", http.StatusTemporaryRedirect)
		})
		entry.Server.BindHandler("/rk/v1/tv/*item", entry.TvEntry.TV)
		entry.Server.BindHandler("/rk/v1/assets/tv/*", entry.TvEntry.AssetsFileHandler())

		entry.TvEntry.Bootstrap(ctx)
	}

	// Default interceptor should be at front
	entry.Server.Use(entry.Interceptors...)

	logger.Info("Bootstrapping GfEntry.", event.ListPayloads()...)

	go func(gfEntry *GfEntry) {
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
	}(entry)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt GfEntry.
func (entry *GfEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	entry.logBasicInfo(event)

	if entry.IsSwEnabled() {
		// Interrupt swagger entry
		entry.SwEntry.Interrupt(ctx)
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
		logger.Info("Interrupting GfEntry.", event.ListPayloads()...)
		if err := entry.Server.Shutdown(); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gf-server.", event.ListPayloads()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
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

// String Stringfy gf entry.
func (entry *GfEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *GfEntry) AddInterceptor(inters ...ghttp.HandlerFunc) {
	entry.Interceptors = append(entry.Interceptors, inters...)
}

// IsTlsEnabled Is TLS enabled?
func (entry *GfEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *GfEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
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

// Add basic fields into event.
func (entry *GfEntry) logBasicInfo(event rkquery.Event) {
	event.AddPayloads(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
		zap.Uint64("port", entry.Port),
	)
}
