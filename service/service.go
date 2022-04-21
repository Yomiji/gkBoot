package service

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"

	"github.com/yomiji/gkBoot/kitDefaults"
)

// UsingDB
//
// Embed this into any Service struct in order to gain the ability to reference
// the custom sql database object.
type UsingDB struct {
	db *sql.DB
}

// SetDatabase
//
// This member function may be very useful for template delegates using the wrapper
// pattern. When injecting delegates to your private service fields, check for
// DatabaseConfigurable and use this to set a passed database from the parent
// to the delegate (for example).
func (u *UsingDB) SetDatabase(db *sql.DB) {
	u.db = db
}

// GetDatabase
//
// This is the workhorse for the UsingDB embed and returns the saved db instance.
func (u UsingDB) GetDatabase() *sql.DB {
	return u.db
}

// Service
//
// Business logic for an endpoint. Execute satisfies the endpoint.Endpoint interface.
type Service interface {
	Execute(ctx context.Context, req any) (any, error)
}

// UpdatableWrappedService
//
// This interface is for services that wrap other services.
type UpdatableWrappedService interface {
	// GetNext
	//
	// Retrieves the wrapped service.
	GetNext() Service
	// UpdateNext
	//
	// Updates the wrapped service. When called, this will update the wrapped service
	UpdateNext(nxt Service)
}

// DatabaseConfigurable
//
// A service implementing this interface is able to use the database supplied by config.WithDatabase
// within its service business logic using GetDatabase member. The database will be supplied from the
// common configuration.
//
// Recommended to just embed the UsingDB struct to your service unless a special wiring is needed.
type DatabaseConfigurable interface {
	SetDatabase(db *sql.DB)
	GetDatabase() *sql.DB
}

// UsingConfig
//
// Embed this into any Service struct in order to gain the ability to reference
// the custom configuration object. Suggest creating a shared wrapper function to convert
// the empty interface to a concrete type.
//   func ToConfig(service ConfigurableService) *viper.Viper {
//       return service.GetConfig().(*viper.Viper)
//   }
type UsingConfig struct {
	config interface{}
}

// GetConfig
//
// Returns the configuration associated with the ConfigurableService.
// Suggest creating a shared wrapper function to convert the empty interface to a concrete type.
//   func ToConfig(service ConfigurableService) *viper.Viper {
//       return service.GetConfig().(*viper.Viper)
//   }
func (u UsingConfig) GetConfig() interface{} {
	return u.config
}

// SetConfig
//
// Sets the configuration associated with the ConfigurableService
//
// This may come in handy when using gkBoot.Wrapper delegates in your private business
// logic. Otherwise, the gkBoot wiring finds this and sets the config automatically.
func (u *UsingConfig) SetConfig(config interface{}) {
	u.config = config
}

// ConfigurableService
//
// The GetConfig method should return a configuration object that can be used by business logic.
//
// Recommended to just embed the UsingConfig struct into your service or embed BasicService.
type ConfigurableService interface {
	GetConfig() interface{}
	SetConfig(config interface{})
}

// HttpEncoder
//
// Objects that implement this interface will pass the defined function to the encoder part of
// the go-kit route definition
type HttpEncoder interface {
	Encode(ctx context.Context, w http.ResponseWriter, response interface{}) error
}

// Wrapper
//
// These functions are used to wrap your services with new functionality. Be sure to compose the service first or
// pass the wrapped service as a delegate field and use it in your package-private business logic.
type Wrapper func(srv Service) Service

// OptionsConfigurable
//
// Services that implement this interface are responsible for a unique set of custom options that are
// applied before each service call. These options are unique to the service
type OptionsConfigurable interface {
	// ServerOptions
	//
	// Returns a list of request options that are only applicable for a single service. These are read when the service
	// is first created and used throughout the service lifetime
	ServerOptions() []kitDefaults.ServerOption
}

// OpenAPICompatible
//
// Indicates that the service adheres to the OpenAPI standard. ExpectedResponses are all of the responses
// that the service will be expected to generate. If Strict mode is enabled (see config.Config) then
// the ExpectedResponses list will be used to validate all service execution.
type OpenAPICompatible interface {
	Service
	ExpectedResponses() MappedResponses
}

// RegisteredResponse
//
// Each RegisteredResponse is an expected response from the given OpenAPICompatible service.
type RegisteredResponse struct {
	ExpectedCode int
	Type         interface{}
	ReflectType  reflect.Type
}

// MappedResponses
//
// type alias: []RegisteredResponse
type MappedResponses []RegisteredResponse

// ResponseType
//
// slimmed down schema used to register service response types
type ResponseType struct {
	Type interface{}
	Code int
}

// ResponseTypes
//
// type alias: []ResponseType
type ResponseTypes []ResponseType

// RegisterResponses
//
// helper method used to generate the proper return value for ExpectedResponses of the OpenAPICompatible interface
func RegisterResponses(types []ResponseType) MappedResponses {
	m := make(MappedResponses, 0)

	for _, typ := range types {
		m = addResponse(m, typ.Type, typ.Code)
	}

	return m
}

func addResponse(responseMap MappedResponses, respType interface{}, statusCode int) MappedResponses {
	reflectType := reflect.ValueOf(respType).Type()

	return append(
		responseMap, RegisteredResponse{
			ExpectedCode: statusCode,
			Type:         respType,
			ReflectType:  reflectType,
		},
	)
}

// IsResponseValid
//
// check if a service response is valid based on its OpenAPICompatible implementation
func IsResponseValid(mappedResp MappedResponses, response interface{}, code int) bool {
	typ := reflect.ValueOf(response).Type()

	for _, resp := range mappedResp {
		if resp.ExpectedCode == code && resp.ReflectType.AssignableTo(typ) {
			return true
		}
	}

	return false
}
