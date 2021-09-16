package requestStructure

import (
	"context"
	"fmt"
	"net/http"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type BasicRequest struct {
	Name  string `request:"header"`
	Slice []int  `request:"header" alias:"Slice-Field"`
	noOp  []bool `request:"header"` // unexported members ignored
}

func (b BasicRequest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func ExampleGenerateRequestDecoder() {
	decoder, err := gkBoot.GenerateRequestDecoder(new(BasicRequest))
	if err != nil {
		panic(err)
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Slice-Field", "1, 2, 3")
	request.Header.Set("noOp", "true, false, true")
	requestObject, _ := decoder(context.TODO(), request)
	basicRequest := requestObject.(*BasicRequest)
	fmt.Printf("Type: %T\n", basicRequest)
	fmt.Printf("basicRequest.Name: %s\n", basicRequest.Name)
	fmt.Printf("basicRequest.Slice: %v\n", basicRequest.Slice)
	fmt.Printf("basicRequest.noOp: %v\n", basicRequest.noOp)
	// Output:
	// Type: *requestStructure.BasicRequest
	// basicRequest.Name: testValue
	// basicRequest.Slice: [1 2 3]
	// basicRequest.noOp: []
}
