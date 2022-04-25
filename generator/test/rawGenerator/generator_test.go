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
	MyPath    string `path:"my_path"`
	MyHeader  string `header:"my_header"`
	MyHeader2 string `header:"my_header2" required:"true"`
	MyQuery   string `query:"my_query"`
	MyQuery2  string `query:"my_query2"`
	MyCookie  string `cookie:"my_cookie"`
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

//func (g GeneratorResponse) JSONSchema() (jsonschema.Schema, error) {
//	title := "GeneratorResponse"
//
//	return jsonschema.Schema{
//		Title: &title,
//	}, nil
//}

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
      - in: query
        name: my_query
        schema:
          type: string
      - in: query
        name: my_query2
        schema:
          type: string
      - in: path
        name: my_path
        required: true
        schema:
          type: string
      - in: cookie
        name: my_cookie
        schema:
          type: string
      - in: header
        name: my_header
        schema:
          type: string
      - in: header
        name: my_header2
        required: true
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
	//yaml, _ := spec.Spec.MarshalYAML()

	fmt.Println(string(file))
	fmt.Println()
	generator.GetOperations([]byte(file))
}
