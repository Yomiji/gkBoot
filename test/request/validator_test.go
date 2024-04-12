package request

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type ValidRequest struct {
	Count         int    `request:"header" alias:"Count" json:"count"`
	ValueCategory string `request:"query" json:"valueCat"`
}

func (v ValidRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "ValidRequest",
		Method:      request.GET,
		Path:        "/validate",
		Description: "Validate a request",
	}
}

func (v ValidRequest) Validate() error {
	if len(v.ValueCategory) < 1 {
		return fmt.Errorf("valueCat in query params is empty or not defined")
	}
	if v.Count < 3 {
		return fmt.Errorf("must have at least 3 in count")
	}
	return nil
}

func ExampleValidator() {
	decoder, err := gkBoot.GenerateRequestDecoder(new(ValidRequest))
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("GET", "http://localhost/validate", nil)
	req.Header.Set("Count", "123")
	val, err := decoder(context.TODO(), req)
	values := val.(*ValidRequest)
	if err != nil {
		fmt.Printf("validation error: %s\n", err.Error())
	}
	fmt.Printf("val.Count: %d\n", values.Count)
	// Output:
	// validation error: valueCat in query params is empty or not defined
	// val.Count: 123
}

func TestValidate(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(ValidRequest))
	if err != nil {
		t.Fail()
	}
	req, _ := http.NewRequest("GET", "http://localhost/validate?valueCat=defined", nil)
	req.Header.Set("Count", "123")
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), req)
		if err != nil {
			t.Fatalf("basic decoder/validator failure: %s", err.Error())
		}
		if _, ok := val.(*ValidRequest); !ok {
			t.Fatalf("type not correct: %T", val)
		}
	}
}

func TestInValidate(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(ValidRequest))
	if err != nil {
		t.Fail()
	}
	req, _ := http.NewRequest("GET", "http://localhost/validate", nil)
	req.Header.Set("Count", "123")
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), req)
		if err == nil {
			t.Fatalf("expected an error to be thrown")
		}
		if _, ok := val.(*ValidRequest); !ok {
			t.Fatalf("type not correct: %T", val)
		}
	}
}
