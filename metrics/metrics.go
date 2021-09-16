package metrics

import (
	"context"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yomiji/gkBoot/service"
)

// Metered
//
// This interface indicates that the service will maintain its ability to serve but also
// enable the processing of metrics after each request. The ProcessRequestMetrics function will be
// called on the deferred stack as a result of the automatic wiring.
type Metered interface {
	// Metrics
	//
	// Returns a set of definitions for the MappedMetrics. This will provide the metrics
	// wrapper with definitions that will be used for your service. This object will be
	// re-used and passed along to the UpdateMetrics function after every request.
	Metrics() *MappedMetrics
	// UpdateMetrics
	//
	// Called after every response sent from the wrapped service. This gathers the necessary
	// data points so that metrics may be derived from the service call. The MappedMetrics
	// contain all metrics mapped for only the implementing service.
	UpdateMetrics(ctx context.Context, request interface{}, response interface{}, startTime time.Time,
	mappedMetrics *MappedMetrics)
	service.Service
}

// MappedMetrics
//
// Contains the concrete metrics defined by the service. This is passed to the configuration
type MappedMetrics struct {
	Counters   map[string]prometheus.Counter
	Histograms map[string]prometheus.Histogram
	Gauges     map[string]prometheus.Gauge
}

type metricsWrapper struct {
	service service.Service
	metrics *MappedMetrics
}

// UpdateNext
//
// Method contained in service.UpdatableWrappedService interface
func (m *metricsWrapper) UpdateNext(nxt service.Service) {
	m.service = nxt
}

// GetNext
//
// Method contained in service.UpdatableWrappedService interface
func (m *metricsWrapper) GetNext() service.Service {
	return m.service
}

func (m *metricsWrapper) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	defer func(startTime time.Time) {
		r := response
		var outerService = getMeteredService(m.service)
		if outerService != nil {
			outerService.UpdateMetrics(ctx, request, r, startTime, m.metrics)
		}
	}(time.Now().UTC())
	return m.service.Execute(ctx, request)
}

func WrapServiceMetrics(service service.Service) service.Service {
	m := new(metricsWrapper)
	if metered, ok := service.(Metered); ok {
		metrics := metered.Metrics()
		for _, counter := range metrics.Counters {
			prometheus.MustRegister(counter)
		}
		for _, histogram := range metrics.Histograms {
			prometheus.MustRegister(histogram)
		}
		for _, gauge := range metrics.Gauges {
			prometheus.MustRegister(gauge)
		}
		m.service = metered
		m.metrics = metrics
	}
	
	return m
}

func getMeteredService(srv service.Service) Metered {
	if metered,ok := srv.(Metered); ok {
		return metered
	}
	if wrapped,ok := srv.(service.UpdatableWrappedService); ok {
		return getMeteredService(wrapped.GetNext())
	}
	return nil
}
