package wiring

import (
	"context"
	"net/http"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/logging"
	"github.com/yomiji/gkBoot/request"
)

type DisabledLogRequest struct {}

func (d DisabledLogRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "DisabledLogRequest",
		Method:      http.MethodGet,
		Path:        "/test2",
		Description: "Disabled Log Test",
	}
}

type DisabledResponse struct {
	Testvalue int `json:"testvalue"`
}

type DisabledService struct {
	gkBoot.BasicService
	logging.LogSkip
}

func (s DisabledService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	return DisabledResponse{ Testvalue: 999}, nil
}
