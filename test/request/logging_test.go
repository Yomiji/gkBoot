package request

import (
	"context"
	"net/http"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/test/tools"
)

type LoggingRequest struct {
	LoggableVal string `request:"header!"`
}

func (l LoggingRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "LoggingRequest",
		Method:      "GET",
		Path:        "/log",
		Description: "A loggable producer",
	}
}

type LoggingResponse struct {
	Response string `json:"response"`
	response.ExpandedLogging
}

type LoggingService struct {
	gkBoot.BasicService
}

func (l LoggingService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	req := request.(*LoggingRequest)
	resp := new(LoggingResponse)
	resp.Log("TestResponse", req.LoggableVal, req.Info(), "RouteInfo")
	resp.Response = req.LoggableVal
	return resp, nil
}

func TestCache(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Log Success", func(subT *testing.T) {
			// fist unique api call
			call1 := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:8080/log", map[string]string{
						"LoggableVal": "123",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			call1()
		},
	)
	tools.Harness([]gkBoot.ServiceRequest{{new(LoggingRequest), new(LoggingService)}}, nil, runners, t)
}
