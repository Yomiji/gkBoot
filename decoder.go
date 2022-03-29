package gkBoot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yomiji/gkBoot/helpers"
	"github.com/yomiji/gkBoot/kitDefaults"
	"github.com/yomiji/gkBoot/request"
)

// GenerateRequestDecoder
//
// When used in go-kit, generates a json decoder function that translates http requests to go concrete objects.
//
// This reads 'request' tags in a concrete object to perform transformations from the relative http
// request parts to the given field in the result object. The 'request' tag may be applied to a field
// of type bool, string, int, int8, int16, int32, int64, float32, float64, complex64, complex128 or
// a slice of any of these (the relative part of the request must list slice values separated by comma).
//
// The 'request' tag itself may be structured thus:
//
//	type ConcreteObject struct {
//	  Value   string           `request:"header"`                     // find in headers with name "Value"
//	  MyInt   int              `request:"header" alias:"New-Integer"` // "New-Integer" in headers not "MyInt"
//    MyBool  bool             `request:"query" json:"myBool"`        // "myBool" in the query params of request
//	  MyFloat float32          `request:"query!"`                     // query param must be present as "MyFloat"
//	  Body    CustomBodyStruct `request:"form"`                       // json request body as an object (json.Unmarshal)
//	}
//
// Note that this function will look for the corresponding field values using the following naming hierarchy:
//  alias -> json -> field name (exported)
// This function will skip over unexported fields
//
// The resulting decoder function always returns a pointer to a new instantiation of the 'obj' argument.
//
// The obj argument may be a reference or a value type
func GenerateRequestDecoder(obj request.HttpRequest) (kitDefaults.DecodeRequestFunc, error) {
	reqObjType := reflect.TypeOf(obj)
	reqObjKind := reqObjType.Kind()

	if reqObjKind == reflect.Ptr {
		reqObjType = reqObjType.Elem()
	}

	if reqObjType.Kind() != reflect.Struct {
		objName := helpers.GetFriendlyRequestName(obj)
		return nil, fmt.Errorf("request object '%s' must be a Struct type", objName)
	}

	wv := reflect.New(reqObjType)
	cv := wv.Interface()
	if _, ok := cv.(jsonBody); ok {
		return func(ctx context.Context, h *http.Request) (interface{}, error) {
			// always get a new blank value on every request
			workingValue := reflect.New(reqObjType)
			concreteValue := workingValue.Interface()
			err := decodeStructBody(ctx, h, workingValue)
			if err != nil {
				return concreteValue, err
			}
			if validator, ok := concreteValue.(request.Validator); ok {
				err = validator.Validate()
			}
			return concreteValue, err
		}, nil
	}

	return func(ctx context.Context, request2 *http.Request) (req interface{}, err error) {
		// always get a new blank value on every request
		workingValue := reflect.New(reqObjType)
		concreteValue := workingValue.Interface()
		err = assignValues(ctx, request2, workingValue)
		if err != nil {
			return concreteValue, err
		}
		if validator, ok := concreteValue.(request.Validator); ok {
			err = validator.Validate()
		}

		return concreteValue, err
	}, nil
}

type jsonBody interface {
	isJsonBody()
}

// JSONBody
//
// When embedded into a request, flags the request as a JSON body to allow for automatic decoding.
type JSONBody struct{}

func (J JSONBody) isJsonBody() {}

func decodeStructBody(ctx context.Context, r *http.Request, workingValuePtr reflect.Value) error {
	baseVal := workingValuePtr
	// if the object is a pointer, get the dereference version. If it is nil, set a zeroed value.
	if baseVal.Kind() == reflect.Ptr {
		if baseVal.IsNil() && baseVal.CanSet() {
			baseVal.Set(reflect.New(baseVal.Type()))
		}
		baseVal = baseVal.Elem()
	}
	baseValType := baseVal.Type()

	// if no field ops, attempt body reading
	// begin to set form values using the interface type via json
	if !baseVal.CanSet() {
		return fmt.Errorf("can't set %s, check exporting", baseValType.Name())
	}
	body := reflect.New(baseVal.Type()).Interface()
	// set req body size limiter if sent to us
	limit := helpers.GetRequestBodyLimit(ctx)
	if limit != nil {
		err := readFormBody(r, body, *limit)
		if err != nil {
			return err
		}
	} else {
		err := readFormBody(r, body, 0)
		if err != nil {
			return err
		}
	}
	baseVal.Set(reflect.ValueOf(body).Elem())

	return nil
}

// HttpDecoder
//
// Objects that implement this interface will pass the defined function to the decoder part of
// the go-kit route definition
type HttpDecoder interface {
	Decode(ctx context.Context, httpRequest *http.Request) (request interface{}, err error)
}

// assignValues
//
// assigns the values of the given struct by iterating over the fields. This only assigns fields that
// are exported and tagged with 'request'
func assignValues(ctx context.Context, r *http.Request, workingValuePtr reflect.Value) error {
	baseVal := workingValuePtr
	// if the object is a pointer, get the dereference version. If it is nil, set a zeroed value.
	if baseVal.Kind() == reflect.Ptr {
		if baseVal.IsNil() && baseVal.CanSet() {
			baseVal.Set(reflect.New(baseVal.Type()))
		}
		baseVal = baseVal.Elem()
	}
	baseValType := baseVal.Type()
	baseValKind := baseValType.Kind()

	// ensure that the first value is always a kind of struct
	if baseValKind != reflect.Struct {
		objName := helpers.GetFriendlyRequestName(baseVal.Interface())
		return errors.New(fmt.Sprintf("request object '%s' must be a Struct type", objName))
	}

	// iterate over all the fields in the struct
	for i := 0; i < baseValType.NumField(); i++ {
		var err error
		fieldDesc := baseValType.Field(i)
		fieldVal := baseVal.Field(i)
		// if it is a pointer we need to init and get the element that is the concrete val
		if fieldDesc.Type.Kind() == reflect.Ptr {
			if fieldVal.IsNil() && fieldVal.CanSet() {
				// initialize pointers all the way down the chain, save the last element for assignment
				for ; fieldVal.Type().Kind() == reflect.Ptr; fieldVal = fieldVal.Elem() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
			} else if fieldVal.CanSet() {
				for ; fieldVal.Type().Kind() == reflect.Ptr; fieldVal = fieldVal.Elem() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
			}
		}
		requestTag, alias, jsonAlias := readTag(fieldDesc)
		if requestTag == "" && (fieldDesc.Type.Kind() == reflect.Struct || (fieldDesc.Anonymous && fieldVal.CanSet())) {
			// recurse if embedded structure
			return assignValues(ctx, r, fieldVal)
		} else if requestTag == "form" {
			// begin to set form values using the interface type via json
			if !fieldVal.CanSet() {
				return fmt.Errorf("field '%s' must be exported if using 'request'", fieldDesc.Name)
			}
			body := reflect.New(fieldVal.Type()).Interface()
			// set req body size limiter if sent to us
			limit := helpers.GetRequestBodyLimit(ctx)
			if limit != nil {
				err = readFormBody(r, body, *limit)
				if err != nil {
					return err
				}
			} else {
				err = readFormBody(r, body, 0)
				if err != nil {
					return err
				}
			}
			fieldVal.Set(reflect.ValueOf(body).Elem())
		} else if requestTag != "" {
			// if its just a normal field type, we can use this common logic to set it
			if !fieldVal.CanSet() {
				return errors.New(fmt.Sprintf("field '%s' must be exported if using 'request'", fieldDesc.Name))
			}
			operation := returnOperationByTagValue(requestTag)
			if operation == nil {
				return fmt.Errorf("unknown 'request' operation: %s", requestTag)
			}
			fieldName := fieldDesc.Name
			destType := fieldDesc.Type
			if jsonAlias != "" {
				fieldName = jsonAlias
			}
			if alias != "" {
				fieldName = alias
			}
			val, err := operation(r, fieldName, destType, strings.HasSuffix(requestTag, "!"))
			if err != nil {
				return err
			}
			fieldVal.Set(val)
		} else {
			continue
		}
	}
	return nil
}

func readTag(field reflect.StructField) (requestPart, alias, jsonAlias string) {
	var ok bool
	var tag string

	if requestPart, alias, jsonAlias, ok = fromSwaggestTag(field); ok {
		return
	}
	if tag, ok = field.Tag.Lookup("request"); ok {
		requestPart = tag
	}
	if tag, ok = field.Tag.Lookup("alias"); ok {
		alias = tag
	}
	if tag, ok = field.Tag.Lookup("json"); ok {
		if tag == "-," {
			jsonAlias = "-"
		} else {
			jsonAlias = strings.Split(tag, ",")[0]
			if jsonAlias == "-" {
				jsonAlias = ""
			}
		}
	}
	return
}

func fromSwaggestTag(field reflect.StructField) (requestPart, alias, jsonAlias string, ok bool) {
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

func returnOperationByTagValue(tagName string) typicalRequestType {
	switch tagName {
	case "cookie", "cookie!":
		return readRequestCookie
	case "header", "header!":
		return readRequestHeader
	case "query", "query!":
		return readRequestQuery
	case "path", "path!":
		return readPathParam
	default:
		return nil
	}
}

func checkRequired(fieldName, strVal string, isRequired bool) error {
	if isRequired {
		if strVal == "" {
			return errors.New(fmt.Sprintf("'%s' is missing a required value", fieldName))
		}
	}
	return nil
}

func checkCookieRequired(fieldName, strVal string, err error, isRequired bool) error {
	if isRequired {
		if strVal == "" {
			return errors.New(fmt.Sprintf("'%s' cookie is missing a required value: %s", fieldName, err))
		}
	}
	return nil
}

type typicalRequestType func(r *http.Request, fieldName string, destType reflect.Type, isRequired bool) (
// returns:
	reflect.Value, error,
)

func readRequestCookie(r *http.Request, fieldName string, destType reflect.Type, isRequired bool) (
// returns:
	reflect.Value, error,
) {
	cookie, err := r.Cookie(fieldName)
	if cookie == nil && isRequired {
		return reflect.Value{}, fmt.Errorf("required cookie not found or not set: %s", fieldName)
	} else if cookie == nil {
		return convertStringToValue("", destType, false)
	}
	if err := checkCookieRequired(fieldName, cookie.Value, err, isRequired); err != nil {
		return reflect.Value{}, err
	}
	return convertStringToValue(cookie.Value, destType, false)
}

func readRequestHeader(r *http.Request, fieldName string, destType reflect.Type, isRequired bool) (
// returns:
	reflect.Value, error,
) {
	headerStringValue := r.Header.Get(fieldName)
	if err := checkRequired(fieldName, headerStringValue, isRequired); err != nil {
		return reflect.Value{}, err
	}
	return convertStringToValue(headerStringValue, destType, false)
}

func readRequestQuery(r *http.Request, fieldName string, destType reflect.Type, isRequired bool) (
// returns:
	reflect.Value, error,
) {
	queryStringValue := r.URL.Query().Get(fieldName)
	if err := checkRequired(fieldName, queryStringValue, isRequired); err != nil {
		return reflect.Value{}, err
	}
	return convertStringToValue(queryStringValue, destType, false)
}

func readPathParam(r *http.Request, fieldName string, destType reflect.Type, isRequired bool) (reflect.Value, error) {
	pathStringValue := chi.URLParam(r, fieldName)
	if err := checkRequired(fieldName, pathStringValue, isRequired); err != nil {
		return reflect.Value{}, err
	}
	return convertStringToValue(pathStringValue, destType, false)
}

func readFormBody(r *http.Request, body interface{}, limit int) error {
	if limit > 0 {
		reader := io.LimitReader(r.Body, int64(limit))
		bytes, err := ioutil.ReadAll(bufio.NewReader(reader))
		if err != nil {
			return err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return err
		}
	} else {
		bytes, err := ioutil.ReadAll(bufio.NewReader(r.Body))
		if err != nil {
			return err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertStringToValue(src string, destType reflect.Type, reReference bool) (reflect.Value, error) {
	kind := destType.Kind()
	switch kind {
	case reflect.Ptr:
		dereferenceType := destType.Elem()
		if reReference {
			val, err := convertStringToValue(src, dereferenceType, reReference)
			if err != nil {
				return reflect.Zero(destType), err
			}
			dereferenceVal := reflect.New(dereferenceType)
			dereferenceVal.Elem().Set(val)
			return dereferenceVal, nil
		}
		return convertStringToValue(src, dereferenceType, reReference)
	}
	if src == "" {
		return reflect.Zero(destType), nil
	}
	switch kind {
	case reflect.String:
		return reflect.ValueOf(src), nil
	case reflect.Int:
		i, err := strconv.ParseInt(src, 10, 64)
		return reflect.ValueOf(int(i)), err
	case reflect.Bool:
		b, err := strconv.ParseBool(src)
		return reflect.ValueOf(b), err
	case reflect.Slice:
		elem := destType.Elem()

		strs := strings.Split(src, ",")
		tempSlice := reflect.MakeSlice(destType, 0, 0)
		for _, v := range strs {
			val, err := convertStringToValue(strings.TrimSpace(v), elem, true)
			if err != nil {
				return reflect.Value{}, errors.New(fmt.Sprintf("value '%s' error: %s", v, err))
			}
			tempSlice = reflect.Append(tempSlice, val)
		}
		return tempSlice, nil
	case reflect.Uint:
		i, err := strconv.ParseUint(src, 10, 64)
		return reflect.ValueOf(uint(i)), err
	case reflect.Float64:
		f, err := strconv.ParseFloat(src, 64)
		return reflect.ValueOf(f), err
	case reflect.Float32:
		f, err := strconv.ParseFloat(src, 32)
		return reflect.ValueOf(float32(f)), err
	case reflect.Int8:
		i, err := strconv.ParseInt(src, 10, 8)
		return reflect.ValueOf(int8(i)), err
	case reflect.Uint8:
		i, err := strconv.ParseInt(src, 10, 8)
		return reflect.ValueOf(uint8(i)), err
	case reflect.Int64:
		i, err := strconv.ParseInt(src, 10, 64)
		return reflect.ValueOf(i), err
	case reflect.Int32:
		i, err := strconv.ParseInt(src, 10, 32)
		return reflect.ValueOf(int32(i)), err
	case reflect.Int16:
		i, err := strconv.ParseInt(src, 10, 16)
		return reflect.ValueOf(int16(i)), err
	case reflect.Uint64:
		i, err := strconv.ParseInt(src, 10, 64)
		return reflect.ValueOf(uint64(i)), err
	case reflect.Uint16:
		i, err := strconv.ParseInt(src, 10, 16)
		return reflect.ValueOf(uint16(i)), err
	case reflect.Uint32:
		i, err := strconv.ParseInt(src, 10, 32)
		return reflect.ValueOf(uint32(i)), err
	case reflect.Complex64:
		c, err := strconv.ParseComplex(src, 64)
		return reflect.ValueOf(complex64(c)), err
	case reflect.Complex128:
		c, err := strconv.ParseComplex(src, 128)
		return reflect.ValueOf(c), err
	default:
		return reflect.Value{}, fmt.Errorf(
			"gkBoot: do not know how to set type %s for request value %s", destType.Name(), src,
		)
	}
}
