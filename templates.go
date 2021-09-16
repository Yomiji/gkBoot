package gkBoot

import (
	"context"
	
	"github.com/yomiji/gkBoot/service"
)

// BasicService
//
// This is the typical service with no DB attached, with an associated Configuration set by WithCustomConfig
//
// It is recommended to use config.WithCustomConfig on gkBoot.Start followed by implementing member Execute function
// of your struct
type BasicService struct {
	service.UsingConfig
}

func (b BasicService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	panic("implement me")
}

// BasicServiceWithDB
//
// This is the typical service with custom config and DB
//
// It is recommended to use config.WithCustomConfig and config.WithDatabase on gkBoot.Start
// followed by implementing member Execute function of your struct
type BasicServiceWithDB struct {
	service.UsingConfig
	service.UsingDB
}

func (b BasicServiceWithDB) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	panic("implement me")
}
