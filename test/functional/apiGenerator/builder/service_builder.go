package builder

import (
	"context"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
	"github.com/yomiji/gkBoot/test/functional/apiGenerator/rest"
)

type AmalgamationRequest struct {
	SpawnError string `query:"error"`
}

func (a AmalgamationRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "AmalgamationRequest",
		Method:      "GET",
		Path:        "/mixin",
		Description: "An amalgamation service",
	}
}

type AmalgamationResponse struct {
	Message string `json:"message"`
}

func (a AmalgamationResponse) StatusCode() int {
	return 201
}

type AmalgamationService struct{}

func (a AmalgamationService) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{
				Type: new(AmalgamationResponse),
				Code: 201,
			},
			{
				Type: new(rest.ErrorResponse),
				Code: 411,
			},
		},
	)
}

func (a AmalgamationService) Execute(_ context.Context, req interface{}) (interface{}, error) {
	r := req.(*AmalgamationRequest)
	if r.SpawnError != "" {
		if r.SpawnError != "error" {
			return struct {
				Errorable bool
			}{Errorable: true}, nil
		} else {
			errType := &rest.ErrorResponse{
				Message: "Fail",
				Code:    411,
			}
			return nil, errType
		}
	}
	
	resp := &AmalgamationResponse{Message: "Success"}
	return resp, nil
}

func NewAmalgamationService() service.Service {
	return gkBoot.NewServiceBuilder(new(AmalgamationService), config.WithStrictAPI()).MixinLogging().Build()
}
