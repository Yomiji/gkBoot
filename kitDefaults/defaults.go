package kitDefaults

import (
	"context"
	"encoding/json"
	"net/http"
	
	"github.com/go-kit/kit/endpoint"
)

// DefaultHttpResponseEncoder
//
// Computes the http response encoding, for different formats, you must attach your own gkBoot.HttpEncoder
// to your gkBoot.Service for each one defined
func DefaultHttpResponseEncoder(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		DefaultHttpErrorEncoder(ctx, f.Failed(), w)
		return nil
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
