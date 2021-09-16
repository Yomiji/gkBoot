package database

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"testing"
	
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/config"
	"github.com/yomiji/gkBoot/test/tools"
)

func TestDatabaseServicesWithOption(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Database Is Present With Option", func(subT *testing.T) {
			resp, err := tools.CallAPI(http.MethodGet, "http://localhost:8080/db", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			responseBody := resp.Body
			defer responseBody.Close()
			bytes, err := ioutil.ReadAll(responseBody)
			if err != nil {
				subT.Fatalf("failed to read body: %s", err.Error())
			}
			if string(bytes) == "false" {
				subT.Fatal("failed response, db not present")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{
			{new(DBRequest), new(DBService)},
		}, []config.GkBootOption{config.WithDatabase(&sql.DB{})}, runners, t,
	)
}

func TestDatabaseServicesNoOption(t *testing.T) {
	runners := tools.NewTestRunner().Test(
		"Database Missing Without Option", func(subT *testing.T) {
			resp, err := tools.CallAPI(http.MethodGet, "http://localhost:8080/db", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			responseBody := resp.Body
			defer responseBody.Close()
			bytes, err := ioutil.ReadAll(responseBody)
			if err != nil {
				subT.Fatalf("failed to read body: %s", err.Error())
			}
			if string(bytes) == "true" {
				subT.Fatal("failed response")
			}
		},
	)
	tools.Harness(
		[]gkBoot.ServiceRequest{
			{new(DBRequest), new(DBService)},
		}, nil, runners, t,
	)
}
