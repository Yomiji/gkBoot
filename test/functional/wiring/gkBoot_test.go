package wiring

import (
	"net/http"
	"testing"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestGkBootStartup(t *testing.T) {
	runnersMap := tools.NewTestRunner().Test(
		"Test Trivial Startup", func(subT *testing.T) {
			resp, err := tools.CallAPI(
				http.MethodGet, "http://localhost:8080/test1", map[string]string{
					"Test-Num": "123",
				}, nil,
			)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			testResponse := TestResponse{}
			err = tools.ReadResponseBody(resp, &testResponse)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
		},
	)
	tools.Harness([]gkBoot.ServiceRequest{{new(TestRequest1), new(TestService1)}}, nil, runnersMap, t)
}

func TestWrappedService(t *testing.T) {
	runnersMap := tools.NewTestRunner().Test(
		"Test Wrapped Service", func(subT *testing.T) {
			resp, err := tools.CallAPI(http.MethodGet, "http://localhost:8080/test2", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			disabledResponse := DisabledResponse{}
			err = tools.ReadResponseBody(resp, &disabledResponse)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
			if !(disabledResponse.Testvalue == 999) {
				subT.Fatalf("failed response validation")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(DisabledLogRequest), new(DisabledService)}},
		[]config.GkBootOption{config.WithServiceWrapper(WrapTestedService(t))}, runnersMap, t,
	)
}

func TestGkBootServiceCapabilities(t *testing.T) {
	configurationSetting := ConfigSettings{
		TestValue1: 123,
		TestValue2: "one",
	}
	runnersMap := tools.NewTestRunner().Test(
		"Test Custom Config", func(subT *testing.T) {
			resp, err := tools.CallAPI(
				http.MethodGet, "http://localhost:8080/test1", map[string]string{
					"Test-Num": "",
				}, nil,
			)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			testResponse := TestResponse{}
			err = tools.ReadResponseBody(resp, &testResponse)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
			if !(testResponse.OptionalResponse1 == 123 && testResponse.OptionalResponse2 == "one") {
				subT.Fatalf("failed response validation")
			}
		},
	).Test(
		"Test Skip Logging", func(subT *testing.T) {
			resp, err := tools.CallAPI(http.MethodGet, "http://localhost:8080/test2", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			disabledResponse := DisabledResponse{}
			err = tools.ReadResponseBody(resp, &disabledResponse)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
			if !(disabledResponse.Testvalue == 999) {
				subT.Fatalf("failed response validation")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{
			{new(TestRequest1), new(ConfigService)},
			{new(DisabledLogRequest), new(DisabledService)},
		},
		[]config.GkBootOption{config.WithCustomConfig(configurationSetting)}, runnersMap, t,
	)
}

type builderConfig struct {
	testVal int
}

func TestBuilderService(t *testing.T) {
	bConfig := builderConfig{testVal: 5678}
	runnersMap := tools.NewTestRunner().Test(
		"Test Builder Service Standalone", func(subT *testing.T) {
			srv := NewBuilderService(bConfig)
			config, err := srv.Execute(nil, nil)
			if err != nil {
				subT.Fatalf("failure")
			}
			if d := config.(builderConfig).testVal; d != 5678 {
				subT.Fatalf("failure, expected 5678, got: %d", d)
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(DisabledLogRequest), new(DisabledService)}},
		[]config.GkBootOption{}, runnersMap, t,
	)
}
