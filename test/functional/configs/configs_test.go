package configs

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/caching"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestOptions(t *testing.T) {
	configService := NewConfService()
	runners := tools.NewTestRunner().Test(
		"Config Using Logging Cache Metrics", func(subT *testing.T) {
			call := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:9090/config?tv1=4444", nil, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			call()
			call()
			call()
			hitCount := configService.CacheHitCounter.Int64()
			if hitCount != 1 {
				subT.Fatalf("failed cache check, expected 1 got %d", hitCount)
			}
			resp, err := tools.CallAPI(http.MethodGet, "http://localhost:9090/m", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			respBody := resp.Body
			defer respBody.Close()
			b, err := ioutil.ReadAll(respBody)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
			if !strings.Contains(string(b), "conf_count 5") {
				subT.Fatal("failed response, metric not found")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(ConfRequest), configService}},
		[]config.GkBootOption{
			caching.WithCache(new(tools.Cache)), config.WithDatabase(nil),
			config.WithCustomConfig(nil), config.WithHttpPort(9090),
			config.WithMetricsPath("/m"),
		}, runners, t,
	)
}
