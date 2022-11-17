package client

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
)

type contextRequestBodyLimitKey int
type contextRequestBodyLimitValue int

type contextHeadersKey int
type contextHeadersValue map[string]interface{}

const (
	requestBodyLimitKey contextRequestBodyLimitKey = -1
	ctxHeadersKey       contextHeadersKey          = -2
)

// GetFriendlyRequestName
//
// Gets the friendly request name of an object. If the object does not implement a known type,
// this function will use reflection to get the objects type name.
func GetFriendlyRequestName(req interface{}) string {
	t := reflect.TypeOf(req)

	if t == nil {
		return ""
	}

	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
	}

	if v, ok := req.(HttpRequest); ok {
		if v.Info().Name == "" {
			return t.Name()
		}
		return v.Info().Name
	}

	return fmt.Sprintf("%s", t.Name())
}

// GetRequestBodyLimit
//
// Gets the request body size limit that was assigned. If this value is not present, the result
// will be nil.
func GetRequestBodyLimit(ctx context.Context) *int {
	limitObj := ctx.Value(requestBodyLimitKey)
	if limit, ok := limitObj.(contextRequestBodyLimitValue); ok {
		limitVal := int(limit)
		return &limitVal
	}
	return nil
}

// GetCtxHeadersFromContext
//
// Retrieve the saved headers from the request context. Returns nil if not found.
func GetCtxHeadersFromContext(ctx context.Context) map[string]interface{} {
	if ctx != nil {
		if v, ok := ctx.Value(ctxHeadersKey).(contextHeadersValue); ok {
			return v
		}
	}
	return nil
}

// InjectCtxHeaders
//
// Inject the given context with the given headers. No-op if context is nil.
func InjectCtxHeaders(ctx *context.Context, headers map[string]interface{}) {
	if ctx == nil {
		return
	}
	ctxVal := context.WithValue(*ctx, ctxHeadersKey, contextHeadersValue(headers))
	*ctx = ctxVal
}

// MixinCtxHeaders
//
// Similar in function to InjectCtxHeaders, except this will use the WithValue pattern
// and return a wrapped context.Context value.
func MixinCtxHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return context.WithValue(ctx, ctxHeadersKey, contextHeadersValue(headers))
}

// GetAdditionalLogs
//
// If a response implements response.ExtendedLog, then this function will retrieve the
// attached logs and return.
func GetAdditionalLogs(resp interface{}) map[string]interface{} {
	if v, ok := resp.(ExtendedLog); ok {
		return v.GetAll()
	}
	return nil
}

func FromSwaggestTag(field reflect.StructField) (requestPart, alias, jsonAlias string, ok bool) {
	swaggestTags := []string{"path", "query", "formData", "cookie", "header"}
	var required bool
	if r, k := field.Tag.Lookup("required"); k {
		if r != "" {
			rBool, _ := strconv.ParseBool(r)
			required = rBool
		} else {
			required = true
		}
	}
	for _, structTag := range swaggestTags {
		var tag string
		if tag, ok = field.Tag.Lookup(structTag); ok {
			switch structTag {
			case "formData":
				requestPart = "form"
			default:
				if required {
					requestPart = structTag + "!"
				} else {
					requestPart = structTag
				}
			}

			alias = tag
			jsonAlias = tag
			ok = true
		}
	}

	return
}
