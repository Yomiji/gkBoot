package rest

import (
	"context"
	"fmt"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/service"
)

type TestRequest struct {
	Name string `header:"Name-Var" description:"a name" required:"true"`
	Cost float32 `query:"cost" required:""`
	NewPath string `path:"path"`
	MayOmit string `query:"mustHave" required:"false"`
	MayErr bool `query:"mayErr"`
	Cookie string `cookie:"testCookie"`
}

func (t TestRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "TestRequest",
		Method:      request.GET,
		Path:        "/test/{path}",
		Description: "Test the Swaggest requests",
	}
}

type TestResponse struct {
	Name string `json:"message"`
	Cost float32 `json:"cost"`
	Path string `json:"path"`
	CookieVal string `json:"cookieVal"`
	response.BasicResponse
	response.ExpandedLogging
}

type TestService struct {
	gkBoot.BasicService
}

func (t TestService) ExpectedResponses() []service.RegisteredResponse {
	return []service.RegisteredResponse{
		{
			ExpectedCode: 200,
			Type:         new(TestResponse),
		},
	}
}

func (t TestService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	req := request.(*TestRequest)
	resp := new(TestResponse)
	
	resp.Log("TestRequest", req)
	resp.NewCode(200)
	
	resp.Name = req.Name
	resp.Cost = req.Cost
	resp.Path = req.NewPath
	resp.CookieVal = req.Cookie
	
	if req.MayErr {
		return resp, fmt.Errorf("new error")
	}
	
	return resp, nil
}
