package gkBoot

import (
	"context"
	"fmt"
	
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/yomiji/gkBoot/kitDefaults"
	"github.com/yomiji/gkBoot/service"
)

func GenerateSpecification(requests []ServiceRequest, optionalReflector *openapi3.Reflector) (
	openapi3.Reflector,
	error,
) {
	reflector := openapi3.Reflector{}
	if optionalReflector != nil {
		reflector = *optionalReflector
	}
	if reflector.Spec == nil {
		reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	}
	for _, request := range requests {
		op := openapi3.Operation{}
		
		// request parts
		method := string(request.Request.Info().Method)
		name := request.Request.Info().Name
		path := request.Request.Info().Path
		
		err := reflector.SetRequest(&op, request.Request, method)
		if err != nil {
			return reflector, err
		}
		
		if srv, ok := request.Service.(service.OpenAPICompatible); ok {
			for _, response := range srv.ExpectedResponses() {
				err = reflector.SetJSONResponse(&op, response.Type, response.ExpectedCode)
				if err != nil {
					return reflector, err
				}
			}
		} else {
			return reflector, fmt.Errorf(
				"service associated with %s is not OpenAPI Compatible",
				name,
			)
		}
		err = reflector.Spec.AddOperation(method, path, op)
		if err != nil {
			return reflector, err
		}
	}
	
	return reflector, nil
}

type openApiResponseValidatorService struct {
	mappedResponses service.MappedResponses
	next service.Service
}

func (o openApiResponseValidatorService) Execute(ctx context.Context, request interface{}) (
	response interface{}, err error,
) {
	response, err = o.next.Execute(ctx, request)
	
	// pass through strict errors without checking API conformity
	if err != nil {
		return response, err
	}
	
	if j, ok2 := response.(kitDefaults.HttpCoder); ok2 {
		if service.IsResponseValid(o.mappedResponses, response, j.StatusCode()) {
			return response, err
		}
	}
	
	return response, fmt.Errorf(
		"possible api violation, type or code not matched for response",
	)
}

func (o openApiResponseValidatorService) GetNext() service.Service {
	return o.next
}

func (o *openApiResponseValidatorService) UpdateNext(nxt service.Service) {
	o.next = nxt
}

func APIValidationWrapper(srv service.Service) service.Service {
	if _,ok := srv.(service.OpenAPICompatible); !ok {
		panic( fmt.Errorf("service is not OpenAPI compatible: %t", srv))
	}
	return &openApiResponseValidatorService{
		mappedResponses: srv.(service.OpenAPICompatible).ExpectedResponses(),
		next: srv,
	}
}
