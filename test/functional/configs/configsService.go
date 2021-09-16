package configs

import (
	"context"
	"encoding/json"
	"math/big"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/metrics"
	"github.com/yomiji/gkBoot/request"
)

type ConfRequest struct {
	TestValue1 int `request:"query" json:"tv1"`
}

func (c ConfRequest) CacheKey() string {
	j,err := json.Marshal(c)
	if err != nil { return "" }
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

func (c ConfService) Metrics() *metrics.MappedMetrics {
	return &metrics.MappedMetrics{
		Counters:   nil,
		Histograms: nil,
		Gauges: map[string]prometheus.Gauge{
			"conf_count": prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "conf_count",
				},
			),
		},
	}
}

func (c ConfService) UpdateMetrics(
	ctx context.Context,
	request interface{},
	response interface{},
	startTime time.Time,
	mappedMetrics *metrics.MappedMetrics,
) {
	mappedMetrics.Gauges["conf_count"].Set(5)
}

func (c ConfService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	c.CacheHitCounter.Add(c.CacheHitCounter, big.NewInt(1))
	return request, nil
}
