package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
)

//TODO: lowercase package name
var t = `
import (
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/response"
	"github.com/yomiji/gkBoot/service"
)

type {{.OperationId}} struct {
	{{- if .PathParams}}
	{{- range $i, $line := makeParamLines .PathParams "path"}}
	{{$line -}}
	{{- end}}
	{{- end}}
	{{- if .QueryParams}}
	{{- range $i, $line := makeParamLines .QueryParams "query"}}
	{{$line -}}
	{{- end}}
	{{- end}}
	{{- if .HeaderParams}}
	{{- range $i, $line := makeParamLines .HeaderParams "header"}}
	{{$line -}}
	{{- end}}
	{{- end}}
	{{- if .CookieParams}}
	{{- range $i, $line := makeParamLines .CookieParams "cookie"}}
	{{$line -}}
	{{- end}}
	{{- end}}
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
{{$operationId := .OperationId}}{{range $key, $elem := .Spec.Responses}}
{{- if not $elem.Value.Content}}
type {{$operationId}}{{$key}} struct {
	response.BasicResponse
}
{{- end}}
{{- end}}
{{- range $i, $type := .ResponseTypes}}
type {{$type.Schema.TypeDecl}} struct {
	{{- range $i, $line := makePropertyLines $type.Schema.Properties}}
	{{$line -}}
	{{- end}}
}
{{- end}}
func (g {{.OperationId}}Service) ExpectedResponses() service.MappedResponses {
	return service.RegisterResponses(
		service.ResponseTypes{
			{{range $key, $elem := .Spec.Responses}}
			{{- if not $elem.Value.Content}}
			{
				Type: {{$operationId}}{{$key}}
				Code: "{{$key}}"
			},
			{{- else if (index $elem.Value.Content "application/json")}}
			{{- $content := (index $elem.Value.Content "application/json")}}
			{
				Type: {{getElemRefName $content.Schema.Ref}}
				Code: "{{$key}}"
			},
			{{- end}}
			{{- end}}
		},
	)
}
`

var m = `
func GkBootHandler() http.Handler {
}
`

func GetOperations(yaml []byte) {
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

	tpl = tpl.Funcs(
		template.FuncMap{
			"makeParamLines":    makeParamLines,
			"makePropertyLines": makePropertyLines,
			"getElemRefName":    getElemRefName,
		},
	)
	tpl, err = tpl.Parse(t)
	if err != nil {
		panic(err)
	}

	obj := struct {
		codegen.OperationDefinition
		ResponseTypes openapi3.Schemas
	}{}
	for _, od := range ods {
		obj.OperationDefinition = od
		obj.ResponseTypes = data.Components.Schemas
		fmt.Println("Properties")
		for k, v := range obj.ResponseTypes {
			fmt.Printf("K: %+v\n\tV:%+v\n", k, v.Value.Properties["file"].Value)
		}
		//for _, v := range obj.ResponseTypes {
		//	fmt.Printf("%+v\n", od.TypeDefinitions)
		//}
		//err := tpl.Execute(os.Stdout, obj)
		//if err != nil {
		//	return
		//}
	}
}

/*
Data Types
The data type of a schema is defined by the type keyword, for example, type: string. OpenAPI defines the following basic types:
string (this includes dates and files)
number
integer
boolean
array
object
*/

func typeToField(gt codegen.Schema, name string, paramType string, pathName string) string {
	goType := gt.GoType
	if gt.RefType != "" {
		goType = gt.RefType
	}
	line := fmt.Sprintf("%s %s `%s:\"%s\"", name, goType, paramType, pathName)
	line += "`"

	return line
}
func makeParamLines(params []codegen.ParameterDefinition, paramType string) []string {
	lines := make([]string, 0, len(params))

	for _, cParam := range params {
		name := cParam.GoName()
		pathName := cParam.ParamName
		param := typeToField(cParam.Schema, name, paramType, pathName)

		lines = append(lines, param)
	}

	return lines
}

func makePropertyLines(props []codegen.Property) []string {
	lines := make([]string, 0, len(props))

	for _, pProp := range props {
		name := pProp.GoFieldName()
		pathName := pProp.JsonFieldName
		param := typeToField(pProp.Schema, name, "json", pathName)

		lines = append(lines, param)
	}

	return lines
}

func getElemRefName(elemRef string) string {
	split := strings.Split(elemRef, "/")

	return split[len(split)-1]
}
