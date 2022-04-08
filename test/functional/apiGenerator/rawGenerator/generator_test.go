package rawGenerator

import (
	"fmt"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/generator"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

type Generator struct {
	MyPath   string `path:"my_path"`
	MyHeader string `header:"my_header"`
}

func (g Generator) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Method:      request.GET,
		Path:        "/path/{my_path}",
		Description: "Just a description",
	}
}

type GeneratorResponse struct {
	File string `json:"file"`
}

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
	services := []gkBoot.ServiceRequest{{new(Generator), new(GeneratorService)}}
	spec, _ := gkBoot.GenerateSpecification(services, nil)
	yaml, _ := spec.Spec.MarshalYAML()

	fmt.Println(string(yaml))
	fmt.Println()
	generator.GetOperations(services)
}
