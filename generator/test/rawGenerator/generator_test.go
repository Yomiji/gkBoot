package rawGenerator

import (
	"fmt"
	"github.com/swaggest/jsonschema-go"
	"reflect"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/generator"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

type Generator struct {
	MyPath   string `path:"my_path"`
	MyHeader string `header:"my_header"`
	MyQuery  string `header:"my_query"`
	MyCookie string `header:"my_cookie"`
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

func (g GeneratorResponse) JSONSchema() (jsonschema.Schema, error) {
	title := "GeneratorResponse"

	return jsonschema.Schema{
		Title: &title,

		ReflectType: reflect.TypeOf(g),
	}, nil
}

type GeneratorService struct {
	gkBoot.BasicService
}

func (g GeneratorService) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{
				Type: new(GeneratorResponse),
				Code: "200",
			},
		},
	)
}

var file = `
openapi: 3.0.3
info:
  title: ""
  version: ""
paths:
  /path/{my_path}:
    get:
      operationId: Generator
      parameters:
      - in: path
        name: my_path
        required: true
        schema:
          type: string
      - in: header
        name: my_header
        schema:
          type: string
      - in: header
        name: my_query
        schema:
          type: string
      - in: header
        name: my_cookie
        schema:
          type: string
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GeneratorResponse'
          description: OK
components:
  schemas:
    GeneratorResponse:
      properties:
        file:
          type: string
      type: object
`

func TestGenerator(t *testing.T) {
	//services := []gkBoot.ServiceRequest{{new(Generator), new(GeneratorService)}}
	//spec, _ := gkBoot.GenerateSpecification(services, nil)

	fmt.Println(file)
	fmt.Println()
	generator.GetOperations([]byte(file))
}
