package metrics

import (
	"context"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/metrics"
	"github.com/yomiji/gkBoot/request"
)

type MetricsRequest struct {
	TestValue1 int `json:"testValue1"`
}

func (m MetricsRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "MetricsTest",
		Method:      "GET",
		Path:        "/metricsTest",
		Description: "Test Metrics Service",
	}
}

type MetricsService struct {
	gkBoot.BasicService
}

func (m MetricsService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	return nil, nil
}

func (m *MetricsService) Metrics() *metrics.MappedMetrics {
	return &metrics.MappedMetrics{
		Counters: map[string]prometheus.Counter{
			/* referenced later in Update */
			"some_random_name": prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: MetricsRequest{}.Info().Name, // prometheus' name for this
					Help: MetricsRequest{}.Info().Description,
				},
			),
		},
	}
}

func (m *MetricsService) UpdateMetrics(
	_ context.Context,
	_ interface{},
	_ interface{},
	_ time.Time,
	mappedMetrics *metrics.MappedMetrics,
) {
	// Called after every execution of the wrapped service
	mappedMetrics.Counters["some_random_name"].Inc()
}
