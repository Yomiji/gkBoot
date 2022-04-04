package metrics

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/yomiji/gkBoot"
)

func TestMetrics(t *testing.T) {
	runners := NewTestRunner().Test(
		"Test Metrics Defaults", func(subT *testing.T) {
			_, err := CallAPI(http.MethodGet, "http://localhost:8080/metricsTest", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			resp, err := CallAPI(http.MethodGet, "http://localhost:8080/metrics", nil, nil)
			if err != nil {
				subT.Fatalf("failed request: %s", err.Error())
			}
			respBody := resp.Body
			defer respBody.Close()
			b, err := ioutil.ReadAll(respBody)
			if err != nil {
				subT.Fatalf("failed response: %s", err.Error())
			}
			if !strings.Contains(string(b), MetricsRequest{}.Info().Name+" 1") {
				subT.Fatal("failed response, metric not found")
			}
		},
	)
	Harness(
		[]gkBoot.ServiceRequest{
			{new(MetricsRequest), new(MetricsService)},
		}, nil, runners, t,
	)
}
