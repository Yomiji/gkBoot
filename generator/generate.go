package generator

import (
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yomiji/gkBoot"
)

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

	for _, od := range ods {
		for _, pathParam := range od.PathParams {
			goType := pathParam.Schema.GoType
			if pathParam.Schema.RefType != "" {
				goType = pathParam.Schema.RefType
			}

			name := pathParam.GoName()
			pathName := pathParam.ParamName
			param := fmt.Sprintf("%s %s `path:\"%s\"", name, goType, pathName)
			if pathParam.Required {
				param += " required:\"true\""
			}
			param += "`"
			fmt.Println(param)
		}
	}
}
