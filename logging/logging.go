package logging

import (
	"context"
	"time"
	
	"github.com/yomiji/gkBoot/helpers"
	"github.com/yomiji/gkBoot/kitDefaults"
	"github.com/yomiji/gkBoot/request"
	"github.com/yomiji/gkBoot/service"
)

// Logger
//
// Defines a logging pattern that is popular in go-kit. Establishing this interface keeps us
// from having to import go-kit here while also maintaining type safety.
type Logger interface {
	Log(elem...interface{}) error
}

type skipLoggable interface {
	skipLogging() bool
}

// LogSkip
//
//  Implements skipLoggable
//
// Embed into services that should skip logging. These services will not be logged at all
type LogSkip struct {}

func (l LogSkip) skipLogging() bool {
	return true
}

type loggingWrappedService struct {
	logger Logger
	next   service.Service
}

func (l *loggingWrappedService) UpdateNext(nxt service.Service) {
	l.next = nxt
}

func (l loggingWrappedService) GetNext() service.Service {
	return l.next
}

func (l loggingWrappedService) Execute(ctx context.Context, req interface{}) (response interface{}, err error) {
	defer func(start time.Time) {
		endTime := time.Now().UTC()
		code := 200
		if v,ok := err.(kitDefaults.CodedError); ok && v != nil && v.StatusCode() != 0 {
			code = v.StatusCode()
		} else if !ok && err != nil {
			code = 500
		}
		ctxHeaders := helpers.GetCtxHeadersFromContext(ctx)
		additionalLogs := helpers.GetAdditionalLogs(response)
		var httpRequestLog []interface{}
		if httpRequest, ok := req.(request.HttpRequest); req != nil && ok {
			httpRequestLog = []interface{}{
				"Name", httpRequest.Info().Name,
				"Description", httpRequest.Info().Description,
			}
		}
		var loggingElements = []interface{}{
			"CtxHeaders", ctxHeaders,
			"Request", req,
			"RequestType", "HTTP",
			"Response", response,
			"AdditionalLogs", additionalLogs,
			"Code", code,
			"Error", err,
			"CallStart", start,
			"CallEnd", endTime,
		}
		l.logger.Log(append(httpRequestLog, loggingElements...)...)
	}(time.Now().UTC())
	return l.next.Execute(ctx, req)
}

// GenerateLoggingWrapper
//
// Called by gkBoot.StartHttpServer, this creates a wrapper that wraps service with a deferred logger. After the
// service executes, a log will be generated
func GenerateLoggingWrapper(logger Logger) service.Wrapper {
	return func(srv service.Service) service.Service {
		if _,ok := srv.(skipLoggable); ok {
			return srv
		}
		return &loggingWrappedService{logger, srv}
	}
}
