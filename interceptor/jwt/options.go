// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfjwt

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gf/interceptor"
	"net/http"
	"reflect"
	"strings"
)

// Interceptor would distinguish auth set based on.
var (
	optionsMap     = make(map[string]*optionSet)
	defaultSkipper = func(*ghttp.Request) bool {
		return false
	}
	errJwtMissing = rkerror.New(
		rkerror.WithHttpCode(http.StatusBadRequest),
		rkerror.WithMessage("missing or malformed jwt"))
	errJwtInvalid = rkerror.New(
		rkerror.WithHttpCode(http.StatusUnauthorized),
		rkerror.WithMessage("invalid or expired jwt"))
)

const (
	// AlgorithmHS256 is default algorithm for jwt
	AlgorithmHS256      = "HS256"
	headerAuthorization = "Authorization"
)

// Skipper default skipper will always return false
type Skipper func(*ghttp.Request) bool

// jwt extractor
type jwtExtractor func(*ghttp.Request) (string, error)

// ParseTokenFunc parse token func
type ParseTokenFunc func(auth string, ctx *ghttp.Request) (*jwt.Token, error)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:        rkgfinter.RpcEntryNameValue,
		EntryType:        rkgfinter.RpcEntryTypeValue,
		Skipper:          defaultSkipper,
		SigningKeys:      make(map[string]interface{}),
		SigningAlgorithm: AlgorithmHS256,
		IgnorePrefix:     make([]string, 0),
		Claims:           jwt.MapClaims{},
		TokenLookup:      "header:" + headerAuthorization,
		AuthScheme:       "Bearer",
	}

	set.KeyFunc = set.defaultKeyFunc
	set.ParseTokenFunc = set.defaultParseToken

	for i := range opts {
		opts[i](set)
	}

	sources := strings.Split(set.TokenLookup, ",")
	for _, source := range sources {
		parts := strings.Split(source, ":")

		switch parts[0] {
		case "query":
			set.extractors = append(set.extractors, jwtFromQuery(parts[1]))
		case "param":
			set.extractors = append(set.extractors, jwtFromParam(parts[1]))
		case "cookie":
			set.extractors = append(set.extractors, jwtFromCookie(parts[1]))
		case "form":
			set.extractors = append(set.extractors, jwtFromForm(parts[1]))
		case "header":
			set.extractors = append(set.extractors, jwtFromHeader(parts[1], set.AuthScheme))
		}
	}

	// default skipper was used, override it with ignoring prefix
	if reflect.ValueOf(set.Skipper).Pointer() == reflect.ValueOf(defaultSkipper).Pointer() {
		set.Skipper = func(ctx *ghttp.Request) bool {
			if ctx == nil || ctx.Request == nil {
				return false
			}

			urlPath := ctx.Request.URL.Path

			for i := range set.IgnorePrefix {
				if strings.HasPrefix(urlPath, set.IgnorePrefix[i]) {
					return true
				}
			}

			return false
		}
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	// EntryName name of entry
	EntryName string
	// EntryType type of entry
	EntryType string
	// Skipper function
	Skipper Skipper
	// IgnorePrefix ignoring paths prefix
	IgnorePrefix []string
	extractors   []jwtExtractor
	// SigningKey Signing key to validate token.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKeys is provided.
	SigningKey interface{}

	// SigningKeys Map of signing keys to validate token with kid field usage.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither user-defined KeyFunc nor SigningKey is provided.
	SigningKeys map[string]interface{}

	// SigningAlgorithm Signing algorithm used to check the token's signing algorithm.
	// Optional. Default value HS256.
	SigningAlgorithm string

	// Claims are extendable claims data defining token content. Used by default ParseTokenFunc implementation.
	// Not used if custom ParseTokenFunc is set.
	// Optional. Default value jwt.MapClaims
	Claims jwt.Claims

	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	// - "form:<name>"
	// Multiply sources example:
	// - "header: Authorization,cookie: myowncookie"
	TokenLookup string

	// AuthScheme to be used in the Authorization header.
	// Optional. Default value "Bearer".
	AuthScheme string

	// KeyFunc defines a user-defined function that supplies the public key for a token validation.
	// The function shall take care of verifying the signing algorithm and selecting the proper key.
	// A user-defined KeyFunc can be useful if tokens are issued by an external party.
	// Used by default ParseTokenFunc implementation.
	//
	// When a user-defined KeyFunc is provided, SigningKey, SigningKeys, and SigningMethod are ignored.
	// This is one of the three options to provide a token validation key.
	// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
	// Required if neither SigningKeys nor SigningKey is provided.
	// Not used if custom ParseTokenFunc is set.
	// Default to an internal implementation verifying the signing algorithm and selecting the proper key.
	KeyFunc jwt.Keyfunc

	// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
	// parsing fails or parsed token is invalid.
	// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
	ParseTokenFunc func(auth string, c *ghttp.Request) (*jwt.Token, error)
}

func (set *optionSet) defaultKeyFunc(t *jwt.Token) (interface{}, error) {
	// check the signing method
	if t.Method.Alg() != set.SigningAlgorithm {
		return nil, fmt.Errorf("unexpected jwt signing algorithm=%v", t.Header["alg"])
	}

	// check kid in token first
	// https://www.rfc-editor.org/rfc/rfc7515#section-4.1.4
	if len(set.SigningKeys) > 0 {
		if kid, ok := t.Header["kid"].(string); ok {
			if key, ok := set.SigningKeys[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
	}

	// return signing key
	return set.SigningKey, nil
}

func (set *optionSet) defaultParseToken(auth string, ctx *ghttp.Request) (*jwt.Token, error) {
	token := new(jwt.Token)
	var err error

	// implementation of jwt.MapClaims
	if _, ok := set.Claims.(jwt.MapClaims); ok {
		token, err = jwt.Parse(auth, set.KeyFunc)
	} else {
		// custom implementation of jwt.Claims
		t := reflect.ValueOf(set.Claims).Type().Elem()
		claims := reflect.New(t).Interface().(jwt.Claims)
		token, err = jwt.ParseWithClaims(auth, claims, set.KeyFunc)
	}

	// return error
	if err != nil {
		return nil, err
	}

	// invalid token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// WithSkipper provide Skipper.
func WithSkipper(skip Skipper) Option {
	return func(opt *optionSet) {
		opt.Skipper = skip
	}
}

// WithSigningKey provide SigningKey.
func WithSigningKey(key interface{}) Option {
	return func(opt *optionSet) {
		if key != nil {
			opt.SigningKey = key
		}
	}
}

// WithSigningKeys provide SigningKey with key and value.
func WithSigningKeys(key string, value interface{}) Option {
	return func(opt *optionSet) {
		if len(key) > 0 {
			opt.SigningKeys[key] = value
		}
	}
}

// WithSigningAlgorithm provide signing algorithm.
// Default is HS256.
func WithSigningAlgorithm(algo string) Option {
	return func(opt *optionSet) {
		if len(algo) > 0 {
			opt.SigningAlgorithm = algo
		}
	}
}

// WithClaims provide jwt.Claims.
func WithClaims(claims jwt.Claims) Option {
	return func(opt *optionSet) {
		opt.Claims = claims
	}
}

// WithTokenLookup provide lookup configs.
// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract token from the request.
// Optional. Default value "header:Authorization".
// Possible values:
// - "header:<name>"
// - "query:<name>"
// - "param:<name>"
// - "cookie:<name>"
// - "form:<name>"
// Multiply sources example:
// - "header: Authorization,cookie: myowncookie"
func WithTokenLookup(lookup string) Option {
	return func(opt *optionSet) {
		if len(lookup) > 0 {
			opt.TokenLookup = lookup
		}
	}
}

// WithAuthScheme provide auth scheme.
// Default is Bearer
func WithAuthScheme(scheme string) Option {
	return func(opt *optionSet) {
		if len(scheme) > 0 {
			opt.AuthScheme = scheme
		}
	}
}

// WithIgnorePrefix provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithIgnorePrefix(paths ...string) Option {
	return func(set *optionSet) {
		set.IgnorePrefix = append(set.IgnorePrefix, paths...)
	}
}

// WithKeyFunc provide user defined key func.
func WithKeyFunc(f jwt.Keyfunc) Option {
	return func(opt *optionSet) {
		opt.KeyFunc = f
	}
}

// WithParseTokenFunc provide user defined token parse func.
func WithParseTokenFunc(f ParseTokenFunc) Option {
	return func(opt *optionSet) {
		opt.ParseTokenFunc = f
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(ctx *ghttp.Request) (string, error) {
		auth := ctx.Request.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
			return auth[l+1:], nil
		}
		return "", errJwtMissing.Err
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(ctx *ghttp.Request) (string, error) {
		if raw := ctx.GetQuery(param); raw != nil {
			token := raw.String()
			if token == "" {
				return "", errJwtMissing.Err
			}
			return token, nil
		}
		return "", errJwtMissing.Err
	}
}

// jwtFromParam returns a `jwtExtractor` that extracts token from the url param string.
func jwtFromParam(param string) jwtExtractor {
	return func(ctx *ghttp.Request) (string, error) {
		if raw := ctx.GetParam(param); raw != nil {
			token := raw.String()
			if token == "" {
				return "", errJwtMissing.Err
			}
			return token, nil
		}

		return "", errJwtMissing.Err
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(ctx *ghttp.Request) (string, error) {
		for _, cookie := range ctx.Cookies() {
			if cookie.Name == name {
				return cookie.Value, nil
			}
		}

		return "", errJwtMissing.Err
	}
}

// jwtFromForm returns a `jwtExtractor` that extracts token from the form field.
func jwtFromForm(name string) jwtExtractor {
	return func(ctx *ghttp.Request) (string, error) {
		field := ctx.FormValue(name)
		if field == "" {
			return "", errJwtMissing.Err
		}
		return field, nil
	}
}
