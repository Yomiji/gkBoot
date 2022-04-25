package rest

import (
	"context"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/service"
)

type TestRequest struct {
	Name    string  `header:"Name-Var" description:"a name" required:"true"`
	Cost    float32 `query:"cost" required:"true"`
	NewPath string  `path:"path"`
	MayOmit string  `query:"mustHave" required:"false"`
	MayErr  bool    `query:"mayErr"`
	Cookie  string  `cookie:"testCookie"`
}

func (t TestRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "TestRequest",
		Method:      request.GET,
		Path:        "/test/{path}",
		Description: "Test the Swaggest requests",
	}
}

type UnregisteredResponse struct {
	ImInvalid string `json:"invalid"`
}

type TestResponse struct {
	Name      string  `json:"message"`
	Cost      float32 `json:"cost"`
	Path      string  `json:"path"`
	CookieVal string  `json:"cookieVal"`
	response.BasicResponse
	response.ExpandedLogging
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e ErrorResponse) StatusCode() int {
	return e.Code
}

func (e ErrorResponse) Error() string {
	return e.Message
}

type TestService struct {
	gkBoot.BasicService
}

func (t TestService) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{
				Type: new(TestResponse),
				Code: 201,
			},
			{
				Type: new(ErrorResponse),
				Code: 411,
			},
		},
	)
}

func (t TestService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	req := request.(*TestRequest)
	resp := new(TestResponse)

	resp.Log("TestRequest", req)
	resp.NewCode(201)

	resp.Name = req.Name
	resp.Cost = req.Cost
	resp.Path = req.NewPath
	resp.CookieVal = req.Cookie

	if req.MayErr {
		errResp := new(ErrorResponse)
		errResp.Message = "new error"
		errResp.Code = 411
		return resp, errResp
	}

	if req.MayOmit != "" {
		return UnregisteredResponse{"invalid"}, nil
	}

	return resp, nil
}
