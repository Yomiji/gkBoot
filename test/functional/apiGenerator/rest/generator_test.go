package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestGenerator(t *testing.T) {
	services := []gkBoot.ServiceRequest{{new(TestRequest), new(TestService)}}

	spec, err := gkBoot.GenerateSpecification(services, nil)
	if err != nil {
		t.Fatalf("expected success, got %s", err)
	}
	yaml, _ := spec.Spec.MarshalYAML()
	t.Log(string(yaml))
}

func TestStrict(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"OpenAPI Conforms to Result", func(subT *testing.T) {
			var httpResponse *http.Response
			var err error
			func() {
				httpResponse, err = tools.CallAPI(
					http.MethodGet, "http://localhost:8080/test/myPath?cost=123.4", map[string]string{
						"Name-Var": "testName",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}()
			responseBytes, err := ioutil.ReadAll(httpResponse.Body)
			if err != nil {
				subT.Fatalf("failed, unexpected err: %s\n", err)
			}
			logicResponse := new(TestResponse)
			json.Unmarshal(responseBytes, logicResponse)
			if logicResponse.Name != "testName" {
				subT.Fatalf("failed, expected testName, got: %s", logicResponse.Name)
			}
			if logicResponse.Cost != 123.4 {
				subT.Fatalf("failed, expected 123.4, got: %f", logicResponse.Cost)
			}
			if logicResponse.Path != "myPath" {
				subT.Fatalf("failed, expected myPath, got: %s", logicResponse.Path)
			}
		},
	).Test(
		"OpenAPI Does not Conform", func(subT *testing.T) {
			var httpResponse *http.Response
			var err error
			func() {
				httpResponse, err = tools.CallAPI(
					http.MethodGet, "http://localhost:8080/test/myPath?cost=123.4&mustHave=invalid", map[string]string{
						"Name-Var": "testName",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}()
			if httpResponse.StatusCode != 500 {
				subT.Fatalf("expected status code == 500")
			}
		},
	).Test(
		"Conforms for correct error type", func(subT *testing.T) {
			var httpResponse *http.Response
			var err error
			func() {
				httpResponse, err = tools.CallAPI(
					http.MethodGet, "http://localhost:8080/test/myPath?cost=123.4&mayErr=true", map[string]string{
						"Name-Var": "testName",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}()
			if httpResponse.StatusCode == 201 {
				subT.Fatalf("expected status code != 201")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(TestRequest), new(TestService)}},
		[]config.GkBootOption{config.WithStrictAPI()}, runners, t,
	)
}
