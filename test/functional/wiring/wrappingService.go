package wiring

import (
	"context"
	"testing"
	
	"github.com/yomiji/gkBoot/service"
)

type TestPassService struct {
	next service.Service
	t *testing.T
}

func (t TestPassService) GetNext() service.Service {
	return t.next
}

func (t *TestPassService) UpdateNext(nxt service.Service) {
	t.next = nxt
}

func (t TestPassService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	t.t.Log("Passed, Wrap Successful\n")
	return t.next.Execute(ctx, request)
}

func WrapTestedService(t *testing.T) service.Wrapper {
	return func(srv service.Service) service.Service {
		tps := new(TestPassService)
		tps.t = t
		tps.next = srv
		return tps
	}
}
