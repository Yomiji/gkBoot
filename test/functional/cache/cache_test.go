package cache

import (
	"net/http"
	"testing"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/caching"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestCache(t *testing.T) {
	cacheService := NewCachableService()
	notCacheService := NewNotCachableService()
	runners := tools.NewTestRunner().Test(
		"Cache Success", func(subT *testing.T) {
			// fist unique api call
			call1 := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:8080/cacheable", map[string]string{
						"Test-Value-1": "123",
						"Test-Value-2": "456",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			// call it twice
			call1()
			call1()
			// second unique api call
			call2 := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:8080/cacheable", map[string]string{
						"Test-Value-1": "321",
						"Test-Value-2": "654",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			// call it 3 times
			call2()
			call2()
			call2()
			if cacheService.CacheHitCount.Int64() != 2 {
				subT.Fatalf("failed, expected 2 cache hits; got: %d\n", cacheService.CacheHitCount.Int64())
			}
		},
	).Test(
		"Cache Miss", func(subT *testing.T) {
			// fist unique api call
			call1 := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:8080/not_cacheable", map[string]string{
						"Test-Value-1": "123",
						"Test-Value-2": "456",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			// call it twice
			call1()
			call1()
			// second unique api call
			call2 := func() {
				_, err := tools.CallAPI(
					http.MethodGet, "http://localhost:8080/not_cacheable", map[string]string{
						"Test-Value-1": "321",
						"Test-Value-2": "654",
					}, nil,
				)
				if err != nil {
					subT.Fatalf("failed request: %s\n", err.Error())
				}
			}
			// call it 3 times
			call2()
			call2()
			call2()
			if notCacheService.CacheHitCount.Int64() != 5 {
				subT.Fatalf("failed, expected 5 regular calls; got: %d\n", notCacheService.CacheHitCount.Int64())
			}
		},
	)
	tools.Harness([]gkBoot.ServiceRequest{{new(CacheableRequest), cacheService},
		{new(NotCacheableRequest), notCacheService}},
	[]config.GkBootOption{caching.WithCache(new(tools.Cache))}, runners, t)
}
