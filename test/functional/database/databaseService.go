package database

import (
	"context"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type DBRequest struct {
}

func (d DBRequest) Info() request.HttpRouteInfo {
	return request.HttpRouteInfo{
		Name:        "DatabaseTest",
		Method:      "GET",
		Path:        "/db",
		Description: "Database Test",
	}
}

type DBService struct {
	gkBoot.BasicServiceWithDB
}

func (d DBService) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	var dbNil = d.GetDatabase() != nil
	return dbNil, nil
}
