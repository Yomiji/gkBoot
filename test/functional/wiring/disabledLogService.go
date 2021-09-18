package wiring

import (
	"context"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/logging"
	"github.com/yomiji/gkBoot/request"
)

type DisabledLogRequest struct {}

func (d DisabledLogRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "DisabledLogRequest",
		Method:      request.GET,
		Path:        "/test2",
		Description: "Disabled Log Test",
	}
}

type DisabledResponse struct {
	Testvalue int `json:"testvalue"`
}

type LogWrappedTarget interface {
	WrapTarget()
}

type DisabledService struct {
	gkBoot.BasicService
	logging.LogSkip
}

func (s DisabledService) WrapTarget() {
	panic("implement me")
}

func (s DisabledService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	return DisabledResponse{ Testvalue: 999}, nil
}
