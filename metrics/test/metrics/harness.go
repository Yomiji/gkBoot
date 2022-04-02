package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/metrics"
	"github.com/yomiji/gkBoot/metrics/config"
)

type TestRunners map[string]func(t *testing.T)

func (r TestRunners) Test(name string, test func(subT *testing.T)) TestRunners {
	r[name] = test
	return r
}

func NewTestRunner() TestRunners {
	return make(map[string]func(*testing.T))
}

func Harness(
	serviceRequests []gkBoot.ServiceRequest,
	bootOption []config.MetricsOption,
	runners TestRunners,
	t *testing.T,
) {
	srv, _ := metrics.Start(
		serviceRequests,
		bootOption...,
	)
	defer srv.Shutdown(context.Background())

	for name, test := range runners {
		t.Run(name, test)
	}
}

func CallAPI(
	method, url string, headers map[string]string, reqBody interface{},
	cookies ...*http.Cookie,
) (
	*http.Response,
	error,
) {
	var reader io.Reader
	if reqBody != nil {
		jBytes, err := json.Marshal(reqBody)
		if err == nil {
			reader = bytes.NewBuffer(jBytes)
		}
	}
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return http.DefaultClient.Do(request)
}

func ReadResponseBody(response *http.Response, respObj interface{}) error {
	bodyReader := response.Body
	defer bodyReader.Close()
	b, err := ioutil.ReadAll(bodyReader)
	if response.StatusCode == 200 {
		if err != nil {
			return err
		}
		return json.Unmarshal(b, respObj)
	} else {
		return fmt.Errorf("code: %d, reason: %s", response.StatusCode, string(b))
	}
}
