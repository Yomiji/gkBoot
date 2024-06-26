package config

import (
	"database/sql"
	"net/http"

	"github.com/yomiji/gkBoot/kitDefaults"
	"github.com/yomiji/gkBoot/logging"
	"github.com/yomiji/gkBoot/service"
)

// TLSConfig
//
// Represents the TLS configuration for the REST service.
//
// Fields:
// - serverCert: The server certificate file path or content.
// - serverKey: The server key file path or content.
//
// Usage Example:
//
//	tls := TLSConfig{
//	  serverCert: "/path/to/server.crt",
//	  serverKey: "/path/to/server.key",
//	}
//
//	bootConfig := BootConfig{
//	  TLS: tls,
//	}
type TLSConfig struct {
	serverCert string
	serverKey  string
}

func (t TLSConfig) IsZero() bool {
	return t.serverCert == "" && t.serverKey == ""
}

func (t TLSConfig) GetCert() string {
	return t.serverCert
}

func (t TLSConfig) GetKey() string {
	return t.serverKey
}

// BootConfig
//
// Used by gkBoot.GkBoot to build the REST service. Each option has a default value.
type BootConfig struct {
	// HttpPort
	//
	//  Default value: 8080
	//
	// Port that the http REST service runs on
	HttpPort *int
	// Logger
	//
	//  Default value: valid JSON gkBoot.Logger
	//
	// The core logging subsystem. Each service is wrapped with this and the output is
	// deferred to the end of each call.
	Logger logging.Logger
	// CustomConfig
	//
	//  Default value: nil
	//
	// A custom configuration. This is passed to each gkBoot.ConfigurableService.
	CustomConfig interface{}
	// Database
	//
	//  Default value: nil
	//
	// A sql database. This is passed to each gkBoot.DatabaseConfigurable gkBoot.Service.
	Database *sql.DB
	// RootPath
	//
	//  Default value: /
	//
	// The path for the root. All gkBoot.HttpRequest are relative to this path.
	RootPath *string
	// HttpOpts
	//
	//  Default value: []
	//
	// A set of http.ServerOption used when constructing route of each service. This is
	// attached to each service passed to gkBoot.GkBoot
	HttpOpts []kitDefaults.ServerOption
	// ServiceWrappers
	//
	//  Default value: []
	//
	// A set of gkBoot.Wrapper used when constructing the services themselves. These
	// wrappers will be applied in the order encountered in the array. Every invocation of
	// WithServiceWrapper will add a new wrapper to wrap around every wired service.
	//
	// The wrapping algorithm will traverse a service tree, following the
	// service.UpdatableWrappedService GetNext function. To prevent wrapping of lower tier
	// protected delegates, do not implement service.UpdatableWrappedService on the owner of the delegate.
	ServiceWrappers []service.Wrapper
	// Decorators
	//
	//  Default value: []
	//
	// A set of functions that will wrap the primary handler of all requests. This can be used for global
	// functionality that must happen on every request (related to the http.Handler).
	//
	// Not to be confused with service.Wrapper, which has a more direct business logic domain.
	// The service.Wrapper can check for and respond to the interfaces implemented on the service,
	// whereas Decorators cannot.
	Decorators []func(handler http.Handler) http.Handler
	// StrictOpenAPI
	//
	// Default value: false
	//
	// When true, all wired services must implement service.OpenAPICompatible interface and all
	// responses from the service must be declared in service.OpenAPICompatible ExpectedResponses function
	StrictOpenAPI bool

	// TLS configures the TLS settings for the REST service.
	TLS TLSConfig
}

// GkBootOption
//
// Option type used during wiring.
type GkBootOption func(config *BootConfig)

// WithTLSConfig sets the TLS configuration for the BootConfig.
// The TLS configuration includes server certificate and key.
// This option is used by gkBoot to build the REST service with TLS enabled.
func WithTLSConfig(cert, key string) GkBootOption {
	return func(config *BootConfig) {
		config.TLS = TLSConfig{serverCert: cert, serverKey: key}
	}
}

// WithCustomConfig
//
// Set the custom config used by each gkBoot.ConfigurableService
func WithCustomConfig(cfg interface{}) GkBootOption {
	return func(config *BootConfig) {
		config.CustomConfig = cfg
	}
}

// WithRootPath
//
// Set the root path of the http server for the REST endpoint
func WithRootPath(path string) GkBootOption {
	return func(config *BootConfig) {
		config.RootPath = &path
	}
}

// WithHttpPort
//
// Set the http port of the server
func WithHttpPort(port int) GkBootOption {
	return func(config *BootConfig) {
		config.HttpPort = &port
	}
}

// WithLogger
//
// Set a custom gkBoot.Logger that will be used by each service automatically on deferment
func WithLogger(logger logging.Logger) GkBootOption {
	return func(config *BootConfig) {
		config.Logger = logger
	}
}

// WithDatabase
//
// Set a common database used and shared by all services
func WithDatabase(db *sql.DB) GkBootOption {
	return func(config *BootConfig) {
		config.Database = db
	}
}

// WithHttpServerOpts
//
// Set server options used by all services on every request
func WithHttpServerOpts(opts ...kitDefaults.ServerOption) GkBootOption {
	return func(config *BootConfig) {
		config.HttpOpts = opts
	}
}

// WithServiceWrapper
//
// Appends the given service.Wrapper to the end of the service wrappers chain. The service wrappers are executed
// at the tail end of the service construction for each service. It may be necessary to include a type
// check.
func WithServiceWrapper(wrapper service.Wrapper) GkBootOption {
	return func(config *BootConfig) {
		config.ServiceWrappers = append(config.ServiceWrappers, wrapper)
	}
}

// WithServiceDecorator
//
// Uses the given decorator wrap all service requests in a handler chain.
// Useful for things like global CORS implementation rules. The decorator will wrap all requests.
func WithServiceDecorator(decorator func(handler http.Handler) http.Handler) GkBootOption {
	return func(config *BootConfig) {
		config.Decorators = append(config.Decorators, decorator)
	}
}

// WithStrictAPI
//
// When used, all services must implement service.OpenAPICompatible interface and all
// responses from the service must be declared in service.OpenAPICompatible ExpectedResponses function
func WithStrictAPI() GkBootOption {
	return func(config *BootConfig) {
		config.StrictOpenAPI = true
	}
}
