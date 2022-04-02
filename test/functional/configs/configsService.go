package configs

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type ConfRequest struct {
	TestValue1 int `request:"query" json:"tv1"`
}

func (c ConfRequest) CacheKey() string {
	j, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(j)
}

func (c ConfRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "ConfigurationTest",
		Method:      request.GET,
		Path:        "/config",
		Description: "Test Configuration Mixes",
	}
}

type ConfService struct {
	CacheHitCounter *big.Int
	gkBoot.BasicService
}

func NewConfService() *ConfService {
	s := new(ConfService)
	s.CacheHitCounter = big.NewInt(0)
	return s
}

func (c ConfService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	c.CacheHitCounter.Add(c.CacheHitCounter, big.NewInt(1))
	return request, nil
}
