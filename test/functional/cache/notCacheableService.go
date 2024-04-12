package cache

import (
	"context"
	"math/big"
	"strconv"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type NotCacheableRequest struct {
	TestValue1 int    `request:"header!" alias:"Test-Value-1" json:"tv1"`
	TestValue2 string `request:"header!" alias:"Test-Value-2" json:"tv2"`
}

func (c NotCacheableRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "CacheableRequest",
		Method:      request.GET,
		Path:        "/not_cacheable",
		Description: "A Cacheable Request",
	}
}

type NotCacheableService struct {
	CacheHitCount *big.Int
	gkBoot.BasicService
}

func NewNotCachableService() *NotCacheableService {
	// use factories like this to compose service object invariants
	c := new(NotCacheableService)
	c.CacheHitCount = big.NewInt(0)
	return c
}

type NotCacheableResponse struct {
	TestResponse1 int `json:"tr1"`
	TestResponse2 int `json:"tr2"`
}

func (c *NotCacheableService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	c.CacheHitCount.Add(c.CacheHitCount, big.NewInt(1))
	req := request.(*NotCacheableRequest)
	tv2Totr1, err := strconv.Atoi(req.TestValue2)
	return NotCacheableResponse{
		TestResponse1: tv2Totr1,
		TestResponse2: req.TestValue1,
	}, err
}
