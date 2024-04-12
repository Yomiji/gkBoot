package wiring

import (
	"context"
	"fmt"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/service"
)

func NewBuilderService(cfg interface{}) service.Service {
	return gkBoot.NewServiceBuilder(new(BuilderService), config.WithCustomConfig(cfg)).
		MixinCustomConfig().
		MixinLogging().
		Build()
}

type BuilderService struct {
	gkBoot.BasicService
}

func (b BuilderService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	if b.GetConfig() == nil {
		return nil, fmt.Errorf("no configuration passed\n")
	}
	return b.GetConfig(), nil
}
