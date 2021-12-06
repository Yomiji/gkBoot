package builder

import (
	"testing"
	
	"github.com/yomiji/gkBoot/test/functional/apiGenerator/rest"
)

func TestServiceBuilder(t *testing.T) {
	t.Run(
		"OpenAPI Conforms to Result", func(subT *testing.T) {
			var resp interface{}
			var err error
			func() {
				resp, err = NewAmalgamationService().Execute(nil, &AmalgamationRequest{SpawnError: ""})
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}()
			logicResponse := resp.(*AmalgamationResponse)
			if logicResponse.Message != "Success" {
				subT.Fatalf("failed, expected Success, got: %s", logicResponse.Message)
			}
		},
	)
	t.Run(
		"OpenAPI Does not Conform", func(subT *testing.T) {
			var err error
			func() {
				_, err = NewAmalgamationService().Execute(nil, &AmalgamationRequest{SpawnError: "failure"})
				if err == nil {
					subT.Fatalf("failed, expected api incompatible")
				}
			}()
		},
	)
	t.Run(
		"Conforms for correct error type", func(subT *testing.T) {
			var err error
			func() {
				_, err = NewAmalgamationService().Execute(nil, &AmalgamationRequest{SpawnError: "error"})
				if err == nil {
					subT.Fatalf("failed request, should return error")
				}
			}()
			logicResponse := err.(*rest.ErrorResponse)
			if logicResponse.Message != "Fail" {
				subT.Fatalf("failed, expected Fail, got: %s", logicResponse.Message)
			}
			if logicResponse.Code != 411 {
				subT.Fatalf("failed, expected 411, got: %d", logicResponse.Code)
			}
		},
	)
}
