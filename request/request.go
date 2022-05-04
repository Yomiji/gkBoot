package request

import (
	"net/http"
)

type Method string

const (
	GET     Method = "GET"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	POST    Method = "POST"
	PATCH   Method = "PATCH"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
)

type HttpRouteInfo struct {
	// Name
	//
	// The name of the request that is sent to the logging middleware. This MUST be unique
	// among all other services that are defined.
	Name string
	// Method
	//
	// HTTP request Method
	Method Method
	// Path
	//  eg. "/path/one" or "v1/path/two"
	// The relative path to the custom root path provided.
	Path string
	// Description
	//
	// A helpful text that describes the service. This will appear in logs.
	Description string
}

// HttpRequest
//
// Indicates essential fields that provide the ability to tag and categorize
// requests for logging, metrics and other functionality.
type HttpRequest interface {
	Info() HttpRouteInfo
}

// Validator
//
// Indicates that the request will have an accompanying validation.
type Validator interface {
	// Validate
	//
	// This method, when implemented, will perform one or more validations of the state of the
	// request after being received from the client. If the state is not accepted, a non-nil result
	// is returned for the accompanying error.
	Validate() error
}

// Decorator
//
// Executes per request, before the request. Use to apply special CORS rules for a single request.
// The returned function will be passed the http.Handler associated with the request.
type Decorator interface {
	UsingDecorator() func(handler http.Handler) http.Handler
}

// OpenAPIExtended
//
// Provides the ability to attach openAPI extensions to a request object. These extensions will be parsed
// by the openapi spec generator and added to the operation.
type OpenAPIExtended interface {
	OpenAPIExtensions() map[string]interface{}
}

// OpenAPISecure
//
// Provides the ability to attach security extensions to a request object. These extensions will be parsed
// by the openapi spec generator and added to the operation.
type OpenAPISecure interface {
	OpenAPISecurity() []map[string][]string
}
