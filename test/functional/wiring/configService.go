package wiring

import (
	"context"

	"github.com/yomiji/gkBoot"
)

type ConfigSettings struct {
	TestValue1 int    `json:"testValue1"`
	TestValue2 string `json:"testValue2"`
}

type ConfigService struct {
	gkBoot.BasicService
}

func (s ConfigService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	return TestResponse{
		OptionalResponse1: s.GetConfig().(ConfigSettings).TestValue1,
		OptionalResponse2: s.GetConfig().(ConfigSettings).TestValue2,
	}, nil
}
