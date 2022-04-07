package rawGenerator

import (
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/generator"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

type GeneratorRequest struct {
	MyPath string `path:"my_path"`
}

func (g GeneratorRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "Generator",
		Method:      request.GET,
		Path:        "/path/{my_path}",
		Description: "",
	}
}

type GeneratorResponse struct{}

type GeneratorService struct {
	gkBoot.BasicService
}

func (g GeneratorService) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{
				Type: new(GeneratorResponse),
				Code: 200,
			},
		},
	)
}

func TestGenerator(t *testing.T) {
	services := []gkBoot.ServiceRequest{{new(GeneratorRequest), new(GeneratorService)}}

	generator.GetOperations(services)
}
