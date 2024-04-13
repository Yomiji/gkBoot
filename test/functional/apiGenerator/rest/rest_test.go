package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestRestService(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Swaggest Test Happy Path", func(subT *testing.T) {
			// fist unique api call
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
	).Test("Omit Name has Err", func(subT *testing.T) {
		// fist unique api call
		var httpResponse *http.Response
		var err error
		func() {
			httpResponse, err = tools.CallAPI(
				http.MethodGet, "http://localhost:8080/test/myPath?cost=123.4", map[string]string{}, nil,
			)
			if err != nil {
				subT.Fatalf("failed request: %s\n", err.Error())
			}
		}()
		if httpResponse.StatusCode == 201 {
			subT.Fatalf("expected status code != 201")
		}
	},
	).Test("Omit Cost has Err", func(subT *testing.T) {
		// fist unique api call
		var httpResponse *http.Response
		var err error
		func() {
			httpResponse, err = tools.CallAPI(
				http.MethodGet, "http://localhost:8080/test/myPath", map[string]string{
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
	).Test("Request is err intentionally", func(subT *testing.T) {
		// fist unique api call
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
	).Test("Request has named cookie", func(subT *testing.T) {
		// fist unique api call
		var httpResponse *http.Response
		var err error
		func() {
			httpResponse, err = tools.CallAPI(
				http.MethodGet, "http://localhost:8080/test/myPath?cost=123.4", map[string]string{
					"Name-Var": "testName",
				}, nil, &http.Cookie{
					Name:  "testCookie",
					Value: "I'm a test cookie",
				},
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
		if logicResponse.CookieVal != "I'm a test cookie" {
			subT.Fatalf("failed, expected  \"I'm a test cookie\", got: %s", logicResponse.CookieVal)
		}
	},
	)
	tools.Harness([]gkBoot.ServiceRequest{{new(TestRequest), new(TestService)}}, nil, runners, t)
}
