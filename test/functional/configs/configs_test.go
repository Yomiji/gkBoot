package configs

import (
	"net/http"
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
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{{new(ConfRequest), configService}},
		[]config.GkBootOption{
			caching.WithCache(new(tools.Cache)), config.WithDatabase(nil),
			config.WithCustomConfig(nil), config.WithHttpPort(9090),
		}, runners, t,
	)
}
