package client

import (
	"context"
	"go/types"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/test/tools"
)

type ZeroTestRequest struct {
	ImZero1    *string  `query:"im_zero_1"`
	ImZero2    int      `query:"im_zero_2"`
	ImZeroForm ZeroForm `formData:"im_zero_form"`
}

func (z ZeroTestRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "ZeroTest",
		Method:      request.POST,
		Path:        "/zero",
		Description: "",
	}
}

type ZeroForm struct {
	AlsoZero *int `json:"also_zero"`
}

type ZeroService struct{}

type ZeroResponse struct {
	response.ErrorResponse
}

func (z ZeroService) Execute(ctx context.Context, req interface{}) (any, error) {
	return req, nil
}

func TestZeroValueClient(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Make Request With Zero Values", func(subT *testing.T) {
			zeroRequest := &ZeroTestRequest{
				ImZero1: nil,
				ImZero2: 0,
				ImZeroForm: ZeroForm{
					AlsoZero: nil,
				},
			}

			err := gkBoot.DoRequest[*ZeroTestRequest, types.Nil](
				"http://localhost:8080", zeroRequest, nil,
			)
			if err != nil {
				subT.Fatalf("failure: %s\n", err.Error())
			}
		},
	)

	tools.Harness(
		[]gkBoot.ServiceRequest{
			{new(ZeroTestRequest), new(ZeroService)},
		}, []config.GkBootOption{}, runners, t,
	)
}
