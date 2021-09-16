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

type CacheableRequest interface {
	CacheKey() string
}

type TTLCacheableRequest interface {
	CacheTimeToLive() time.Duration
}

type RequestCache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error)
}

type gkCache struct {
	cache RequestCache
	next service.Service
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
			if timeable, ok := request.(TTLCacheableRequest); ok {
				ttl = timeable.CacheTimeToLive()
			}
			_, err = g.cache.Put(ctx, cacheable.CacheKey(), res, ttl)
			if err != nil {
				if v,ok := response.(logging.Logger); ok {
					_ = v.Log("caching", CacheSaveFailed)
				}
			}
			return res, err2
		} else {
			return val, nil
		}
	}
	return g.next.Execute(ctx, request)
}

func WithCache(cache RequestCache) config.GkBootOption {
	return func(config *config.BootConfig) {
		config.ServiceWrappers = append(config.ServiceWrappers, func(srv service.Service) service.Service {
			gkC := new(gkCache)
			gkC.next = srv
			gkC.cache = cache
			return gkC
		})
	}
}
