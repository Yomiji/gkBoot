package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/yomiji/gkBoot/response"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type CacheableRequest struct {
	TestValue1 int    `request:"header!" alias:"Test-Value-1" json:"tv1"`
	TestValue2 string `request:"header!" alias:"Test-Value-2" json:"tv2"`
}

func (c CacheableRequest) Validate() error {
	if _, e := strconv.Atoi(c.TestValue2); e != nil {
		return fmt.Errorf("invalid value for Test-Value-2: %s", e)
	}

	return nil
}

func (c CacheableRequest) CacheKey() string {
	j, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(j)
}

func (c CacheableRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "CacheableRequest",
		Method:      request.GET,
		Path:        "/cacheable",
		Description: "A Cacheable Request",
	}
}

type CacheableService struct {
	CacheHitCount *big.Int
	gkBoot.BasicService
}

func NewCachableService() *CacheableService {
	// use factories like this to compose service object invariants
	c := new(CacheableService)
	c.CacheHitCount = big.NewInt(0)
	return c
}

type CacheableResponse struct {
	TestResponse1          int `json:"tr1"`
	TestResponse2          int `json:"tr2"`
	response.ErrorResponse `json:"-"`
}

func (c *CacheableService) Execute(ctx context.Context, request interface{}) (interface{}, error) {
	c.CacheHitCount.Add(c.CacheHitCount, big.NewInt(1))
	req := request.(*CacheableRequest)
	tv2Totr1, err := strconv.Atoi(req.TestValue2)

	if req.TestValue1 == 999 {
		resp := CacheableResponse{
			TestResponse1: -1,
			TestResponse2: -1,
		}
		resp.NewError(403, "unauthorized")

		return resp, nil
	}
	return CacheableResponse{
		TestResponse1: tv2Totr1,
		TestResponse2: req.TestValue1,
	}, err
}
