package helpers

import (
	"context"
	"testing"

	"github.com/yomiji/gkBoot/helpers"
)

func TestHeadersInject(t *testing.T) {
	var ctx = context.Background()
	helpers.InjectCtxHeaders(&ctx, map[string]interface{}{"testInjection": 123})
	if headers := helpers.GetCtxHeadersFromContext(ctx); headers == nil {
		t.Fatal("Unable to retrieve headers from context injection: nil")
	} else {
		if testInjection, ok := headers["testInjection"]; !ok || testInjection != 123 {
			t.Fatal("arguments expected do not exist")
		}
	}
}

func TestRateLimiter(t *testing.T) {
	var ctx = context.Background()
	helpers.SetRequestBodyLimit(&ctx, 1000)
	if limit := helpers.GetRequestBodyLimit(ctx); limit == nil {
		t.Fatal("Unable to retrieve limit from context injection: nil")
	} else {
		if *limit != 1000 {
			t.Fatal("expected limit not valid")
		}
	}
}
