package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/helpers"
	"github.com/yomiji/gkBoot/kitDefaults"
	"github.com/yomiji/gkBoot/logging"
	"github.com/yomiji/gkBoot/metrics/config"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

var loggingWrapper service.Wrapper

// Start
//
// Starts the http server for GkBoot. Returns the running http.Server and a blocking function
// that waits until a signal (syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM) is sent.
func Start(serviceRequests []gkBoot.ServiceRequest, option ...config.MetricsOption) (*http.Server, <-chan struct{}) {
	customConfig := &config.MetricsConfig{}
	for _, opt := range option {
		opt(customConfig)
	}
	if loggingWrapper == nil && customConfig.Logger == nil {
		logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		customConfig.Logger = logger
	}
	loggingWrapper = logging.GenerateLoggingWrapper(customConfig.Logger)

	r := chi.NewRouter()
	var useMetrics bool

	rmain := chi.NewRouter()

	for _, sr := range serviceRequests {
		if _, ok := sr.Service.(Metered); ok {
			useMetrics = true
		}
		r.Method(
			string(sr.Request.Info().Method), sr.Request.Info().Path, buildHttpRoute(
				sr, customConfig,
				customConfig.HttpOpts...,
			),
		)
	}

	if useMetrics {
		var metricsPath = "/metrics"
		if customConfig.MetricsPath != nil {
			metricsPath = *customConfig.MetricsPath
		}
		r.Method(
			http.MethodGet,
			metricsPath, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{
					// Opt into OpenMetrics to support exemplars.
					EnableOpenMetrics: customConfig.EnableOpenMetrics,
				},
			),
		)
	}
	var rootPath = "/"

	if customConfig.RootPath != nil {
		rootPath = *customConfig.RootPath
	}
	rmain.Mount(rootPath, r)

	var err error
	var httpPort = 8080
	if customConfig.HttpPort != nil {
		httpPort = *customConfig.HttpPort
	}

	portString := makePortString(&httpPort)

	// apply all global decorators
	var decoratedRouter http.Handler = rmain
	for _, decorator := range customConfig.Decorators {
		decoratedRouter = decorator(decoratedRouter)
	}
	srv := &http.Server{Handler: decoratedRouter, Addr: portString}

	errs := make(chan error)
	go func(srv *http.Server) {
		err = srv.ListenAndServe()
		if err != nil {
			errs <- err
		}
	}(srv)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	doneChan := make(chan struct{})
	go func() {
		// blocks until <-errs
		if customConfig.Logger != nil {
			level.Error(customConfig.Logger).Log("exit", <-errs)
		}
		doneChan <- struct{}{}
	}()
	return srv, doneChan
}

// StartServer
//
//  Convenience method.
//
// If the service and blocker of Start are unnecessary, this conveniently does all of that for us.
func StartServer(serviceRequests []gkBoot.ServiceRequest, option ...config.MetricsOption) {
	_, blocker := Start(serviceRequests, option...)
	<-blocker
}

func MakeHandler(serviceRequests []gkBoot.ServiceRequest, option ...config.MetricsOption) (
	http.Handler,
	*config.MetricsConfig,
) {
	customConfig := &config.MetricsConfig{}
	for _, opt := range option {
		opt(customConfig)
	}
	if loggingWrapper == nil && customConfig.Logger == nil {
		logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		customConfig.Logger = logger
	}
	loggingWrapper = logging.GenerateLoggingWrapper(customConfig.Logger)
	var r = chi.NewRouter()
	var useMetrics bool

	for _, sr := range serviceRequests {
		if _, ok := sr.Service.(Metered); ok {
			useMetrics = true
		}
		r.Method(
			string(sr.Request.Info().Method), sr.Request.Info().Path, buildHttpRoute(
				sr, customConfig,
				customConfig.HttpOpts...,
			),
		)
	}

	if useMetrics {
		var metricsPath = "/metrics"
		if customConfig.MetricsPath != nil {
			metricsPath = *customConfig.MetricsPath
		}
		r.Handle(
			metricsPath, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{
					// Opt into OpenMetrics to support exemplars.
					EnableOpenMetrics: customConfig.EnableOpenMetrics,
				},
			),
		)
	}
	var rootPath = "/"

	if customConfig.RootPath != nil {
		rootPath = *customConfig.RootPath
	}

	rmain := chi.NewRouter()

	rmain.Mount(rootPath, r)

	// apply all global decorators
	var decoratedRouter http.Handler = rmain
	for _, decorator := range customConfig.Decorators {
		decoratedRouter = decorator(decoratedRouter)
	}

	return decoratedRouter, customConfig
}

func StartWithHandler(serviceRequests []gkBoot.ServiceRequest, option ...config.MetricsOption) (
	*http.Server,
	<-chan struct{},
) {
	var err error
	handler, config := MakeHandler(serviceRequests, option...)

	var httpPort = 8080
	if config.HttpPort != nil {
		httpPort = *config.HttpPort
	}

	portString := makePortString(&httpPort)

	srv := &http.Server{Handler: handler, Addr: portString}

	errs := make(chan error)
	go func(srv *http.Server) {
		err = srv.ListenAndServe()
		if err != nil {
			errs <- err
		}
	}(srv)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	doneChan := make(chan struct{})
	go func() {
		// blocks until <-errs
		if config.Logger != nil {
			level.Error(config.Logger).Log("exit", <-errs)
		}
		doneChan <- struct{}{}
	}()
	return srv, doneChan
}

func StartServerWithHandler(serviceRequests []gkBoot.ServiceRequest, option ...config.MetricsOption) {
	_, blocker := StartWithHandler(serviceRequests, option...)
	<-blocker
}

func getCustomDecoder(sr gkBoot.ServiceRequest) (kitDefaults.DecodeRequestFunc, error) {
	if customDecoder, ok := sr.Request.(gkBoot.HttpDecoder); ok {
		return customDecoder.Decode, nil
	}
	return gkBoot.GenerateRequestDecoder(sr.Request)
}

func getCustomEncoder(sr gkBoot.ServiceRequest) kitDefaults.EncodeResponseFunc {
	if customEncoder, ok := sr.Service.(service.HttpEncoder); ok {
		return customEncoder.Encode
	}
	return kitDefaults.DefaultHttpResponseEncoder
}

type serviceBuilder struct {
	srv    service.Service
	config *config.MetricsConfig
}

// NewServiceBuilder
//
// This will wire up a service-only object that can re-use logging and metrics wrappers established elsewhere while
// maintaining the REST Request-Service established pattern. The associated Mixin chains must be called to identify
// which functionality is wired in.
//
// Unavailable config options (using the following will not do anything):
//  config.WithServiceDecorator
//  config.WithHttpServerOpts
//  config.WithOpenMetrics
//  config.WithMetricsPath
//  config.WithHttpPort
//  config.WithRootPath
//  config.WithStrictAPI
func NewServiceBuilder(srv service.Service, option ...config.MetricsOption) *serviceBuilder {
	nsb := new(serviceBuilder)
	nsb.srv = srv
	nsb.config = &config.MetricsConfig{}
	for _, opt := range option {
		opt(nsb.config)
	}

	for _, wrapper := range nsb.config.ServiceWrappers {
		nsb.srv = wrapRootService(nsb.srv, wrapper)
	}

	return nsb
}

// MixinLogging
//
// Wrap logging around the service.
func (s *serviceBuilder) MixinLogging() *serviceBuilder {
	if loggingWrapper == nil {
		logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		s.config.Logger = logger
		loggingWrapper = logging.GenerateLoggingWrapper(s.config.Logger)
	}
	s.srv = wrapRootService(s.srv, loggingWrapper)
	return s
}

// MixinCustomConfig
//
// Inject custom configuration object into the service.
func (s *serviceBuilder) MixinCustomConfig() *serviceBuilder {
	if configurableService, ok := s.srv.(service.ConfigurableService); ok {
		configurableService.SetConfig(s.config.CustomConfig)
	}
	return s
}

// MixinDatabase
//
// Inject database into the service.
func (s *serviceBuilder) MixinDatabase() *serviceBuilder {
	if databaseService, ok := s.srv.(service.DatabaseConfigurable); ok {
		databaseService.SetDatabase(s.config.Database)
	}
	return s
}

// MixinMetrics
//
// Wrap the metrics around the service
func (s *serviceBuilder) MixinMetrics() *serviceBuilder {
	if metered, ok := s.srv.(Metered); ok {
		s.srv = wrapRootService(metered, WrapServiceMetrics)
	}
	return s
}

// MixinCustomWrapper
//
// Wrap the service with the given wrapper.
func (s *serviceBuilder) MixinCustomWrapper(wrapper service.Wrapper) *serviceBuilder {
	s.srv = wrapRootService(s.srv, wrapper)
	return s
}

// Build
//
// Build the final service and return the result.
func (s *serviceBuilder) Build() service.Service {
	if s.config.StrictOpenAPI {
		s.srv = wrapRootService(s.srv, gkBoot.APIValidationWrapper)
	}

	return s.srv
}

func buildHttpRoute(
	sr gkBoot.ServiceRequest, bConfig *config.MetricsConfig,
	opts ...kitDefaults.ServerOption,
) http.Handler {
	var decoder kitDefaults.DecodeRequestFunc
	var encoder kitDefaults.EncodeResponseFunc
	var err error
	var req = sr.Request.(request.HttpRequest)

	if configurableService, ok := sr.Service.(service.ConfigurableService); ok {
		configurableService.SetConfig(bConfig.CustomConfig)
	}

	if databaseService, ok := sr.Service.(service.DatabaseConfigurable); ok {
		databaseService.SetDatabase(bConfig.Database)
	}

	var serviceOptions = make([]kitDefaults.ServerOption, 0)
	copy(serviceOptions, opts)

	if ros, ok := sr.Service.(service.OptionsConfigurable); ok {
		serviceOptions = append(serviceOptions, ros.ServerOptions()...)
	}

	if metered, ok := sr.Service.(Metered); ok {
		sr.Service = wrapRootService(metered, WrapServiceMetrics)
	}

	sr.Service = wrapRootService(sr.Service, loggingWrapper)

	for _, wrapper := range bConfig.ServiceWrappers {
		sr.Service = wrapRootService(sr.Service, wrapper)
	}

	if bConfig.StrictOpenAPI {
		sr.Service = wrapRootService(sr.Service, gkBoot.APIValidationWrapper)
	}

	if decoder, err = getCustomDecoder(sr); err != nil {
		_ = bConfig.Logger.Log("err", fmt.Sprintf("decoder generation failed for %s", req.Info().Name))
	}
	encoder = getCustomEncoder(sr)

	router := chi.NewRouter()
	router.Handle(
		req.Info().Path, kitDefaults.NewServer(
			sr.Service.Execute,
			decoder,
			encoder,
			append(
				serviceOptions, kitDefaults.ServerErrorEncoder(
					func(ctx context.Context, err error, w http.ResponseWriter) {
						if err != nil {
							if bConfig.Logger != nil {
								ctxHeaders := helpers.GetCtxHeadersFromContext(ctx)
								bConfig.Logger.Log(
									"Request", req.Info(),
									"RequestType", "HTTP", "Error", err, "CtxHeaders", ctxHeaders,
								)
							}

							sc := w.Header().Get("Status-Code")

							// if no status code given, use the coder interface or, if missing, use 500
							if sc == "" {
								if j, ok := err.(kitDefaults.HttpCoder); ok {
									w.WriteHeader(j.StatusCode())
								} else {
									if _, err := strconv.Atoi(sc); err != nil {
										w.WriteHeader(http.StatusInternalServerError)
									}
								}
							}
						}
					},
				),
			)...,
		),
	)

	var decoratedRouter http.Handler = router
	if decoratedRequest, ok := sr.Request.(request.Decorator); ok {
		decoratedRouter = decoratedRequest.UsingDecorator()(decoratedRouter)
	}
	return decoratedRouter
}

func makePortString(port *int) string {
	if port != nil {
		return ":" + strconv.Itoa(*port)
	}
	panic(fmt.Errorf("unknown port or port nil: %+d", port))
}

func wrapRootService(svc service.Service, wrapper service.Wrapper) service.Service {
	if wrapper == nil {
		return svc
	}
	var temp interface{} = svc
	switch v := temp.(type) {
	case service.UpdatableWrappedService:
		next := wrapRootService(v.GetNext(), wrapper)
		v.UpdateNext(next)
		return svc
	case service.Service:
		return wrapper(v)
	default:
		return svc
	}
}
