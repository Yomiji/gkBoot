package kitDefaults

import (
	"context"
	"encoding/json"
	"net/http"
)

// Failer may be implemented by Go kit response types that contain business
// logic error details. If Failed returns a non-nil error, the Go kit transport
// layer may interpret this as a business logic error, and may encode it
// differently than a regular, successful response.
//
// It's not necessary for your response types to implement Failer, but it may
// help for more sophisticated use cases. The addsvc example shows how Failer
// should be used by a complete application.
type Failer interface {
	Failed() error
}

// HttpCoder is any response that returns a status code
type HttpCoder interface {
	StatusCode() int
}

// DecodeRequestFunc any decoder that takes on this format, goal is to translate http.Request to
// API request object
type DecodeRequestFunc func(context.Context, *http.Request) (request interface{}, err error)

// EncodeResponseFunc any encoder that takes on this format, goal is to translate API response
// object to the http.ResponseWriter
type EncodeResponseFunc func(context.Context, http.ResponseWriter, interface{}) error

// DefaultHttpResponseEncoder
//
// Computes the http response encoding, for different formats, you must attach your own gkBoot.HttpEncoder
// to your gkBoot.Service for each one defined
func DefaultHttpResponseEncoder(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(Failer); ok && f.Failed() != nil {
		DefaultHttpErrorEncoder(ctx, f.Failed(), w)
		
		return nil
	} else if coder, ok := response.(HttpCoder); ok {
		code := coder.StatusCode()
		
		// overwrite default nonsense code
		if code == 0 {
			code = 200
		}
		
		w.WriteHeader(code)
	}
	
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	return json.NewEncoder(w).Encode(response)
}

// DefaultHttpErrorEncoder
//
// Computes the default http error response. When implementing custom gkBoot.HttpEncoder, ensure
// to implement your own error encoder handler.
func DefaultHttpErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	if errorWithCode, ok := err.(CodedError); ok {
		w.WriteHeader(errorWithCode.StatusCode())
	}
	_ = json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

type errorWrapper struct {
	Error string `json:"error"`
}

// CodedError
//
// Indicates an error capable of storing a status code. Used with default (en/de)-coders
type CodedError interface {
	StatusCode() int
}
