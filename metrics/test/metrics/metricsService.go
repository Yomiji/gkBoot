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

type MetricsService[T MetricsRequest, V any] struct {
	gkBoot.BasicService
}

func (m *MetricsService[T, V]) Execute(ctx context.Context, request T) (response V, err error) {
	return nil, nil
}

func (m *MetricsService[T, V]) Metrics() *metrics.MappedMetrics {
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

func (m *MetricsService[T, V]) UpdateMetrics(
	_ context.Context,
	_ interface{},
	_ interface{},
	_ time.Time,
	mappedMetrics *metrics.MappedMetrics,
) {
	// Called after every execution of the wrapped service
	mappedMetrics.Counters["some_random_name"].Inc()
}
