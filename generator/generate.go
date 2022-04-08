package generator

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yomiji/gkBoot"
)

//TODO: lowercase package name
var t = `
import (
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

type {{.OperationId}} struct {
	{{if .PathParams}}{{makeParamLines .PathParams "path"}}{{- end}}
	{{if .QueryParams}}{{makeParamLines .QueryParams "query"}}{{end -}}
	{{if .HeaderParams}}{{makeParamLines .HeaderParams "header"}}{{end -}}
	{{if .CookieParams}}{{makeParamLines .CookieParams "cookie"}}{{- end}}
}

func (g {{.OperationId}}) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Method:      request.{{.Method}},
		Path:        "{{.Path}}",
		Description: "{{.Spec.Description}}",
	}
}

type {{.OperationId}}Service struct {
	gkBoot.BasicService
}

{{range .Spec.Responses}}
type {{
{{end}}
func (g {{.OperationId}}Service) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{{range .Spec.Responses}}
			{
				Type: {{if .Ref}}{{end}}
			}
			{{end}}
		},
	)
}
`

//TODO: range over response types
//TODO: fix response names in schema output
//  ensure schema is same on input to generate go code
//  responses that do not have any fields are StatusCoders

func GetOperations(sr []gkBoot.ServiceRequest) {
	spec, err := gkBoot.GenerateSpecification(sr, nil)
	if err != nil {
		panic(err)
	}
	yaml, _ := spec.Spec.MarshalYAML()
	loader := openapi3.NewLoader()

	data, err := loader.LoadFromData(yaml)
	if err != nil {
		panic(err)
	}
	ods, err := codegen.OperationDefinitions(data)
	if err != nil {
		panic(err)
	}

	tpl := template.New("gkBoot")
	if err != nil {
		panic(err)
	}

	tpl = tpl.Funcs(template.FuncMap{"makeParamLines": makeParamLines})
	tpl, err = tpl.Parse(t)
	if err != nil {
		panic(err)
	}

	for _, od := range ods {
		err := tpl.Execute(os.Stdout, od)
		if err != nil {
			return
		}
	}
}

func makeParamLines(params []codegen.ParameterDefinition, paramType string) string {
	lines := make([]string, 0, len(params))
	for _, cParam := range params {
		goType := cParam.Schema.GoType
		if cParam.Schema.RefType != "" {
			goType = cParam.Schema.RefType
		}

		name := cParam.GoName()
		pathName := cParam.ParamName
		param := fmt.Sprintf("%s %s `%s:\"%s\"", name, goType, paramType, pathName)
		if cParam.Required {
			param += " required:\"true\""
		}
		param += "`"
		lines = append(lines, param)
	}

	return strings.Join(lines, "\n")
}
