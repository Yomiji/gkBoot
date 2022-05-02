package caching

import (
	"context"
	"fmt"
	"time"

	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/logging"
	"github.com/yomiji/gkBoot/service"
)

var CacheSaveFailed = fmt.Errorf("cache failed to save")

// CacheableRequest
//
// Any object implementing CacheKey can represent a cache key for the purposes of a gkBoot cache
type CacheableRequest interface {
	CacheKey() string
}

// TimedCache
//
// Any object implementing CacheDuration will be a timed cache
type TimedCache interface {
	CacheDuration() time.Duration
}

// RequestCache
//
// Any object implementing Get and Put of this interface will qualify for being a RequestCache
type RequestCache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error)
}

type gkCache struct {
	cache RequestCache
	next  service.Service
}

func (g gkCache) GetNext() service.Service {
	return g.next
}

func (g *gkCache) UpdateNext(nxt service.Service) {
	g.next = nxt
}

func (g gkCache) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	if cacheable, ok := request.(CacheableRequest); ok {
		if val, err := g.cache.Get(ctx, cacheable.CacheKey()); err != nil {
			res, err2 := g.next.Execute(ctx, request)
			if err2 != nil {
				return res, err2
			}
			var ttl time.Duration = 0
			if timeable, ok := request.(TimedCache); ok {
				ttl = timeable.CacheDuration()
			}
			_, err = g.cache.Put(ctx, cacheable.CacheKey(), res, ttl)
			if err != nil {
				if v, ok := res.(logging.Logger); ok {
					_ = v.Log("caching", CacheSaveFailed, "err", err)
				}
			}
			return res, err2
		} else {
			return val, nil
		}
	}
	return g.next.Execute(ctx, request)
}

// WithCache
//
// Use to enable caching throughout the app. Cache will only be applied where the requests implement
// at least CacheableRequest interface
func WithCache(cache RequestCache) config.GkBootOption {
	return func(config *config.BootConfig) {
		config.ServiceWrappers = append(config.ServiceWrappers, NewCacheWrapper(cache))
	}
}

// NewCacheWrapper
//
// Enable caching throughout the app. Cache will only be applied where the requests implement
// at least CacheableRequest interface. This is a convenience function to create a wrapper.
func NewCacheWrapper(cache RequestCache) service.Wrapper {
	return func(srv service.Service) service.Service {
		gkC := new(gkCache)
		gkC.next = srv
		gkC.cache = cache
		return gkC
	}
}
