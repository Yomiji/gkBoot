package wiring

import (
	"context"
	"net/http"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type TestRequest1 struct {
	TestNum int `request:"header" alias:"Test-Num" json:"testNum"`
}

func (t TestRequest1) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "TestRequest1",
		Method:      http.MethodGet,
		Path:        "/test1",
		Description: "N/A",
	}
}

type TestResponse struct {
	TestNumIs int `json:"testNum"`
	OptionalResponse1 int `json:"optional1"`
	OptionalResponse2 string `json:"optional2"`
}

type TestService1 struct {
	gkBoot.BasicService
}

func (s TestService1) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	return TestResponse{
		TestNumIs: request.(*TestRequest1).TestNum,
	}, nil
}
