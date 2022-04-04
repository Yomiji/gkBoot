package client

import (
	"context"
	"strings"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/caching"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/test/tools"
)

type ClientTypesTestAnonymousStruct struct{}

type ClientTypesTestRequest struct {
	Type1 int     `path:"type_1"`
	Type2 *string `request:"query"`
	Type3 struct {
		Type4 int `json:"type_4"`
	} `header:"type_3"`
	Type5 complex128          `header:"type_5"`
	Type6 ClientTypesTestForm `required:"true" request:"form"`
	*ClientTypesTestAnonymousStruct
}

func (c ClientTypesTestRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "ClientTypes",
		Method:      request.POST,
		Path:        "/client/{type_1}/test",
		Description: "A test of types and positions",
	}
}

type ClientTypesTestForm struct {
	Type7 **int   `json:"type_7,omitempty"`
	Type8 float64 `json:"type_8,omitempty"`
	Type9 float32 `json:"type_9,omitempty"`
}

type ClientTypesResponse struct {
	Type7 **int   `json:"type_7,omitempty"`
	Type8 float64 `json:"type_8,omitempty"`
	Type9 float32 `json:"type_9,omitempty"`
	response.ErrorResponse
}

type ClientTypesTestService struct {
	gkBoot.BasicService
}

func (c ClientTypesTestService) Execute(ctx context.Context, r any) (any, error) {
	req := r.(*ClientTypesTestRequest)

	return req.Type6, nil
}

func TestClientTypes(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Get Multiple Types", func(subT *testing.T) {
			fancyInt := 333
			fancyString := "my test"
			fancyIntPtr := &fancyInt
			Request := ClientTypesTestRequest{
				Type1: fancyInt,
				Type2: &fancyString,
				Type3: struct {
					Type4 int `json:"type_4"`
				}{Type4: 21},
				Type5: 1 + 3i,
				Type6: ClientTypesTestForm{
					Type7: &fancyIntPtr,
					Type8: 3e-5,
					Type9: 1e-3,
				},
			}
			Response := &ClientTypesResponse{}
			err := gkBoot.DoRequest("http://localhost:8080", Request, Response)
			if err != nil {
				subT.Fatalf("unexpected error: %s", err)
			}

			if **Response.Type7 != 333 {
				subT.Fatalf("failed to find value, expected 333 got %d", **Response.Type7)
			}
		},
	).Test(
		"Get Header Value", func(subT *testing.T) {
			fancyInt := 333
			fancyString := "my test"
			fancyIntPtr := &fancyInt
			Request := ClientTypesTestRequest{
				Type1: fancyInt,
				Type2: &fancyString,
				Type3: struct {
					Type4 int `json:"type_4"`
				}{Type4: 21},
				Type5: 1 + 3i,
				Type6: ClientTypesTestForm{
					Type7: &fancyIntPtr,
					Type8: 3e-5,
					Type9: 1e-3,
				},
			}
			ClientRequest, err := gkBoot.GenerateClientRequest("http://localhost:8080", Request)
			if err != nil {
				subT.Fatalf("unexpected error: %s", err)
			}

			if !strings.Contains(ClientRequest.Header.Get("type_3"), "type_4") {
				subT.Fatalf(
					"failed to find value in header, expected 'type_3' got '%s'",
					ClientRequest.Header.Get("type_3"),
				)
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(ClientTypesTestRequest), new(ClientTypesTestService)}},
		[]config.GkBootOption{caching.WithCache(new(tools.Cache))}, runners, t,
	)
}
