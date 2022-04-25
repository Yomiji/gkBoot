package client

import (
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/caching"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/functional/cache"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestGenerateClientRequest(t *testing.T) {
	cacheService := cache.NewCachableService()
	runners := tools.NewTestRunner().Test(
		"Cache Client Success", func(subT *testing.T) {
			req := new(cache.CacheableRequest)
			req.TestValue1 = 123
			req.TestValue2 = "456"

			resp := new(cache.CacheableResponse)
			err := gkBoot.DoRequest("http://localhost:8080", req, resp)
			if err != nil {
				subT.Fatalf("err encountered: %s", err)
			}

			if resp.TestResponse1 != 456 {
				subT.Fatalf("could not parse correctly, expected 456, got %d", resp.TestResponse1)
			}
		},
	).Test(
		"Cache Client Validation Fails", func(subT *testing.T) {
			req := new(cache.CacheableRequest)
			req.TestValue1 = 123
			req.TestValue2 = "value"

			_, err := gkBoot.GenerateClientRequest("http://localhost:8080", req)
			if err == nil {
				subT.Fatalf("err was expected, got none")
			}
		},
	).Test(
		"Cache Client Graceful Response", func(subT *testing.T) {
			req := new(cache.CacheableRequest)
			req.TestValue1 = 999
			req.TestValue2 = "999"

			resp := new(cache.CacheableResponse)
			r, err := gkBoot.GenerateClientRequest("http://localhost:8080", req)
			if err != nil {
				subT.Fatalf("unexpected err: %s", err)
			}

			err = gkBoot.DoGeneratedRequest(r, resp)
			if err != nil {
				subT.Fatalf("unexpected err: %s", err)
			}

			if resp.StatusCode() != 403 {
				subT.Fatalf("expected 403 status code, got %d", resp.StatusCode())
			}
		},
	)

	tools.Harness(
		[]gkBoot.ServiceRequest{{new(cache.CacheableRequest), cacheService}},
		[]config.GkBootOption{caching.WithCache(new(tools.Cache))}, runners, t,
	)
}
